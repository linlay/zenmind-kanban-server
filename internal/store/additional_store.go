package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"zenmind-kanban-server/internal/kanban"
)

func (s *Store) ListUsers(ctx context.Context) ([]kanban.UserAccount, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, EMAIL_, DISPLAY_NAME_, AVATAR_URL_, STATUS_, CREATED_AT_, UPDATED_AT_, DELETED_AT_
		FROM user_account
		WHERE DELETED_AT_ IS NULL
		ORDER BY DISPLAY_NAME_ ASC, EMAIL_ ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []kanban.UserAccount{}
	for rows.Next() {
		var item kanban.UserAccount
		var avatarURL, deletedAt sql.NullString
		var createdAt, updatedAt string
		if err := rows.Scan(&item.ID, &item.Email, &item.DisplayName, &avatarURL, &item.Status, &createdAt, &updatedAt, &deletedAt); err != nil {
			return nil, err
		}
		item.AvatarURL = stringPtr(avatarURL)
		var err error
		item.CreatedAt, err = parseTime(createdAt)
		if err != nil {
			return nil, err
		}
		item.UpdatedAt, err = parseTime(updatedAt)
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

func (s *Store) ListTeamMembers(ctx context.Context) ([]kanban.TeamMember, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT TEAM_ID_, USER_ID_, ROLE_, INVITED_BY_, JOINED_AT_, LEFT_AT_
		FROM team_member
		WHERE LEFT_AT_ IS NULL
		ORDER BY TEAM_ID_ ASC, ROLE_ ASC, USER_ID_ ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []kanban.TeamMember{}
	for rows.Next() {
		var item kanban.TeamMember
		var invitedBy, leftAt sql.NullString
		var joinedAt string
		if err := rows.Scan(&item.TeamID, &item.UserID, &item.Role, &invitedBy, &joinedAt, &leftAt); err != nil {
			return nil, err
		}
		item.InvitedBy = stringPtr(invitedBy)
		var err error
		item.JoinedAt, err = parseTime(joinedAt)
		if err != nil {
			return nil, err
		}
		item.LeftAt, err = parseOptionalTime(leftAt)
		if err != nil {
			return nil, err
		}
		values = append(values, item)
	}
	return values, rows.Err()
}

func (s *Store) ListAgents(ctx context.Context) ([]kanban.Agent, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, AGENT_KEY_, NAME_, DESCRIPTION_, ROLE_, MODEL_, MODEL_VERSION_, ENABLED_,
			CREATED_BY_, UPDATED_BY_, CREATED_AT_, UPDATED_AT_, DELETED_AT_
		FROM agent
		WHERE DELETED_AT_ IS NULL
		ORDER BY ENABLED_ DESC, NAME_ ASC, AGENT_KEY_ ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []kanban.Agent{}
	for rows.Next() {
		var item kanban.Agent
		var model, modelVersion, createdBy, updatedBy, deletedAt sql.NullString
		var enabled int
		var createdAt, updatedAt string
		if err := rows.Scan(&item.ID, &item.AgentKey, &item.Name, &item.Description, &item.Role, &model, &modelVersion, &enabled, &createdBy, &updatedBy, &createdAt, &updatedAt, &deletedAt); err != nil {
			return nil, err
		}
		item.Model = stringPtr(model)
		item.ModelVersion = stringPtr(modelVersion)
		item.Enabled = enabled == 1
		item.CreatedBy = stringPtr(createdBy)
		item.UpdatedBy = stringPtr(updatedBy)
		var err error
		item.CreatedAt, err = parseTime(createdAt)
		if err != nil {
			return nil, err
		}
		item.UpdatedAt, err = parseTime(updatedAt)
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

func (s *Store) ListIssueLabels(ctx context.Context) ([]kanban.IssueLabel, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, PROJECT_ID_, KEY_, NAME_, COLOR_, CREATED_BY_, UPDATED_BY_, CREATED_AT_, UPDATED_AT_, DELETED_AT_
		FROM issue_label
		WHERE DELETED_AT_ IS NULL
		ORDER BY PROJECT_ID_ ASC, NAME_ ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []kanban.IssueLabel{}
	for rows.Next() {
		var item kanban.IssueLabel
		var createdBy, updatedBy, deletedAt sql.NullString
		var createdAt, updatedAt string
		if err := rows.Scan(&item.ID, &item.ProjectID, &item.Key, &item.Name, &item.Color, &createdBy, &updatedBy, &createdAt, &updatedAt, &deletedAt); err != nil {
			return nil, err
		}
		item.CreatedBy = stringPtr(createdBy)
		item.UpdatedBy = stringPtr(updatedBy)
		var err error
		item.CreatedAt, err = parseTime(createdAt)
		if err != nil {
			return nil, err
		}
		item.UpdatedAt, err = parseTime(updatedAt)
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

func (s *Store) ListIssueLabelsForProjects(ctx context.Context, projectIDs []string) ([]kanban.IssueLabel, error) {
	ids := normalizedIDList(projectIDs)
	if len(ids) == 0 {
		return []kanban.IssueLabel{}, nil
	}
	where, args := sqlInClause("PROJECT_ID_", ids)
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, PROJECT_ID_, KEY_, NAME_, COLOR_, CREATED_BY_, UPDATED_BY_, CREATED_AT_, UPDATED_AT_, DELETED_AT_
		FROM issue_label
		WHERE DELETED_AT_ IS NULL AND `+where+`
		ORDER BY PROJECT_ID_ ASC, NAME_ ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []kanban.IssueLabel{}
	for rows.Next() {
		var item kanban.IssueLabel
		var createdBy, updatedBy, deletedAt sql.NullString
		var createdAt, updatedAt string
		if err := rows.Scan(&item.ID, &item.ProjectID, &item.Key, &item.Name, &item.Color, &createdBy, &updatedBy, &createdAt, &updatedAt, &deletedAt); err != nil {
			return nil, err
		}
		item.CreatedBy = stringPtr(createdBy)
		item.UpdatedBy = stringPtr(updatedBy)
		var err error
		item.CreatedAt, err = parseTime(createdAt)
		if err != nil {
			return nil, err
		}
		item.UpdatedAt, err = parseTime(updatedAt)
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

func (s *Store) ListIssueLabelLinks(ctx context.Context) ([]kanban.IssueLabelLink, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ISSUE_ID_, LABEL_ID_
		FROM issue_label_link
		ORDER BY ISSUE_ID_ ASC, LABEL_ID_ ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []kanban.IssueLabelLink{}
	for rows.Next() {
		var item kanban.IssueLabelLink
		if err := rows.Scan(&item.IssueID, &item.LabelID); err != nil {
			return nil, err
		}
		values = append(values, item)
	}
	return values, rows.Err()
}

func (s *Store) ListIssueLabelLinksForIssues(ctx context.Context, issueIDs []string) ([]kanban.IssueLabelLink, error) {
	ids := normalizedIDList(issueIDs)
	if len(ids) == 0 {
		return []kanban.IssueLabelLink{}, nil
	}
	where, args := sqlInClause("ISSUE_ID_", ids)
	rows, err := s.db.QueryContext(ctx, `
		SELECT ISSUE_ID_, LABEL_ID_
		FROM issue_label_link
		WHERE `+where+`
		ORDER BY ISSUE_ID_ ASC, LABEL_ID_ ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []kanban.IssueLabelLink{}
	for rows.Next() {
		var item kanban.IssueLabelLink
		if err := rows.Scan(&item.IssueID, &item.LabelID); err != nil {
			return nil, err
		}
		values = append(values, item)
	}
	return values, rows.Err()
}

func (s *Store) ListIssueDependencies(ctx context.Context) ([]kanban.IssueDependency, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, FROM_ISSUE_ID_, TO_ISSUE_ID_, TYPE_, CREATED_BY_, CREATED_AT_, DELETED_AT_
		FROM issue_dependency
		WHERE DELETED_AT_ IS NULL
		ORDER BY FROM_ISSUE_ID_ ASC, TO_ISSUE_ID_ ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []kanban.IssueDependency{}
	for rows.Next() {
		var item kanban.IssueDependency
		var createdBy, deletedAt sql.NullString
		var createdAt string
		if err := rows.Scan(&item.ID, &item.FromIssueID, &item.ToIssueID, &item.Type, &createdBy, &createdAt, &deletedAt); err != nil {
			return nil, err
		}
		item.CreatedBy = stringPtr(createdBy)
		var err error
		item.CreatedAt, err = parseTime(createdAt)
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

func (s *Store) ListIssueDependenciesForIssues(ctx context.Context, issueIDs []string) ([]kanban.IssueDependency, error) {
	ids := normalizedIDList(issueIDs)
	if len(ids) == 0 {
		return []kanban.IssueDependency{}, nil
	}
	placeholders := sqlPlaceholders(len(ids))
	args := append(sqlArgs(ids), sqlArgs(ids)...)
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, FROM_ISSUE_ID_, TO_ISSUE_ID_, TYPE_, CREATED_BY_, CREATED_AT_, DELETED_AT_
		FROM issue_dependency
		WHERE DELETED_AT_ IS NULL
			AND (FROM_ISSUE_ID_ IN (`+placeholders+`) OR TO_ISSUE_ID_ IN (`+placeholders+`))
		ORDER BY FROM_ISSUE_ID_ ASC, TO_ISSUE_ID_ ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []kanban.IssueDependency{}
	for rows.Next() {
		var item kanban.IssueDependency
		var createdBy, deletedAt sql.NullString
		var createdAt string
		if err := rows.Scan(&item.ID, &item.FromIssueID, &item.ToIssueID, &item.Type, &createdBy, &createdAt, &deletedAt); err != nil {
			return nil, err
		}
		item.CreatedBy = stringPtr(createdBy)
		var err error
		item.CreatedAt, err = parseTime(createdAt)
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

func (s *Store) ListReviews(ctx context.Context) ([]kanban.Review, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, ISSUE_ID_, AGENT_RUN_ID_, REVIEW_TYPE_, REVIEWER_ID_, STATUS_, REQUESTED_BY_,
			REQUESTED_AT_, SUBMITTED_AT_, SUMMARY_, CREATED_AT_, UPDATED_AT_, DELETED_AT_
		FROM review
		WHERE DELETED_AT_ IS NULL
		ORDER BY UPDATED_AT_ DESC, ID_ ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []kanban.Review{}
	for rows.Next() {
		item, err := scanReview(rows)
		if err != nil {
			return nil, err
		}
		values = append(values, item)
	}
	return values, rows.Err()
}

func (s *Store) ListReviewsForIssues(ctx context.Context, issueIDs []string) ([]kanban.Review, error) {
	ids := normalizedIDList(issueIDs)
	if len(ids) == 0 {
		return []kanban.Review{}, nil
	}
	where, args := sqlInClause("ISSUE_ID_", ids)
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, ISSUE_ID_, AGENT_RUN_ID_, REVIEW_TYPE_, REVIEWER_ID_, STATUS_, REQUESTED_BY_,
			REQUESTED_AT_, SUBMITTED_AT_, SUMMARY_, CREATED_AT_, UPDATED_AT_, DELETED_AT_
		FROM review
		WHERE DELETED_AT_ IS NULL AND `+where+`
		ORDER BY UPDATED_AT_ DESC, ID_ ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []kanban.Review{}
	for rows.Next() {
		item, err := scanReview(rows)
		if err != nil {
			return nil, err
		}
		values = append(values, item)
	}
	return values, rows.Err()
}

func (s *Store) ListReviewComments(ctx context.Context) ([]kanban.ReviewComment, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, REVIEW_ID_, ISSUE_ID_, AUTHOR_ID_, AUTHOR_AGENT_, BODY_, ANCHOR_JSON_, CREATED_AT_, UPDATED_AT_, DELETED_AT_
		FROM review_comment
		WHERE DELETED_AT_ IS NULL
		ORDER BY CREATED_AT_ ASC, ID_ ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []kanban.ReviewComment{}
	for rows.Next() {
		item, err := scanReviewComment(rows)
		if err != nil {
			return nil, err
		}
		values = append(values, item)
	}
	return values, rows.Err()
}

func (s *Store) ListReviewCommentsForIssues(ctx context.Context, issueIDs []string) ([]kanban.ReviewComment, error) {
	ids := normalizedIDList(issueIDs)
	if len(ids) == 0 {
		return []kanban.ReviewComment{}, nil
	}
	where, args := sqlInClause("ISSUE_ID_", ids)
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, REVIEW_ID_, ISSUE_ID_, AUTHOR_ID_, AUTHOR_AGENT_, BODY_, ANCHOR_JSON_, CREATED_AT_, UPDATED_AT_, DELETED_AT_
		FROM review_comment
		WHERE DELETED_AT_ IS NULL AND `+where+`
		ORDER BY CREATED_AT_ ASC, ID_ ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []kanban.ReviewComment{}
	for rows.Next() {
		item, err := scanReviewComment(rows)
		if err != nil {
			return nil, err
		}
		values = append(values, item)
	}
	return values, rows.Err()
}

func (s *Store) ListAgentRuns(ctx context.Context, issueID string) ([]kanban.AgentRun, error) {
	query := `
		SELECT ID_, ISSUE_ID_, AGENT_ID_, WORKER_AGENT_, DELEGATED_BY_, CHAT_ID_, RUN_ID_, SESSION_ID_,
			STATUS_, CONFIDENCE_, INPUT_TOKENS_, OUTPUT_TOKENS_, COST_MICROS_,
			STARTED_AT_, FINISHED_AT_, ERROR_MESSAGE_, CREATED_AT_, UPDATED_AT_
		FROM agent_run
	`
	args := []any{}
	if strings.TrimSpace(issueID) != "" {
		query += ` WHERE ISSUE_ID_ = ?`
		args = append(args, strings.TrimSpace(issueID))
	}
	query += ` ORDER BY CREATED_AT_ DESC, ID_ ASC`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []kanban.AgentRun{}
	for rows.Next() {
		item, err := scanAgentRun(rows)
		if err != nil {
			return nil, err
		}
		values = append(values, item)
	}
	return values, rows.Err()
}

func (s *Store) ListAgentRunsForIssues(ctx context.Context, issueIDs []string) ([]kanban.AgentRun, error) {
	ids := normalizedIDList(issueIDs)
	if len(ids) == 0 {
		return []kanban.AgentRun{}, nil
	}
	where, args := sqlInClause("ISSUE_ID_", ids)
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, ISSUE_ID_, AGENT_ID_, WORKER_AGENT_, DELEGATED_BY_, CHAT_ID_, RUN_ID_, SESSION_ID_,
			STATUS_, CONFIDENCE_, INPUT_TOKENS_, OUTPUT_TOKENS_, COST_MICROS_,
			STARTED_AT_, FINISHED_AT_, ERROR_MESSAGE_, CREATED_AT_, UPDATED_AT_
		FROM agent_run
		WHERE `+where+`
		ORDER BY CREATED_AT_ DESC, ID_ ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []kanban.AgentRun{}
	for rows.Next() {
		item, err := scanAgentRun(rows)
		if err != nil {
			return nil, err
		}
		values = append(values, item)
	}
	return values, rows.Err()
}

func (s *Store) ListAgentToolCalls(ctx context.Context, agentRunID string) ([]kanban.AgentToolCall, error) {
	query := `
		SELECT ID_, AGENT_RUN_ID_, TOOL_NAME_, TOOL_TYPE_, STATUS_, INPUT_JSON_, OUTPUT_JSON_, STARTED_AT_, FINISHED_AT_, ERROR_MESSAGE_
		FROM agent_tool_call
	`
	args := []any{}
	if strings.TrimSpace(agentRunID) != "" {
		query += ` WHERE AGENT_RUN_ID_ = ?`
		args = append(args, strings.TrimSpace(agentRunID))
	}
	query += ` ORDER BY COALESCE(STARTED_AT_, ID_) ASC, ID_ ASC`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []kanban.AgentToolCall{}
	for rows.Next() {
		item, err := scanAgentToolCall(rows)
		if err != nil {
			return nil, err
		}
		values = append(values, item)
	}
	return values, rows.Err()
}

func (s *Store) ListAgentToolCallsForRuns(ctx context.Context, agentRunIDs []string) ([]kanban.AgentToolCall, error) {
	ids := normalizedIDList(agentRunIDs)
	if len(ids) == 0 {
		return []kanban.AgentToolCall{}, nil
	}
	where, args := sqlInClause("AGENT_RUN_ID_", ids)
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, AGENT_RUN_ID_, TOOL_NAME_, TOOL_TYPE_, STATUS_, INPUT_JSON_, OUTPUT_JSON_, STARTED_AT_, FINISHED_AT_, ERROR_MESSAGE_
		FROM agent_tool_call
		WHERE `+where+`
		ORDER BY COALESCE(STARTED_AT_, ID_) ASC, ID_ ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []kanban.AgentToolCall{}
	for rows.Next() {
		item, err := scanAgentToolCall(rows)
		if err != nil {
			return nil, err
		}
		values = append(values, item)
	}
	return values, rows.Err()
}

func (s *Store) ListRecentEvents(ctx context.Context, boardID string, limit int) ([]kanban.EventLogItem, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT ID_, BOARD_ID_, PROJECT_ID_, ISSUE_ID_, REVISION_, EVENT_TYPE_, ACTOR_ID_, ACTOR_AGENT_, PAYLOAD_JSON_, CREATED_AT_
		FROM event_log
		WHERE BOARD_ID_ = ?
		ORDER BY REVISION_ DESC, ID_ DESC
		LIMIT ?
	`, normalizeBoardID(boardID), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []kanban.EventLogItem{}
	for rows.Next() {
		var item kanban.EventLogItem
		var projectID, issueID, actorID, actorAgent sql.NullString
		var payloadJSON, createdAt string
		if err := rows.Scan(&item.ID, &item.BoardID, &projectID, &issueID, &item.Revision, &item.EventType, &actorID, &actorAgent, &payloadJSON, &createdAt); err != nil {
			return nil, err
		}
		item.ProjectID = stringPtr(projectID)
		item.IssueID = stringPtr(issueID)
		item.ActorID = stringPtr(actorID)
		item.ActorAgent = stringPtr(actorAgent)
		item.Payload = map[string]any{}
		_ = json.Unmarshal([]byte(payloadJSON), &item.Payload)
		var err error
		item.CreatedAt, err = parseTime(createdAt)
		if err != nil {
			return nil, err
		}
		values = append(values, item)
	}
	return values, rows.Err()
}

func normalizedIDList(ids []string) []string {
	seen := map[string]bool{}
	values := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		values = append(values, id)
	}
	sort.Strings(values)
	return values
}

func sqlInClause(column string, ids []string) (string, []any) {
	return column + " IN (" + sqlPlaceholders(len(ids)) + ")", sqlArgs(ids)
}

func sqlPlaceholders(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.TrimRight(strings.Repeat("?,", count), ",")
}

func sqlArgs(ids []string) []any {
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return args
}

func (s *Store) CreateUser(ctx context.Context, input kanban.UserInput, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.user.created", "", "", actor, input, func(tx *sql.Tx, revision int64) error {
		now := nowText()
		status := normalizeUserStatus(input.Status)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO user_account (ID_, EMAIL_, DISPLAY_NAME_, AVATAR_URL_, STATUS_, CREATED_AT_, UPDATED_AT_)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, newID("user"), strings.TrimSpace(input.Email), strings.TrimSpace(input.DisplayName), trimmedPtr(input.AvatarURL), status, now, now)
		return err
	})
}

func (s *Store) UpdateUser(ctx context.Context, id string, input kanban.UserUpdateInput, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.user.updated", "", "", actor, input, func(tx *sql.Tx, revision int64) error {
		result, err := tx.ExecContext(ctx, `
			UPDATE user_account
			SET EMAIL_ = CASE WHEN ? THEN ? ELSE EMAIL_ END,
				DISPLAY_NAME_ = CASE WHEN ? THEN ? ELSE DISPLAY_NAME_ END,
				AVATAR_URL_ = CASE WHEN ? THEN ? ELSE AVATAR_URL_ END,
				STATUS_ = CASE WHEN ? THEN ? ELSE STATUS_ END,
				UPDATED_AT_ = ?
			WHERE ID_ = ? AND DELETED_AT_ IS NULL
		`, boolInt(input.Email != nil), trimmedValue(input.Email), boolInt(input.DisplayName != nil), trimmedValue(input.DisplayName),
			boolInt(input.AvatarURL != nil), trimmedPtr(input.AvatarURL), boolInt(input.Status != nil), normalizeUserStatus(input.Status), nowText(), id)
		return ensureAffected(result, err)
	})
}

func (s *Store) DeleteUser(ctx context.Context, id string, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.user.deleted", "", "", actor, map[string]string{"id": id}, func(tx *sql.Tx, revision int64) error {
		result, err := tx.ExecContext(ctx, `
			UPDATE user_account SET DELETED_AT_ = ?, UPDATED_AT_ = ? WHERE ID_ = ? AND DELETED_AT_ IS NULL
		`, nowText(), nowText(), id)
		return ensureAffected(result, err)
	})
}

func (s *Store) CreateTeam(ctx context.Context, input kanban.TeamInput, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.team.created", "", "", actor, input, func(tx *sql.Tx, revision int64) error {
		now := nowText()
		name := strings.TrimSpace(input.Name)
		slug := normalizedSlug(input.Slug, name)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO team (ID_, SLUG_, NAME_, DESCRIPTION_, CREATED_BY_, UPDATED_BY_, CREATED_AT_, UPDATED_AT_)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, newID("team"), slug, name, trimmedValue(input.Description), nullIfEmpty(actor), nullIfEmpty(actor), now, now)
		return err
	})
}

