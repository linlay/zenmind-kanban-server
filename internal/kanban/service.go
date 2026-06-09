package kanban

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Repository interface {
	Path() string
	ListWorkflowCatalog(ctx context.Context) (WorkflowCatalog, error)
	ListProjects(ctx context.Context) ([]Project, error)
	ListIssues(ctx context.Context, boardID string, projectID string) ([]Issue, int64, error)
	ListUsers(ctx context.Context) ([]UserAccount, error)
	ListTeamMembers(ctx context.Context) ([]TeamMember, error)
	ListAgents(ctx context.Context) ([]Agent, error)
	ListIssueLabels(ctx context.Context) ([]IssueLabel, error)
	ListIssueLabelsForProjects(ctx context.Context, projectIDs []string) ([]IssueLabel, error)
	ListIssueLabelLinks(ctx context.Context) ([]IssueLabelLink, error)
	ListIssueLabelLinksForIssues(ctx context.Context, issueIDs []string) ([]IssueLabelLink, error)
	ListIssueDependencies(ctx context.Context) ([]IssueDependency, error)
	ListIssueDependenciesForIssues(ctx context.Context, issueIDs []string) ([]IssueDependency, error)
	ListReviews(ctx context.Context) ([]Review, error)
	ListReviewsForIssues(ctx context.Context, issueIDs []string) ([]Review, error)
	ListReviewComments(ctx context.Context) ([]ReviewComment, error)
	ListReviewCommentsForIssues(ctx context.Context, issueIDs []string) ([]ReviewComment, error)
	ListAgentRuns(ctx context.Context, issueID string) ([]AgentRun, error)
	ListAgentRunsForIssues(ctx context.Context, issueIDs []string) ([]AgentRun, error)
	ListAgentToolCalls(ctx context.Context, agentRunID string) ([]AgentToolCall, error)
	ListAgentToolCallsForRuns(ctx context.Context, agentRunIDs []string) ([]AgentToolCall, error)
	ListRecentEvents(ctx context.Context, boardID string, limit int) ([]EventLogItem, error)
	GetIssue(ctx context.Context, boardID string, issueID string) (*Issue, error)
	ReplaceIssue(ctx context.Context, boardID string, issue Issue, eventType string, actor string) (int64, error)
	SoftDeleteIssue(ctx context.Context, boardID string, issueID string, actor string) (int64, error)
	CreateProject(ctx context.Context, project Project, actor string) (int64, error)
	UpdateProject(ctx context.Context, projectID string, input ProjectUpdateInput, actor string) (int64, error)
	MoveProject(ctx context.Context, input ProjectMoveInput, actor string) (int64, error)
	CreateUser(ctx context.Context, input UserInput, actor string) (int64, error)
	UpdateUser(ctx context.Context, id string, input UserUpdateInput, actor string) (int64, error)
	DeleteUser(ctx context.Context, id string, actor string) (int64, error)
	CreateTeam(ctx context.Context, input TeamInput, actor string) (int64, error)
	UpdateTeam(ctx context.Context, id string, input TeamUpdateInput, actor string) (int64, error)
	DeleteTeam(ctx context.Context, id string, actor string) (int64, error)
	AddTeamMember(ctx context.Context, input TeamMemberInput, actor string) (int64, error)
	UpdateTeamMember(ctx context.Context, teamID string, userID string, input TeamMemberUpdateInput, actor string) (int64, error)
	RemoveTeamMember(ctx context.Context, teamID string, userID string, actor string) (int64, error)
	CreateAgent(ctx context.Context, input AgentInput, actor string) (int64, error)
	UpdateAgent(ctx context.Context, id string, input AgentUpdateInput, actor string) (int64, error)
	DeleteAgent(ctx context.Context, id string, actor string) (int64, error)
	GrantProjectPermission(ctx context.Context, input ProjectPermissionInput, actor string) (int64, error)
	UpdateProjectPermission(ctx context.Context, id string, input ProjectPermissionUpdateInput, actor string) (int64, error)
	RevokeProjectPermission(ctx context.Context, id string, actor string) (int64, error)
	CreateIssueLabel(ctx context.Context, input IssueLabelInput, actor string) (int64, error)
	UpdateIssueLabel(ctx context.Context, id string, input IssueLabelUpdateInput, actor string) (int64, error)
	DeleteIssueLabel(ctx context.Context, id string, actor string) (int64, error)
	SetIssueLabels(ctx context.Context, input IssueLabelsSetInput, actor string) (int64, error)
	CreateIssueDependency(ctx context.Context, input IssueDependencyInput, actor string) (int64, error)
	DeleteIssueDependency(ctx context.Context, id string, actor string) (int64, error)
	CreateReview(ctx context.Context, input ReviewInput, actor string) (int64, error)
	UpdateReview(ctx context.Context, id string, input ReviewUpdateInput, actor string) (int64, error)
	DeleteReview(ctx context.Context, id string, actor string) (int64, error)
	CreateReviewComment(ctx context.Context, input ReviewCommentInput, actor string) (int64, error)
	UpdateReviewComment(ctx context.Context, id string, input ReviewCommentUpdateInput, actor string) (int64, error)
	DeleteReviewComment(ctx context.Context, id string, actor string) (int64, error)
	UpsertWorkflow(ctx context.Context, wf Workflow) error
	SoftDeleteWorkflow(ctx context.Context, id string) error
}

type Service struct {
	repo Repository
}

const staleRunAfter = 30 * time.Minute

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func snapshotProjectIDs(projects []Project, projectID string) []string {
	if projectID == "" {
		projectID = DefaultProjectID
	}
	children := map[string][]Project{}
	for _, project := range projects {
		parentID := ""
		if project.ParentID != nil {
			parentID = *project.ParentID
		}
		children[parentID] = append(children[parentID], project)
	}
	ids := []string{}
	seen := map[string]bool{}
	add := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			return
		}
		seen[id] = true
		ids = append(ids, id)
	}
	add(DefaultProjectID)
	var visit func(id string)
	visit = func(id string) {
		add(id)
		for _, child := range children[id] {
			visit(child.ID)
		}
	}
	visit(projectID)
	return ids
}

