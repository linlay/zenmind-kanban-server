// Seed-demo 通过直接 SQL 向看板数据库批量写入演示数据（50 个项目 + 100 个 issues）。
//
// 用法:
//
//	go run ./cmd/seed-demo/
package main

import (
	"context"
	crand "crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"zenmind-kanban-server/internal/config"
	"zenmind-kanban-server/internal/kanban"
	"zenmind-kanban-server/internal/store"

	_ "modernc.org/sqlite"
)

// projectDef 描述一个带层级关系的项目
type projectDef struct {
	Name      string
	Children  []projectDef
	WithIssue bool // 只有顶层父节点为 true，分配 20 个 issue
}

var projectTree = []projectDef{
	{
		Name:      "电商平台重构",
		WithIssue: true,
		Children: []projectDef{
			{
				Name: "前端商城",
				Children: []projectDef{
					{Name: "商品详情页"},
					{Name: "购物车模块"},
				},
			},
			{
				Name: "后端服务",
				Children: []projectDef{
					{Name: "订单系统"},
					{Name: "支付网关"},
				},
			},
			{Name: "用户中心"},
		},
	},
	{
		Name:      "数据中台建设",
		WithIssue: true,
		Children: []projectDef{
			{Name: "数据采集层"},
			{Name: "数据治理"},
			{Name: "数据服务"},
		},
	},
	{
		Name:      "智能客服机器人",
		WithIssue: true,
		Children: []projectDef{
			{Name: "NLP 引擎"},
			{Name: "对话管理"},
			{Name: "知识库"},
		},
	},
	{
		Name:      "移动端 App 改版",
		WithIssue: true,
		Children: []projectDef{
			{Name: "iOS 客户端"},
			{Name: "Android 客户端"},
			{Name: "小程序端"},
		},
	},
	{
		Name:      "DevOps 流水线",
		WithIssue: true,
		Children: []projectDef{
			{Name: "CI/CD 平台"},
			{Name: "监控告警"},
			{Name: "日志中心"},
		},
	},
	{Name: "OA 办公自动化"},
	{Name: "风控决策引擎"},
	{Name: "推荐系统优化"},
	{Name: "物流调度平台"},
	{Name: "财务对账系统"},
	{Name: "内容管理 CMS"},
	{Name: "直播带货平台"},
	{Name: "在线教育门户"},
	{Name: "医疗预约挂号"},
	{Name: "智慧园区管理"},
	{Name: "物联网设备平台"},
	{Name: "游戏运营后台"},
	{Name: "广告投放系统"},
	{Name: "第三方支付网关"},
	{Name: "短信通道聚合"},
	{Name: "电子合同签署"},
	{Name: "代码审查平台"},
	{Name: "API 网关升级"},
	{Name: "配置中心重构"},
	{Name: "分布式任务调度"},
	{Name: "消息推送中台"},
	{Name: "用户画像标签"},
	{Name: "A/B 实验平台"},
	{Name: "工单流转系统"},
	{Name: "远程桌面网关"},
	{Name: "视频转码流水线"},
}

