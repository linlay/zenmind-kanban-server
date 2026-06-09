package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"zenmind-kanban-server/internal/kanban"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

//go:embed workflow_seed.sql
var workflowSeedSQL string

//go:embed seed_defaults.sql
var seedDefaultsSQL string

type Store struct {
	db   *sql.DB
	path string
}

func Open(ctx context.Context, databasePath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, err
	}
	store := &Store{db: db, path: databasePath}
	if err := store.migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Path() string {
	return s.path
}

func (s *Store) migrate(ctx context.Context) error {
	if err := executeSchemaScript(ctx, s.db); err != nil {
		return err
	}
	statements := []string{
		`PRAGMA busy_timeout = 3000`,
		`PRAGMA journal_mode = WAL`,
		`PRAGMA foreign_keys = ON`,
		`CREATE TABLE IF NOT EXISTS user_account (
			ID_ TEXT PRIMARY KEY,
			EMAIL_ TEXT NOT NULL UNIQUE,
			DISPLAY_NAME_ TEXT NOT NULL,
			AVATAR_URL_ TEXT,
			STATUS_ TEXT NOT NULL DEFAULT 'active',
			CREATED_AT_ TEXT NOT NULL,
			UPDATED_AT_ TEXT NOT NULL,
			DELETED_AT_ TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS team (
			ID_ TEXT PRIMARY KEY,
			SLUG_ TEXT NOT NULL UNIQUE,
			NAME_ TEXT NOT NULL,
			DESCRIPTION_ TEXT NOT NULL DEFAULT '',
			CREATED_BY_ TEXT,
			UPDATED_BY_ TEXT,
			CREATED_AT_ TEXT NOT NULL,
			UPDATED_AT_ TEXT NOT NULL,
			DELETED_AT_ TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS team_member (
			TEAM_ID_ TEXT NOT NULL REFERENCES team(ID_),
			USER_ID_ TEXT NOT NULL REFERENCES user_account(ID_),
			ROLE_ TEXT NOT NULL CHECK (ROLE_ IN ('owner','admin','member')),
			INVITED_BY_ TEXT,
			JOINED_AT_ TEXT NOT NULL,
			LEFT_AT_ TEXT,
			PRIMARY KEY (TEAM_ID_, USER_ID_)
		)`,
		`CREATE TABLE IF NOT EXISTS workflow (
			ID_ TEXT PRIMARY KEY,
			KEY_ TEXT NOT NULL UNIQUE,
			NAME_ TEXT NOT NULL,
			DESCRIPTION_ TEXT NOT NULL DEFAULT '',
			IS_DEFAULT_ INTEGER NOT NULL DEFAULT 0 CHECK (IS_DEFAULT_ IN (0, 1)),
			TRANSITION_MODE_ TEXT NOT NULL DEFAULT 'strict' CHECK (TRANSITION_MODE_ IN ('strict', 'free')),
			CREATED_BY_ TEXT,
			UPDATED_BY_ TEXT,
			CREATED_AT_ TEXT NOT NULL,
			UPDATED_AT_ TEXT NOT NULL,
			DELETED_AT_ TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS workflow_stage_def (
			ID_ TEXT PRIMARY KEY,
			KEY_ TEXT NOT NULL UNIQUE,
			NAME_ TEXT NOT NULL,
			DESCRIPTION_ TEXT NOT NULL DEFAULT '',
			IS_SYSTEM_ INTEGER NOT NULL DEFAULT 1 CHECK (IS_SYSTEM_ IN (0, 1)),
			CREATED_AT_ TEXT NOT NULL,
			UPDATED_AT_ TEXT NOT NULL,
			DELETED_AT_ TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS workflow_status_def (
			ID_ TEXT PRIMARY KEY,
			KEY_ TEXT NOT NULL UNIQUE,
			NAME_ TEXT NOT NULL,
			COLUMN_KEY_ TEXT NOT NULL CHECK (COLUMN_KEY_ IN ('backlog','todo','in_progress','in_review','completed')),
			DESCRIPTION_ TEXT NOT NULL DEFAULT '',
			IS_SYSTEM_ INTEGER NOT NULL DEFAULT 1 CHECK (IS_SYSTEM_ IN (0, 1)),
			CREATED_AT_ TEXT NOT NULL,
			UPDATED_AT_ TEXT NOT NULL,
			DELETED_AT_ TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS workflow_stage (
			ID_ TEXT PRIMARY KEY,
			WORKFLOW_ID_ TEXT NOT NULL REFERENCES workflow(ID_),
			STAGE_DEF_ID_ TEXT NOT NULL REFERENCES workflow_stage_def(ID_),
			KEY_ TEXT NOT NULL,
			NAME_ TEXT NOT NULL,
			POSITION_ INTEGER NOT NULL,
			IS_START_ INTEGER NOT NULL DEFAULT 0 CHECK (IS_START_ IN (0, 1)),
			IS_END_ INTEGER NOT NULL DEFAULT 0 CHECK (IS_END_ IN (0, 1)),
			UNIQUE (WORKFLOW_ID_, KEY_)
		)`,
		`CREATE TABLE IF NOT EXISTS workflow_status (
			ID_ TEXT PRIMARY KEY,
			WORKFLOW_ID_ TEXT NOT NULL REFERENCES workflow(ID_),
			STATUS_DEF_ID_ TEXT NOT NULL REFERENCES workflow_status_def(ID_),
			KEY_ TEXT NOT NULL,
			NAME_ TEXT NOT NULL,
			COLUMN_KEY_ TEXT NOT NULL CHECK (COLUMN_KEY_ IN ('backlog','todo','in_progress','in_review','completed')),
			POSITION_ INTEGER NOT NULL,
			IS_START_ INTEGER NOT NULL DEFAULT 0 CHECK (IS_START_ IN (0, 1)),
			IS_TERMINAL_ INTEGER NOT NULL DEFAULT 0 CHECK (IS_TERMINAL_ IN (0, 1)),
			REVIEW_REQUIRED_ INTEGER NOT NULL DEFAULT 0 CHECK (REVIEW_REQUIRED_ IN (0, 1)),
			UNIQUE (WORKFLOW_ID_, KEY_)
		)`,
		`CREATE TABLE IF NOT EXISTS workflow_transition (
			ID_ TEXT PRIMARY KEY,
			WORKFLOW_ID_ TEXT NOT NULL REFERENCES workflow(ID_),
			FROM_STAGE_ID_ TEXT NOT NULL REFERENCES workflow_stage(ID_),
			FROM_STATUS_ID_ TEXT NOT NULL REFERENCES workflow_status(ID_),
			TO_STAGE_ID_ TEXT NOT NULL REFERENCES workflow_stage(ID_),
			TO_STATUS_ID_ TEXT NOT NULL REFERENCES workflow_status(ID_),
			ACTION_KEY_ TEXT NOT NULL,
			NAME_ TEXT NOT NULL,
			ACTOR_TYPE_ TEXT NOT NULL CHECK (ACTOR_TYPE_ IN ('human','agent','system')),
			REQUIRES_REVIEW_ INTEGER NOT NULL DEFAULT 0 CHECK (REQUIRES_REVIEW_ IN (0, 1)),
			POSITION_ INTEGER NOT NULL DEFAULT 0,
			CREATED_AT_ TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS project (
			ID_ TEXT PRIMARY KEY,
			PARENT_ID_ TEXT REFERENCES project(ID_),
			SLUG_ TEXT NOT NULL,
			KEY_ TEXT NOT NULL DEFAULT '',
			NAME_ TEXT NOT NULL,
			DESCRIPTION_ TEXT NOT NULL DEFAULT '',
			PATH_ TEXT NOT NULL UNIQUE,
			DEPTH_ INTEGER NOT NULL DEFAULT 0,
			POSITION_ REAL NOT NULL DEFAULT 0,
			VISIBILITY_ TEXT NOT NULL DEFAULT 'workspace' CHECK (VISIBILITY_ IN ('private','team','workspace')),
			DEFAULT_WORKFLOW_ID_ TEXT NOT NULL REFERENCES workflow(ID_),
			ARCHIVED_AT_ TEXT,
			CREATED_BY_ TEXT,
			UPDATED_BY_ TEXT,
			CREATED_AT_ TEXT NOT NULL,
			UPDATED_AT_ TEXT NOT NULL,
			DELETED_AT_ TEXT,
			UNIQUE (PARENT_ID_, SLUG_)
		)`,
		`CREATE TABLE IF NOT EXISTS project_closure (
			ANCESTOR_ID_ TEXT NOT NULL REFERENCES project(ID_),
			DESCENDANT_ID_ TEXT NOT NULL REFERENCES project(ID_),
			DEPTH_ INTEGER NOT NULL,
			PRIMARY KEY (ANCESTOR_ID_, DESCENDANT_ID_)
		)`,
		`CREATE TABLE IF NOT EXISTS project_permission (
			ID_ TEXT PRIMARY KEY,
			PROJECT_ID_ TEXT NOT NULL REFERENCES project(ID_),
			PRINCIPAL_TYPE_ TEXT NOT NULL CHECK (PRINCIPAL_TYPE_ IN ('user','team')),
			PRINCIPAL_ID_ TEXT NOT NULL,
			ROLE_ TEXT NOT NULL CHECK (ROLE_ IN ('owner','admin','maintainer','developer','reviewer','viewer')),
			INHERIT_TO_CHILDREN_ INTEGER NOT NULL DEFAULT 1 CHECK (INHERIT_TO_CHILDREN_ IN (0, 1)),
			CREATED_BY_ TEXT,
			CREATED_AT_ TEXT NOT NULL,
			DELETED_AT_ TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS agent (
			ID_ TEXT PRIMARY KEY,
			AGENT_KEY_ TEXT NOT NULL UNIQUE,
			NAME_ TEXT NOT NULL,
			DESCRIPTION_ TEXT NOT NULL DEFAULT '',
			ROLE_ TEXT NOT NULL DEFAULT '',
			MODEL_ TEXT,
			MODEL_VERSION_ TEXT,
			ENABLED_ INTEGER NOT NULL DEFAULT 1 CHECK (ENABLED_ IN (0, 1)),
			CREATED_BY_ TEXT,
			UPDATED_BY_ TEXT,
			CREATED_AT_ TEXT NOT NULL,
			UPDATED_AT_ TEXT NOT NULL,
			DELETED_AT_ TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS review (
			ID_ TEXT PRIMARY KEY,
			ISSUE_ID_ TEXT NOT NULL,
			AGENT_RUN_ID_ TEXT,
			REVIEW_TYPE_ TEXT NOT NULL CHECK (REVIEW_TYPE_ IN ('agent_output','peer')),
			REVIEWER_ID_ TEXT,
			STATUS_ TEXT NOT NULL CHECK (STATUS_ IN ('pending','approved','changes_requested','rejected')),
			REQUESTED_BY_ TEXT,
			REQUESTED_AT_ TEXT NOT NULL,
			SUBMITTED_AT_ TEXT,
			SUMMARY_ TEXT NOT NULL DEFAULT '',
			CREATED_AT_ TEXT NOT NULL,
			UPDATED_AT_ TEXT NOT NULL,
			DELETED_AT_ TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS issue (
			ID_ TEXT PRIMARY KEY,
			PROJECT_ID_ TEXT NOT NULL REFERENCES project(ID_),
			WORKFLOW_ID_ TEXT NOT NULL REFERENCES workflow(ID_),
			STAGE_ID_ TEXT NOT NULL REFERENCES workflow_stage(ID_),
			STATUS_ID_ TEXT NOT NULL REFERENCES workflow_status(ID_),
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
			ACTIVE_REVIEW_ID_ TEXT REFERENCES review(ID_),
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
		)`,
		`CREATE TABLE IF NOT EXISTS issue_label (
			ID_ TEXT PRIMARY KEY,
			PROJECT_ID_ TEXT NOT NULL REFERENCES project(ID_),
			KEY_ TEXT NOT NULL,
			NAME_ TEXT NOT NULL,
			COLOR_ TEXT NOT NULL DEFAULT '',
			CREATED_BY_ TEXT,
			UPDATED_BY_ TEXT,
			CREATED_AT_ TEXT NOT NULL,
			UPDATED_AT_ TEXT NOT NULL,
			DELETED_AT_ TEXT,
			UNIQUE (PROJECT_ID_, KEY_)
		)`,
		`CREATE TABLE IF NOT EXISTS issue_label_link (
			ISSUE_ID_ TEXT NOT NULL REFERENCES issue(ID_),
			LABEL_ID_ TEXT NOT NULL REFERENCES issue_label(ID_),
			PRIMARY KEY (ISSUE_ID_, LABEL_ID_)
		)`,
		`CREATE TABLE IF NOT EXISTS issue_attachment (
			ID_ TEXT PRIMARY KEY,
			ISSUE_ID_ TEXT NOT NULL REFERENCES issue(ID_),
			KIND_ TEXT NOT NULL DEFAULT '',
			NAME_ TEXT NOT NULL DEFAULT '',
			MIME_TYPE_ TEXT NOT NULL DEFAULT '',
			SIZE_BYTES_ INTEGER NOT NULL DEFAULT 0,
			URL_ TEXT,
			TEXT_ TEXT,
			METADATA_JSON_ TEXT NOT NULL DEFAULT '{}',
			CREATED_BY_ TEXT,
			CREATED_BY_AGENT_ TEXT,
			CREATED_AT_ TEXT NOT NULL,
			DELETED_AT_ TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS issue_dependency (
			ID_ TEXT PRIMARY KEY,
			FROM_ISSUE_ID_ TEXT NOT NULL REFERENCES issue(ID_),
			TO_ISSUE_ID_ TEXT NOT NULL REFERENCES issue(ID_),
			TYPE_ TEXT NOT NULL DEFAULT 'blocks',
			CREATED_BY_ TEXT,
			CREATED_AT_ TEXT NOT NULL,
			DELETED_AT_ TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS issue_automation (
			ID_ TEXT PRIMARY KEY,
			ISSUE_ID_ TEXT NOT NULL REFERENCES issue(ID_),
			EXTERNAL_AUTOMATION_ID_ TEXT,
			ENABLED_ INTEGER NOT NULL DEFAULT 0 CHECK (ENABLED_ IN (0, 1)),
			CRON_ TEXT,
			TIMEZONE_ TEXT,
			MESSAGE_ TEXT,
			CREATED_BY_ TEXT,
			UPDATED_BY_ TEXT,
			CREATED_AT_ TEXT NOT NULL,
			UPDATED_AT_ TEXT NOT NULL,
			DELETED_AT_ TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS agent_run (
			ID_ TEXT PRIMARY KEY,
			ISSUE_ID_ TEXT NOT NULL REFERENCES issue(ID_),
			AGENT_ID_ TEXT REFERENCES agent(ID_),
			WORKER_AGENT_ TEXT,
			DELEGATED_BY_ TEXT,
			CHAT_ID_ TEXT,
			RUN_ID_ TEXT,
			SESSION_ID_ TEXT,
			STATUS_ TEXT NOT NULL CHECK (STATUS_ IN ('queued','running','succeeded','failed','cancelled','completed')),
			CONFIDENCE_ REAL,
			INPUT_TOKENS_ INTEGER,
			OUTPUT_TOKENS_ INTEGER,
			COST_MICROS_ INTEGER,
			STARTED_AT_ TEXT,
			FINISHED_AT_ TEXT,
			ERROR_MESSAGE_ TEXT,
			CREATED_AT_ TEXT NOT NULL,
			UPDATED_AT_ TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS agent_tool_call (
			ID_ TEXT PRIMARY KEY,
			AGENT_RUN_ID_ TEXT NOT NULL REFERENCES agent_run(ID_),
			TOOL_NAME_ TEXT NOT NULL,
			TOOL_TYPE_ TEXT NOT NULL DEFAULT '',
			STATUS_ TEXT NOT NULL DEFAULT '',
			INPUT_JSON_ TEXT NOT NULL DEFAULT '{}',
			OUTPUT_JSON_ TEXT NOT NULL DEFAULT '{}',
			STARTED_AT_ TEXT,
			FINISHED_AT_ TEXT,
			ERROR_MESSAGE_ TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS review_comment (
			ID_ TEXT PRIMARY KEY,
			REVIEW_ID_ TEXT NOT NULL REFERENCES review(ID_),
			ISSUE_ID_ TEXT NOT NULL REFERENCES issue(ID_),
			AUTHOR_ID_ TEXT,
			AUTHOR_AGENT_ TEXT,
			BODY_ TEXT NOT NULL,
			ANCHOR_JSON_ TEXT NOT NULL DEFAULT '{}',
			CREATED_AT_ TEXT NOT NULL,
			UPDATED_AT_ TEXT NOT NULL,
			DELETED_AT_ TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS board (
			ID_ TEXT PRIMARY KEY,
			PROJECT_ID_ TEXT NOT NULL REFERENCES project(ID_),
			KEY_ TEXT NOT NULL,
			NAME_ TEXT NOT NULL,
			CREATED_BY_ TEXT,
			UPDATED_BY_ TEXT,
			CREATED_AT_ TEXT NOT NULL,
			UPDATED_AT_ TEXT NOT NULL,
			DELETED_AT_ TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS board_meta (
			BOARD_ID_ TEXT NOT NULL,
			KEY_ TEXT NOT NULL,
			VALUE_ TEXT NOT NULL,
			PRIMARY KEY (BOARD_ID_, KEY_)
		)`,
		`CREATE TABLE IF NOT EXISTS event_log (
			ID_ INTEGER PRIMARY KEY AUTOINCREMENT,
			BOARD_ID_ TEXT NOT NULL,
			PROJECT_ID_ TEXT,
			ISSUE_ID_ TEXT,
			REVISION_ INTEGER NOT NULL,
			EVENT_TYPE_ TEXT NOT NULL,
			ACTOR_ID_ TEXT,
			ACTOR_AGENT_ TEXT,
			PAYLOAD_JSON_ TEXT NOT NULL DEFAULT '{}',
			CREATED_AT_ TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS desktop_client (
			SESSION_ID_ TEXT PRIMARY KEY,
			DEVICE_ID_ TEXT,
			CURRENT_USER_ID_ TEXT,
			CURRENT_USER_NAME_ TEXT,
			CAPABILITIES_JSON_ TEXT NOT NULL DEFAULT '[]',
			SELECTED_PROJECT_ID_ TEXT,
			CONNECTED_AT_ TEXT NOT NULL,
			LAST_SEEN_AT_ TEXT NOT NULL
		)`,
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	if err := s.migrateLegacyProjects(ctx); err != nil {
		return err
	}
	if err := s.repairLegacyProjectClosure(ctx); err != nil {
		return err
	}
	if err := s.migrateAddSeverityColumn(ctx); err != nil {
		return err
	}
	if err := s.migrateWorkflowStatusColumns(ctx); err != nil {
		return err
	}
	if err := s.migrateAddWorkflowTransitionModeColumn(ctx); err != nil {
		return err
	}
	if err := s.migrateDropLegacyIssueTypeColumn(ctx); err != nil {
		return err
	}
	if err := s.migrateLegacyIssues(ctx); err != nil {
		return err
	}
	if err := s.migrateLegacyRevision(ctx); err != nil {
		return err
	}
	if err := s.migrateAddDesktopClientMetadataColumns(ctx); err != nil {
		return err
	}
	if err := s.rebuildProjectClosure(ctx); err != nil {
		return err
	}
	indexStatements := []string{
		`CREATE INDEX IF NOT EXISTS idx_project_parent_position
			ON project(PARENT_ID_, POSITION_, ID_)
			WHERE DELETED_AT_ IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_project_closure_descendant
			ON project_closure(DESCENDANT_ID_, ANCESTOR_ID_)`,
		`CREATE INDEX IF NOT EXISTS idx_project_permission_project
			ON project_permission(PROJECT_ID_, PRINCIPAL_TYPE_, PRINCIPAL_ID_)
			WHERE DELETED_AT_ IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_issue_project_status_position
			ON issue(PROJECT_ID_, STATUS_ID_, POSITION_, ID_)
			WHERE DELETED_AT_ IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_issue_workflow_stage_status
			ON issue(WORKFLOW_ID_, STAGE_ID_, STATUS_ID_)
			WHERE DELETED_AT_ IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_issue_active_run
			ON issue(ACTIVE_RUN_ID_)
			WHERE ACTIVE_RUN_ID_ IS NOT NULL AND DELETED_AT_ IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_issue_automation_issue
			ON issue_automation(ISSUE_ID_)
			WHERE DELETED_AT_ IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_issue_attachment_issue
			ON issue_attachment(ISSUE_ID_, CREATED_AT_, ID_)
			WHERE DELETED_AT_ IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_issue_label_project
			ON issue_label(PROJECT_ID_, NAME_, ID_)
			WHERE DELETED_AT_ IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_issue_dependency_from
			ON issue_dependency(FROM_ISSUE_ID_, TO_ISSUE_ID_, ID_)
			WHERE DELETED_AT_ IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_issue_dependency_to
			ON issue_dependency(TO_ISSUE_ID_, FROM_ISSUE_ID_, ID_)
			WHERE DELETED_AT_ IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_review_issue
			ON review(ISSUE_ID_, UPDATED_AT_, ID_)
			WHERE DELETED_AT_ IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_review_comment_issue
			ON review_comment(ISSUE_ID_, CREATED_AT_, ID_)
			WHERE DELETED_AT_ IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_agent_run_issue
			ON agent_run(ISSUE_ID_, CREATED_AT_, ID_)`,
		`CREATE INDEX IF NOT EXISTS idx_agent_tool_call_run
			ON agent_tool_call(AGENT_RUN_ID_, STARTED_AT_, ID_)`,
		`CREATE INDEX IF NOT EXISTS idx_agent_run_external
			ON agent_run(RUN_ID_, CHAT_ID_, STATUS_)`,
		`CREATE INDEX IF NOT EXISTS idx_event_log_board_revision
			ON event_log(BOARD_ID_, REVISION_)`,
		`CREATE INDEX IF NOT EXISTS idx_workflow_transition_from
			ON workflow_transition(WORKFLOW_ID_, FROM_STAGE_ID_, FROM_STATUS_ID_, ACTION_KEY_)`,
	}
	for _, statement := range indexStatements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func executeSQLScript(ctx context.Context, db *sql.DB, script string, label string) error {
	for _, statement := range splitSQL(script) {
		statement = strings.TrimSpace(statement)
		if statement == "" || strings.HasPrefix(statement, "--") {
			continue
		}
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("%s: %w\nSQL: %s", label, err, statement)
		}
	}
	return nil
}

func executeSchemaScript(ctx context.Context, db *sql.DB) error {
	for _, statement := range splitSQL(schemaSQL) {
		statement = strings.TrimSpace(statement)
		if statement == "" || strings.HasPrefix(statement, "--") {
			continue
		}
		if _, err := db.ExecContext(ctx, statement); err != nil {
			upper := strings.ToUpper(statement)
			if strings.HasPrefix(upper, "CREATE INDEX") && strings.Contains(err.Error(), "no such column") {
				continue
			}
			return fmt.Errorf("schema: %w\nSQL: %s", err, statement)
		}
	}
	return nil
}

func (s *Store) migrateAddSeverityColumn(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `ALTER TABLE issue ADD COLUMN SEVERITY_ TEXT NOT NULL DEFAULT 'medium'`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column") {
		return err
	}
	return nil
}