func snapshotIssueIDs(issues []Issue) []string {
	ids := make([]string, 0, len(issues))
	seen := map[string]bool{}
	for _, issue := range issues {
		id := strings.TrimSpace(issue.ID)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	return ids
}

func snapshotAgentRunIDs(runs []AgentRun) []string {
	ids := make([]string, 0, len(runs))
	seen := map[string]bool{}
	for _, run := range runs {
		id := strings.TrimSpace(run.ID)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	return ids
}

func (s *Service) Snapshot(ctx context.Context, boardID string, projectID string, desktopStatus DesktopStatus) (ListResult, error) {
	boardID = normalizeBoardID(boardID)
	projectID = normalizeProjectID(projectID)
	catalog, err := s.repo.ListWorkflowCatalog(ctx)
	if err != nil {
		return ListResult{}, err
	}
	projects, err := s.repo.ListProjects(ctx)
	if err != nil {
		return ListResult{}, err
	}
	if !projectExists(projects, projectID) {
		projectID = DefaultProjectID
	}
	issues, revision, err := s.repo.ListIssues(ctx, boardID, projectID)
	if err != nil {
		return ListResult{}, err
	}
	if !desktopStatus.Online {
		expired, err := s.expireStaleRuns(ctx, boardID, projectID, catalog, issues)
		if err != nil {
			return ListResult{}, err
		}
		if expired {
			issues, revision, err = s.repo.ListIssues(ctx, boardID, projectID)
			if err != nil {
				return ListResult{}, err
			}
		}
	}
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return ListResult{}, err
	}
	teamMembers, err := s.repo.ListTeamMembers(ctx)
	if err != nil {
		return ListResult{}, err
	}
	agents, err := s.repo.ListAgents(ctx)
	if err != nil {
		return ListResult{}, err
	}
	projectIDs := snapshotProjectIDs(projects, projectID)
	issueIDs := snapshotIssueIDs(issues)
	issueLabels, err := s.repo.ListIssueLabelsForProjects(ctx, projectIDs)
	if err != nil {
		return ListResult{}, err
	}
	issueLabelLinks, err := s.repo.ListIssueLabelLinksForIssues(ctx, issueIDs)
	if err != nil {
		return ListResult{}, err
	}
	issueDependencies, err := s.repo.ListIssueDependenciesForIssues(ctx, issueIDs)
	if err != nil {
		return ListResult{}, err
	}
	reviews, err := s.repo.ListReviewsForIssues(ctx, issueIDs)
	if err != nil {
		return ListResult{}, err
	}
	reviewComments, err := s.repo.ListReviewCommentsForIssues(ctx, issueIDs)
	if err != nil {
		return ListResult{}, err
	}
	agentRuns, err := s.repo.ListAgentRunsForIssues(ctx, issueIDs)
	if err != nil {
		return ListResult{}, err
	}
	agentToolCalls, err := s.repo.ListAgentToolCallsForRuns(ctx, snapshotAgentRunIDs(agentRuns))
	if err != nil {
		return ListResult{}, err
	}
	recentEvents, err := s.repo.ListRecentEvents(ctx, boardID, 100)
	if err != nil {
		return ListResult{}, err
	}
	return ListResult{
		OK:                  true,
		Message:             "issue 看板已加载。",
		BoardID:             boardID,
		ProjectID:           projectID,
		Revision:            revision,
		Complete:            true,
		Scope:               "project",
		Projects:            projects,
		Issues:              issues,
		Users:               users,
		Workflows:           catalog.Workflows,
		WorkflowStageDefs:   catalog.WorkflowStageDefs,
		WorkflowStatusDefs:  catalog.WorkflowStatusDefs,
		WorkflowStages:      catalog.WorkflowStages,
		WorkflowStatuses:    catalog.WorkflowStatuses,
		WorkflowTransitions: catalog.WorkflowTransitions,
		Teams:               catalog.Teams,
		TeamMembers:         teamMembers,
		ProjectPermissions:  catalog.ProjectPermissions,
		IssueLabels:         issueLabels,
		IssueLabelLinks:     issueLabelLinks,
		IssueDependencies:   issueDependencies,
		Reviews:             reviews,
		ReviewComments:      reviewComments,
		Agents:              agents,
		AgentRuns:           agentRuns,
		AgentToolCalls:      agentToolCalls,
		RecentEvents:        recentEvents,
		DesktopStatus:       desktopStatus,
		StoragePath:         s.repo.Path(),
	}, nil
}

func (s *Service) expireStaleRuns(ctx context.Context, boardID string, projectID string, catalog WorkflowCatalog, issues []Issue) (bool, error) {
	now := time.Now().UTC()
	expired := false
	for _, issue := range issues {
		if issue.RunID == nil || strings.TrimSpace(*issue.RunID) == "" {
			continue
		}
		if issue.UpdatedAt.IsZero() || now.Sub(issue.UpdatedAt) < staleRunAfter {
			continue
		}
		next := issue
		next.RunID = nil
		failed := RunStateFailed
		next.RunState = &failed
		nextStatus := StatusTodo
		statusRef := resolveStatus(catalog, next.WorkflowID, nil, &nextStatus)
		next.Status = nextStatus
		next.StatusID = statusRef.ID
		next.StatusKey = statusRef.Key
		next.StatusName = statusRef.Name
		next.ColumnKey = statusRef.ColumnKey
		next.UpdatedAt = now
		systemActor := "system"
		next.UpdatedBy = &systemActor
		if _, err := s.repo.ReplaceIssue(ctx, boardID, next, "kanban.issue.run.stale", systemActor); err != nil {
			return false, err
		}
		expired = true
	}
	return expired, nil
}

func (s *Service) ListProjectIssues(ctx context.Context, boardID string, projectID string) (IssuesResult, error) {
	boardID = normalizeBoardID(boardID)
	projectID = normalizeProjectID(projectID)
	projects, err := s.repo.ListProjects(ctx)
	if err != nil {
		return IssuesResult{}, err
	}
	if !projectExists(projects, projectID) {
		projectID = DefaultProjectID
	}
	issues, revision, err := s.repo.ListIssues(ctx, boardID, projectID)
	if err != nil {
		return IssuesResult{}, err
	}
	return IssuesResult{
		OK:        true,
		Message:   "issue 列表已加载。",
		BoardID:   boardID,
		ProjectID: projectID,
		Revision:  revision,
		Count:     len(issues),
		Issues:    SortIssues(issues),
	}, nil
}

func (s *Service) CreateIssue(ctx context.Context, boardID string, input IssueInput, actor string) (ChangeResult, error) {
	boardID = normalizeBoardID(boardID)
	projectID := normalizeProjectID("")
	if input.ProjectID != nil {
		projectID = normalizeProjectID(*input.ProjectID)
	}
	projects, err := s.repo.ListProjects(ctx)
	if err != nil {
		return ChangeResult{}, err
	}
	if !projectExists(projects, projectID) {
		projectID = DefaultProjectID
	}
	catalog, err := s.repo.ListWorkflowCatalog(ctx)
	if err != nil {
		return ChangeResult{}, err
	}
	project := findProject(projects, projectID)
	workflowID := resolveWorkflowID(catalog, project, input.WorkflowID)
	stage := resolveStage(catalog, workflowID, input.StageID, nil)
	current, revision, err := s.repo.ListIssues(ctx, boardID, projectID)
	if err != nil {
		return ChangeResult{}, err
	}
	title := Trimmed(input.Title)
	if title == "" {
		return changeError(boardID, projectID, revision, current, "issue 标题不能为空。"), nil
	}
	status := StatusBacklog
	if input.Status != nil {
		if normalized, ok := NormalizeStatus(*input.Status); ok {
			status = normalized
		}
	}
	statusRef := resolveStatus(catalog, workflowID, input.StatusID, &status)
	status = Status(statusRef.Key)
	priority := PriorityMedium
	if input.Priority != nil {
		if normalized, ok := NormalizePriority(*input.Priority); ok {
			priority = normalized
		}
	}
	severity := SeverityMedium
	if input.Severity != nil {
		if normalized, ok := NormalizeSeverity(*input.Severity); ok {
			severity = normalized
		}
	}
	var runState *RunState
	if input.RunState != nil {
		normalized, ok := NormalizeRunState(*input.RunState)
		if !ok {
			return changeError(boardID, projectID, revision, current, "issue 运行状态无效。"), nil
		}
		runState = normalized
	}
	now := time.Now().UTC()
	issue := Issue{
		BoardID:            boardID,
		ProjectID:          projectID,
		WorkflowID:         workflowID,
		StageID:            stage.ID,
		StageKey:           stage.Key,
		StageName:          stage.Name,
		StatusID:           statusRef.ID,
		StatusKey:          statusRef.Key,
		StatusName:         statusRef.Name,
		ColumnKey:          statusRef.ColumnKey,
		ID:                 createIssueID(),
		Title:              title,
		Description:        NormalizeDescription(input.Description),
		Status:             status,
		Priority:           priority,
		Severity:           severity,
		AssigneeAgentKey:   NullableTrimmed(input.AssigneeAgentKey),
		AssigneeID:         NullableTrimmed(input.AssigneeID),
		WorkerType:         NullableTrimmed(input.WorkerType),
		WorkerID:           NullableTrimmed(input.WorkerID),
		WorkerAgent:        NullableTrimmed(firstString(input.WorkerAgent, input.AssigneeAgentKey)),
		ReviewerID:         NullableTrimmed(input.ReviewerID),
		ReviewRequired:     input.ReviewRequired != nil && *input.ReviewRequired || status == StatusInReview,
		Position:           NextPosition(current, status),
		RunState:           runState,
		AutomationID:       NullableTrimmed(input.AutomationID),
		AutomationEnabled:  input.AutomationEnabled != nil && *input.AutomationEnabled,
		AutomationCron:     NullableTrimmed(input.AutomationCron),
		AutomationMessage:  NullableTrimmed(input.AutomationMessage),
		AutomationTimezone: NullableTrimmed(input.AutomationTimezone),
		AttachmentChatID:   NullableTrimmed(input.AttachmentChatID),
		Attachments:        normalizeAttachments(input.Attachments),
		CreatedAt:          now,
		UpdatedAt:          now,
		CreatedBy:          actorPtr(actor),
		UpdatedBy:          actorPtr(actor),
	}
	if issue.WorkerAgent != nil && issue.WorkerType == nil {
		workerType := "agent"
		issue.WorkerType = &workerType
	}
	if issue.AssigneeAgentKey == nil {
		issue.AssigneeAgentKey = issue.WorkerAgent
	}
	nextRevision, err := s.repo.ReplaceIssue(ctx, boardID, issue, "kanban.issue.created", actor)
	if err != nil {
		return ChangeResult{}, err
	}
	issues, _, err := s.repo.ListIssues(ctx, boardID, projectID)
	if err != nil {
		return ChangeResult{}, err
	}
	issue.Revision = nextRevision
	return ChangeResult{
		OK:        true,
		Message:   "issue 已创建。",
		BoardID:   boardID,
		ProjectID: projectID,
		Revision:  nextRevision,
		Complete:  true,
		Scope:     "project",
		Issue:     findIssue(issues, issue.ID),
		Issues:    issues,
	}, nil
}

func (s *Service) UpdateIssue(ctx context.Context, boardID string, projectID string, issueID string, input IssueUpdateInput, actor string) (ChangeResult, error) {
	return s.updateIssue(ctx, boardID, projectID, issueID, input, actor, false, "issue 已更新。", "kanban.issue.updated")
}

func (s *Service) StartRun(ctx context.Context, boardID string, projectID string, issueID string, agentKey *string, result StartRunResult, actor string) (ChangeResult, error) {
	if !result.OK {
		return s.resultWithCurrentIssues(ctx, boardID, projectID, false, result.Message)
	}
	if result.ChatID == nil || strings.TrimSpace(*result.ChatID) == "" || result.RunID == nil || strings.TrimSpace(*result.RunID) == "" {
		return s.resultWithCurrentIssues(ctx, boardID, projectID, false, "desktop 未返回 chatId/runId。")
	}
	resolvedAgentKey := agentKey
	if resolvedAgentKey == nil {
		resolvedAgentKey = result.AgentKey
	}
	running := string(RunStateRunning)
	inProgress := string(StatusInProgress)
	workerType := "agent"
	input := IssueUpdateInput{
		Status:           &inProgress,
		AssigneeAgentKey: resolvedAgentKey,
		WorkerType:       &workerType,
		WorkerAgent:      resolvedAgentKey,
		ChatID:           result.ChatID,
		RunID:            result.RunID,
		RunState:         &running,
	}
	return s.updateIssue(ctx, boardID, projectID, issueID, input, actor, true, "issue 已分配给智能体。", "kanban.issue.updated")
}

func (s *Service) MoveIssue(ctx context.Context, boardID string, projectID string, input MoveInput, actor string) (ChangeResult, error) {
	boardID = normalizeBoardID(boardID)
	projectID = normalizeProjectID(projectID)
	current, revision, err := s.repo.ListIssues(ctx, boardID, projectID)
	if err != nil {
		return ChangeResult{}, err
	}
	issue := findIssue(current, input.ID)
	if issue == nil {
		return changeError(boardID, projectID, revision, current, "issue 不存在。"), nil
	}
	if conflict := revisionConflict(boardID, projectID, revision, current, issue, input.BaseIssueRevision); conflict != nil {
		return *conflict, nil
	}
	targetStatus, ok := NormalizeStatus(input.Status)
	if !ok || !ValidPosition(input.Position) {
		return changeError(boardID, projectID, revision, current, "issue 移动参数无效。"), nil
	}
	catalog, err := s.repo.ListWorkflowCatalog(ctx)
	if err != nil {
		return ChangeResult{}, err
	}
	targetStatusRef := resolveStatus(catalog, issue.WorkflowID, nil, &targetStatus)
	if issue.RunID != nil {
		return changeError(boardID, projectID, revision, current, "智能体正在回答，完成后才能切换状态。"), nil
	}
	next := *issue
	next.Status = targetStatus
	next.StatusID = targetStatusRef.ID
	next.StatusKey = targetStatusRef.Key
	next.StatusName = targetStatusRef.Name
	next.ColumnKey = targetStatusRef.ColumnKey
	next.ReviewRequired = targetStatus == StatusInReview || next.ReviewRequired
	next.Position = input.Position
	if targetStatus != issue.Status {
		next.RunState = nil
	}
	next.UpdatedAt = time.Now().UTC()
	next.UpdatedBy = actorPtr(actor)
	nextRevision, err := s.repo.ReplaceIssue(ctx, boardID, next, "kanban.issue.updated", actor)
	if err != nil {
		return ChangeResult{}, err
	}
	issues, _, err := s.repo.ListIssues(ctx, boardID, projectID)
	if err != nil {
		return ChangeResult{}, err
	}
	return ChangeResult{
		OK:        true,
		Message:   "issue 已移动。",
		BoardID:   boardID,
		ProjectID: projectID,
		Revision:  nextRevision,
		Complete:  true,
		Scope:     "project",
		Issue:     findIssue(issues, input.ID),
		Issues:    issues,
	}, nil
}

func (s *Service) DeleteIssue(ctx context.Context, boardID string, projectID string, issueID string, baseIssueRevision *int64, actor string) (ChangeResult, error) {
	boardID = normalizeBoardID(boardID)
	projectID = normalizeProjectID(projectID)
	current, revision, err := s.repo.ListIssues(ctx, boardID, projectID)
	if err != nil {
		return ChangeResult{}, err
	}
	issue := findIssue(current, issueID)
	if issue == nil {
		return changeError(boardID, projectID, revision, current, "issue 不存在。"), nil
	}
	if conflict := revisionConflict(boardID, projectID, revision, current, issue, baseIssueRevision); conflict != nil {
		return *conflict, nil
	}
	nextRevision, err := s.repo.SoftDeleteIssue(ctx, boardID, issueID, actor)
	if err != nil {
		return ChangeResult{}, err
	}
	issues, _, err := s.repo.ListIssues(ctx, boardID, projectID)
	if err != nil {
		return ChangeResult{}, err
	}
	return ChangeResult{
		OK:             true,
		Message:        "issue 已删除。",
		BoardID:        boardID,
		ProjectID:      projectID,
		Revision:       nextRevision,
		Complete:       true,
		Scope:          "project",
		DeletedIssueID: issueID,
		Issues:         issues,
	}, nil
}

func (s *Service) SyncAssistantEvent(ctx context.Context, boardID string, projectID string, event AssistantEvent, actor string) (ChangeResult, error) {
	boardID = normalizeBoardID(boardID)
	projectID = normalizeProjectID(projectID)
	current, revision, err := s.repo.ListIssues(ctx, boardID, projectID)
	if err != nil {
		return ChangeResult{}, err
	}
	runState := runStateFromAssistantEvent(event)
	if runState == nil {
		return changeError(boardID, projectID, revision, current, "无需同步的智能体事件。"), nil
	}
	var issue *Issue
	if event.RunID != nil {
		trimmed := strings.TrimSpace(*event.RunID)
		for i := range current {
			if current[i].RunID != nil && *current[i].RunID == trimmed {
				issue = &current[i]
				break
			}
		}
	}
	if issue == nil && event.ChatID != nil {
		trimmed := strings.TrimSpace(*event.ChatID)
		for i := range current {
			if current[i].ChatID != nil && *current[i].ChatID == trimmed && current[i].Status == StatusInProgress {
				issue = &current[i]
				break
			}
		}
	}
	if issue == nil {
		return changeError(boardID, projectID, revision, current, "task 运行不存在。"), nil
	}
	catalog, err := s.repo.ListWorkflowCatalog(ctx)
	if err != nil {
		return ChangeResult{}, err
	}
	next := *issue
	next.RunID = nil
	next.RunState = runState
	if event.ChatID != nil && strings.TrimSpace(*event.ChatID) != "" {
		next.ChatID = event.ChatID
	}
	if *runState == RunStateCompleted {
		nextStatus := StatusCompleted
		if next.ReviewRequired || next.ReviewerID != nil {
			nextStatus = StatusInReview
		}
		statusRef := resolveStatus(catalog, next.WorkflowID, nil, &nextStatus)
		next.Status = nextStatus
		next.StatusID = statusRef.ID
		next.StatusKey = statusRef.Key
		next.StatusName = statusRef.Name
		next.ColumnKey = statusRef.ColumnKey
		next.ReviewRequired = next.ReviewRequired || next.Status == StatusInReview
	} else if *runState == RunStateFailed || *runState == RunStateCancelled {
		nextStatus := StatusTodo
		statusRef := resolveStatus(catalog, next.WorkflowID, nil, &nextStatus)
		next.Status = nextStatus
		next.StatusID = statusRef.ID
		next.StatusKey = statusRef.Key
		next.StatusName = statusRef.Name
		next.ColumnKey = statusRef.ColumnKey
	}
	next.UpdatedAt = time.Now().UTC()
	next.UpdatedBy = actorPtr(actor)
	nextRevision, err := s.repo.ReplaceIssue(ctx, boardID, next, "kanban.issue.updated", actor)
	if err != nil {
		return ChangeResult{}, err
	}
	issues, _, err := s.repo.ListIssues(ctx, boardID, projectID)
	if err != nil {
		return ChangeResult{}, err
	}
	return ChangeResult{
		OK:        true,
		Message:   "task 运行状态已更新。",
		BoardID:   boardID,
		ProjectID: projectID,
		Revision:  nextRevision,
		Complete:  true,
		Scope:     "project",
		Issue:     findIssue(issues, issue.ID),
		Issues:    issues,
	}, nil
}

func (s *Service) updateIssue(
	ctx context.Context,
	boardID string,
	projectID string,
	issueID string,
	input IssueUpdateInput,
	actor string,
	allowCompletedTransition bool,
	message string,
	eventType string,
) (ChangeResult, error) {
	boardID = normalizeBoardID(boardID)
	projectID = normalizeProjectID(projectID)
	current, revision, err := s.repo.ListIssues(ctx, boardID, projectID)
	if err != nil {
		return ChangeResult{}, err
	}
	issue := findIssue(current, issueID)
	if issue == nil {
		return changeError(boardID, projectID, revision, current, "issue 不存在。"), nil
	}
	if conflict := revisionConflict(boardID, projectID, revision, current, issue, input.BaseIssueRevision); conflict != nil {
		return *conflict, nil
	}
	next := *issue
	next.Attachments = append([]Attachment(nil), issue.Attachments...)
	next.UpdatedAt = time.Now().UTC()
	next.UpdatedBy = actorPtr(actor)

	catalog, err := s.repo.ListWorkflowCatalog(ctx)
	if err != nil {
		return ChangeResult{}, err
	}
	targetWorkflowID := next.WorkflowID
	if targetWorkflowID == "" {
		targetWorkflowID = DefaultWorkflowID
	}
	if input.WorkflowID != nil && workflowExists(catalog, *input.WorkflowID) {
		targetWorkflowID = strings.TrimSpace(*input.WorkflowID)
	}
	var requestedStatus *Status
	if input.Status != nil {
		status, ok := NormalizeStatus(*input.Status)
		if ok {
			requestedStatus = &status
		}
	}
	if input.StatusID != nil {
		statusRef := resolveStatus(catalog, targetWorkflowID, input.StatusID, nil)
		status := Status(statusRef.Key)
		requestedStatus = &status
	}
	clearsActiveRun := input.RunID != nil && strings.TrimSpace(*input.RunID) == ""
	if issue.RunID != nil && requestedStatus != nil && *requestedStatus != issue.Status && !clearsActiveRun && !allowCompletedTransition {
		return changeError(boardID, projectID, revision, current, "智能体正在回答，完成后才能切换状态。"), nil
	}
	if requestedStatus != nil && *requestedStatus == StatusCompleted && issue.Status != StatusCompleted && !clearsActiveRun && !allowCompletedTransition {
		return changeError(boardID, projectID, revision, current, nonDragCompletedTransitionMessage), nil
	}

	if input.Title != nil {
		title := Trimmed(*input.Title)
		if title == "" {
			return changeError(boardID, projectID, revision, current, "issue 标题不能为空。"), nil
		}
		next.Title = title
	}
	if input.ProjectID != nil {
		nextProjectID := normalizeProjectID(*input.ProjectID)
		projects, err := s.repo.ListProjects(ctx)
		if err != nil {
			return ChangeResult{}, err
		}
		if !projectExists(projects, nextProjectID) {
			return changeError(boardID, projectID, revision, current, "项目不存在。"), nil
		}
		if nextProjectID != next.ProjectID {
			targetIssues, _, err := s.repo.ListIssues(ctx, boardID, nextProjectID)
			if err != nil {
				return ChangeResult{}, err
			}
			next.ProjectID = nextProjectID
			next.Position = NextPosition(targetIssues, next.Status)
		}
	}
	if input.Description != nil {
		next.Description = NormalizeDescription(input.Description)
	}
	if next.WorkflowID != targetWorkflowID {
		next.WorkflowID = targetWorkflowID
		if input.StageID == nil {
			next.StageID = ""
		}
	}
	stage := resolveStage(catalog, next.WorkflowID, input.StageID, &next.StageID)
	next.StageID = stage.ID
	next.StageKey = stage.Key
	next.StageName = stage.Name
	if requestedStatus != nil {
		next.Status = *requestedStatus
		statusRef := resolveStatus(catalog, next.WorkflowID, input.StatusID, requestedStatus)
		next.StatusID = statusRef.ID
		next.StatusKey = statusRef.Key
		next.StatusName = statusRef.Name
		next.ColumnKey = statusRef.ColumnKey
		next.ReviewRequired = next.ReviewRequired || next.Status == StatusInReview
	} else {
		statusRef := resolveStatus(catalog, next.WorkflowID, &next.StatusID, &next.Status)
		next.StatusID = statusRef.ID
		next.StatusKey = statusRef.Key
		next.StatusName = statusRef.Name
		next.ColumnKey = statusRef.ColumnKey
	}
	if input.Priority != nil {
		if priority, ok := NormalizePriority(*input.Priority); ok {
			next.Priority = priority
		}
	}
	if input.Severity != nil {
		if severity, ok := NormalizeSeverity(*input.Severity); ok {
			next.Severity = severity
		}
	}
	if input.AssigneeID != nil {
		next.AssigneeID = NullableTrimmed(input.AssigneeID)
	}
	if input.AssigneeAgentKey != nil {
		next.AssigneeAgentKey = NullableTrimmed(input.AssigneeAgentKey)
		next.WorkerAgent = next.AssigneeAgentKey
		if next.WorkerAgent != nil {
			workerType := "agent"
			next.WorkerType = &workerType
			next.WorkerID = nil
		}
	}
	if input.WorkerType != nil {
		next.WorkerType = NullableTrimmed(input.WorkerType)
	}
	if input.WorkerID != nil {
		next.WorkerID = NullableTrimmed(input.WorkerID)
		if next.WorkerID != nil {
			workerType := "human"
			next.WorkerType = &workerType
			next.WorkerAgent = nil
			next.AssigneeAgentKey = nil
		}
	}
	if input.WorkerAgent != nil {
		next.WorkerAgent = NullableTrimmed(input.WorkerAgent)
		next.AssigneeAgentKey = next.WorkerAgent
		if next.WorkerAgent != nil {
			workerType := "agent"
			next.WorkerType = &workerType
			next.WorkerID = nil
		}
	}
	if input.ReviewerID != nil {
		next.ReviewerID = NullableTrimmed(input.ReviewerID)
	}
	if input.ReviewRequired != nil {
		next.ReviewRequired = *input.ReviewRequired
	}
	if input.ChatID != nil {
		next.ChatID = NullableTrimmed(input.ChatID)
	}
	if input.RunID != nil {
		next.RunID = NullableTrimmed(input.RunID)
	}
	if input.RunState != nil {
		runState, ok := NormalizeRunState(*input.RunState)
		if !ok {
			return changeError(boardID, projectID, revision, current, "issue 运行状态无效。"), nil
		}
		next.RunState = runState
	} else if input.RunID != nil {
		if next.RunID != nil {
			state := RunStateRunning
			next.RunState = &state
		} else if next.Status == StatusCompleted {
			state := RunStateCompleted
			next.RunState = &state
		} else if issue.RunID != nil && next.Status == StatusTodo {
			state := RunStateFailed
			next.RunState = &state
		}
	} else if input.Status != nil && next.Status != issue.Status && next.RunID == nil {
		next.RunState = nil
	}
	if input.AutomationID != nil {
		next.AutomationID = NullableTrimmed(input.AutomationID)
	}
	if input.AutomationEnabled != nil {
		next.AutomationEnabled = *input.AutomationEnabled
	}
	if input.AutomationCron != nil {
		next.AutomationCron = NullableTrimmed(input.AutomationCron)
	}
	if input.AutomationMessage != nil {
		next.AutomationMessage = NullableTrimmed(input.AutomationMessage)
	}
	if input.AutomationTimezone != nil {
		next.AutomationTimezone = NullableTrimmed(input.AutomationTimezone)
	}
	if input.AttachmentChatID != nil {
		next.AttachmentChatID = NullableTrimmed(input.AttachmentChatID)
	}
	if input.Attachments != nil {
		next.Attachments = normalizeAttachments(input.Attachments)
	}

	nextRevision, err := s.repo.ReplaceIssue(ctx, boardID, next, eventType, actor)
	if err != nil {
		return ChangeResult{}, err
	}
	issues, _, err := s.repo.ListIssues(ctx, boardID, projectID)
	if err != nil {
		return ChangeResult{}, err
	}
	return ChangeResult{
		OK:        true,
		Message:   message,
		BoardID:   boardID,
		ProjectID: projectID,
		Revision:  nextRevision,
		Complete:  true,
		Scope:     "project",
		Issue:     findIssue(issues, issueID),
		Issues:    issues,
	}, nil
}

func (s *Service) CreateProject(ctx context.Context, input ProjectInput, actor string) (ProjectChangeResult, error) {
	parentID := DefaultProjectID
	if input.ParentID != nil {
		parentID = normalizeProjectID(*input.ParentID)
	}
	projects, err := s.repo.ListProjects(ctx)
	if err != nil {
		return ProjectChangeResult{}, err
	}
	if !projectExists(projects, parentID) {
		return projectChangeError(0, projects, parentID, "父项目不存在。"), nil
	}
	name := NormalizeProjectName(input.Name)
	if name == "" {
		return projectChangeError(0, projects, parentID, "项目名称不能为空。"), nil
	}
	slug := ProjectSlugFromName(name)
	if input.Slug != nil {
		if normalized := NormalizeProjectSlug(*input.Slug); normalized != "" {
			slug = normalized
		}
	}
	visibility := normalizeProjectVisibility(input.Visibility)
	if visibility == "" {
		return projectChangeError(0, projects, parentID, "项目可见性无效。"), nil
	}
	position := float64(len(projects) + 1)
	if input.Position != nil && ValidPosition(*input.Position) {
		position = *input.Position
	}
	catalog, err := s.repo.ListWorkflowCatalog(ctx)
	if err != nil {
		return ProjectChangeResult{}, err
	}
	defaultWorkflowID := ""
	if input.DefaultWorkflowID != nil && workflowExists(catalog, *input.DefaultWorkflowID) {
		defaultWorkflowID = strings.TrimSpace(*input.DefaultWorkflowID)
	}
	now := time.Now().UTC()
	project := Project{
		ID:                createProjectID(),
		ParentID:          &parentID,
		Slug:              slug,
		Name:              name,
		Description:       NormalizeDescription(input.Description),
		Visibility:        visibility,
		DefaultWorkflowID: defaultWorkflowID,
		Position:          position,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	revision, err := s.repo.CreateProject(ctx, project, actor)
	if err != nil {
		return ProjectChangeResult{}, err
	}
	projects, err = s.repo.ListProjects(ctx)
	if err != nil {
		return ProjectChangeResult{}, err
	}
	return ProjectChangeResult{
		OK:        true,
		Message:   "项目已创建。",
		BoardID:   DefaultBoardID,
		ProjectID: project.ID,
		Revision:  revision,
		Project:   findProject(projects, project.ID),
		Projects:  projects,
	}, nil
}

func (s *Service) UpdateProject(ctx context.Context, projectID string, input ProjectUpdateInput, actor string) (ProjectChangeResult, error) {
	projectID = normalizeProjectID(projectID)
	projects, err := s.repo.ListProjects(ctx)
	if err != nil {
		return ProjectChangeResult{}, err
	}
	if !projectExists(projects, projectID) {
		return projectChangeError(0, projects, projectID, "项目不存在。"), nil
	}
	if projectID == DefaultProjectID {
		return projectChangeError(0, projects, projectID, "默认项目不能重命名。"), nil
	}
	if input.Name != nil && NormalizeProjectName(*input.Name) == "" {
		return projectChangeError(0, projects, projectID, "项目名称不能为空。"), nil
	}
	if input.Slug != nil && NormalizeProjectSlug(*input.Slug) == "" {
		return projectChangeError(0, projects, projectID, "项目 slug 不能为空。"), nil
	}
	if input.Visibility != nil && normalizeProjectVisibility(input.Visibility) == "" {
		return projectChangeError(0, projects, projectID, "项目可见性无效。"), nil
	}
	if input.DefaultWorkflowID != nil {
		catalog, err := s.repo.ListWorkflowCatalog(ctx)
		if err != nil {
			return ProjectChangeResult{}, err
		}
		if !workflowExists(catalog, *input.DefaultWorkflowID) {
			return projectChangeError(0, projects, projectID, "项目 workflow 不存在。"), nil
		}
	}
	revision, err := s.repo.UpdateProject(ctx, projectID, input, actor)
	if err != nil {
		return ProjectChangeResult{}, err
	}
	projects, err = s.repo.ListProjects(ctx)
	if err != nil {
		return ProjectChangeResult{}, err
	}
	return ProjectChangeResult{
		OK:        true,
		Message:   "项目已更新。",
		BoardID:   DefaultBoardID,
		ProjectID: projectID,
		Revision:  revision,
		Project:   findProject(projects, projectID),
		Projects:  projects,
	}, nil
}

func (s *Service) MoveProject(ctx context.Context, input ProjectMoveInput, actor string) (ProjectChangeResult, error) {
	input.ID = normalizeProjectID(input.ID)
	if input.ParentID == nil || strings.TrimSpace(*input.ParentID) == "" {
		parentID := DefaultProjectID
		input.ParentID = &parentID
	} else {
		parentID := normalizeProjectID(*input.ParentID)
		input.ParentID = &parentID
	}
	projects, err := s.repo.ListProjects(ctx)
	if err != nil {
		return ProjectChangeResult{}, err
	}
	if input.ID == DefaultProjectID {
		return projectChangeError(0, projects, input.ID, "默认项目不能移动。"), nil
	}
	if !projectExists(projects, input.ID) || !projectExists(projects, *input.ParentID) {
		return projectChangeError(0, projects, input.ID, "项目不存在。"), nil
	}
	if input.Position != nil && !ValidPosition(*input.Position) {
		return projectChangeError(0, projects, input.ID, "项目位置无效。"), nil
	}
	revision, err := s.repo.MoveProject(ctx, input, actor)
	if err != nil {
		return ProjectChangeResult{}, err
	}
	projects, err = s.repo.ListProjects(ctx)
	if err != nil {
		return ProjectChangeResult{}, err
	}
	return ProjectChangeResult{
		OK:        true,
		Message:   "项目已移动。",
		BoardID:   DefaultBoardID,
		ProjectID: input.ID,
		Revision:  revision,
		Project:   findProject(projects, input.ID),
		Projects:  projects,
	}, nil
}

func (s *Service) resultWithCurrentIssues(ctx context.Context, boardID string, projectID string, ok bool, message string) (ChangeResult, error) {
	boardID = normalizeBoardID(boardID)
	projectID = normalizeProjectID(projectID)
	issues, revision, err := s.repo.ListIssues(ctx, boardID, projectID)
	if err != nil {
		return ChangeResult{}, err
	}
	if message == "" {
		message = "操作失败。"
	}
	return ChangeResult{OK: ok, Message: message, BoardID: boardID, ProjectID: projectID, Revision: revision, Complete: true, Scope: "project", Issues: issues}, nil
}

func changeError(boardID string, projectID string, revision int64, issues []Issue, message string) ChangeResult {
	return changeErrorCode(boardID, projectID, revision, issues, "", message)
}

func changeErrorCode(boardID string, projectID string, revision int64, issues []Issue, code string, message string) ChangeResult {
	return ChangeResult{
		OK:        false,
		Code:      code,
		Message:   message,
		BoardID:   boardID,
		ProjectID: projectID,
		Revision:  revision,
		Complete:  true,
		Scope:     "project",
		Issues:    SortIssues(issues),
	}
}

func revisionConflict(boardID string, projectID string, revision int64, issues []Issue, issue *Issue, baseIssueRevision *int64) *ChangeResult {
	if issue == nil || baseIssueRevision == nil || *baseIssueRevision <= 0 || *baseIssueRevision == issue.Revision {
		return nil
	}
	result := changeErrorCode(boardID, projectID, revision, issues, "conflict", "任务已被其他端更新，请刷新后重试。")
	return &result
}

func projectChangeError(revision int64, projects []Project, projectID string, message string) ProjectChangeResult {
	return ProjectChangeResult{
		OK:        false,
		Message:   message,
		BoardID:   DefaultBoardID,
		ProjectID: projectID,
		Revision:  revision,
		Projects:  projects,
	}
}

func findIssue(issues []Issue, issueID string) *Issue {
	for i := range issues {
		if issues[i].ID == issueID {
			return &issues[i]
		}
	}
	return nil
}

func findProject(projects []Project, projectID string) *Project {
	for i := range projects {
		if projects[i].ID == projectID {
			return &projects[i]
		}
	}
	return nil
}

func normalizeBoardID(boardID string) string {
	boardID = strings.TrimSpace(boardID)
	if boardID == "" {
		return DefaultBoardID
	}
	return boardID
}

func normalizeProjectID(projectID string) string {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return DefaultProjectID
	}
	return projectID
}

func normalizeProjectVisibility(value *string) string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return "workspace"
	}
	switch strings.ToLower(strings.TrimSpace(*value)) {
	case "private", "team", "workspace":
		return strings.ToLower(strings.TrimSpace(*value))
	default:
		return ""
	}
}

