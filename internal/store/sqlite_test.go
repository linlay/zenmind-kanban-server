package store

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
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