func (s *Store) migrateWorkflowStatusColumns(ctx context.Context) error {
	for _, stmt := range []string{
		`ALTER TABLE workflow_status ADD COLUMN STAGE_ID_ TEXT`,
		`ALTER TABLE workflow_status ADD COLUMN IS_ACTIVE_ INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE workflow_transition ADD COLUMN IS_ACTIVE_ INTEGER NOT NULL DEFAULT 1`,
	} {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil && !strings.Contains(err.Error(), "duplicate column") {
			return err
		}
	}
	return nil
}

func (s *Store) migrateAddWorkflowTransitionModeColumn(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `ALTER TABLE workflow ADD COLUMN TRANSITION_MODE_ TEXT NOT NULL DEFAULT 'strict' CHECK (TRANSITION_MODE_ IN ('strict', 'free'))`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column") {
		return err
	}
	return nil
}

func (s *Store) migrateAddDesktopClientMetadataColumns(ctx context.Context) error {
	columns := []string{
		`ALTER TABLE desktop_client ADD COLUMN DEVICE_ID_ TEXT`,
		`ALTER TABLE desktop_client ADD COLUMN CURRENT_USER_ID_ TEXT`,
		`ALTER TABLE desktop_client ADD COLUMN CURRENT_USER_NAME_ TEXT`,
	}
	for _, statement := range columns {
		if _, err := s.db.ExecContext(ctx, statement); err != nil && !strings.Contains(err.Error(), "duplicate column") {
			return err
		}
	}
	return nil
}

