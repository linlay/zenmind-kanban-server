package kanban

import (
	"context"
	"fmt"
	"strings"
)

// SyncDesktopIssues 处理桌面端批量上行的 issue 同步(desktop.issue.sync)。
// 冲突以云端 revision 为准:baseIssueRevision 不匹配时返回 conflict 与云端权威版本,
// 单条失败不会中断整个批次。
func (s *Service) SyncDesktopIssues(ctx context.Context, boardID string, input DesktopIssueSyncInput, actor string) (DesktopIssueSyncResult, error) {
	boardID = normalizeBoardID(boardID)
	projectID := normalizeProjectID(input.ProjectID)
	result := DesktopIssueSyncResult{
		OK:        true,
		BoardID:   boardID,
		ProjectID: projectID,
		Results:   []DesktopIssueSyncItemResult{},
	}
	projects, err := s.repo.ListProjects(ctx)
	if err != nil {
		return DesktopIssueSyncResult{}, err
	}
	if !projectExists(projects, projectID) {
		result.OK = false
		result.Message = "云端项目不存在。"
		return result, nil
	}
	binding, err := s.FindProjectBindingForDevice(ctx, projectID, input.DeviceID)
	if err != nil {
		return DesktopIssueSyncResult{}, err
	}
	if binding != nil && (binding.ControlMode == "disabled" || binding.Status == "paused") {
		result.OK = false
		result.Message = "绑定已停用或暂停，无法同步。"
		return result, nil
	}
	for _, upsert := range input.Upserts {
		item := s.syncDesktopUpsert(ctx, boardID, projectID, binding, upsert, actor)
		result.Results = append(result.Results, item)
	}
	for _, deletion := range input.Deletes {
		item := s.syncDesktopDelete(ctx, boardID, projectID, deletion, actor)
		result.Results = append(result.Results, item)
	}
	_, revision, err := s.repo.ListIssues(ctx, boardID, projectID)
	if err != nil {
		return DesktopIssueSyncResult{}, err
	}
	result.Revision = revision
	return result, nil
}

func (s *Service) syncDesktopUpsert(
	ctx context.Context,
	boardID string,
	projectID string,
	binding *ProjectBinding,
	upsert DesktopIssueSyncUpsert,
	actor string,
) DesktopIssueSyncItemResult {
	item := DesktopIssueSyncItemResult{
		LocalIssueID:  strings.TrimSpace(upsert.LocalIssueID),
		RemoteIssueID: strings.TrimSpace(upsert.RemoteIssueID),
	}
	if item.LocalIssueID == "" {
		item.Status = "skipped"
		item.Message = "缺少本地 issue ID。"
		return item
	}
	input := upsert.Input
	input.ProjectID = &projectID
	if item.RemoteIssueID == "" {
		change, err := s.CreateIssue(ctx, boardID, input, actor)
		if err != nil {
			item.Status = "error"
			item.Message = err.Error()
			return item
		}
		if !change.OK || change.Issue == nil {
			item.Status = "error"
			item.Message = change.Message
			return item
		}
		item.Status = "created"
		item.RemoteIssueID = change.Issue.ID
		item.Issue = change.Issue
		if binding != nil && binding.SyncPolicy == "select" {
			if err := s.repo.AddProjectBindingIssue(ctx, binding.ID, change.Issue.ID, "desktop_sync"); err != nil {
				item.Message = fmt.Sprintf("issue 已创建，但加入同步清单失败：%v", err)
			}
		}
		return item
	}
	existing, err := s.repo.GetIssue(ctx, boardID, item.RemoteIssueID)
	if err != nil {
		item.Status = "error"
		item.Message = err.Error()
		return item
	}
	if existing == nil || existing.DeletedAt != nil {
		item.Status = "conflict"
		item.Message = "云端 issue 已删除。"
		return item
	}
	if upsert.BaseIssueRevision > 0 && upsert.BaseIssueRevision != existing.Revision {
		item.Status = "conflict"
		item.Message = "云端版本已更新，已返回最新内容。"
		item.Issue = existing
		return item
	}
	updateInput := issueInputToUpdateInput(input)
	updateInput.BaseIssueRevision = nil
	change, err := s.UpdateIssue(ctx, boardID, projectID, item.RemoteIssueID, updateInput, actor)
	if err != nil {
		item.Status = "error"
		item.Message = err.Error()
		return item
	}
	if !change.OK {
		item.Status = "error"
		item.Message = change.Message
		if change.Issue != nil {
			item.Issue = change.Issue
		}
		return item
	}
	item.Status = "updated"
	item.Issue = change.Issue
	return item
}

func (s *Service) syncDesktopDelete(
	ctx context.Context,
	boardID string,
	projectID string,
	deletion DesktopIssueSyncDelete,
	actor string,
) DesktopIssueSyncItemResult {
	item := DesktopIssueSyncItemResult{
		LocalIssueID:  strings.TrimSpace(deletion.LocalIssueID),
		RemoteIssueID: strings.TrimSpace(deletion.RemoteIssueID),
	}
	if item.RemoteIssueID == "" {
		item.Status = "skipped"
		item.Message = "缺少云端 issue ID。"
		return item
	}
	existing, err := s.repo.GetIssue(ctx, boardID, item.RemoteIssueID)
	if err != nil {
		item.Status = "error"
		item.Message = err.Error()
		return item
	}
	if existing == nil || existing.DeletedAt != nil {
		item.Status = "deleted"
		item.Message = "云端 issue 已不存在。"
		return item
	}
	if deletion.BaseIssueRevision > 0 && deletion.BaseIssueRevision != existing.Revision {
		item.Status = "conflict"
		item.Message = "云端版本已更新，删除被拒绝。"
		item.Issue = existing
		return item
	}
	change, err := s.DeleteIssue(ctx, boardID, projectID, item.RemoteIssueID, nil, actor)
	if err != nil {
		item.Status = "error"
		item.Message = err.Error()
		return item
	}
	if !change.OK {
		item.Status = "error"
		item.Message = change.Message
		return item
	}
	item.Status = "deleted"
	return item
}

func issueInputToUpdateInput(input IssueInput) IssueUpdateInput {
	update := IssueUpdateInput{
		ProjectID:          input.ProjectID,
		WorkflowID:         input.WorkflowID,
		StageID:            input.StageID,
		StatusID:           input.StatusID,
		Description:        input.Description,
		Status:             input.Status,
		Priority:           input.Priority,
		Severity:           input.Severity,
		AssigneeAgentKey:   input.AssigneeAgentKey,
		AssigneeID:         input.AssigneeID,
		WorkerType:         input.WorkerType,
		WorkerID:           input.WorkerID,
		WorkerAgent:        input.WorkerAgent,
		ReviewerID:         input.ReviewerID,
		ReviewRequired:     input.ReviewRequired,
		RunState:           input.RunState,
		AutomationID:       input.AutomationID,
		AutomationEnabled:  input.AutomationEnabled,
		AutomationCron:     input.AutomationCron,
		AutomationMessage:  input.AutomationMessage,
		AutomationTimezone: input.AutomationTimezone,
		AttachmentChatID:   input.AttachmentChatID,
		Attachments:        input.Attachments,
	}
	if title := strings.TrimSpace(input.Title); title != "" {
		update.Title = &title
	}
	return update
}
