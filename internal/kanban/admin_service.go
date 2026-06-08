package kanban

import (
	"context"
	"strings"
	"time"
)

func (s *Service) CreateUser(ctx context.Context, input UserInput, actor string) (MutationResult, error) {
	if strings.TrimSpace(input.Email) == "" || strings.TrimSpace(input.DisplayName) == "" {
		return mutationError("用户邮箱和名称不能为空。"), nil
	}
	revision, err := s.repo.CreateUser(ctx, input, actor)
	return mutationResult(revision, "用户已创建。"), err
}

func (s *Service) UpdateUser(ctx context.Context, id string, input UserUpdateInput, actor string) (MutationResult, error) {
	if strings.TrimSpace(id) == "" {
		return mutationError("用户不存在。"), nil
	}
	revision, err := s.repo.UpdateUser(ctx, id, input, actor)
	return mutationResult(revision, "用户已更新。"), err
}

func (s *Service) DeleteUser(ctx context.Context, id string, actor string) (MutationResult, error) {
	if strings.TrimSpace(id) == "" {
		return mutationError("用户不存在。"), nil
	}
	revision, err := s.repo.DeleteUser(ctx, id, actor)
	return mutationResult(revision, "用户已删除。"), err
}

func (s *Service) CreateTeam(ctx context.Context, input TeamInput, actor string) (MutationResult, error) {
	if strings.TrimSpace(input.Name) == "" {
		return mutationError("团队名称不能为空。"), nil
	}
	revision, err := s.repo.CreateTeam(ctx, input, actor)
	return mutationResult(revision, "团队已创建。"), err
}

func (s *Service) UpdateTeam(ctx context.Context, id string, input TeamUpdateInput, actor string) (MutationResult, error) {
	if strings.TrimSpace(id) == "" {
		return mutationError("团队不存在。"), nil
	}
	revision, err := s.repo.UpdateTeam(ctx, id, input, actor)
	return mutationResult(revision, "团队已更新。"), err
}

func (s *Service) DeleteTeam(ctx context.Context, id string, actor string) (MutationResult, error) {
	if strings.TrimSpace(id) == "" {
		return mutationError("团队不存在。"), nil
	}
	revision, err := s.repo.DeleteTeam(ctx, id, actor)
	return mutationResult(revision, "团队已删除。"), err
}

func (s *Service) AddTeamMember(ctx context.Context, input TeamMemberInput, actor string) (MutationResult, error) {
	if strings.TrimSpace(input.TeamID) == "" || strings.TrimSpace(input.UserID) == "" || !validTeamRole(input.Role) {
		return mutationError("团队成员参数无效。"), nil
	}
	revision, err := s.repo.AddTeamMember(ctx, input, actor)
	return mutationResult(revision, "团队成员已添加。"), err
}

func (s *Service) UpdateTeamMember(ctx context.Context, teamID string, userID string, input TeamMemberUpdateInput, actor string) (MutationResult, error) {
	if strings.TrimSpace(teamID) == "" || strings.TrimSpace(userID) == "" {
		return mutationError("团队成员不存在。"), nil
	}
	if input.Role != nil && !validTeamRole(*input.Role) {
		return mutationError("团队成员角色无效。"), nil
	}
	revision, err := s.repo.UpdateTeamMember(ctx, teamID, userID, input, actor)
	return mutationResult(revision, "团队成员已更新。"), err
}

func (s *Service) RemoveTeamMember(ctx context.Context, teamID string, userID string, actor string) (MutationResult, error) {
	if strings.TrimSpace(teamID) == "" || strings.TrimSpace(userID) == "" {
		return mutationError("团队成员不存在。"), nil
	}
	revision, err := s.repo.RemoveTeamMember(ctx, teamID, userID, actor)
	return mutationResult(revision, "团队成员已移除。"), err
}

func (s *Service) CreateAgent(ctx context.Context, input AgentInput, actor string) (MutationResult, error) {
	if strings.TrimSpace(input.AgentKey) == "" || strings.TrimSpace(input.Name) == "" {
		return mutationError("智能体 key 和名称不能为空。"), nil
	}
	revision, err := s.repo.CreateAgent(ctx, input, actor)
	return mutationResult(revision, "智能体已创建。"), err
}

func (s *Service) UpdateAgent(ctx context.Context, id string, input AgentUpdateInput, actor string) (MutationResult, error) {
	if strings.TrimSpace(id) == "" {
		return mutationError("智能体不存在。"), nil
	}
	revision, err := s.repo.UpdateAgent(ctx, id, input, actor)
	return mutationResult(revision, "智能体已更新。"), err
}