func (s *Store) migrateDropLegacyIssueTypeColumn(ctx context.Context) error {
	columns, err := s.tableColumns(ctx, "issue")
	if err != nil {
		return err
	}
	if !columns["TYPE_ID_"] {
		return nil
	}
	_, err = s.db.ExecContext(ctx, `ALTER TABLE issue DROP COLUMN TYPE_ID_`)
	return err
}

func (s *Store) EnsureDefaultProject(ctx context.Context) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO project (
			ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_,
			VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_
		)
		VALUES (?, NULL, 'default', 'DEFAULT', 'All Projects', '', 'default', 0, 0, 'workspace', ?, ?, ?)
		ON CONFLICT(ID_) DO NOTHING
	`, kanban.DefaultProjectID, kanban.DefaultWorkflowID, now, now)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
		VALUES (?, ?, 0)
		ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING
	`, kanban.DefaultProjectID, kanban.DefaultProjectID)
	return err
}

func (s *Store) EnsureDefaultBoard(ctx context.Context) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO board (ID_, PROJECT_ID_, KEY_, NAME_, CREATED_AT_, UPDATED_AT_)
		VALUES (?, ?, 'default', 'Default Board', ?, ?)
		ON CONFLICT(ID_) DO NOTHING
	`, kanban.DefaultBoardID, kanban.DefaultProjectID, now, now)
	return err
}

func (s *Store) SeedDefaults(ctx context.Context) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	sql := strings.ReplaceAll(seedDefaultsSQL, "__NOW__", now)
	return executeSQLScript(ctx, s.db, sql, "seed defaults")
}

func (s *Store) SeedWorkflowCatalog(ctx context.Context) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	sql := strings.ReplaceAll(workflowSeedSQL, "__NOW__", now)
	return executeSQLScript(ctx, s.db, sql, "seed workflow")
}

func splitSQL(s string) []string {
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
func (s *Store) migrateLegacyProjects(ctx context.Context) error {
	exists, err := s.tableExists(ctx, "projects")
	if err != nil || !exists {
		return err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, parent_id, slug, name, path, depth, position, archived_at, created_at, updated_at, deleted_at
		FROM projects
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id, slug, name, path, createdAt, updatedAt string
		var parentID, archivedAt, deletedAt sql.NullString
		var depth int
		var position float64
		if err := rows.Scan(&id, &parentID, &slug, &name, &path, &depth, &position, &archivedAt, &createdAt, &updatedAt, &deletedAt); err != nil {
			return err
		}
		if _, err := s.db.ExecContext(ctx, `
			INSERT INTO project (
				ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_,
				VISIBILITY_, DEFAULT_WORKFLOW_ID_, ARCHIVED_AT_, CREATED_AT_, UPDATED_AT_, DELETED_AT_
			)
			VALUES (?, ?, ?, ?, ?, '', ?, ?, ?, 'workspace', ?, ?, ?, ?, ?)
			ON CONFLICT(ID_) DO NOTHING
		`, id, nullStringValue(parentID), slug, strings.ToUpper(slug), name, path, depth, position,
			kanban.DefaultWorkflowID, nullStringValue(archivedAt), createdAt, updatedAt, nullStringValue(deletedAt)); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}

func (s *Store) repairLegacyProjectClosure(ctx context.Context) error {
	exists, err := s.tableExists(ctx, "project_closure")
	if err != nil || !exists {
		return err
	}
	columns, err := s.tableColumns(ctx, "project_closure")
	if err != nil {
		return err
	}
	if columns["ANCESTOR_ID_"] && columns["DESCENDANT_ID_"] && columns["DEPTH_"] {
		return nil
	}

	if _, err := s.db.ExecContext(ctx, `DROP TABLE project_closure`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE project_closure (
		ANCESTOR_ID_ TEXT NOT NULL REFERENCES project(ID_),
		DESCENDANT_ID_ TEXT NOT NULL REFERENCES project(ID_),
		DEPTH_ INTEGER NOT NULL,
		PRIMARY KEY (ANCESTOR_ID_, DESCENDANT_ID_)
	)`); err != nil {
		return err
	}
	return nil
}

func (s *Store) rebuildProjectClosure(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, PARENT_ID_
		FROM project
		WHERE DELETED_AT_ IS NULL
	`)
	if err != nil {
		return err
	}
	parents := map[string]string{}
	for rows.Next() {
		var id string
		var parentID sql.NullString
		if err := rows.Scan(&id, &parentID); err != nil {
			_ = rows.Close()
			return err
		}
		if parentID.Valid {
			parents[id] = parentID.String
		} else {
			parents[id] = ""
		}
	}
	if err := rows.Close(); err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM project_closure`); err != nil {
		_ = tx.Rollback()
		return err
	}
	for descendantID := range parents {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
			VALUES (?, ?, 0)
		`, descendantID, descendantID); err != nil {
			_ = tx.Rollback()
			return err
		}
		seen := map[string]bool{descendantID: true}
		parentID := parents[descendantID]
		for depth := 1; parentID != ""; depth++ {
			if seen[parentID] {
				_ = tx.Rollback()
				return fmt.Errorf("project parent cycle detected at %s", parentID)
			}
			if _, ok := parents[parentID]; !ok {
				break
			}
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
				VALUES (?, ?, ?)
			`, parentID, descendantID, depth); err != nil {
				_ = tx.Rollback()
				return err
			}
			seen[parentID] = true
			parentID = parents[parentID]
		}
	}
	return tx.Commit()
}

func (s *Store) migrateLegacyIssues(ctx context.Context) error {
	exists, err := s.tableExists(ctx, "task_board_issues")
	if err != nil || !exists {
		return err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT board_id, project_id, id, title, description, status, priority, assignee_agent_key,
			position, chat_id, run_id, run_state, automation_id, automation_enabled,
			automation_cron, automation_message, automation_timezone, attachment_chat_id,
			attachments_json, created_at, updated_at, revision, deleted_at, created_by, updated_by
		FROM task_board_issues
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var boardID, projectID, id, title, description, status, priority string
		var assigneeAgentKey, chatID, runID, runState, automationID sql.NullString
		var automationCron, automationMessage, automationTimezone, attachmentChatID sql.NullString
		var attachmentsJSON string
		var createdAt, updatedAt string
		var deletedAt, createdBy, updatedBy sql.NullString
		var position float64
		var automationEnabled int
		var revision int64
		if err := rows.Scan(
			&boardID, &projectID, &id, &title, &description, &status, &priority, &assigneeAgentKey,
			&position, &chatID, &runID, &runState, &automationID, &automationEnabled,
			&automationCron, &automationMessage, &automationTimezone, &attachmentChatID,
			&attachmentsJSON, &createdAt, &updatedAt, &revision, &deletedAt, &createdBy, &updatedBy,
		); err != nil {
			return err
		}
		if projectID == "" {
			projectID = kanban.DefaultProjectID
		}
		statusKey := status
		if normalized, ok := kanban.NormalizeStatus(status); ok {
			statusKey = string(normalized)
		}
		stageID := workflowStartStageID(kanban.DefaultWorkflowID)
		statusID := workflowStatusID(kanban.DefaultWorkflowID, statusKey)
		workerType := any(nil)
		workerAgent := any(nil)
		if assigneeAgentKey.Valid && strings.TrimSpace(assigneeAgentKey.String) != "" {
			workerType = "agent"
			workerAgent = strings.TrimSpace(assigneeAgentKey.String)
		}
		activeRunID := any(nil)
		if runID.Valid && strings.TrimSpace(runID.String) != "" {
			activeRunID = newID("agent-run")
			agentStatus := "running"
			if runState.Valid {
				agentStatus = runState.String
			}
			if _, err := s.db.ExecContext(ctx, `
				INSERT INTO agent_run (
					ID_, ISSUE_ID_, WORKER_AGENT_, DELEGATED_BY_, CHAT_ID_, RUN_ID_, STATUS_,
					STARTED_AT_, CREATED_AT_, UPDATED_AT_
				)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				ON CONFLICT(ID_) DO NOTHING
			`, activeRunID, id, workerAgent, nullStringValue(createdBy), nullStringValue(chatID),
				strings.TrimSpace(runID.String), normalizeAgentRunStatus(agentStatus), createdAt, createdAt, updatedAt); err != nil {
				return err
			}
		}
		if _, err := s.db.ExecContext(ctx, `
			INSERT INTO issue (
				ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_,
				TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_,
				WORKER_TYPE_, WORKER_AGENT_, REVIEW_REQUIRED_, ACTIVE_RUN_ID_, REVISION_,
				CREATED_BY_, UPDATED_BY_, CREATED_AT_, UPDATED_AT_, DELETED_AT_
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(ID_) DO NOTHING
		`, id, projectID, kanban.DefaultWorkflowID, stageID, statusID,
			title, description, priority, "medium", position, workerType, workerAgent, boolInt(statusKey == string(kanban.StatusInReview)),
			activeRunID, revision, nullStringValue(createdBy), nullStringValue(updatedBy), createdAt, updatedAt, nullStringValue(deletedAt)); err != nil {
			return err
		}
		if automationID.Valid || automationEnabled == 1 || automationCron.Valid || automationMessage.Valid || automationTimezone.Valid {
			autoID := nullStringValue(automationID)
			if autoID == nil {
				autoID = "automation-" + id
			}
			if _, err := s.db.ExecContext(ctx, `
				INSERT INTO issue_automation (
					ID_, ISSUE_ID_, EXTERNAL_AUTOMATION_ID_, ENABLED_, CRON_, TIMEZONE_, MESSAGE_,
					CREATED_BY_, UPDATED_BY_, CREATED_AT_, UPDATED_AT_
				)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				ON CONFLICT(ID_) DO NOTHING
			`, autoID, id, nullStringValue(automationID), automationEnabled, nullStringValue(automationCron),
				nullStringValue(automationTimezone), nullStringValue(automationMessage), nullStringValue(createdBy),
				nullStringValue(updatedBy), createdAt, updatedAt); err != nil {
				return err
			}
		}
		var attachments []kanban.Attachment
		if strings.TrimSpace(attachmentsJSON) != "" {
			_ = json.Unmarshal([]byte(attachmentsJSON), &attachments)
		}
		if len(attachments) > 0 {
			if err := s.replaceAttachments(ctx, nil, id, attachments, nullStringPtr(createdBy), nil, createdAt); err != nil {
				return err
			}
		}
		_ = boardID
		_ = attachmentChatID
	}
	return rows.Err()
}