func (s *Store) UpdateTeam(ctx context.Context, id string, input kanban.TeamUpdateInput, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.team.updated", "", "", actor, input, func(tx *sql.Tx, revision int64) error {
		result, err := tx.ExecContext(ctx, `
			UPDATE team
			SET SLUG_ = CASE WHEN ? THEN ? ELSE SLUG_ END,
				NAME_ = CASE WHEN ? THEN ? ELSE NAME_ END,
				DESCRIPTION_ = CASE WHEN ? THEN ? ELSE DESCRIPTION_ END,
				UPDATED_BY_ = ?,
				UPDATED_AT_ = ?
			WHERE ID_ = ? AND DELETED_AT_ IS NULL
		`, boolInt(input.Slug != nil), normalizedSlug(input.Slug, ""), boolInt(input.Name != nil), trimmedValue(input.Name),
			boolInt(input.Description != nil), trimmedValue(input.Description), nullIfEmpty(actor), nowText(), id)
		return ensureAffected(result, err)
	})
}

func (s *Store) DeleteTeam(ctx context.Context, id string, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.team.deleted", "", "", actor, map[string]string{"id": id}, func(tx *sql.Tx, revision int64) error {
		now := nowText()
		if _, err := tx.ExecContext(ctx, `UPDATE team_member SET LEFT_AT_ = ? WHERE TEAM_ID_ = ? AND LEFT_AT_ IS NULL`, now, id); err != nil {
			return err
		}
		result, err := tx.ExecContext(ctx, `
			UPDATE team SET DELETED_AT_ = ?, UPDATED_BY_ = ?, UPDATED_AT_ = ? WHERE ID_ = ? AND DELETED_AT_ IS NULL
		`, now, nullIfEmpty(actor), now, id)
		return ensureAffected(result, err)
	})
}