func (s *Service) DeleteAgent(ctx context.Context, id string, actor string) (MutationResult, error) {
	if strings.TrimSpace(id) == "" {
		return mutationError("智能体不存在。"), nil
	}
	revision, err := s.repo.DeleteAgent(ctx, id, actor)
	return mutationResult(revision, "智能体已删除。"), err
}

func (s *Service) GrantProjectPermission(ctx context.Context, input ProjectPermissionInput, actor string) (MutationResult, error) {
	if strings.TrimSpace(input.ProjectID) == "" || strings.TrimSpace(input.PrincipalID) == "" ||
		!validPrincipalType(input.PrincipalType) || !validProjectRole(input.Role) {
		return mutationError("项目权限参数无效。"), nil
	}
	revision, err := s.repo.GrantProjectPermission(ctx, input, actor)
	return mutationResult(revision, "项目权限已授权。"), err
}

func (s *Service) UpdateProjectPermission(ctx context.Context, id string, input ProjectPermissionUpdateInput, actor string) (MutationResult, error) {
	if strings.TrimSpace(id) == "" {
		return mutationError("项目权限不存在。"), nil
	}
	if input.Role != nil && !validProjectRole(*input.Role) {
		return mutationError("项目权限角色无效。"), nil
	}
	revision, err := s.repo.UpdateProjectPermission(ctx, id, input, actor)
	return mutationResult(revision, "项目权限已更新。"), err
}

func (s *Service) RevokeProjectPermission(ctx context.Context, id string, actor string) (MutationResult, error) {
	if strings.TrimSpace(id) == "" {
		return mutationError("项目权限不存在。"), nil
	}
	revision, err := s.repo.RevokeProjectPermission(ctx, id, actor)
	return mutationResult(revision, "项目权限已撤销。"), err
}

func (s *Service) CreateIssueLabel(ctx context.Context, input IssueLabelInput, actor string) (MutationResult, error) {
	if strings.TrimSpace(input.Name) == "" {
		return mutationError("标签名称不能为空。"), nil
	}
	revision, err := s.repo.CreateIssueLabel(ctx, input, actor)
	return mutationResult(revision, "标签已创建。"), err
}

func (s *Service) UpdateIssueLabel(ctx context.Context, id string, input IssueLabelUpdateInput, actor string) (MutationResult, error) {
	if strings.TrimSpace(id) == "" {
		return mutationError("标签不存在。"), nil
	}
	revision, err := s.repo.UpdateIssueLabel(ctx, id, input, actor)
	return mutationResult(revision, "标签已更新。"), err
}

func (s *Service) DeleteIssueLabel(ctx context.Context, id string, actor string) (MutationResult, error) {
	if strings.TrimSpace(id) == "" {
		return mutationError("标签不存在。"), nil
	}
	revision, err := s.repo.DeleteIssueLabel(ctx, id, actor)
	return mutationResult(revision, "标签已删除。"), err
}

func (s *Service) SetIssueLabels(ctx context.Context, input IssueLabelsSetInput, actor string) (MutationResult, error) {
	if strings.TrimSpace(input.IssueID) == "" {
		return mutationError("issue 不存在。"), nil
	}
	revision, err := s.repo.SetIssueLabels(ctx, input, actor)
	return mutationResult(revision, "issue 标签已更新。"), err
}

func (s *Service) CreateIssueDependency(ctx context.Context, input IssueDependencyInput, actor string) (MutationResult, error) {
	if strings.TrimSpace(input.FromIssueID) == "" || strings.TrimSpace(input.ToIssueID) == "" || input.FromIssueID == input.ToIssueID {
		return mutationError("issue 依赖参数无效。"), nil
	}
	revision, err := s.repo.CreateIssueDependency(ctx, input, actor)
	return mutationResult(revision, "issue 依赖已创建。"), err
}

func (s *Service) DeleteIssueDependency(ctx context.Context, id string, actor string) (MutationResult, error) {
	if strings.TrimSpace(id) == "" {
		return mutationError("issue 依赖不存在。"), nil
	}
	revision, err := s.repo.DeleteIssueDependency(ctx, id, actor)
	return mutationResult(revision, "issue 依赖已删除。"), err
}

func (s *Service) CreateReview(ctx context.Context, input ReviewInput, actor string) (MutationResult, error) {
	if strings.TrimSpace(input.IssueID) == "" {
		return mutationError("review 缺少 issue。"), nil
	}
	revision, err := s.repo.CreateReview(ctx, input, actor)
	return mutationResult(revision, "review 已创建。"), err
}

func (s *Service) UpdateReview(ctx context.Context, id string, input ReviewUpdateInput, actor string) (MutationResult, error) {
	if strings.TrimSpace(id) == "" {
		return mutationError("review 不存在。"), nil
	}
	if input.Status != nil && !validReviewStatus(*input.Status) {
		return mutationError("review 状态无效。"), nil
	}
	revision, err := s.repo.UpdateReview(ctx, id, input, actor)
	return mutationResult(revision, "review 已更新。"), err
}

