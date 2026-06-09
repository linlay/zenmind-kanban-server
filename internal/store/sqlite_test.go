package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"zenmind-kanban-server/internal/kanban"
)

func TestMigrationCreatesDomainTablesWithUppercaseColumns(t *testing.T) {
	sqliteStore, err := Open(context.Background(), filepath.Join(t.TempDir(), "kanban.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer sqliteStore.Close()

	expectedTables := []string{
		"user_account",
		"team",
		"team_member",
		"project",
		"project_closure",
		"project_permission",
		"workflow",
		"workflow_stage_def",
		"workflow_status_def",
		"workflow_stage",
		"workflow_status",
		"workflow_transition",
		"issue",
		"issue_label",
		"issue_label_link",
		"issue_attachment",
		"issue_dependency",
		"issue_automation",
		"agent",
		"agent_run",
		"agent_tool_call",
		"review",
		"review_comment",
		"board",
		"board_meta",
		"event_log",
		"desktop_client",
	}
	for _, table := range expectedTables {
		var count int
		if err := sqliteStore.db.QueryRowContext(context.Background(), `
			SELECT COUNT(*)
			FROM sqlite_master
			WHERE type = 'table' AND name = ?
		`, table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("expected table %s to exist", table)
		}

		rows, err := sqliteStore.db.QueryContext(context.Background(), fmt.Sprintf(`PRAGMA table_info(%s)`, table))
		if err != nil {
			t.Fatal(err)
		}
		columnCount := 0
		for rows.Next() {
			var cid int
			var name string
			var columnType string
			var notNull int
			var defaultValue any
			var pk int
			if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
				_ = rows.Close()
				t.Fatal(err)
			}
			columnCount++
			if name != strings.ToUpper(name) || !strings.HasSuffix(name, "_") {
				_ = rows.Close()
				t.Fatalf("expected uppercase underscore column in %s, got %s", table, name)
			}
		}
		if err := rows.Close(); err != nil {
			t.Fatal(err)
		}
		if columnCount == 0 {
			t.Fatalf("expected table %s to have columns", table)
		}
	}
}

func TestSchemaCreatesFinalWorkflowColumns(t *testing.T) {
	sqliteStore, err := Open(context.Background(), filepath.Join(t.TempDir(), "kanban.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer sqliteStore.Close()

	expected := map[string][]string{
		"workflow":            {"TRANSITION_MODE_"},
		"workflow_status":     {"STAGE_ID_", "IS_ACTIVE_"},
		"workflow_transition": {"IS_ACTIVE_"},
		"issue":               {"SEVERITY_"},
		"desktop_client":      {"DEVICE_ID_", "CURRENT_USER_ID_", "CURRENT_USER_NAME_"},
	}
	for table, columns := range expected {
		for _, column := range columns {
			var count int
			if err := sqliteStore.db.QueryRowContext(context.Background(), `
				SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?
			`, table, column).Scan(&count); err != nil {
				t.Fatal(err)
			}
			if count != 1 {
				t.Fatalf("expected %s.%s to exist", table, column)
			}
		}
	}
}

func TestWorkflowStatusContractForIssueStatusFields(t *testing.T) {
	ctx := context.Background()
	sqliteStore, err := Open(ctx, filepath.Join(t.TempDir(), "kanban.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer sqliteStore.Close()

	if err := sqliteStore.SeedWorkflowCatalog(ctx); err != nil {
		t.Fatal(err)
	}
	if err := sqliteStore.SeedDefaults(ctx); err != nil {
		t.Fatal(err)
	}

	allowedColumnKeys := map[string]bool{
		"backlog": true, "todo": true, "in_progress": true, "in_review": true, "completed": true,
	}
	rows, err := sqliteStore.db.QueryContext(ctx, `
		SELECT KEY_, NAME_, COLUMN_KEY_
		FROM workflow_status
	`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var key, name, columnKey string
		if err := rows.Scan(&key, &name, &columnKey); err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(name) == "" {
			t.Fatalf("workflow_status %s has empty NAME_", key)
		}
		if strings.TrimSpace(key) == "" {
			t.Fatal("workflow_status has empty KEY_")
		}
		if !allowedColumnKeys[columnKey] {
			t.Fatalf("unexpected workflow_status COLUMN_KEY_: %s", columnKey)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	issue := kanban.Issue{
		BoardID:        kanban.DefaultBoardID,
		ProjectID:      kanban.DefaultProjectID,
		WorkflowID:     kanban.DefaultWorkflowID,
		StageID:        "workflow-standard-requirement-stage-solution_design",
		StatusID:       "workflow-standard-requirement-status-waiting_approval",
		ID:             "status-contract-issue",
		Title:          "Status contract issue",
		Description:    "",
		Status:         kanban.StatusInReview,
		Priority:       kanban.PriorityMedium,
		Severity:       kanban.SeverityMedium,
		Position:       1,
		ReviewRequired: true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if _, err := sqliteStore.ReplaceIssue(ctx, kanban.DefaultBoardID, issue, "test.issue.created", "test"); err != nil {
		t.Fatal(err)
	}
	issues, _, err := sqliteStore.ListIssues(ctx, kanban.DefaultBoardID, kanban.DefaultProjectID)
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	got := issues[0]
	if got.StatusID != issue.StatusID {
		t.Fatalf("statusId mismatch: got %s want %s", got.StatusID, issue.StatusID)
	}
	if got.StatusKey != "waiting_approval" {
		t.Fatalf("statusKey mismatch: got %s", got.StatusKey)
	}
	if got.StatusName != "等待批准" {
		t.Fatalf("statusName mismatch: got %s", got.StatusName)
	}
	if got.Status != kanban.StatusInReview {
		t.Fatalf("status mismatch: got %s", got.Status)
	}
	if got.ColumnKey != "in_review" {
		t.Fatalf("columnKey mismatch: got %s", got.ColumnKey)
	}
}

func TestDemoSQLResolvesIssueStatusFields(t *testing.T) {
	ctx := context.Background()
	sqliteStore, err := Open(ctx, filepath.Join(t.TempDir(), "kanban.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer sqliteStore.Close()

	if err := sqliteStore.SeedWorkflowCatalog(ctx); err != nil {
		t.Fatal(err)
	}
	if err := sqliteStore.SeedDefaults(ctx); err != nil {
		t.Fatal(err)
	}
	nowTime := time.Date(2026, time.June, 9, 6, 39, 39, 123456789, time.UTC)
	now := nowTime.Format(time.RFC3339Nano)
	for _, file := range []string{
		filepath.Join("..", "..", "cmd", "demo", "01_projects.sql"),
		filepath.Join("..", "..", "cmd", "demo", "02_issues.sql"),
	} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		sql := strings.ReplaceAll(string(data), "__NOW__", now)
		if err := executeSQLScript(ctx, sqliteStore.db, sql, "demo sql"); err != nil {
			t.Fatal(err)
		}
	}

	var issueCount int
	if err := sqliteStore.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM issue WHERE CREATED_BY_ = 'demo'
	`).Scan(&issueCount); err != nil {
		t.Fatal(err)
	}
	if issueCount != 200 {
		t.Fatalf("expected 200 demo issues to be inserted, got %d", issueCount)
	}

	var rootlessDemoProjectCount int
	if err := sqliteStore.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM project
		WHERE ID_ LIKE 'demo-proj-%'
			AND PARENT_ID_ IS NULL
	`).Scan(&rootlessDemoProjectCount); err != nil {
		t.Fatal(err)
	}
	if rootlessDemoProjectCount != 0 {
		t.Fatalf("expected demo projects to be under default root, got %d rootless", rootlessDemoProjectCount)
	}

	var defaultClosureCount int
	if err := sqliteStore.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM project_closure
		WHERE ANCESTOR_ID_ = 'default'
			AND DESCENDANT_ID_ LIKE 'demo-proj-%'
	`).Scan(&defaultClosureCount); err != nil {
		t.Fatal(err)
	}
	if defaultClosureCount != 50 {
		t.Fatalf("expected default closure to include 50 demo projects, got %d", defaultClosureCount)
	}

	defaultIssues, _, err := sqliteStore.ListIssues(ctx, kanban.DefaultBoardID, kanban.DefaultProjectID)
	if err != nil {
		t.Fatal(err)
	}
	if len(defaultIssues) != 200 {
		t.Fatalf("expected default project to aggregate 200 demo issues, got %d", len(defaultIssues))
	}

	projectIssueStats, err := sqliteStore.ListProjectIssueStats(ctx, kanban.DefaultBoardID)
	if err != nil {
		t.Fatal(err)
	}
	defaultStatCount := -1
	for _, stat := range projectIssueStats {
		if stat.ProjectID == kanban.DefaultProjectID {
			defaultStatCount = stat.IssueCount
			break
		}
	}
	if defaultStatCount != 200 {
		t.Fatalf("expected default project issue stats to aggregate 200 demo issues, got %d", defaultStatCount)
	}

	var longTitleCount int
	if err := sqliteStore.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM issue
		WHERE CREATED_BY_ = 'demo'
			AND length(TITLE_) >= 70
	`).Scan(&longTitleCount); err != nil {
		t.Fatal(err)
	}
	if longTitleCount < 30 {
		t.Fatalf("expected demo issues to include many long titles, got %d", longTitleCount)
	}

	var titleContainsIDCount int
	if err := sqliteStore.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM issue
		WHERE CREATED_BY_ = 'demo'
			AND instr(TITLE_, ID_) > 0
	`).Scan(&titleContainsIDCount); err != nil {
		t.Fatal(err)
	}
	if titleContainsIDCount != 0 {
		t.Fatalf("expected demo issue titles not to include their ids, got %d", titleContainsIDCount)
	}

	var invalidIssueIDCount int
	if err := sqliteStore.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM issue
		WHERE CREATED_BY_ = 'demo'
			AND (ID_ GLOB '*[^0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ]*' OR ID_ GLOB 'DEMO*')
	`).Scan(&invalidIssueIDCount); err != nil {
		t.Fatal(err)
	}
	if invalidIssueIDCount != 0 {
		t.Fatalf("expected demo issue ids to be uppercase base36 strings, got %d invalid", invalidIssueIDCount)
	}

	var sampleBase36IDCount int
	timestampTick := nowTime.UnixMilli() / 100
	expectedBase36IDs := []string{
		strings.ToUpper(strconv.FormatInt(timestampTick, 36)),
		strings.ToUpper(strconv.FormatInt(timestampTick+9, 36)),
		strings.ToUpper(strconv.FormatInt(timestampTick+35, 36)),
		strings.ToUpper(strconv.FormatInt(timestampTick+199, 36)),
	}
	if err := sqliteStore.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM issue
		WHERE CREATED_BY_ = 'demo'
			AND ID_ IN (?, ?, ?, ?)
	`, expectedBase36IDs[0], expectedBase36IDs[1], expectedBase36IDs[2], expectedBase36IDs[3]).Scan(&sampleBase36IDCount); err != nil {
		t.Fatal(err)
	}
	if sampleBase36IDCount != 4 {
		t.Fatalf("expected demo issues to include timestamp base36 ids %v; got %d", expectedBase36IDs, sampleBase36IDCount)
	}

	var missingStatusCount int
	if err := sqliteStore.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM issue i
		LEFT JOIN workflow_status wst ON wst.ID_ = i.STATUS_ID_
		WHERE i.CREATED_BY_ = 'demo'
			AND (wst.ID_ IS NULL OR wst.KEY_ = '' OR wst.NAME_ = '')
	`).Scan(&missingStatusCount); err != nil {
		t.Fatal(err)
	}
	if missingStatusCount != 0 {
		t.Fatalf("expected all demo issues to resolve statusKey/statusName, got %d missing", missingStatusCount)
	}

	expectedStatuses := map[string]string{
		"waiting_answer":              "等待回答",
		"waiting_submit":              "等待提交",
		"waiting_approval":            "等待批准",
		"success":                     "成功",
		"failed":                      "失败",
		"interrupted":                 "中断",
		"testing_waiting_submit":      "测试等待提交",
		"testing_in_progress":         "测试中",
		"testing_waiting_approval":    "测试等待批准",
		"testing_success":             "测试成功",
		"testing_failed":              "测试失败",
		"testing_interrupted":         "测试中断",
		"deployment_waiting_submit":   "部署等待提交",
		"deployment_in_progress":      "部署中",
		"deployment_waiting_approval": "部署等待批准",
		"deployment_success":          "部署成功",
		"deployment_failed":           "部署失败",
		"deployment_interrupted":      "部署中断",
	}
	for statusKey, statusName := range expectedStatuses {
		var count int
		if err := sqliteStore.db.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM issue i
			JOIN workflow_status wst ON wst.ID_ = i.STATUS_ID_
			WHERE i.CREATED_BY_ = 'demo'
				AND wst.KEY_ = ?
				AND wst.NAME_ = ?
		`, statusKey, statusName).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count == 0 {
			t.Fatalf("expected demo issues to include statusKey=%s statusName=%s", statusKey, statusName)
		}
	}

	var invalidColumnKeyCount int
	if err := sqliteStore.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM issue i
		JOIN workflow_status wst ON wst.ID_ = i.STATUS_ID_
		WHERE i.CREATED_BY_ = 'demo'
			AND wst.COLUMN_KEY_ NOT IN ('backlog','todo','in_progress','in_review','completed')
	`).Scan(&invalidColumnKeyCount); err != nil {
		t.Fatal(err)
	}
	if invalidColumnKeyCount != 0 {
		t.Fatalf("expected demo issues to use only board column status values, got %d invalid", invalidColumnKeyCount)
	}
}

func TestMigrationRepairsLegacyProjectClosureColumns(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "kanban.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(context.Background(), `
		CREATE TABLE project_closure (
			ancestor_id TEXT NOT NULL,
			descendant_id TEXT NOT NULL,
			depth INTEGER NOT NULL
		)
	`); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO project_closure (ancestor_id, descendant_id, depth)
		VALUES ('legacy-root', 'legacy-child', 1)
	`); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	// Insert projects so rebuildProjectClosure has data to work with.
	if _, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS project (
			ID_ TEXT PRIMARY KEY,
			PARENT_ID_ TEXT,
			SLUG_ TEXT NOT NULL,
			KEY_ TEXT NOT NULL DEFAULT '',
			NAME_ TEXT NOT NULL,
			DESCRIPTION_ TEXT NOT NULL DEFAULT '',
			PATH_ TEXT NOT NULL UNIQUE,
			DEPTH_ INTEGER NOT NULL DEFAULT 0,
			POSITION_ REAL NOT NULL DEFAULT 0,
			VISIBILITY_ TEXT NOT NULL DEFAULT 'workspace',
			DEFAULT_WORKFLOW_ID_ TEXT NOT NULL DEFAULT '',
			ARCHIVED_AT_ TEXT,
			CREATED_BY_ TEXT,
			UPDATED_BY_ TEXT,
			CREATED_AT_ TEXT NOT NULL,
			UPDATED_AT_ TEXT NOT NULL,
			DELETED_AT_ TEXT
		)
	`); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, PATH_, DEPTH_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
		VALUES ('proj-1', NULL, 'one', 'ONE', 'Project One', 'one', 0, '', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')
	`); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	sqliteStore, err := Open(context.Background(), dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer sqliteStore.Close()

	columns := map[string]bool{}
	rows, err := sqliteStore.db.QueryContext(context.Background(), `PRAGMA table_info(project_closure)`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var cid int
		var name string
		var columnType string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			_ = rows.Close()
			t.Fatal(err)
		}
		columns[name] = true
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	for _, column := range []string{"ANCESTOR_ID_", "DESCENDANT_ID_", "DEPTH_"} {
		if !columns[column] {
			t.Fatalf("expected repaired project_closure to include %s", column)
		}
	}

	var selfClosureCount int
	if err := sqliteStore.db.QueryRowContext(context.Background(), `
		SELECT COUNT(*) FROM project_closure
		WHERE ANCESTOR_ID_ = DESCENDANT_ID_ AND DEPTH_ = 0
	`).Scan(&selfClosureCount); err != nil {
		t.Fatal(err)
	}
	if selfClosureCount == 0 {
		t.Fatalf("expected project_closure to be rebuilt from valid projects")
	}
}

func TestMigrationAddsWorkflowTransitionModeColumn(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "kanban.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(context.Background(), `
		CREATE TABLE workflow (
			ID_ TEXT PRIMARY KEY,
			KEY_ TEXT NOT NULL UNIQUE,
			NAME_ TEXT NOT NULL,
			DESCRIPTION_ TEXT NOT NULL DEFAULT '',
			IS_DEFAULT_ INTEGER NOT NULL DEFAULT 0 CHECK (IS_DEFAULT_ IN (0, 1)),
			CREATED_BY_ TEXT,
			UPDATED_BY_ TEXT,
			CREATED_AT_ TEXT NOT NULL,
			UPDATED_AT_ TEXT NOT NULL,
			DELETED_AT_ TEXT
		)
	`); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	sqliteStore, err := Open(context.Background(), dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer sqliteStore.Close()

	var transitionMode string
	if err := sqliteStore.db.QueryRowContext(context.Background(), `
		SELECT dflt_value FROM pragma_table_info('workflow') WHERE name = 'TRANSITION_MODE_'
	`).Scan(&transitionMode); err != nil {
		t.Fatal(err)
	}
	if transitionMode != "'strict'" {
		t.Fatalf("expected TRANSITION_MODE_ default 'strict', got %q", transitionMode)
	}
}

func TestMigrationDropsLegacyIssueTypeColumn(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "kanban.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(context.Background(), `
		CREATE TABLE issue (
			ID_ TEXT PRIMARY KEY,
			PROJECT_ID_ TEXT NOT NULL,
			WORKFLOW_ID_ TEXT NOT NULL,
			TYPE_ID_ TEXT NOT NULL,
			STAGE_ID_ TEXT NOT NULL,
			STATUS_ID_ TEXT NOT NULL,
			TITLE_ TEXT NOT NULL CHECK (length(trim(TITLE_)) > 0),
			DESCRIPTION_ TEXT NOT NULL DEFAULT '',
			PRIORITY_ TEXT NOT NULL CHECK (PRIORITY_ IN ('high','medium','low')),
			SEVERITY_ TEXT NOT NULL DEFAULT 'medium' CHECK (SEVERITY_ IN ('critical','high','medium','low')),
			POSITION_ REAL NOT NULL,
			ASSIGNEE_ID_ TEXT,
			WORKER_TYPE_ TEXT CHECK (WORKER_TYPE_ IN ('human','agent')),
			WORKER_ID_ TEXT,
			WORKER_AGENT_ TEXT,
			REVIEWER_ID_ TEXT,
			REVIEW_REQUIRED_ INTEGER NOT NULL DEFAULT 0 CHECK (REVIEW_REQUIRED_ IN (0, 1)),
			ACTIVE_REVIEW_ID_ TEXT,
			ACTIVE_RUN_ID_ TEXT,
			REVISION_ INTEGER NOT NULL DEFAULT 0,
			CREATED_BY_ TEXT,
			UPDATED_BY_ TEXT,
			CREATED_BY_AGENT_ TEXT,
			UPDATED_BY_AGENT_ TEXT,
			CREATED_AT_ TEXT NOT NULL,
			UPDATED_AT_ TEXT NOT NULL,
			DELETED_AT_ TEXT,
			CHECK (
				WORKER_TYPE_ IS NULL OR
				(WORKER_TYPE_ = 'human' AND WORKER_ID_ IS NOT NULL AND WORKER_AGENT_ IS NULL) OR
				(WORKER_TYPE_ = 'agent' AND WORKER_AGENT_ IS NOT NULL AND WORKER_ID_ IS NULL)
			)
		)
	`); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	sqliteStore, err := Open(context.Background(), dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer sqliteStore.Close()

	var typeColumnCount int
	if err := sqliteStore.db.QueryRowContext(context.Background(), `
		SELECT COUNT(*) FROM pragma_table_info('issue') WHERE name = 'TYPE_ID_'
	`).Scan(&typeColumnCount); err != nil {
		t.Fatal(err)
	}
	if typeColumnCount != 0 {
		t.Fatalf("expected legacy TYPE_ID_ column to be removed, got %d", typeColumnCount)
	}
}