func projectExists(projects []Project, projectID string) bool {
	return findProject(projects, projectID) != nil
}

func resolveWorkflowID(catalog WorkflowCatalog, project *Project, workflowID *string) string {
	if workflowID != nil && workflowExists(catalog, *workflowID) {
		return strings.TrimSpace(*workflowID)
	}
	if project != nil && workflowExists(catalog, project.DefaultWorkflowID) {
		return project.DefaultWorkflowID
	}
	if workflowExists(catalog, DefaultWorkflowID) {
		return DefaultWorkflowID
	}
	for _, workflow := range catalog.Workflows {
		if workflow.IsDefault {
			return workflow.ID
		}
	}
	if len(catalog.Workflows) > 0 {
		return catalog.Workflows[0].ID
	}
	return DefaultWorkflowID
}

func resolveStage(catalog WorkflowCatalog, workflowID string, requestedID *string, currentID *string) WorkflowStage {
	if requestedID != nil {
		if stage := findWorkflowStage(catalog, workflowID, *requestedID); stage != nil {
			return *stage
		}
	}
	if currentID != nil {
		if stage := findWorkflowStage(catalog, workflowID, *currentID); stage != nil {
			return *stage
		}
	}
	for _, stage := range catalog.WorkflowStages {
		if stage.WorkflowID == workflowID && stage.IsStart {
			return stage
		}
	}
	for _, stage := range catalog.WorkflowStages {
		if stage.WorkflowID == workflowID {
			return stage
		}
	}
	return WorkflowStage{ID: "workflow-standard-requirement-stage-requirement_clarification", WorkflowID: workflowID, Key: "requirement_clarification", Name: "需求澄清", IsStart: true}
}