func (s *Store) AddTeamMember(ctx context.Context, input kanban.TeamMemberInput, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.teamMember.added", "", "", actor, input, func(tx *sql.Tx, revision int64) error {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO team_member (TEAM_ID_, USER_ID_, ROLE_, INVITED_BY_, JOINED_AT_, LEFT_AT_)
			VALUES (?, ?, ?, ?, ?, NULL)
			ON CONFLICT(TEAM_ID_, USER_ID_) DO UPDATE SET ROLE_ = excluded.ROLE_, INVITED_BY_ = excluded.INVITED_BY_, LEFT_AT_ = NULL
		`, strings.TrimSpace(input.TeamID), strings.TrimSpace(input.UserID), normalizeTeamRole(input.Role), trimmedPtr(input.InvitedBy), nowText())
		return err
	})
}

func (s *Store) UpdateTeamMember(ctx context.Context, teamID string, userID string, input kanban.TeamMemberUpdateInput, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.teamMember.updated", "", "", actor, input, func(tx *sql.Tx, revision int64) error {
		leftAt := any(nil)
		if input.LeftAt != nil {
			leftAt = normalizedTimeValue(*input.LeftAt)
		}
		result, err := tx.ExecContext(ctx, `
			UPDATE team_member
			SET ROLE_ = CASE WHEN ? THEN ? ELSE ROLE_ END,
				LEFT_AT_ = CASE WHEN ? THEN ? ELSE LEFT_AT_ END
			WHERE TEAM_ID_ = ? AND USER_ID_ = ?
		`, boolInt(input.Role != nil), normalizeTeamRole(trimmedValue(input.Role)), boolInt(input.LeftAt != nil), leftAt, teamID, userID)
		return ensureAffected(result, err)
	})
}

func (s *Store) RemoveTeamMember(ctx context.Context, teamID string, userID string, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.teamMember.removed", "", "", actor, map[string]string{"teamId": teamID, "userId": userID}, func(tx *sql.Tx, revision int64) error {
		result, err := tx.ExecContext(ctx, `UPDATE team_member SET LEFT_AT_ = ? WHERE TEAM_ID_ = ? AND USER_ID_ = ? AND LEFT_AT_ IS NULL`, nowText(), teamID, userID)
		return ensureAffected(result, err)
	})
}

func (s *Store) CreateAgent(ctx context.Context, input kanban.AgentInput, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.agent.created", "", "", actor, input, func(tx *sql.Tx, revision int64) error {
		now := nowText()
		enabled := true
		if input.Enabled != nil {
			enabled = *input.Enabled
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO agent (
				ID_, AGENT_KEY_, NAME_, DESCRIPTION_, ROLE_, MODEL_, MODEL_VERSION_, ENABLED_,
				CREATED_BY_, UPDATED_BY_, CREATED_AT_, UPDATED_AT_
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, newID("agent"), strings.TrimSpace(input.AgentKey), strings.TrimSpace(input.Name), trimmedValue(input.Description),
			trimmedValue(input.Role), trimmedPtr(input.Model), trimmedPtr(input.ModelVersion), boolInt(enabled), nullIfEmpty(actor), nullIfEmpty(actor), now, now)
		return err
	})
}

func (s *Store) UpdateAgent(ctx context.Context, id string, input kanban.AgentUpdateInput, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.agent.updated", "", "", actor, input, func(tx *sql.Tx, revision int64) error {
		result, err := tx.ExecContext(ctx, `
			UPDATE agent
			SET AGENT_KEY_ = CASE WHEN ? THEN ? ELSE AGENT_KEY_ END,
				NAME_ = CASE WHEN ? THEN ? ELSE NAME_ END,
				DESCRIPTION_ = CASE WHEN ? THEN ? ELSE DESCRIPTION_ END,
				ROLE_ = CASE WHEN ? THEN ? ELSE ROLE_ END,
				MODEL_ = CASE WHEN ? THEN ? ELSE MODEL_ END,
				MODEL_VERSION_ = CASE WHEN ? THEN ? ELSE MODEL_VERSION_ END,
				ENABLED_ = CASE WHEN ? THEN ? ELSE ENABLED_ END,
				UPDATED_BY_ = ?,
				UPDATED_AT_ = ?
			WHERE ID_ = ? AND DELETED_AT_ IS NULL
		`, boolInt(input.AgentKey != nil), trimmedValue(input.AgentKey), boolInt(input.Name != nil), trimmedValue(input.Name),
			boolInt(input.Description != nil), trimmedValue(input.Description), boolInt(input.Role != nil), trimmedValue(input.Role),
			boolInt(input.Model != nil), trimmedPtr(input.Model), boolInt(input.ModelVersion != nil), trimmedPtr(input.ModelVersion),
			boolInt(input.Enabled != nil), boolInt(input.Enabled != nil && *input.Enabled), nullIfEmpty(actor), nowText(), id)
		return ensureAffected(result, err)
	})
}

func (s *Store) DeleteAgent(ctx context.Context, id string, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.agent.deleted", "", "", actor, map[string]string{"id": id}, func(tx *sql.Tx, revision int64) error {
		result, err := tx.ExecContext(ctx, `
			UPDATE agent SET DELETED_AT_ = ?, UPDATED_BY_ = ?, UPDATED_AT_ = ? WHERE ID_ = ? AND DELETED_AT_ IS NULL
		`, nowText(), nullIfEmpty(actor), nowText(), id)
		return ensureAffected(result, err)
	})
}

func (s *Store) GrantProjectPermission(ctx context.Context, input kanban.ProjectPermissionInput, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.projectPermission.granted", input.ProjectID, "", actor, input, func(tx *sql.Tx, revision int64) error {
		inherit := true
		if input.InheritToChildren != nil {
			inherit = *input.InheritToChildren
		}
		now := nowText()
		var existingID string
		err := tx.QueryRowContext(ctx, `
			SELECT ID_
			FROM project_permission
			WHERE PROJECT_ID_ = ? AND PRINCIPAL_TYPE_ = ? AND PRINCIPAL_ID_ = ? AND DELETED_AT_ IS NULL
			LIMIT 1
		`, input.ProjectID, normalizePrincipalType(input.PrincipalType), input.PrincipalID).Scan(&existingID)
		if err == nil {
			_, err = tx.ExecContext(ctx, `
				UPDATE project_permission SET ROLE_ = ?, INHERIT_TO_CHILDREN_ = ? WHERE ID_ = ?
			`, normalizeProjectRole(input.Role), boolInt(inherit), existingID)
			return err
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO project_permission (
				ID_, PROJECT_ID_, PRINCIPAL_TYPE_, PRINCIPAL_ID_, ROLE_, INHERIT_TO_CHILDREN_, CREATED_BY_, CREATED_AT_
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, newID("perm"), input.ProjectID, normalizePrincipalType(input.PrincipalType), input.PrincipalID,
			normalizeProjectRole(input.Role), boolInt(inherit), nullIfEmpty(actor), now)
		return err
	})
}

func (s *Store) UpdateProjectPermission(ctx context.Context, id string, input kanban.ProjectPermissionUpdateInput, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.projectPermission.updated", "", "", actor, input, func(tx *sql.Tx, revision int64) error {
		result, err := tx.ExecContext(ctx, `
			UPDATE project_permission
			SET ROLE_ = CASE WHEN ? THEN ? ELSE ROLE_ END,
				INHERIT_TO_CHILDREN_ = CASE WHEN ? THEN ? ELSE INHERIT_TO_CHILDREN_ END
			WHERE ID_ = ? AND DELETED_AT_ IS NULL
		`, boolInt(input.Role != nil), normalizeProjectRole(trimmedValue(input.Role)),
			boolInt(input.InheritToChildren != nil), boolInt(input.InheritToChildren != nil && *input.InheritToChildren), id)
		return ensureAffected(result, err)
	})
}

func (s *Store) RevokeProjectPermission(ctx context.Context, id string, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.projectPermission.revoked", "", "", actor, map[string]string{"id": id}, func(tx *sql.Tx, revision int64) error {
		result, err := tx.ExecContext(ctx, `UPDATE project_permission SET DELETED_AT_ = ? WHERE ID_ = ? AND DELETED_AT_ IS NULL`, nowText(), id)
		return ensureAffected(result, err)
	})
}

func (s *Store) CreateIssueLabel(ctx context.Context, input kanban.IssueLabelInput, actor string) (int64, error) {
	projectID := kanban.DefaultProjectID
	if input.ProjectID != nil && strings.TrimSpace(*input.ProjectID) != "" {
		projectID = strings.TrimSpace(*input.ProjectID)
	}
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.issueLabel.created", projectID, "", actor, input, func(tx *sql.Tx, revision int64) error {
		now := nowText()
		key := normalizedSlug(input.Key, input.Name)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO issue_label (ID_, PROJECT_ID_, KEY_, NAME_, COLOR_, CREATED_BY_, UPDATED_BY_, CREATED_AT_, UPDATED_AT_)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, newID("label"), projectID, key, strings.TrimSpace(input.Name), trimmedValue(input.Color), nullIfEmpty(actor), nullIfEmpty(actor), now, now)
		return err
	})
}

func (s *Store) UpdateIssueLabel(ctx context.Context, id string, input kanban.IssueLabelUpdateInput, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.issueLabel.updated", "", "", actor, input, func(tx *sql.Tx, revision int64) error {
		result, err := tx.ExecContext(ctx, `
			UPDATE issue_label
			SET KEY_ = CASE WHEN ? THEN ? ELSE KEY_ END,
				NAME_ = CASE WHEN ? THEN ? ELSE NAME_ END,
				COLOR_ = CASE WHEN ? THEN ? ELSE COLOR_ END,
				UPDATED_BY_ = ?,
				UPDATED_AT_ = ?
			WHERE ID_ = ? AND DELETED_AT_ IS NULL
		`, boolInt(input.Key != nil), normalizedSlug(input.Key, ""), boolInt(input.Name != nil), trimmedValue(input.Name),
			boolInt(input.Color != nil), trimmedValue(input.Color), nullIfEmpty(actor), nowText(), id)
		return ensureAffected(result, err)
	})
}

func (s *Store) DeleteIssueLabel(ctx context.Context, id string, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.issueLabel.deleted", "", "", actor, map[string]string{"id": id}, func(tx *sql.Tx, revision int64) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM issue_label_link WHERE LABEL_ID_ = ?`, id); err != nil {
			return err
		}
		result, err := tx.ExecContext(ctx, `
			UPDATE issue_label SET DELETED_AT_ = ?, UPDATED_BY_ = ?, UPDATED_AT_ = ? WHERE ID_ = ? AND DELETED_AT_ IS NULL
		`, nowText(), nullIfEmpty(actor), nowText(), id)
		return ensureAffected(result, err)
	})
}

