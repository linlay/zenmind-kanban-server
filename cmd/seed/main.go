// Seed 命令用于人工初始化看板数据库种子数据（工作流、默认项目、默认看板）。
// 仅在数据库为空时运行一次，重复运行安全（ON CONFLICT DO NOTHING）。
//
// 用法:
//
//	go run ./cmd/seed/
//	ZENMIND_KANBAN_DB=/path/to/kanban.db go run ./cmd/seed/
package main

import (
	"context"
	"fmt"
	"os"

	"zenmind-kanban-server/internal/config"
	"zenmind-kanban-server/internal/store"
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

	fmt.Printf("正在向 %s 写入种子数据...\n", st.Path())

	if err := st.SeedWorkflowCatalog(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "写入工作流种子数据失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("  ✓ 工作流目录")

	if err := st.SeedDefaults(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "写入默认项目和看板失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("  ✓ 默认项目")
	fmt.Println("  ✓ 默认看板")

	fmt.Println("种子数据写入完成。")
}
