// add-issues 向看板数据库追加 900 个 issues（使总数达到 1000），无需重启服务。
//
// 用法:
//
//	go run ./cmd/add-issues/
package main

import (
	"database/sql"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"zenmind-kanban-server/internal/config"
	"zenmind-kanban-server/internal/kanban"

	_ "modernc.org/sqlite"
)

var issueTitlePrefixes = []string{
	"优化", "修复", "实现", "重构", "设计", "调研", "集成", "部署", "测试",
	"文档", "监控", "配置", "迁移", "升级", "清理", "适配", "安全", "性能",
	"日志", "缓存", "接口", "数据", "用户", "权限", "通知", "审核", "报表",
}

var issueTitleNouns = []string{
	"登录", "订单", "商品", "购物车", "支付", "搜索", "首页", "个人中心", "设置",
	"消息", "评论", "上传", "导出", "导入", "推送", "同步", "备份", "恢复",
	"认证", "授权", "限流", "降级", "熔断", "路由", "负载", "健康检查",
	"数据库", "缓存", "队列", "任务", "定时器", "WebSocket", "API",
	"表单", "表格", "图表", "地图", "富文本", "文件", "图片", "视频",
	"通知", "邮件", "短信", "审批", "工作流", "报表", "仪表盘", "日志",
}

var priorityWeights = []string{
	"medium", "medium", "medium", "high", "low",
}

var severityWeights = []string{
	"medium", "medium", "medium", "high", "medium", "low", "critical",
}

var statusWeights = []string{
	"backlog", "backlog", "backlog", "todo", "todo", "in_progress", "in_review", "completed",
}

func main() {
	cfg := config.Load()
	db, err := sql.Open("sqlite", cfg.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	db.Exec("PRAGMA busy_timeout = 3000")
	db.Exec("PRAGMA foreign_keys = ON")

	// 获取所有项目 ID
	rows, err := db.Query("SELECT ID_ FROM project WHERE ID_ != ? AND DELETED_AT_ IS NULL", kanban.DefaultProjectID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "查询项目失败: %v\n", err)
		os.Exit(1)
	}
	var projectIDs []string
	for rows.Next() {
		var id string
		rows.Scan(&id)
		projectIDs = append(projectIDs, id)
	}
	rows.Close()

	if len(projectIDs) == 0 {
		fmt.Fprintln(os.Stderr, "没有可用的项目")
		os.Exit(1)
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	stageID := kanban.DefaultWorkflowID + "-stage-requirement_clarification"

	// 每个项目计算当前最大 position
	type projectInfo struct {
		id       string
		maxPos   float64
	}
	projects := make([]projectInfo, len(projectIDs))
	for i, pid := range projectIDs {
		projects[i].id = pid
		db.QueryRow("SELECT COALESCE(MAX(POSITION_), 0) FROM issue WHERE PROJECT_ID_ = ? AND DELETED_AT_ IS NULL", pid).Scan(&projects[i].maxPos)
	}

	targetTotal := 1000

	// 查询当前 issue 总数
	var currentCount int
	db.QueryRow("SELECT COUNT(*) FROM issue WHERE DELETED_AT_ IS NULL").Scan(&currentCount)
	toAdd := targetTotal - currentCount
	if toAdd <= 0 {
		fmt.Printf("当前已有 %d 个 issues，无需追加。\n", currentCount)
		return
	}

	fmt.Printf("当前 %d 个 issues，将追加 %d 个 → 目标 %d\n", currentCount, toAdd, targetTotal)

	for i := 0; i < toAdd; i++ {
		idx := i % len(projects)
		projects[idx].maxPos++
		pid := projects[idx].id
		pos := projects[idx].maxPos

		title := genTitle(i)
		issueID := genIssueID()
		priority := priorityWeights[rand.IntN(len(priorityWeights))]
		severity := severityWeights[rand.IntN(len(severityWeights))]
		status := statusWeights[rand.IntN(len(statusWeights))]
		statusID := kanban.DefaultWorkflowID + "-status-" + status

		_, err := db.Exec(`
			INSERT INTO issue (
				ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_,
				TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_,
				REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_
			) VALUES (?, ?, ?, ?, ?, ?, '', ?, ?, ?, 0, 0, ?, ?)
		`, issueID, pid, kanban.DefaultWorkflowID,
			stageID, statusID, title, priority, severity, pos, now, now)
		if err != nil {
			fmt.Fprintf(os.Stderr, "插入 issue 失败: %v\n", err)
			os.Exit(1)
		}

		if (i+1)%100 == 0 {
			fmt.Printf("  ✓ 已插入 %d/%d\n", i+1, toAdd)
		}
	}
	fmt.Println("追加完成。")
}

func genTitle(index int) string {
	p := issueTitlePrefixes[rand.IntN(len(issueTitlePrefixes))]
	n := issueTitleNouns[rand.IntN(len(issueTitleNouns))]
	return fmt.Sprintf("%s%s功能（#%d）", p, n, index+1)
}

var (
	lastIssueTick   int64
	lastIssueTickMu sync.Mutex
)

func genIssueID() string {
	tick := time.Now().UnixMilli() / 100
	lastIssueTickMu.Lock()
	if tick > lastIssueTick {
		lastIssueTick = tick
	} else {
		lastIssueTick++
	}
	id := strings.ToUpper(strconv.FormatInt(lastIssueTick, 36))
	lastIssueTickMu.Unlock()
	return id
}