func (s *Store) migrateLegacyRevision(ctx context.Context) error {
	exists, err := s.tableExists(ctx, "task_board_meta")
	if err != nil || !exists {
		return err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT board_id, key, value FROM task_board_meta`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var boardID, key, value string
		if err := rows.Scan(&boardID, &key, &value); err != nil {
			return err
		}
		if _, err := s.db.ExecContext(ctx, `
			INSERT INTO board_meta (BOARD_ID_, KEY_, VALUE_)
			VALUES (?, ?, ?)
			ON CONFLICT(BOARD_ID_, KEY_) DO NOTHING
		`, boardID, key, value); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (s *Store) tableExists(ctx context.Context, table string) (bool, error) {
	var name string
	err := s.db.QueryRowContext(ctx, `
		SELECT name FROM sqlite_master
		WHERE type = 'table' AND name = ?
	`, table).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (s *Store) tableColumns(ctx context.Context, table string) (map[string]bool, error) {
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns := map[string]bool{}
	for rows.Next() {
		var cid int
		var name string
		var columnType string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			return nil, err
		}
		columns[name] = true
	}
	return columns, rows.Err()
}
func (s *Store) UpsertWorkflow(ctx context.Context, wf kanban.Workflow) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	isDefault := wf.ID == kanban.DefaultWorkflowID
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO workflow (
			ID_, KEY_, NAME_, DESCRIPTION_, IS_DEFAULT_, TRANSITION_MODE_, CREATED_AT_, UPDATED_AT_
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(ID_) DO UPDATE SET
			KEY_ = excluded.KEY_,
			NAME_ = excluded.NAME_,
			DESCRIPTION_ = excluded.DESCRIPTION_,
			IS_DEFAULT_ = excluded.IS_DEFAULT_,
			TRANSITION_MODE_ = excluded.TRANSITION_MODE_,
			UPDATED_AT_ = excluded.UPDATED_AT_
	`, wf.ID, wf.Key, wf.Name, wf.Description, boolInt(isDefault), wf.TransitionMode, now, now)
	return err
}

func (s *Store) SoftDeleteWorkflow(ctx context.Context, id string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx, `
		UPDATE workflow SET DELETED_AT_ = ? WHERE ID_ = ?
	`, now, id)
	return err
}

func (s *Store) ListWorkflowCatalog(ctx context.Context) (kanban.WorkflowCatalog, error) {
	var catalog kanban.WorkflowCatalog
	var err error
	catalog.Workflows, err = s.listWorkflows(ctx)
	if err != nil {
		return catalog, err
	}
	catalog.WorkflowStageDefs, err = s.listWorkflowStageDefs(ctx)
	if err != nil {
		return catalog, err
	}
	catalog.WorkflowStatusDefs, err = s.listWorkflowStatusDefs(ctx)
	if err != nil {
		return catalog, err
	}
	catalog.WorkflowStages, err = s.listWorkflowStages(ctx)
	if err != nil {
		return catalog, err
	}
	catalog.WorkflowStatuses, err = s.listWorkflowStatuses(ctx)
	if err != nil {
		return catalog, err
	}
	catalog.WorkflowTransitions, err = s.listWorkflowTransitions(ctx)
	if err != nil {
		return catalog, err
	}
	catalog.Teams, err = s.listTeams(ctx)
	if err != nil {
		return catalog, err
	}
	catalog.ProjectPermissions, err = s.listProjectPermissions(ctx)
	if err != nil {
		return catalog, err
	}
	return catalog, nil
}

func (s *Store) listWorkflows(ctx context.Context) ([]kanban.Workflow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, KEY_, NAME_, DESCRIPTION_, IS_DEFAULT_, TRANSITION_MODE_, CREATED_AT_, UPDATED_AT_, DELETED_AT_
		FROM workflow
		WHERE DELETED_AT_ IS NULL
		ORDER BY IS_DEFAULT_ DESC, NAME_ ASC, ID_ ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var values []kanban.Workflow
	for rows.Next() {
		var item kanban.Workflow
		var isDefault int
		var transitionMode string
		var createdAt, updatedAt string
		var deletedAt sql.NullString
		if err := rows.Scan(&item.ID, &item.Key, &item.Name, &item.Description, &isDefault, &transitionMode, &createdAt, &updatedAt, &deletedAt); err != nil {
			return nil, err
		}
		item.IsDefault = isDefault == 1
		if transitionMode == "" {
			transitionMode = "strict"
		}
		item.TransitionMode = transitionMode
		var err error
		item.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			return nil, err
		}
		item.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
		if err != nil {
			return nil, err
		}
		item.DeletedAt, err = parseOptionalTime(deletedAt)
		if err != nil {
			return nil, err
		}
		values = append(values, item)
	}
	return values, rows.Err()
}

func (s *Store) listWorkflowStageDefs(ctx context.Context) ([]kanban.WorkflowStageDef, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, KEY_, NAME_, DESCRIPTION_, IS_SYSTEM_, CREATED_AT_, UPDATED_AT_, DELETED_AT_
		FROM workflow_stage_def
		WHERE DELETED_AT_ IS NULL
		ORDER BY KEY_ ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var values []kanban.WorkflowStageDef
	for rows.Next() {
		var item kanban.WorkflowStageDef
		var isSystem int
		var createdAt, updatedAt string
		var deletedAt sql.NullString
		if err := rows.Scan(&item.ID, &item.Key, &item.Name, &item.Description, &isSystem, &createdAt, &updatedAt, &deletedAt); err != nil {
			return nil, err
		}
		item.IsSystem = isSystem == 1
		var err error
		item.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			return nil, err
		}
		item.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
		if err != nil {
			return nil, err
		}
		item.DeletedAt, err = parseOptionalTime(deletedAt)
		if err != nil {
			return nil, err
		}
		values = append(values, item)
	}
	return values, rows.Err()
}

func (s *Store) listWorkflowStatusDefs(ctx context.Context) ([]kanban.WorkflowStatusDef, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, KEY_, NAME_, COLUMN_KEY_, DESCRIPTION_, IS_SYSTEM_, CREATED_AT_, UPDATED_AT_, DELETED_AT_
		FROM workflow_status_def
		WHERE DELETED_AT_ IS NULL
		ORDER BY CASE COLUMN_KEY_
			WHEN 'backlog' THEN 0
			WHEN 'todo' THEN 1
			WHEN 'in_progress' THEN 2
			WHEN 'in_review' THEN 3
			WHEN 'completed' THEN 4
			ELSE 99
		END
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var values []kanban.WorkflowStatusDef
	for rows.Next() {
		var item kanban.WorkflowStatusDef
		var isSystem int
		var createdAt, updatedAt string
		var deletedAt sql.NullString
		if err := rows.Scan(&item.ID, &item.Key, &item.Name, &item.ColumnKey, &item.Description, &isSystem, &createdAt, &updatedAt, &deletedAt); err != nil {
			return nil, err
		}
		item.IsSystem = isSystem == 1
		var err error
		item.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			return nil, err
		}
		item.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
		if err != nil {
			return nil, err
		}
		item.DeletedAt, err = parseOptionalTime(deletedAt)
		if err != nil {
			return nil, err
		}
		values = append(values, item)
	}
	return values, rows.Err()
}

func (s *Store) listWorkflowStages(ctx context.Context) ([]kanban.WorkflowStage, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, WORKFLOW_ID_, STAGE_DEF_ID_, KEY_, NAME_, POSITION_, IS_START_, IS_END_
		FROM workflow_stage
		ORDER BY WORKFLOW_ID_ ASC, POSITION_ ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var values []kanban.WorkflowStage
	for rows.Next() {
		var item kanban.WorkflowStage
		var isStart, isEnd int
		if err := rows.Scan(&item.ID, &item.WorkflowID, &item.StageDefID, &item.Key, &item.Name, &item.Position, &isStart, &isEnd); err != nil {
			return nil, err
		}
		item.IsStart = isStart == 1
		item.IsEnd = isEnd == 1
		values = append(values, item)
	}
	return values, rows.Err()
}

func (s *Store) listWorkflowStatuses(ctx context.Context) ([]kanban.WorkflowStatus, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, WORKFLOW_ID_, STATUS_DEF_ID_, KEY_, NAME_, COLUMN_KEY_, POSITION_,
			IS_START_, IS_TERMINAL_, REVIEW_REQUIRED_
		FROM workflow_status
		ORDER BY WORKFLOW_ID_ ASC, POSITION_ ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var values []kanban.WorkflowStatus
	for rows.Next() {
		var item kanban.WorkflowStatus
		var isStart, isTerminal, reviewRequired int
		if err := rows.Scan(&item.ID, &item.WorkflowID, &item.StatusDefID, &item.Key, &item.Name, &item.ColumnKey, &item.Position, &isStart, &isTerminal, &reviewRequired); err != nil {
			return nil, err
		}
		item.IsStart = isStart == 1
		item.IsTerminal = isTerminal == 1
		item.ReviewRequired = reviewRequired == 1
		values = append(values, item)
	}
	return values, rows.Err()
}