func resolveStatus(catalog WorkflowCatalog, workflowID string, requestedID *string, requestedStatus *Status) WorkflowStatus {
	if requestedID != nil {
		if status := findWorkflowStatus(catalog, workflowID, *requestedID); status != nil {
			return *status
		}
	}
	if requestedStatus != nil {
		for _, status := range catalog.WorkflowStatuses {
			if status.WorkflowID == workflowID && status.Key == string(*requestedStatus) {
				return status
			}
		}
	}
	for _, status := range catalog.WorkflowStatuses {
		if status.WorkflowID == workflowID && status.IsStart {
			return status
		}
	}
	for _, status := range catalog.WorkflowStatuses {
		if status.WorkflowID == workflowID {
			return status
		}
	}
	return WorkflowStatus{ID: "workflow-standard-requirement-status-backlog", WorkflowID: workflowID, Key: string(StatusBacklog), Name: "新建", ColumnKey: string(StatusBacklog), IsStart: true}
}

func workflowExists(catalog WorkflowCatalog, workflowID string) bool {
	workflowID = strings.TrimSpace(workflowID)
	for _, workflow := range catalog.Workflows {
		if workflow.ID == workflowID {
			return true
		}
	}
	return false
}

func findWorkflowStage(catalog WorkflowCatalog, workflowID string, stageID string) *WorkflowStage {
	stageID = strings.TrimSpace(stageID)
	for i := range catalog.WorkflowStages {
		stage := &catalog.WorkflowStages[i]
		if stage.WorkflowID == workflowID && (stage.ID == stageID || stage.Key == stageID) {
			return stage
		}
	}
	return nil
}