func (s *Store) SetIssueLabels(ctx context.Context, input kanban.IssueLabelsSetInput, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.issue.labels.updated", "", input.IssueID, actor, input, func(tx *sql.Tx, revision int64) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM issue_label_link WHERE ISSUE_ID_ = ?`, input.IssueID); err != nil {
			return err
		}
		seen := map[string]bool{}
		for _, labelID := range input.LabelIDs {
			labelID = strings.TrimSpace(labelID)
			if labelID == "" || seen[labelID] {
				continue
			}
			seen[labelID] = true
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO issue_label_link (ISSUE_ID_, LABEL_ID_) VALUES (?, ?)
			`, input.IssueID, labelID); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) CreateIssueDependency(ctx context.Context, input kanban.IssueDependencyInput, actor string) (int64, error) {
	dependencyType := strings.TrimSpace(trimmedValue(input.Type))
	if dependencyType == "" {
		dependencyType = "blocks"
	}
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.issueDependency.created", "", input.FromIssueID, actor, input, func(tx *sql.Tx, revision int64) error {
		var count int
		if err := tx.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM issue_dependency
			WHERE FROM_ISSUE_ID_ = ? AND TO_ISSUE_ID_ = ? AND TYPE_ = ? AND DELETED_AT_ IS NULL
		`, input.FromIssueID, input.ToIssueID, dependencyType).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO issue_dependency (ID_, FROM_ISSUE_ID_, TO_ISSUE_ID_, TYPE_, CREATED_BY_, CREATED_AT_)
			VALUES (?, ?, ?, ?, ?, ?)
		`, newID("dep"), input.FromIssueID, input.ToIssueID, dependencyType, nullIfEmpty(actor), nowText())
		return err
	})
}

func (s *Store) DeleteIssueDependency(ctx context.Context, id string, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.issueDependency.deleted", "", "", actor, map[string]string{"id": id}, func(tx *sql.Tx, revision int64) error {
		result, err := tx.ExecContext(ctx, `UPDATE issue_dependency SET DELETED_AT_ = ? WHERE ID_ = ? AND DELETED_AT_ IS NULL`, nowText(), id)
		return ensureAffected(result, err)
	})
}

func (s *Store) CreateReview(ctx context.Context, input kanban.ReviewInput, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.review.created", "", input.IssueID, actor, input, func(tx *sql.Tx, revision int64) error {
		now := nowText()
		reviewType := strings.TrimSpace(trimmedValue(input.ReviewType))
		if reviewType == "" {
			reviewType = "peer"
		}
		status := normalizeReviewStatus(input.Status)
		reviewID := newID("review")
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO review (
				ID_, ISSUE_ID_, AGENT_RUN_ID_, REVIEW_TYPE_, REVIEWER_ID_, STATUS_,
				REQUESTED_BY_, REQUESTED_AT_, SUMMARY_, CREATED_AT_, UPDATED_AT_
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, reviewID, input.IssueID, trimmedPtr(input.AgentRunID), reviewType, trimmedPtr(input.ReviewerID), status,
			nullIfEmpty(actor), now, trimmedValue(input.Summary), now, now); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `
			UPDATE issue SET ACTIVE_REVIEW_ID_ = ?, REVIEW_REQUIRED_ = 1, UPDATED_AT_ = ?, REVISION_ = ?
			WHERE ID_ = ? AND DELETED_AT_ IS NULL
		`, reviewID, now, revision, input.IssueID)
		return err
	})
}

func (s *Store) UpdateReview(ctx context.Context, id string, input kanban.ReviewUpdateInput, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.review.updated", "", "", actor, input, func(tx *sql.Tx, revision int64) error {
		status := normalizeReviewStatus(input.Status)
		submittedAt := any(nil)
		if input.SubmittedAt != nil {
			submittedAt = normalizedTimeValue(*input.SubmittedAt)
		} else if input.Status != nil && status != "pending" {
			submittedAt = nowText()
		}
		result, err := tx.ExecContext(ctx, `
			UPDATE review
			SET REVIEWER_ID_ = CASE WHEN ? THEN ? ELSE REVIEWER_ID_ END,
				STATUS_ = CASE WHEN ? THEN ? ELSE STATUS_ END,
				SUBMITTED_AT_ = CASE WHEN ? THEN ? ELSE SUBMITTED_AT_ END,
				SUMMARY_ = CASE WHEN ? THEN ? ELSE SUMMARY_ END,
				UPDATED_AT_ = ?
			WHERE ID_ = ? AND DELETED_AT_ IS NULL
		`, boolInt(input.ReviewerID != nil), trimmedPtr(input.ReviewerID), boolInt(input.Status != nil), status,
			boolInt(submittedAt != nil), submittedAt, boolInt(input.Summary != nil), trimmedValue(input.Summary), nowText(), id)
		if err := ensureAffected(result, err); err != nil {
			return err
		}
		if input.Status != nil && status != "pending" {
			_, err = tx.ExecContext(ctx, `
				UPDATE issue
				SET ACTIVE_REVIEW_ID_ = NULL,
					REVIEW_REQUIRED_ = CASE WHEN ? = 'approved' THEN 0 ELSE REVIEW_REQUIRED_ END,
					UPDATED_AT_ = ?,
					REVISION_ = ?
				WHERE ACTIVE_REVIEW_ID_ = ?
			`, status, nowText(), revision, id)
		}
		return err
	})
}

func (s *Store) DeleteReview(ctx context.Context, id string, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.review.deleted", "", "", actor, map[string]string{"id": id}, func(tx *sql.Tx, revision int64) error {
		now := nowText()
		if _, err := tx.ExecContext(ctx, `UPDATE review_comment SET DELETED_AT_ = ?, UPDATED_AT_ = ? WHERE REVIEW_ID_ = ? AND DELETED_AT_ IS NULL`, now, now, id); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE issue SET ACTIVE_REVIEW_ID_ = NULL, UPDATED_AT_ = ?, REVISION_ = ? WHERE ACTIVE_REVIEW_ID_ = ?`, now, revision, id); err != nil {
			return err
		}
		result, err := tx.ExecContext(ctx, `UPDATE review SET DELETED_AT_ = ?, UPDATED_AT_ = ? WHERE ID_ = ? AND DELETED_AT_ IS NULL`, now, now, id)
		return ensureAffected(result, err)
	})
}