func (s *Store) listWorkflowTransitions(ctx context.Context) ([]kanban.WorkflowTransition, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, WORKFLOW_ID_, FROM_STAGE_ID_, FROM_STATUS_ID_, TO_STAGE_ID_, TO_STATUS_ID_,
			ACTION_KEY_, NAME_, ACTOR_TYPE_, REQUIRES_REVIEW_, POSITION_, CREATED_AT_
		FROM workflow_transition
		ORDER BY WORKFLOW_ID_ ASC, POSITION_ ASC, ID_ ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var values []kanban.WorkflowTransition
	for rows.Next() {
		var item kanban.WorkflowTransition
		var requiresReview int
		var createdAt string
		if err := rows.Scan(&item.ID, &item.WorkflowID, &item.FromStageID, &item.FromStatusID, &item.ToStageID, &item.ToStatusID, &item.ActionKey, &item.Name, &item.ActorType, &requiresReview, &item.Position, &createdAt); err != nil {
			return nil, err
		}
		item.RequiresReview = requiresReview == 1
		var err error
		item.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			return nil, err
		}
		values = append(values, item)
	}
	return values, rows.Err()
}

func (s *Store) listTeams(ctx context.Context) ([]kanban.Team, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, SLUG_, NAME_, DESCRIPTION_, CREATED_BY_, UPDATED_BY_, CREATED_AT_, UPDATED_AT_, DELETED_AT_
		FROM team
		WHERE DELETED_AT_ IS NULL
		ORDER BY NAME_ ASC, ID_ ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []kanban.Team{}
	for rows.Next() {
		var item kanban.Team
		var createdBy, updatedBy, deletedAt sql.NullString
		var createdAt, updatedAt string
		if err := rows.Scan(&item.ID, &item.Slug, &item.Name, &item.Description, &createdBy, &updatedBy, &createdAt, &updatedAt, &deletedAt); err != nil {
			return nil, err
		}
		item.CreatedBy = stringPtr(createdBy)
		item.UpdatedBy = stringPtr(updatedBy)
		var err error
		item.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			return nil, err
		}
		item.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
		if err != nil {
			return nil, err
		}
		item.DeletedAt, err = parseOptionalTime(deletedAt)
		if err != nil {
			return nil, err
		}
		values = append(values, item)
	}
	return values, rows.Err()
}

func (s *Store) listProjectPermissions(ctx context.Context) ([]kanban.ProjectPermission, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, PROJECT_ID_, PRINCIPAL_TYPE_, PRINCIPAL_ID_, ROLE_, INHERIT_TO_CHILDREN_,
			CREATED_BY_, CREATED_AT_, DELETED_AT_
		FROM project_permission
		WHERE DELETED_AT_ IS NULL
		ORDER BY PROJECT_ID_ ASC, ROLE_ ASC, ID_ ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []kanban.ProjectPermission{}
	for rows.Next() {
		var item kanban.ProjectPermission
		var inherit int
		var createdBy, deletedAt sql.NullString
		var createdAt string
		if err := rows.Scan(&item.ID, &item.ProjectID, &item.PrincipalType, &item.PrincipalID, &item.Role, &inherit, &createdBy, &createdAt, &deletedAt); err != nil {
			return nil, err
		}
		item.InheritToChildren = inherit == 1
		item.CreatedBy = stringPtr(createdBy)
		var err error
		item.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			return nil, err
		}
		item.DeletedAt, err = parseOptionalTime(deletedAt)
		if err != nil {
			return nil, err
		}
		values = append(values, item)
	}
	return values, rows.Err()
}

func (s *Store) ListProjects(ctx context.Context) ([]kanban.Project, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_,
			VISIBILITY_, DEFAULT_WORKFLOW_ID_, ARCHIVED_AT_, CREATED_AT_, UPDATED_AT_, DELETED_AT_,
			CREATED_BY_, UPDATED_BY_
		FROM project
		WHERE DELETED_AT_ IS NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var projects []kanban.Project
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sortProjects(projects), nil
}