func (s *Service) DeleteReview(ctx context.Context, id string, actor string) (MutationResult, error) {
	if strings.TrimSpace(id) == "" {
		return mutationError("review 不存在。"), nil
	}
	revision, err := s.repo.DeleteReview(ctx, id, actor)
	return mutationResult(revision, "review 已删除。"), err
}

func (s *Service) CreateReviewComment(ctx context.Context, input ReviewCommentInput, actor string) (MutationResult, error) {
	if strings.TrimSpace(input.ReviewID) == "" || strings.TrimSpace(input.IssueID) == "" || strings.TrimSpace(input.Body) == "" {
		return mutationError("review comment 参数无效。"), nil
	}
	revision, err := s.repo.CreateReviewComment(ctx, input, actor)
	return mutationResult(revision, "review comment 已创建。"), err
}

func (s *Service) UpdateReviewComment(ctx context.Context, id string, input ReviewCommentUpdateInput, actor string) (MutationResult, error) {
	if strings.TrimSpace(id) == "" {
		return mutationError("review comment 不存在。"), nil
	}
	revision, err := s.repo.UpdateReviewComment(ctx, id, input, actor)
	return mutationResult(revision, "review comment 已更新。"), err
}

func (s *Service) DeleteReviewComment(ctx context.Context, id string, actor string) (MutationResult, error) {
	if strings.TrimSpace(id) == "" {
		return mutationError("review comment 不存在。"), nil
	}
	revision, err := s.repo.DeleteReviewComment(ctx, id, actor)
	return mutationResult(revision, "review comment 已删除。"), err
}

func (s *Service) TransitionIssue(ctx context.Context, boardID string, projectID string, input TransitionInput, actor string) (ChangeResult, error) {
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
	if issue.RunID != nil {
		return changeError(boardID, projectID, revision, current, "智能体正在回答，完成后才能切换状态。"), nil
	}
	catalog, err := s.repo.ListWorkflowCatalog(ctx)
	if err != nil {
		return ChangeResult{}, err
	}

	wf := findWorkflowByID(catalog, issue.WorkflowID)
	if wf != nil && wf.TransitionMode == "free" {
		return s.freeTransition(ctx, boardID, projectID, current, revision, issue, catalog, input, actor)
	}

	transition := findTransition(catalog, *issue, input)
	if transition == nil {
		return changeError(boardID, projectID, revision, current, "当前阶段没有这个 workflow 动作。"), nil
	}
	targetStage := findWorkflowStage(catalog, issue.WorkflowID, transition.ToStageID)
	targetStatus := findWorkflowStatus(catalog, issue.WorkflowID, transition.ToStatusID)
	if targetStage == nil || targetStatus == nil {
		return changeError(boardID, projectID, revision, current, "workflow 目标阶段不存在。"), nil
	}
	next := *issue
	next.StageID = targetStage.ID
	next.StageKey = targetStage.Key
	next.StageName = targetStage.Name
	next.StatusID = targetStatus.ID
	next.StatusKey = targetStatus.Key
	next.StatusName = targetStatus.Name
	next.ColumnKey = targetStatus.ColumnKey
	next.Status = Status(targetStatus.Key)
	next.ReviewRequired = next.Status == StatusInReview
	if next.Status != issue.Status {
		next.RunState = nil
	}
	next.Position = NextPosition(current, next.Status)
	if input.Position != nil && ValidPosition(*input.Position) {
		next.Position = *input.Position
	}
	next.UpdatedAt = time.Now().UTC()
	next.UpdatedBy = actorPtr(actor)
	nextRevision, err := s.repo.ReplaceIssue(ctx, boardID, next, "kanban.issue.transitioned", actor)
	if err != nil {
		return ChangeResult{}, err
	}
	issues, _, err := s.repo.ListIssues(ctx, boardID, projectID)
	if err != nil {
		return ChangeResult{}, err
	}
	return ChangeResult{
		OK:        true,
		Message:   transition.Name + " 已完成。",
		BoardID:   boardID,
		ProjectID: projectID,
		Revision:  nextRevision,
		Issue:     findIssue(issues, issue.ID),
		Issues:    issues,
	}, nil
}