func (s *Store) CreateReviewComment(ctx context.Context, input kanban.ReviewCommentInput, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.reviewComment.created", "", input.IssueID, actor, input, func(tx *sql.Tx, revision int64) error {
		anchorJSON, _ := json.Marshal(input.Anchor)
		if len(anchorJSON) == 0 || string(anchorJSON) == "null" {
			anchorJSON = []byte("{}")
		}
		now := nowText()
		_, err := tx.ExecContext(ctx, `
			INSERT INTO review_comment (
				ID_, REVIEW_ID_, ISSUE_ID_, AUTHOR_ID_, AUTHOR_AGENT_, BODY_, ANCHOR_JSON_, CREATED_AT_, UPDATED_AT_
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, newID("comment"), input.ReviewID, input.IssueID, trimmedPtr(input.AuthorID), trimmedPtr(input.AuthorAgent),
			strings.TrimSpace(input.Body), string(anchorJSON), now, now)
		return err
	})
}

func (s *Store) UpdateReviewComment(ctx context.Context, id string, input kanban.ReviewCommentUpdateInput, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.reviewComment.updated", "", "", actor, input, func(tx *sql.Tx, revision int64) error {
		anchorJSON := "{}"
		if input.Anchor != nil {
			data, _ := json.Marshal(input.Anchor)
			anchorJSON = string(data)
		}
		result, err := tx.ExecContext(ctx, `
			UPDATE review_comment
			SET BODY_ = CASE WHEN ? THEN ? ELSE BODY_ END,
				ANCHOR_JSON_ = CASE WHEN ? THEN ? ELSE ANCHOR_JSON_ END,
				UPDATED_AT_ = ?
			WHERE ID_ = ? AND DELETED_AT_ IS NULL
		`, boolInt(input.Body != nil), trimmedValue(input.Body), boolInt(input.Anchor != nil), anchorJSON, nowText(), id)
		return ensureAffected(result, err)
	})
}

func (s *Store) DeleteReviewComment(ctx context.Context, id string, actor string) (int64, error) {
	return s.withRevision(ctx, kanban.DefaultBoardID, "kanban.reviewComment.deleted", "", "", actor, map[string]string{"id": id}, func(tx *sql.Tx, revision int64) error {
		result, err := tx.ExecContext(ctx, `UPDATE review_comment SET DELETED_AT_ = ?, UPDATED_AT_ = ? WHERE ID_ = ? AND DELETED_AT_ IS NULL`, nowText(), nowText(), id)
		return ensureAffected(result, err)
	})
}

func scanReview(scanner issueScanner) (kanban.Review, error) {
	var item kanban.Review
	var agentRunID, reviewerID, requestedBy, submittedAt, deletedAt sql.NullString
	var requestedAt, createdAt, updatedAt string
	err := scanner.Scan(&item.ID, &item.IssueID, &agentRunID, &item.ReviewType, &reviewerID, &item.Status, &requestedBy,
		&requestedAt, &submittedAt, &item.Summary, &createdAt, &updatedAt, &deletedAt)
	if err != nil {
		return item, err
	}
	item.AgentRunID = stringPtr(agentRunID)
	item.ReviewerID = stringPtr(reviewerID)
	item.RequestedBy = stringPtr(requestedBy)
	item.RequestedAt, err = parseTime(requestedAt)
	if err != nil {
		return item, err
	}
	item.SubmittedAt, err = parseOptionalTime(submittedAt)
	if err != nil {
		return item, err
	}
	item.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return item, err
	}
	item.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return item, err
	}
	item.DeletedAt, err = parseOptionalTime(deletedAt)
	return item, err
}

func scanReviewComment(scanner issueScanner) (kanban.ReviewComment, error) {
	var item kanban.ReviewComment
	var authorID, authorAgent, deletedAt sql.NullString
	var anchorJSON, createdAt, updatedAt string
	err := scanner.Scan(&item.ID, &item.ReviewID, &item.IssueID, &authorID, &authorAgent, &item.Body, &anchorJSON, &createdAt, &updatedAt, &deletedAt)
	if err != nil {
		return item, err
	}
	item.AuthorID = stringPtr(authorID)
	item.AuthorAgent = stringPtr(authorAgent)
	item.Anchor = map[string]any{}
	_ = json.Unmarshal([]byte(anchorJSON), &item.Anchor)
	item.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return item, err
	}
	item.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return item, err
	}
	item.DeletedAt, err = parseOptionalTime(deletedAt)
	return item, err
}

func scanAgentRun(scanner issueScanner) (kanban.AgentRun, error) {
	var item kanban.AgentRun
	var agentID, workerAgent, delegatedBy, chatID, runID, sessionID, errorMessage sql.NullString
	var confidence sql.NullFloat64
	var inputTokens, outputTokens, costMicros sql.NullInt64
	var startedAt, finishedAt sql.NullString
	var createdAt, updatedAt string
	err := scanner.Scan(&item.ID, &item.IssueID, &agentID, &workerAgent, &delegatedBy, &chatID, &runID, &sessionID,
		&item.Status, &confidence, &inputTokens, &outputTokens, &costMicros, &startedAt, &finishedAt, &errorMessage, &createdAt, &updatedAt)
	if err != nil {
		return item, err
	}
	item.AgentID = stringPtr(agentID)
	item.WorkerAgent = stringPtr(workerAgent)
	item.DelegatedBy = stringPtr(delegatedBy)
	item.ChatID = stringPtr(chatID)
	item.RunID = stringPtr(runID)
	item.SessionID = stringPtr(sessionID)
	item.ErrorMessage = stringPtr(errorMessage)
	item.Confidence = floatPtr(confidence)
	item.InputTokens = intPtr(inputTokens)
	item.OutputTokens = intPtr(outputTokens)
	item.CostMicros = intPtr(costMicros)
	var err2 error
	item.StartedAt, err2 = parseOptionalTime(startedAt)
	if err2 != nil {
		return item, err2
	}
	item.FinishedAt, err2 = parseOptionalTime(finishedAt)
	if err2 != nil {
		return item, err2
	}
	item.CreatedAt, err2 = parseTime(createdAt)
	if err2 != nil {
		return item, err2
	}
	item.UpdatedAt, err2 = parseTime(updatedAt)
	return item, err2
}

func scanAgentToolCall(scanner issueScanner) (kanban.AgentToolCall, error) {
	var item kanban.AgentToolCall
	var startedAt, finishedAt, errorMessage sql.NullString
	var inputJSON, outputJSON string
	err := scanner.Scan(&item.ID, &item.AgentRunID, &item.ToolName, &item.ToolType, &item.Status, &inputJSON, &outputJSON, &startedAt, &finishedAt, &errorMessage)
	if err != nil {
		return item, err
	}
	item.Input = map[string]any{}
	item.Output = map[string]any{}
	_ = json.Unmarshal([]byte(inputJSON), &item.Input)
	_ = json.Unmarshal([]byte(outputJSON), &item.Output)
	item.StartedAt, err = parseOptionalTime(startedAt)
	if err != nil {
		return item, err
	}
	item.FinishedAt, err = parseOptionalTime(finishedAt)
	if err != nil {
		return item, err
	}
	item.ErrorMessage = stringPtr(errorMessage)
	return item, nil
}

func parseTime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, value)
}

func nowText() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func trimmedValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func trimmedPtr(value *string) any {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	return strings.TrimSpace(*value)
}

func normalizedSlug(value *string, fallback string) string {
	raw := trimmedValue(value)
	if raw == "" {
		raw = fallback
	}
	normalized := kanban.ProjectSlugFromName(raw)
	if normalized == "" {
		return "item"
	}
	return normalized
}

func normalizeUserStatus(value *string) string {
	switch strings.ToLower(trimmedValue(value)) {
	case "inactive", "disabled":
		return "inactive"
	default:
		return "active"
	}
}

func normalizeTeamRole(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "owner", "admin":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "member"
	}
}

func normalizePrincipalType(value string) string {
	if strings.ToLower(strings.TrimSpace(value)) == "team" {
		return "team"
	}
	return "user"
}

func normalizeProjectRole(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "owner", "admin", "maintainer", "developer", "reviewer":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "viewer"
	}
}

func normalizeReviewStatus(value *string) string {
	switch strings.ToLower(trimmedValue(value)) {
	case "approved", "changes_requested", "rejected":
		return strings.ToLower(trimmedValue(value))
	default:
		return "pending"
	}
}

func normalizedTimeValue(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if _, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return value
	}
	return nowText()
}

func floatPtr(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}
	next := value.Float64
	return &next
}

func intPtr(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	next := value.Int64
	return &next
}

func ensureAffected(result sql.Result, err error) error {
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
}