func (s *Store) ListIssues(ctx context.Context, boardID string, projectID string) ([]kanban.Issue, int64, error) {
	if boardID == "" {
		boardID = kanban.DefaultBoardID
	}
	if projectID == "" {
		projectID = kanban.DefaultProjectID
	}
	revision, err := s.Revision(ctx, boardID)
	if err != nil {
		return nil, 0, err
	}
	rows, err := s.db.QueryContext(ctx, issueSelectSQL(`
		JOIN project_closure pc ON pc.DESCENDANT_ID_ = i.PROJECT_ID_ AND pc.ANCESTOR_ID_ = ?
		WHERE i.DELETED_AT_ IS NULL
		ORDER BY ws.POSITION_ ASC, wst.POSITION_ ASC, i.POSITION_ ASC, i.ID_ ASC
	`), projectID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	issues, err := scanIssues(rows)
	if err != nil {
		return nil, 0, err
	}
	if err := s.attachmentsForIssues(ctx, issues); err != nil {
		return nil, 0, err
	}
	return kanban.SortIssues(issues), revision, nil
}

func (s *Store) GetIssue(ctx context.Context, boardID string, issueID string) (*kanban.Issue, error) {
	rows, err := s.db.QueryContext(ctx, issueSelectSQL(`
		WHERE i.ID_ = ? AND i.DELETED_AT_ IS NULL
	`), issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	issues, err := scanIssues(rows)
	if err != nil {
		return nil, err
	}
	if len(issues) == 0 {
		return nil, nil
	}
	if err := s.attachmentsForIssues(ctx, issues); err != nil {
		return nil, err
	}
	issues[0].BoardID = normalizeBoardID(boardID)
	return &issues[0], nil
}

func (s *Store) ReplaceIssue(ctx context.Context, boardID string, issue kanban.Issue, eventType string, actor string) (int64, error) {
	return s.withRevision(ctx, boardID, eventType, issue.ProjectID, issue.ID, actor, issue, func(tx *sql.Tx, revision int64) error {
		if issue.ProjectID == "" {
			issue.ProjectID = kanban.DefaultProjectID
		}
		if issue.WorkflowID == "" {
			issue.WorkflowID = kanban.DefaultWorkflowID
		}

		if issue.StageID == "" {
			issue.StageID = workflowStartStageID(issue.WorkflowID)
		}
		if issue.StatusID == "" {
			statusKey := string(issue.Status)
			if statusKey == "" {
				statusKey = string(kanban.StatusBacklog)
			}
			issue.StatusID = workflowStatusID(issue.WorkflowID, statusKey)
		}
		if issue.WorkerAgent == nil {
			issue.WorkerAgent = issue.AssigneeAgentKey
		}
		if issue.WorkerType == nil && issue.WorkerAgent != nil {
			workerType := "agent"
			issue.WorkerType = &workerType
		}
		if issue.AssigneeAgentKey == nil {
			issue.AssigneeAgentKey = issue.WorkerAgent
		}
		issue.Revision = revision
		now := issue.UpdatedAt
		if now.IsZero() {
			now = time.Now().UTC()
			issue.UpdatedAt = now
		}
		if issue.CreatedAt.IsZero() {
			issue.CreatedAt = now
		}
		if err := persistRunTx(ctx, tx, &issue, actor, now); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO issue (
				ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_,
				TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_,
				ASSIGNEE_ID_, WORKER_TYPE_, WORKER_ID_, WORKER_AGENT_, REVIEWER_ID_,
				REVIEW_REQUIRED_, ACTIVE_REVIEW_ID_, ACTIVE_RUN_ID_, REVISION_,
				CREATED_BY_, UPDATED_BY_, CREATED_BY_AGENT_, UPDATED_BY_AGENT_,
				CREATED_AT_, UPDATED_AT_, DELETED_AT_
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(ID_) DO UPDATE SET
				PROJECT_ID_ = excluded.PROJECT_ID_,
				WORKFLOW_ID_ = excluded.WORKFLOW_ID_,
				STAGE_ID_ = excluded.STAGE_ID_,
				STATUS_ID_ = excluded.STATUS_ID_,
				TITLE_ = excluded.TITLE_,
				DESCRIPTION_ = excluded.DESCRIPTION_,
				PRIORITY_ = excluded.PRIORITY_,
				SEVERITY_ = excluded.SEVERITY_,
				POSITION_ = excluded.POSITION_,
				ASSIGNEE_ID_ = excluded.ASSIGNEE_ID_,
				WORKER_TYPE_ = excluded.WORKER_TYPE_,
				WORKER_ID_ = excluded.WORKER_ID_,
				WORKER_AGENT_ = excluded.WORKER_AGENT_,
				REVIEWER_ID_ = excluded.REVIEWER_ID_,
				REVIEW_REQUIRED_ = excluded.REVIEW_REQUIRED_,
				ACTIVE_REVIEW_ID_ = excluded.ACTIVE_REVIEW_ID_,
				ACTIVE_RUN_ID_ = excluded.ACTIVE_RUN_ID_,
				REVISION_ = excluded.REVISION_,
				UPDATED_BY_ = excluded.UPDATED_BY_,
				UPDATED_BY_AGENT_ = excluded.UPDATED_BY_AGENT_,
				UPDATED_AT_ = excluded.UPDATED_AT_,
				DELETED_AT_ = excluded.DELETED_AT_
		`, issueParams(issue)...)
		if err != nil {
			return err
		}
		if err := replaceAutomationTx(ctx, tx, issue, actor, now); err != nil {
			return err
		}
		return replaceAttachmentsTx(ctx, tx, issue.ID, issue.Attachments, issue.CreatedBy, issue.CreatedByAgent, now.Format(time.RFC3339Nano))
	})
}

func (s *Store) SoftDeleteIssue(ctx context.Context, boardID string, issueID string, actor string) (int64, error) {
	payload := map[string]string{"id": issueID}
	return s.withRevision(ctx, boardID, "kanban.issue.deleted", "", issueID, actor, payload, func(tx *sql.Tx, revision int64) error {
		now := time.Now().UTC().Format(time.RFC3339Nano)
		result, err := tx.ExecContext(ctx, `
			UPDATE issue
			SET DELETED_AT_ = ?, UPDATED_AT_ = ?, REVISION_ = ?, UPDATED_BY_ = ?
			WHERE ID_ = ? AND DELETED_AT_ IS NULL
		`, now, now, revision, nullIfEmpty(actor), issueID)
		if err != nil {
			return err
		}
		count, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if count == 0 {
			return sql.ErrNoRows
		}
		return nil
	})
}

func (s *Store) Revision(ctx context.Context, boardID string) (int64, error) {
	var value string
	err := s.db.QueryRowContext(ctx, `
		SELECT VALUE_ FROM board_meta
		WHERE BOARD_ID_ = ? AND KEY_ = 'revision'
	`, normalizeBoardID(boardID)).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	var revision int64
	_, err = fmt.Sscan(value, &revision)
	return revision, err
}

func (s *Store) SaveDesktopClient(ctx context.Context, sessionID string, deviceID string, currentUserID string, currentUserName string, capabilities []string, selectedProjectID string) error {
	data, err := json.Marshal(capabilities)
	if err != nil {
		return err
	}
	if selectedProjectID == "" {
		selectedProjectID = kanban.DefaultProjectID
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO desktop_client (SESSION_ID_, DEVICE_ID_, CURRENT_USER_ID_, CURRENT_USER_NAME_, CAPABILITIES_JSON_, SELECTED_PROJECT_ID_, CONNECTED_AT_, LAST_SEEN_AT_)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(SESSION_ID_) DO UPDATE SET
			DEVICE_ID_ = excluded.DEVICE_ID_,
			CURRENT_USER_ID_ = excluded.CURRENT_USER_ID_,
			CURRENT_USER_NAME_ = excluded.CURRENT_USER_NAME_,
			CAPABILITIES_JSON_ = excluded.CAPABILITIES_JSON_,
			SELECTED_PROJECT_ID_ = excluded.SELECTED_PROJECT_ID_,
			LAST_SEEN_AT_ = excluded.LAST_SEEN_AT_
	`, sessionID, nullIfEmpty(deviceID), nullIfEmpty(currentUserID), nullIfEmpty(currentUserName), string(data), selectedProjectID, now, now)
	return err
}

func (s *Store) RemoveDesktopClient(ctx context.Context, sessionID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM desktop_client WHERE SESSION_ID_ = ?`, sessionID)
	return err
}

func (s *Store) CreateProject(ctx context.Context, project kanban.Project, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.project.created", project.ID, "", actor, project, func(tx *sql.Tx, revision int64) error {
		if project.ParentID == nil || *project.ParentID == "" {
			parentID := kanban.DefaultProjectID
			project.ParentID = &parentID
		}
		parent, err := projectByIDTx(ctx, tx, *project.ParentID)
		if err != nil {
			return err
		}
		project.Path = childProjectPath(&parent, project.Slug)
		project.Depth = parent.Depth + 1
		if project.Key == "" {
			project.Key = strings.ToUpper(project.Slug)
		}
		if project.Visibility == "" {
			project.Visibility = "workspace"
		}
		if project.DefaultWorkflowID == "" {
			project.DefaultWorkflowID = parent.DefaultWorkflowID
			if project.DefaultWorkflowID == "" {
				project.DefaultWorkflowID = kanban.DefaultWorkflowID
			}
		}
		if project.CreatedAt.IsZero() {
			project.CreatedAt = time.Now().UTC()
		}
		if project.UpdatedAt.IsZero() {
			project.UpdatedAt = project.CreatedAt
		}
		project.CreatedBy = actorPtr(actor)
		project.UpdatedBy = actorPtr(actor)
		_, err = tx.ExecContext(ctx, `
			INSERT INTO project (
				ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_,
				VISIBILITY_, DEFAULT_WORKFLOW_ID_, ARCHIVED_AT_, CREATED_BY_, UPDATED_BY_,
				CREATED_AT_, UPDATED_AT_, DELETED_AT_
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, project.ID, project.ParentID, project.Slug, project.Key, project.Name, project.Description, project.Path,
			project.Depth, project.Position, project.Visibility, project.DefaultWorkflowID, timePtrString(project.ArchivedAt),
			project.CreatedBy, project.UpdatedBy, project.CreatedAt.UTC().Format(time.RFC3339Nano),
			project.UpdatedAt.UTC().Format(time.RFC3339Nano), timePtrString(project.DeletedAt))
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
			SELECT ANCESTOR_ID_, ?, DEPTH_ + 1
			FROM project_closure
			WHERE DESCENDANT_ID_ = ?
		`, project.ID, parent.ID); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
			VALUES (?, ?, 0)
		`, project.ID, project.ID)
		return err
	})
}

func (s *Store) UpdateProject(ctx context.Context, projectID string, input kanban.ProjectUpdateInput, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.project.updated", projectID, "", actor, input, func(tx *sql.Tx, revision int64) error {
		project, err := projectByIDTx(ctx, tx, projectID)
		if err != nil {
			return err
		}
		name := project.Name
		if input.Name != nil {
			name = kanban.NormalizeProjectName(*input.Name)
		}
		slug := project.Slug
		if input.Slug != nil {
			slug = kanban.NormalizeProjectSlug(*input.Slug)
		} else if input.Name != nil {
			slug = kanban.ProjectSlugFromName(name)
		}
		description := project.Description
		if input.Description != nil {
			description = kanban.NormalizeDescription(input.Description)
		}
		visibility := project.Visibility
		if input.Visibility != nil {
			visibility = strings.TrimSpace(*input.Visibility)
		}
		defaultWorkflowID := project.DefaultWorkflowID
		if input.DefaultWorkflowID != nil && strings.TrimSpace(*input.DefaultWorkflowID) != "" {
			defaultWorkflowID = strings.TrimSpace(*input.DefaultWorkflowID)
		}
		var parent *kanban.Project
		if project.ParentID != nil {
			next, err := projectByIDTx(ctx, tx, *project.ParentID)
			if err != nil {
				return err
			}
			parent = &next
		}
		newPath := slug
		newDepth := 0
		if parent != nil {
			newPath = childProjectPath(parent, slug)
			newDepth = parent.Depth + 1
		}
		now := time.Now().UTC().Format(time.RFC3339Nano)
		descendants, err := descendantProjectsTx(ctx, tx, projectID)
		if err != nil {
			return err
		}
		for _, descendant := range descendants {
			path := newPath
			if descendant.ID != projectID {
				suffix := strings.TrimPrefix(descendant.Path, project.Path+"/")
				path = newPath + "/" + suffix
			}
			depth := newDepth + (descendant.Depth - project.Depth)
			if descendant.ID == projectID {
				if _, err = tx.ExecContext(ctx, `
					UPDATE project
					SET NAME_ = ?, SLUG_ = ?, KEY_ = ?, DESCRIPTION_ = ?, VISIBILITY_ = ?,
						DEFAULT_WORKFLOW_ID_ = ?, PATH_ = ?, DEPTH_ = ?, UPDATED_AT_ = ?, UPDATED_BY_ = ?
					WHERE ID_ = ?
				`, name, slug, strings.ToUpper(slug), description, visibility, defaultWorkflowID, path, depth, now, nullIfEmpty(actor), descendant.ID); err != nil {
					return err
				}
				continue
			}
			if _, err = tx.ExecContext(ctx, `
				UPDATE project
				SET PATH_ = ?, DEPTH_ = ?, UPDATED_AT_ = ?, UPDATED_BY_ = ?
				WHERE ID_ = ?
			`, path, depth, now, nullIfEmpty(actor), descendant.ID); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) MoveProject(ctx context.Context, input kanban.ProjectMoveInput, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.project.moved", input.ID, "", actor, input, func(tx *sql.Tx, revision int64) error {
		project, err := projectByIDTx(ctx, tx, input.ID)
		if err != nil {
			return err
		}
		if input.ParentID == nil || *input.ParentID == "" {
			parentID := kanban.DefaultProjectID
			input.ParentID = &parentID
		}
		parent, err := projectByIDTx(ctx, tx, *input.ParentID)
		if err != nil {
			return err
		}
		var cycleCount int
		if err = tx.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM project_closure
			WHERE ANCESTOR_ID_ = ? AND DESCENDANT_ID_ = ?
		`, project.ID, parent.ID).Scan(&cycleCount); err != nil {
			return err
		}
		if cycleCount > 0 {
			return errors.New("不能把项目移动到自身或子项目下")
		}
		position := project.Position
		if input.Position != nil {
			position = *input.Position
		}
		newPath := childProjectPath(&parent, project.Slug)
		newDepth := parent.Depth + 1
		now := time.Now().UTC().Format(time.RFC3339Nano)
		descendants, err := descendantProjectsTx(ctx, tx, project.ID)
		if err != nil {
			return err
		}
		for _, descendant := range descendants {
			path := newPath
			if descendant.ID != project.ID {
				suffix := strings.TrimPrefix(descendant.Path, project.Path+"/")
				path = newPath + "/" + suffix
			}
			depth := newDepth + (descendant.Depth - project.Depth)
			if descendant.ID == project.ID {
				if _, err = tx.ExecContext(ctx, `
					UPDATE project
					SET PARENT_ID_ = ?, POSITION_ = ?, PATH_ = ?, DEPTH_ = ?, UPDATED_AT_ = ?, UPDATED_BY_ = ?
					WHERE ID_ = ?
				`, parent.ID, position, path, depth, now, nullIfEmpty(actor), descendant.ID); err != nil {
					return err
				}
				continue
			}
			if _, err = tx.ExecContext(ctx, `
				UPDATE project
				SET PATH_ = ?, DEPTH_ = ?, UPDATED_AT_ = ?, UPDATED_BY_ = ?
				WHERE ID_ = ?
			`, path, depth, now, nullIfEmpty(actor), descendant.ID); err != nil {
				return err
			}
		}
		if _, err = tx.ExecContext(ctx, `
			DELETE FROM project_closure
			WHERE DESCENDANT_ID_ IN (
				SELECT DESCENDANT_ID_ FROM project_closure WHERE ANCESTOR_ID_ = ?
			)
			AND ANCESTOR_ID_ NOT IN (
				SELECT DESCENDANT_ID_ FROM project_closure WHERE ANCESTOR_ID_ = ?
			)
		`, project.ID, project.ID); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
			SELECT super.ANCESTOR_ID_, sub.DESCENDANT_ID_, super.DEPTH_ + sub.DEPTH_ + 1
			FROM project_closure AS super
			CROSS JOIN project_closure AS sub
			WHERE super.DESCENDANT_ID_ = ? AND sub.ANCESTOR_ID_ = ?
		`, parent.ID, project.ID)
		return err
	})
}

func (s *Store) withRevision(
	ctx context.Context,
	boardID string,
	eventType string,
	projectID string,
	issueID string,
	actor string,
	payload any,
	fn func(tx *sql.Tx, revision int64) error,
) (int64, error) {
	boardID = normalizeBoardID(boardID)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	revision, err := revisionTx(ctx, tx, boardID)
	if err != nil {
		return 0, err
	}
	revision++
	if err = fn(tx, revision); err != nil {
		return 0, err
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO event_log (
			BOARD_ID_, PROJECT_ID_, ISSUE_ID_, REVISION_, EVENT_TYPE_, ACTOR_ID_, PAYLOAD_JSON_, CREATED_AT_
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, boardID, nullIfEmpty(projectID), nullIfEmpty(issueID), revision, eventType, nullIfEmpty(actor), string(payloadJSON), now); err != nil {
		return 0, err
	}
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO board_meta (BOARD_ID_, KEY_, VALUE_)
		VALUES (?, 'revision', ?)
		ON CONFLICT(BOARD_ID_, KEY_) DO UPDATE SET VALUE_ = excluded.VALUE_
	`, boardID, fmt.Sprint(revision)); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return revision, nil
}

func revisionTx(ctx context.Context, tx *sql.Tx, boardID string) (int64, error) {
	var value string
	err := tx.QueryRowContext(ctx, `
		SELECT VALUE_ FROM board_meta
		WHERE BOARD_ID_ = ? AND KEY_ = 'revision'
	`, boardID).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	var revision int64
	_, err = fmt.Sscan(value, &revision)
	return revision, err
}

type issueScanner interface {
	Scan(dest ...any) error
}

func projectByIDTx(ctx context.Context, tx *sql.Tx, projectID string) (kanban.Project, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_,
			VISIBILITY_, DEFAULT_WORKFLOW_ID_, ARCHIVED_AT_, CREATED_AT_, UPDATED_AT_, DELETED_AT_,
			CREATED_BY_, UPDATED_BY_
		FROM project
		WHERE ID_ = ? AND DELETED_AT_ IS NULL
	`, projectID)
	return scanProject(row)
}

