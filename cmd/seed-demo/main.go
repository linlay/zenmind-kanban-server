// Seed-demo 从 SQL 文件向看板数据库批量写入演示数据（50 个项目 + 100 个 issues）。
// SQL 文件位于 cmd/demo/ 目录下。
//
// 用法:
//
//	go run ./cmd/seed-demo/
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"zenmind-kanban-server/internal/config"
	"zenmind-kanban-server/internal/store"

	_ "modernc.org/sqlite"
)

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

	// 获取 SQL 文件目录
	demoDir := demoSQLDir()
	files := []string{"01_projects.sql", "02_issues.sql"}
	for _, f := range files {
		path := filepath.Join(demoDir, f)
		if err := executeSQLFile(db, path); err != nil {
			fmt.Fprintf(os.Stderr, "执行 %s 失败: %v\n", f, err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ %s\n", f)
	}

	fmt.Println("演示数据写入完成。")
}

func demoSQLDir() string {
	if dir := os.Getenv("ZENMIND_KANBAN_DEMO_DIR"); dir != "" {
		return dir
	}
	cwd, _ := os.Getwd()
	candidate := filepath.Join(cwd, "cmd", "demo")
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return candidate
}

func executeSQLFile(db *sql.DB, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取文件: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	sql := strings.ReplaceAll(string(data), "__NOW__", now)

	for _, statement := range splitSQLStatements(sql) {
		statement = strings.TrimSpace(statement)
		if statement == "" || strings.HasPrefix(statement, "--") {
			continue
		}
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("%w\n  SQL: %s", err, truncate(statement, 120))
		}
	}
	return nil
}

func splitSQLStatements(s string) []string {
	var parts []string
	current := ""
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		current += line + "\n"
		if strings.HasSuffix(trimmed, ";") {
			parts = append(parts, current)
			current = ""
		}
	}
	if strings.TrimSpace(current) != "" {
		parts = append(parts, current)
	}
	return parts
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}