// 100 个 issue 标题
var issueTitles = []string{
	"登录页响应式适配移动端",
	"订单列表分页加载性能优化",
	"商品详情页图片懒加载",
	"购物车数量增减动画效果",
	"用户头像上传裁剪功能",
	"首页 banner 轮播图自动播放",
	"搜索框联想词下拉列表",
	"个人中心修改密码安全性校验",
	"注册页面短信验证码倒计时",
	"支付结果页面的状态轮询",
	"消息通知红点数字标记",
	"下拉刷新加载更多内容",
	"富文本编辑器工具栏布局",
	"评论列表嵌套回复展开收起",
	"多选删除批量操作确认框",
	"表格列宽拖拽调整大小",
	"表单必填项红色星号提示",
	"侧边栏菜单折叠展开动画",
	"面包屑导航动态生成路径",
	"文件拖拽上传进度条展示",
	"时间选择器日期范围联动",
	"图表折线图数据点悬浮提示",
	"暗黑模式主题切换持久化",
	"多语言国际化文案抽取整理",
	"水印防截图叠加层实现",
	"二维码生成与长按识别",
	"地图选点位置搜索周边",
	"语音输入转文字实时识别",
	"指纹/面容生物认证登录",
	"快捷方式桌面图标生成",
	"第三方 OAuth 授权回调处理",
	"接口请求超时重试策略优化",
	"WebSocket 断线重连心跳机制",
	"离线缓存数据同步冲突合并",
	"灰度发布流量百分比动态调整",
	"慢 SQL 监控告警阈值配置",
	"日志采集链路追踪采样率",
	"数据库连接池大小自适应",
	"Redis 热 key 自动发现迁移",
	"消息队列积压消费者扩容",
	"服务降级熔断半开恢复策略",
	"API 限流令牌桶算法实现",
	"幂等性去重防重复提交",
	"分布式锁续期 watchdog 机制",
	"配置热更新不重启生效",
	"定时任务分片并行执行",
	"数据脱敏敏感字段掩码规则",
	"加密传输国密 SM4 算法支持",
	"审计日志操作留痕不可篡改",
	"白名单 IP 访问控制策略",
	"单点登录 Session 共享方案",
	"跨域 CORS 预检请求缓存",
	"文件分片断点续传合并校验",
	"PDF 预览水印页码渲染",
	"Excel 导入导出大数据量优化",
	"邮件模板变量替换发送队列",
	"站内信已读未读状态管理",
	"操作日志按天归档清理策略",
	"版本发布 Changelog 自动生成",
	"自动化测试用例覆盖率报表",
	"代码扫描 Sonar 规则定制",
	"容器镜像分层构建缓存优化",
	"K8s 滚动更新就绪探针配置",
	"负载均衡会话保持一致性哈希",
	"CDN 缓存预热刷新策略",
	"数据库主从切换故障转移",
	"备份恢复定期演练自动化",
	"灰度 A/B 实验分流规则引擎",
	"移动端推送长连接保活优化",
	"应用内升级强制更新弹窗",
	"崩溃日志符号化堆栈解析",
	"性能埋点启动耗时阶段分析",
	"内存泄漏循环引用检测工具",
	"电量优化后台任务合并调度",
	"网络库 HTTP/3 QUIC 协议适配",
	"视频播放器预加载缓冲策略",
	"直播推流画质自适应码率",
	"即时通讯消息已读回执",
	"社交动态时间线 Feed 流聚合",
	"好友推荐共同联系人算法",
	"附近的人 GeoHash 经纬度编码",
	"红包雨高并发抢购队列设计",
	"秒杀库存扣减 Redis Lua 原子操作",
	"优惠券过期提醒定时扫描",
	"积分商城兑换物流跟踪",
	"拼团活动成团超时自动退款",
	"商品评价图片审核敏感词过滤",
	"售后退款原路返回到账时效",
	"发票抬头智能识别 OCR 提取",
	"会员等级升降规则引擎计算",
	"签到日历连续天数补签卡",
	"数据导出异步任务进度查询",
	"短信通道自动切换故障转移",
	"第三方回调验签防伪造重放",
	"接口文档 Swagger 自动生成",
	"CI/CD 构建产物版本号注入",
	"域名切换灰度 DNS 权重调整",
	"服务注册发现健康检查剔除",
	"链路追踪 Span 上下文透传",
	"日志级别运行时动态调整",
}