func descendantProjectsTx(ctx context.Context, tx *sql.Tx, projectID string) ([]kanban.Project, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT p.ID_, p.PARENT_ID_, p.SLUG_, p.KEY_, p.NAME_, p.DESCRIPTION_, p.PATH_, p.DEPTH_, p.POSITION_,
			p.VISIBILITY_, p.DEFAULT_WORKFLOW_ID_, p.ARCHIVED_AT_, p.CREATED_AT_, p.UPDATED_AT_, p.DELETED_AT_,
			p.CREATED_BY_, p.UPDATED_BY_
		FROM project p
		JOIN project_closure pc ON pc.DESCENDANT_ID_ = p.ID_
		WHERE pc.ANCESTOR_ID_ = ? AND p.DELETED_AT_ IS NULL
		ORDER BY pc.DEPTH_ ASC, p.PATH_ ASC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var projects []kanban.Project
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, rows.Err()
}

func scanProject(scanner issueScanner) (kanban.Project, error) {
	var project kanban.Project
	var parentID, archivedAt, deletedAt, createdBy, updatedBy sql.NullString
	var createdAt, updatedAt string
	err := scanner.Scan(
		&project.ID,
		&parentID,
		&project.Slug,
		&project.Key,
		&project.Name,
		&project.Description,
		&project.Path,
		&project.Depth,
		&project.Position,
		&project.Visibility,
		&project.DefaultWorkflowID,
		&archivedAt,
		&createdAt,
		&updatedAt,
		&deletedAt,
		&createdBy,
		&updatedBy,
	)
	if err != nil {
		return project, err
	}
	project.ParentID = stringPtr(parentID)
	project.CreatedBy = stringPtr(createdBy)
	project.UpdatedBy = stringPtr(updatedBy)
	parsedCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return project, err
	}
	parsedUpdatedAt, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return project, err
	}
	project.CreatedAt = parsedCreatedAt
	project.UpdatedAt = parsedUpdatedAt
	project.ArchivedAt, err = parseOptionalTime(archivedAt)
	if err != nil {
		return project, err
	}
	project.DeletedAt, err = parseOptionalTime(deletedAt)
	return project, err
}

func sortProjects(projects []kanban.Project) []kanban.Project {
	children := map[string][]kanban.Project{}
	for _, project := range projects {
		parentID := ""
		if project.ParentID != nil {
			parentID = *project.ParentID
		}
		children[parentID] = append(children[parentID], project)
	}
	for parentID := range children {
		sort.SliceStable(children[parentID], func(i, j int) bool {
			left := children[parentID][i]
			right := children[parentID][j]
			if left.Position != right.Position {
				return left.Position < right.Position
			}
			if left.Name != right.Name {
				return left.Name < right.Name
			}
			return left.ID < right.ID
		})
	}
	var sorted []kanban.Project
	var visit func(parentID string)
	visit = func(parentID string) {
		for _, project := range children[parentID] {
			sorted = append(sorted, project)
			visit(project.ID)
		}
	}
	visit("")
	if len(sorted) != len(projects) {
		seen := map[string]bool{}
		for _, project := range sorted {
			seen[project.ID] = true
		}
		for _, project := range projects {
			if !seen[project.ID] {
				sorted = append(sorted, project)
			}
		}
	}
	return sorted
}

func childProjectPath(parent *kanban.Project, slug string) string {
	if parent == nil || parent.ID == kanban.DefaultProjectID {
		return slug
	}
	return strings.Trim(parent.Path, "/") + "/" + slug
}

func issueSelectSQL(where string) string {
	return `
		SELECT i.ID_, i.PROJECT_ID_, COALESCE(p.PATH_, ''), COALESCE(p.NAME_, ''),
			i.WORKFLOW_ID_,
			i.STAGE_ID_, COALESCE(ws.KEY_, ''), COALESCE(ws.NAME_, ''),
			i.STATUS_ID_, COALESCE(wst.KEY_, ''), COALESCE(wst.NAME_, ''), COALESCE(wst.COLUMN_KEY_, ''),
			i.TITLE_, i.DESCRIPTION_, i.PRIORITY_, COALESCE(i.SEVERITY_, 'medium'), i.POSITION_,
			i.ASSIGNEE_ID_, i.WORKER_TYPE_, i.WORKER_ID_, i.WORKER_AGENT_, i.REVIEWER_ID_,
			i.REVIEW_REQUIRED_, i.ACTIVE_REVIEW_ID_, i.ACTIVE_RUN_ID_,
			ar.CHAT_ID_, CASE WHEN ar.STATUS_ = 'running' THEN ar.RUN_ID_ ELSE NULL END, ar.STATUS_,
			ia.ID_, COALESCE(ia.ENABLED_, 0), ia.CRON_, ia.MESSAGE_, ia.TIMEZONE_,
			i.CREATED_AT_, i.UPDATED_AT_, i.REVISION_, i.DELETED_AT_,
			i.CREATED_BY_, i.UPDATED_BY_, i.CREATED_BY_AGENT_, i.UPDATED_BY_AGENT_
		FROM issue i
		LEFT JOIN project p ON p.ID_ = i.PROJECT_ID_
		LEFT JOIN workflow_stage ws ON ws.ID_ = i.STAGE_ID_
		LEFT JOIN workflow_status wst ON wst.ID_ = i.STATUS_ID_
		LEFT JOIN agent_run ar ON ar.ID_ = i.ACTIVE_RUN_ID_
		LEFT JOIN issue_automation ia ON ia.ISSUE_ID_ = i.ID_ AND ia.DELETED_AT_ IS NULL
	` + where
}

func scanIssues(rows *sql.Rows) ([]kanban.Issue, error) {
	var issues []kanban.Issue
	for rows.Next() {
		issue, err := scanIssue(rows)
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}
	return issues, rows.Err()
}

func scanIssue(scanner issueScanner) (kanban.Issue, error) {
	var issue kanban.Issue
	var assigneeID, workerType, workerID, workerAgent, reviewerID sql.NullString
	var activeReviewID, activeRunID, chatID, runID, runState sql.NullString
	var automationID, automationCron, automationMessage, automationTimezone sql.NullString
	var createdAt, updatedAt string
	var deletedAt, createdBy, updatedBy, createdByAgent, updatedByAgent sql.NullString
	var severity sql.NullString
	var reviewRequired, automationEnabled int
	err := scanner.Scan(
		&issue.ID,
		&issue.ProjectID,
		&issue.ProjectPath,
		&issue.ProjectName,
		&issue.WorkflowID,
		&issue.StageID,
		&issue.StageKey,
		&issue.StageName,
		&issue.StatusID,
		&issue.StatusKey,
		&issue.StatusName,
		&issue.ColumnKey,
		&issue.Title,
		&issue.Description,
		&issue.Priority,
		&severity,
		&issue.Position,
		&assigneeID,
		&workerType,
		&workerID,
		&workerAgent,
		&reviewerID,
		&reviewRequired,
		&activeReviewID,
		&activeRunID,
		&chatID,
		&runID,
		&runState,
		&automationID,
		&automationEnabled,
		&automationCron,
		&automationMessage,
		&automationTimezone,
		&createdAt,
		&updatedAt,
		&issue.Revision,
		&deletedAt,
		&createdBy,
		&updatedBy,
		&createdByAgent,
		&updatedByAgent,
	)
	if err != nil {
		return issue, err
	}
	issue.BoardID = kanban.DefaultBoardID
	if issue.StatusKey == "" {
		issue.StatusKey = string(kanban.StatusBacklog)
	}
	if issue.ColumnKey == "" {
		issue.ColumnKey = issue.StatusKey
	}
	issue.Status = kanban.Status(issue.ColumnKey)
	if severity.Valid {
		issue.Severity = kanban.Severity(severity.String)
	} else {
		issue.Severity = kanban.SeverityMedium
	}
	issue.AssigneeID = stringPtr(assigneeID)
	issue.WorkerType = stringPtr(workerType)
	issue.WorkerID = stringPtr(workerID)
	issue.WorkerAgent = stringPtr(workerAgent)
	issue.AssigneeAgentKey = issue.WorkerAgent
	issue.ReviewerID = stringPtr(reviewerID)
	issue.ReviewRequired = reviewRequired == 1
	issue.ActiveReviewID = stringPtr(activeReviewID)
	issue.ActiveRunID = stringPtr(activeRunID)
	issue.ChatID = stringPtr(chatID)
	issue.RunID = stringPtr(runID)
	if runState.Valid {
		state := kanban.RunState(runState.String)
		issue.RunState = &state
	}
	issue.AutomationID = stringPtr(automationID)
	issue.AutomationEnabled = automationEnabled == 1
	issue.AutomationCron = stringPtr(automationCron)
	issue.AutomationMessage = stringPtr(automationMessage)
	issue.AutomationTimezone = stringPtr(automationTimezone)
	issue.CreatedBy = stringPtr(createdBy)
	issue.UpdatedBy = stringPtr(updatedBy)
	issue.CreatedByAgent = stringPtr(createdByAgent)
	issue.UpdatedByAgent = stringPtr(updatedByAgent)
	parsedCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return issue, err
	}
	parsedUpdatedAt, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return issue, err
	}
	issue.CreatedAt = parsedCreatedAt
	issue.UpdatedAt = parsedUpdatedAt
	issue.DeletedAt, err = parseOptionalTime(deletedAt)
	if err != nil {
		return issue, err
	}
	issue.Attachments = []kanban.Attachment{}
	return issue, nil
}