func (s *Service) freeTransition(
	ctx context.Context,
	boardID string,
	projectID string,
	current []Issue,
	revision int64,
	issue *Issue,
	catalog WorkflowCatalog,
	input TransitionInput,
	actor string,
) (ChangeResult, error) {
	var targetStage *WorkflowStage
	var targetStatus *WorkflowStatus

	if input.TransitionID != nil || input.ActionKey != nil {
		transition := findTransition(catalog, *issue, input)
		if transition != nil {
			targetStage = findWorkflowStage(catalog, issue.WorkflowID, transition.ToStageID)
			targetStatus = findWorkflowStatus(catalog, issue.WorkflowID, transition.ToStatusID)
		}
	}

	if targetStage == nil && input.StageID != nil {
		targetStage = findWorkflowStage(catalog, issue.WorkflowID, *input.StageID)
	}
	if targetStage == nil {
		targetStage = findWorkflowStage(catalog, issue.WorkflowID, issue.StageID)
	}

	if targetStatus == nil && input.StatusID != nil {
		targetStatus = findWorkflowStatus(catalog, issue.WorkflowID, *input.StatusID)
	}
	if targetStatus == nil {
		targetStatus = findWorkflowStatus(catalog, issue.WorkflowID, issue.StatusID)
	}

	if targetStage == nil || targetStatus == nil {
		return changeError(boardID, projectID, revision, current, "自由流转缺少目标阶段或状态。"), nil
	}

	next := *issue
	next.StageID = targetStage.ID
	next.StageKey = targetStage.Key
	next.StageName = targetStage.Name
	next.StatusID = targetStatus.ID
	next.StatusKey = targetStatus.Key
	next.StatusName = targetStatus.Name
	next.ColumnKey = targetStatus.ColumnKey
	next.Status = Status(targetStatus.Key)
	next.ReviewRequired = next.Status == StatusInReview
	if next.Status != issue.Status {
		next.RunState = nil
	}
	next.Position = NextPosition(current, next.Status)
	if input.Position != nil && ValidPosition(*input.Position) {
		next.Position = *input.Position
	}
	next.UpdatedAt = time.Now().UTC()
	next.UpdatedBy = actorPtr(actor)
	nextRevision, err := s.repo.ReplaceIssue(ctx, boardID, next, "kanban.issue.transitioned", actor)
	if err != nil {
		return ChangeResult{}, err
	}
	issues, _, err := s.repo.ListIssues(ctx, boardID, projectID)
	if err != nil {
		return ChangeResult{}, err
	}
	return ChangeResult{
		OK:        true,
		Message:   "自由流转已完成。",
		BoardID:   boardID,
		ProjectID: projectID,
		Revision:  nextRevision,
		Issue:     findIssue(issues, issue.ID),
		Issues:    issues,
	}, nil
}

func findWorkflowByID(catalog WorkflowCatalog, workflowID string) *Workflow {
	for i := range catalog.Workflows {
		if catalog.Workflows[i].ID == workflowID {
			return &catalog.Workflows[i]
		}
	}
	return nil
}

func (s *Service) ListActivity(ctx context.Context, boardID string, limit int) ([]EventLogItem, error) {
	return s.repo.ListRecentEvents(ctx, normalizeBoardID(boardID), limit)
}

func (s *Service) ListAgentRuns(ctx context.Context, issueID string) ([]AgentRun, error) {
	return s.repo.ListAgentRuns(ctx, issueID)
}

func (s *Service) ListAgentToolCalls(ctx context.Context, agentRunID string) ([]AgentToolCall, error) {
	return s.repo.ListAgentToolCalls(ctx, agentRunID)
}

func findTransition(catalog WorkflowCatalog, issue Issue, input TransitionInput) *WorkflowTransition {
	actionKey := ""
	if input.ActionKey != nil {
		actionKey = strings.TrimSpace(*input.ActionKey)
	}
	transitionID := ""
	if input.TransitionID != nil {
		transitionID = strings.TrimSpace(*input.TransitionID)
	}
	for i := range catalog.WorkflowTransitions {
		transition := &catalog.WorkflowTransitions[i]
		if transition.WorkflowID != issue.WorkflowID ||
			transition.FromStageID != issue.StageID ||
			transition.FromStatusID != issue.StatusID {
			continue
		}
		if transitionID != "" && transition.ID == transitionID {
			return transition
		}
		if actionKey != "" && transition.ActionKey == actionKey {
			return transition
		}
	}
	return nil
}

func mutationResult(revision int64, message string) MutationResult {
	return MutationResult{OK: true, Message: message, BoardID: DefaultBoardID, ProjectID: DefaultProjectID, Revision: revision}
}

func mutationError(message string) MutationResult {
	return MutationResult{OK: false, Message: message, BoardID: DefaultBoardID, ProjectID: DefaultProjectID}
}

func validTeamRole(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "owner", "admin", "member":
		return true
	default:
		return false
	}
}

func validPrincipalType(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "user", "team":
		return true
	default:
		return false
	}
}

func validProjectRole(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "owner", "admin", "maintainer", "developer", "reviewer", "viewer":
		return true
	default:
		return false
	}
}

func validReviewStatus(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "pending", "approved", "changes_requested", "rejected":
		return true
	default:
		return false
	}
}