func main() {
	cfg := config.Load()
	ctx := context.Background()

	st, err := store.Open(ctx, cfg.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	fmt.Printf("正在向 %s 写入演示数据...\n", st.Path())

	if err := st.SeedWorkflowCatalog(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "写入工作流种子数据失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("  ✓ 工作流目录")

	if err := st.EnsureDefaultProject(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "确保默认项目失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("  ✓ 默认项目")

	if err := st.EnsureDefaultBoard(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "确保默认看板失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("  ✓ 默认看板")

	// 直接开数据库连接执行原始 SQL
	db, err := sql.Open("sqlite", cfg.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开数据库连接失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	db.Exec("PRAGMA busy_timeout = 3000")
	db.Exec("PRAGMA foreign_keys = ON")

	// 清空已有演示数据
	db.Exec("DELETE FROM issue WHERE PROJECT_ID_ != ?", kanban.DefaultProjectID)
	db.Exec("DELETE FROM project_closure")
	db.Exec("DELETE FROM project WHERE ID_ != ?", kanban.DefaultProjectID)
	db.Exec("INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_) VALUES (?, ?, 0)",
		kanban.DefaultProjectID, kanban.DefaultProjectID)

	now := time.Now().UTC().Format(time.RFC3339Nano)
	projectCount := 0
	totalIssues := 0
	issueIndex := 0

	for _, node := range projectTree {
		n, ic := insertProject(ctx, db, node, kanban.DefaultProjectID, now, &issueIndex)
		projectCount += n
		totalIssues += ic
	}

	fmt.Printf("  ✓ 已创建 %d 个演示项目（含层级关系）\n", projectCount)
	fmt.Printf("  ✓ 已创建 %d 个 issues\n", totalIssues)
	fmt.Println("演示数据写入完成。")
}

// insertProject 通过原始 SQL 递归插入项目树，直接维护 project + project_closure + issue
func insertProject(
	ctx context.Context,
	db *sql.DB,
	node projectDef,
	parentID string,
	now string,
	issueIndex *int,
) (int, int) {
	id := genProjectID()
	slug := kanban.ProjectSlugFromName(node.Name)

	// 计算 path
	var path string
	if parentID == kanban.DefaultProjectID {
		path = slug
	} else {
		var parentPath string
		db.QueryRow("SELECT PATH_ FROM project WHERE ID_ = ?", parentID).Scan(&parentPath)
		path = strings.Trim(parentPath, "/") + "/" + slug
	}

	// 计算 depth
	depth := 1
	if parentID != kanban.DefaultProjectID {
		db.QueryRow("SELECT DEPTH_ FROM project WHERE ID_ = ?", parentID).Scan(&depth)
		depth++
	}

	// 插入 project
	db.Exec(`
		INSERT INTO project (
			ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_,
			VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_
		) VALUES (?, ?, ?, ?, ?, '', ?, ?, 0, 'workspace', ?, ?, ?)
	`, id, parentID, slug, strings.ToUpper(slug), node.Name, path, depth, kanban.DefaultWorkflowID, now, now)

	// project_closure: self
	db.Exec(`INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_) VALUES (?, ?, 0)`, id, id)

	// project_closure: ancestors → self
	db.Exec(`
		INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
		SELECT ANCESTOR_ID_, ?, DEPTH_ + 1
		FROM project_closure
		WHERE DESCENDANT_ID_ = ?
	`, id, parentID)

	childProjects := 1
	childIssues := 0

	// 只有标记了 WithIssue 的顶层项目才分配 issue（前 5 个，各 20 个）
	needIssues := node.WithIssue && *issueIndex < 100
	if needIssues {
		count := 20
		priorities := []string{"high", "medium", "low"}
		severities := []string{"critical", "high", "medium", "low"}
		statuses := []string{"backlog", "todo", "in_progress", "in_review", "completed"}
		stageID := kanban.DefaultWorkflowID + "-stage-requirement_clarification"

		for j := 0; j < count && *issueIndex+j < len(issueTitles); j++ {
			title := issueTitles[*issueIndex+j]
			issueID := genIssueID()
			status := statuses[rand.IntN(len(statuses))]
			statusID := kanban.DefaultWorkflowID + "-status-" + status

			db.Exec(`
				INSERT INTO issue (
					ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_,
					TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_,
					REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_
				) VALUES (?, ?, ?, ?, ?, ?, '', ?, ?, ?, 0, 0, ?, ?)
			`, issueID, id, kanban.DefaultWorkflowID,
				stageID, statusID, title,
				priorities[rand.IntN(len(priorities))],
				severities[rand.IntN(len(severities))],
				float64(j+1), now, now)
			childIssues++
		}
		*issueIndex += count
	}

	// 递归子项目
	for _, child := range node.Children {
		n, ic := insertProject(ctx, db, child, id, now, issueIndex)
		childProjects += n
		childIssues += ic
	}

	return childProjects, childIssues
}

// ========== ID 生成 ==========

func genProjectID() string {
	var random [4]byte
	if _, err := crand.Read(random[:]); err != nil {
		return strings.ToUpper(hex.EncodeToString([]byte(time.Now().Format("150405.000000000"))))
	}
	return strings.ToUpper(strings.TrimLeft(time.Now().UTC().Format("20060102150405"), "0") + hex.EncodeToString(random[:]))
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