func (s *Store) attachmentsForIssues(ctx context.Context, issues []kanban.Issue) error {
	if len(issues) == 0 {
		return nil
	}
	ids := make([]string, 0, len(issues))
	for _, issue := range issues {
		ids = append(ids, issue.ID)
	}
	ids = normalizedIDList(ids)
	if len(ids) == 0 {
		return nil
	}
	where, args := sqlInClause("ISSUE_ID_", ids)
	rows, err := s.db.QueryContext(ctx, `
		SELECT ISSUE_ID_, METADATA_JSON_, KIND_, NAME_, MIME_TYPE_, SIZE_BYTES_, URL_, TEXT_
		FROM issue_attachment
		WHERE DELETED_AT_ IS NULL AND `+where+`
		ORDER BY ISSUE_ID_ ASC, CREATED_AT_ ASC, ID_ ASC
	`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	byIssueID := map[string][]kanban.Attachment{}
	for rows.Next() {
		var issueID, metadataJSON, kind, name, mimeType string
		var sizeBytes int64
		var url, text sql.NullString
		if err := rows.Scan(&issueID, &metadataJSON, &kind, &name, &mimeType, &sizeBytes, &url, &text); err != nil {
			return err
		}
		attachment := kanban.Attachment{}
		if strings.TrimSpace(metadataJSON) != "" {
			_ = json.Unmarshal([]byte(metadataJSON), &attachment)
		}
		if attachment == nil {
			attachment = kanban.Attachment{}
		}
		setAttachmentDefault(attachment, "kind", kind)
		setAttachmentDefault(attachment, "name", name)
		setAttachmentDefault(attachment, "mimeType", mimeType)
		if sizeBytes > 0 && attachment["sizeBytes"] == nil {
			attachment["sizeBytes"] = sizeBytes
		}
		if url.Valid && attachment["url"] == nil {
			attachment["url"] = url.String
		}
		if text.Valid && attachment["text"] == nil {
			attachment["text"] = text.String
		}
		byIssueID[issueID] = append(byIssueID[issueID], attachment)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for i := range issues {
		issues[i].Attachments = byIssueID[issues[i].ID]
		if issues[i].Attachments == nil {
			issues[i].Attachments = []kanban.Attachment{}
		}
	}
	return nil
}

func issueParams(issue kanban.Issue) []any {
	if issue.Severity == "" {
		issue.Severity = kanban.SeverityMedium
	}
	if issue.WorkerAgent == nil {
		issue.WorkerAgent = issue.AssigneeAgentKey
	}
	if issue.AssigneeAgentKey == nil {
		issue.AssigneeAgentKey = issue.WorkerAgent
	}
	if issue.WorkerType == nil && issue.WorkerAgent != nil {
		workerType := "agent"
		issue.WorkerType = &workerType
	}
	return []any{
		issue.ID,
		issue.ProjectID,
		issue.WorkflowID,
		issue.StageID,
		issue.StatusID,
		issue.Title,
		issue.Description,
		issue.Priority,
		issue.Severity,
		issue.Position,
		issue.AssigneeID,
		issue.WorkerType,
		issue.WorkerID,
		issue.WorkerAgent,
		issue.ReviewerID,
		boolInt(issue.ReviewRequired),
		issue.ActiveReviewID,
		issue.ActiveRunID,
		issue.Revision,
		issue.CreatedBy,
		issue.UpdatedBy,
		issue.CreatedByAgent,
		issue.UpdatedByAgent,
		issue.CreatedAt.UTC().Format(time.RFC3339Nano),
		issue.UpdatedAt.UTC().Format(time.RFC3339Nano),
		timePtrString(issue.DeletedAt),
	}
}

func persistRunTx(ctx context.Context, tx *sql.Tx, issue *kanban.Issue, actor string, now time.Time) error {
	nowText := now.UTC().Format(time.RFC3339Nano)
	if issue.RunID != nil {
		if issue.ActiveRunID == nil {
			activeRunID := newID("agent-run")
			issue.ActiveRunID = &activeRunID
		}
		status := "running"
		if issue.RunState != nil {
			status = normalizeAgentRunStatus(string(*issue.RunState))
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO agent_run (
				ID_, ISSUE_ID_, WORKER_AGENT_, DELEGATED_BY_, CHAT_ID_, RUN_ID_, STATUS_,
				STARTED_AT_, CREATED_AT_, UPDATED_AT_
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(ID_) DO UPDATE SET
				WORKER_AGENT_ = excluded.WORKER_AGENT_,
				CHAT_ID_ = excluded.CHAT_ID_,
				RUN_ID_ = excluded.RUN_ID_,
				STATUS_ = excluded.STATUS_,
				UPDATED_AT_ = excluded.UPDATED_AT_
		`, *issue.ActiveRunID, issue.ID, issue.WorkerAgent, nullIfEmpty(actor), issue.ChatID, issue.RunID, status, nowText, nowText, nowText)
		return err
	}
	if issue.ActiveRunID != nil && issue.RunState != nil && *issue.RunState != kanban.RunStateRunning {
		status := normalizeAgentRunStatus(string(*issue.RunState))
		if _, err := tx.ExecContext(ctx, `
			UPDATE agent_run
			SET STATUS_ = ?, FINISHED_AT_ = ?, UPDATED_AT_ = ?
			WHERE ID_ = ?
		`, status, nowText, nowText, *issue.ActiveRunID); err != nil {
			return err
		}
	}
	return nil
}

func replaceAutomationTx(ctx context.Context, tx *sql.Tx, issue kanban.Issue, actor string, now time.Time) error {
	hasAutomation := issue.AutomationID != nil || issue.AutomationEnabled || issue.AutomationCron != nil || issue.AutomationMessage != nil || issue.AutomationTimezone != nil
	if !hasAutomation {
		_, err := tx.ExecContext(ctx, `
			UPDATE issue_automation
			SET DELETED_AT_ = ?, UPDATED_AT_ = ?, UPDATED_BY_ = ?
			WHERE ISSUE_ID_ = ? AND DELETED_AT_ IS NULL
		`, now.UTC().Format(time.RFC3339Nano), now.UTC().Format(time.RFC3339Nano), nullIfEmpty(actor), issue.ID)
		return err
	}
	automationID := issue.AutomationID
	if automationID == nil || strings.TrimSpace(*automationID) == "" {
		generated := "automation-" + issue.ID
		automationID = &generated
	}
	nowText := now.UTC().Format(time.RFC3339Nano)
	_, err := tx.ExecContext(ctx, `
		INSERT INTO issue_automation (
			ID_, ISSUE_ID_, EXTERNAL_AUTOMATION_ID_, ENABLED_, CRON_, TIMEZONE_, MESSAGE_,
			CREATED_BY_, UPDATED_BY_, CREATED_AT_, UPDATED_AT_
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(ID_) DO UPDATE SET
			EXTERNAL_AUTOMATION_ID_ = excluded.EXTERNAL_AUTOMATION_ID_,
			ENABLED_ = excluded.ENABLED_,
			CRON_ = excluded.CRON_,
			TIMEZONE_ = excluded.TIMEZONE_,
			MESSAGE_ = excluded.MESSAGE_,
			UPDATED_BY_ = excluded.UPDATED_BY_,
			UPDATED_AT_ = excluded.UPDATED_AT_,
			DELETED_AT_ = NULL
	`, *automationID, issue.ID, issue.AutomationID, boolInt(issue.AutomationEnabled),
		issue.AutomationCron, issue.AutomationTimezone, issue.AutomationMessage, nullIfEmpty(actor), nullIfEmpty(actor), nowText, nowText)
	return err
}

func (s *Store) replaceAttachments(ctx context.Context, tx *sql.Tx, issueID string, attachments []kanban.Attachment, createdBy *string, createdByAgent *string, createdAt string) error {
	if tx != nil {
		return replaceAttachmentsTx(ctx, tx, issueID, attachments, createdBy, createdByAgent, createdAt)
	}
	dbTx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = dbTx.Rollback()
		}
	}()
	if err = replaceAttachmentsTx(ctx, dbTx, issueID, attachments, createdBy, createdByAgent, createdAt); err != nil {
		return err
	}
	return dbTx.Commit()
}

func replaceAttachmentsTx(ctx context.Context, tx *sql.Tx, issueID string, attachments []kanban.Attachment, createdBy *string, createdByAgent *string, createdAt string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM issue_attachment WHERE ISSUE_ID_ = ?`, issueID); err != nil {
		return err
	}
	if createdAt == "" {
		createdAt = time.Now().UTC().Format(time.RFC3339Nano)
	}
	for i, attachment := range attachments {
		metadata, _ := json.Marshal(attachment)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO issue_attachment (
				ID_, ISSUE_ID_, KIND_, NAME_, MIME_TYPE_, SIZE_BYTES_, URL_, TEXT_, METADATA_JSON_,
				CREATED_BY_, CREATED_BY_AGENT_, CREATED_AT_
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, attachmentID(issueID, i, attachment), issueID, attachmentString(attachment, "kind"),
			attachmentString(attachment, "name"), attachmentString(attachment, "mimeType"),
			attachmentSize(attachment), nullIfEmpty(attachmentString(attachment, "url")),
			nullIfEmpty(attachmentString(attachment, "text")), string(metadata), createdBy, createdByAgent, createdAt)
		if err != nil {
			return err
		}
	}
	return nil
}

var workflowStartStageKeys = map[string]string{
	"workflow-standard-requirement":   "requirement_clarification",
	"workflow-bug-fix":                "issue_triage",
	"workflow-optimization-iteration": "current_assessment",
	"workflow-free-task":              "todo",
	"workflow-graphic-publish":        "topic_planning",
}

func workflowStartStageID(workflowID string) string {
	if key, ok := workflowStartStageKeys[workflowID]; ok {
		return workflowStageID(workflowID, key)
	}
	return workflowStageID(kanban.DefaultWorkflowID, "requirement_clarification")
}

func statusDefID(key string) string {
	return "status-def-" + key
}

func stageDefID(key string) string {
	return "stage-def-" + key
}

func workflowStageID(workflowID string, stageKey string) string {
	return workflowID + "-stage-" + stageKey
}

func workflowStatusID(workflowID string, statusKey string) string {
	return workflowID + "-status-" + statusKey
}

func normalizeAgentRunStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "queued":
		return "queued"
	case "completed", "succeeded", "success":
		return "completed"
	case "failed", "error":
		return "failed"
	case "cancelled", "canceled":
		return "cancelled"
	default:
		return "running"
	}
}

func normalizeBoardID(boardID string) string {
	boardID = strings.TrimSpace(boardID)
	if boardID == "" {
		return kanban.DefaultBoardID
	}
	return boardID
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func nullIfEmpty(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func nullStringValue(value sql.NullString) any {
	if !value.Valid || strings.TrimSpace(value.String) == "" {
		return nil
	}
	return value.String
}

func nullStringPtr(value sql.NullString) *string {
	if !value.Valid || strings.TrimSpace(value.String) == "" {
		return nil
	}
	next := value.String
	return &next
}

func stringPtr(value sql.NullString) *string {
	if !value.Valid || value.String == "" {
		return nil
	}
	next := value.String
	return &next
}

func actorPtr(actor string) *string {
	actor = strings.TrimSpace(actor)
	if actor == "" {
		return nil
	}
	return &actor
}

func parseOptionalTime(value sql.NullString) (*time.Time, error) {
	if !value.Valid || value.String == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value.String)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func timePtrString(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func newID(prefix string) string {
	var random [6]byte
	if _, err := rand.Read(random[:]); err != nil {
		return prefix + "-" + strings.ToLower(hex.EncodeToString([]byte(time.Now().Format("150405.000000000"))))
	}
	return prefix + "-" + strings.ToLower(time.Now().UTC().Format("20060102150405")+hex.EncodeToString(random[:]))
}

func attachmentID(issueID string, index int, attachment kanban.Attachment) string {
	if raw, ok := attachment["id"].(string); ok && strings.TrimSpace(raw) != "" {
		return issueID + "-attachment-" + strings.TrimSpace(raw)
	}
	return issueID + "-attachment-" + strconv.Itoa(index+1)
}

func attachmentString(attachment kanban.Attachment, key string) string {
	if value, ok := attachment[key].(string); ok {
		return value
	}
	return ""
}

func attachmentSize(attachment kanban.Attachment) int64 {
	switch value := attachment["sizeBytes"].(type) {
	case int64:
		return value
	case int:
		return int64(value)
	case float64:
		return int64(value)
	default:
		return 0
	}
}

func setAttachmentDefault(attachment kanban.Attachment, key string, value string) {
	if value != "" && attachment[key] == nil {
		attachment[key] = value
	}
}