func findWorkflowStatus(catalog WorkflowCatalog, workflowID string, statusID string) *WorkflowStatus {
	statusID = strings.TrimSpace(statusID)
	for i := range catalog.WorkflowStatuses {
		status := &catalog.WorkflowStatuses[i]
		if status.WorkflowID == workflowID && (status.ID == statusID || status.Key == statusID) {
			return status
		}
	}
	return nil
}

func firstString(values ...*string) *string {
	for _, value := range values {
		if value != nil && strings.TrimSpace(*value) != "" {
			return value
		}
	}
	return nil
}

func normalizeAttachments(attachments []Attachment) []Attachment {
	if attachments == nil {
		return []Attachment{}
	}
	return attachments
}

func actorPtr(actor string) *string {
	actor = strings.TrimSpace(actor)
	if actor == "" {
		return nil
	}
	return &actor
}

var (
	lastIssueTickMu sync.Mutex
	lastIssueTick   int64
)

func createIssueID() string {
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

func createProjectID() string {
	var random [4]byte
	if _, err := rand.Read(random[:]); err != nil {
		return strings.ToUpper(hex.EncodeToString([]byte(time.Now().Format("150405.000000000"))))
	}
	return strings.ToUpper(strings.TrimLeft(time.Now().UTC().Format("20060102150405"), "0") + hex.EncodeToString(random[:]))
}

func runStateFromAssistantEvent(event AssistantEvent) *RunState {
	if event.Type == "done" || event.Type == "run.complete" {
		state := RunStateCompleted
		return &state
	}
	status := ""
	if event.Status != nil {
		status = strings.ToLower(strings.TrimSpace(*event.Status))
	}
	if event.Type == "run.cancel" ||
		event.Type == "run.cancelled" ||
		event.Type == "run.canceled" ||
		event.Type == "task.cancel" ||
		event.Type == "task.cancelled" ||
		event.Type == "task.canceled" ||
		event.Type == "stopped" ||
		event.Type == "run.stopped" ||
		event.Type == "run.interrupt" ||
		status == "cancelled" ||
		status == "canceled" ||
		status == "stopped" {
		state := RunStateCancelled
		return &state
	}
	if event.Type == "error" ||
		event.Type == "run.error" ||
		event.Type == "run.fail" ||
		event.Type == "run.failed" ||
		event.Type == "task.fail" ||
		event.Type == "task.failed" ||
		event.Type == "run.expired" ||
		status == "error" ||
		status == "failed" ||
		status == "fail" ||
		status == "timeout" {
		state := RunStateFailed
		return &state
	}
	return nil
}

var ErrDesktopUnavailable = errors.New("desktop client unavailable")
