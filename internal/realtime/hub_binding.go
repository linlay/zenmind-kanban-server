package realtime

import (
	"context"
	"encoding/json"
	"strings"

	"zenmind-kanban-server/internal/kanban"
)

// resolveDispatchTarget 未显式指定目标会话且多台 desktop 在线时,
// 若恰有一台设备对该项目持有 controlMode=dispatch 的 active 绑定,则自动选中该设备的会话。
// 其余情况保持原值(单台在线/显式指定均由 pickDesktop 原逻辑处理)。
func (h *Hub) resolveDispatchTarget(projectID string, targetDesktopSessionID string) string {
	targetDesktopSessionID = strings.TrimSpace(targetDesktopSessionID)
	if targetDesktopSessionID != "" {
		return targetDesktopSessionID
	}
	h.mu.RLock()
	multiple := len(h.desktopSessions) > 1
	sessionsByDevice := map[string]string{}
	for id, session := range h.desktopSessions {
		if deviceID := strings.TrimSpace(session.deviceID); deviceID != "" {
			sessionsByDevice[deviceID] = id
		}
	}
	h.mu.RUnlock()
	if !multiple {
		return targetDesktopSessionID
	}
	bindings, err := h.service.ListProjectBindings(context.Background(), projectID)
	if err != nil || !bindings.OK {
		return targetDesktopSessionID
	}
	candidate := ""
	for _, binding := range bindings.Bindings {
		if binding.ControlMode != "dispatch" || binding.Status != "active" {
			continue
		}
		sessionID, online := sessionsByDevice[binding.DeviceID]
		if !online {
			continue
		}
		if candidate != "" && candidate != sessionID {
			return targetDesktopSessionID
		}
		candidate = sessionID
	}
	if candidate != "" {
		return candidate
	}
	return targetDesktopSessionID
}

// bindingAllowsDispatch 判定绑定是否允许向设备派发任务。
// 无绑定时维持历史行为(放行);有绑定时要求 controlMode=dispatch 且 status=active。
func bindingAllowsDispatch(binding *kanban.ProjectBinding) (bool, string) {
	if binding == nil {
		return true, ""
	}
	if binding.Status == "paused" {
		return false, "该设备与此项目的绑定已暂停，无法派发任务。"
	}
	switch binding.ControlMode {
	case "dispatch":
		return true, ""
	case "observe":
		return false, "该设备对此项目为只读观察模式，无法派发任务。"
	case "disabled":
		return false, "该设备与此项目的绑定已停用，无法派发任务。"
	default:
		return true, ""
	}
}

// resolveDispatchBinding 在派发前查找目标 desktop 会话对应的绑定并执行 controlMode 校验。
// 返回的 binding 用于把 localProjectId 注入转发 payload;校验失败时返回错误信息。
func (h *Hub) resolveDispatchBinding(projectID string, targetDesktopSessionID string) (*kanban.ProjectBinding, string) {
	deviceID := ""
	h.mu.RLock()
	if targetDesktopSessionID != "" {
		if session := h.desktopSessions[strings.TrimSpace(targetDesktopSessionID)]; session != nil {
			deviceID = session.deviceID
		}
	} else if len(h.desktopSessions) == 1 {
		for _, session := range h.desktopSessions {
			deviceID = session.deviceID
		}
	}
	h.mu.RUnlock()
	if deviceID == "" {
		return nil, ""
	}
	binding, err := h.service.FindProjectBindingForDevice(context.Background(), projectID, deviceID)
	if err != nil {
		h.logger.Warn("failed to resolve dispatch binding", "error", err)
		return nil, ""
	}
	if ok, reason := bindingAllowsDispatch(binding); !ok {
		return binding, reason
	}
	return binding, ""
}

// dispatchBindingContext 注入到 desktop 转发 payload 的绑定上下文。
func dispatchBindingContext(binding *kanban.ProjectBinding) map[string]any {
	if binding == nil {
		return nil
	}
	return map[string]any{
		"id":               binding.ID,
		"localProjectId":   binding.LocalProjectID,
		"localDisplayName": binding.LocalDisplayName,
	}
}

// pinDispatchedIssue 派发成功后把 issue 钉进绑定,防止 future/select 策略的快照墓碑误删。
func (h *Hub) pinDispatchedIssue(binding *kanban.ProjectBinding, issueID string) {
	if binding == nil || strings.TrimSpace(issueID) == "" {
		return
	}
	if err := h.service.PinProjectBindingIssue(context.Background(), binding.ID, issueID, "dispatch"); err != nil {
		h.logger.Warn("failed to pin dispatched issue", "error", err, "binding", binding.ID, "issue", issueID)
	}
}

// applyDesktopBindingScope 按绑定的 syncPolicy 过滤发往 desktop 的快照。
// 在 applyDesktopSnapshotScope 之后调用:
//   - 无绑定 → 原样(兼容存量流程,全量同步)
//   - disabled/paused → 清空 issues 且 Complete=false(关闭墓碑,防止本地副本被误删)
//   - all → 原样
//   - future → CreatedAt >= SyncSinceAt 的 issue ∪ 钉住的 issue
//   - select → 仅钉住的 issue
func (h *Hub) applyDesktopBindingScope(session *Session, snapshot *kanban.ListResult) {
	if session.role != "desktop" {
		return
	}
	deviceID := strings.TrimSpace(session.deviceID)
	if deviceID == "" {
		return
	}
	var binding *kanban.ProjectBinding
	for i := range snapshot.ProjectBindings {
		candidate := &snapshot.ProjectBindings[i]
		if candidate.DeviceID == deviceID && candidate.ProjectID == snapshot.ProjectID {
			binding = candidate
			break
		}
	}
	if binding == nil {
		return
	}
	if binding.ControlMode == "disabled" || binding.Status == "paused" {
		snapshot.Complete = false
		filterSnapshotIssues(snapshot, func(kanban.Issue) bool { return false })
		return
	}
	switch binding.SyncPolicy {
	case "all":
		return
	case "future":
		pinned := snapshotPinnedIssueIDs(snapshot, binding.ID)
		filterSnapshotIssues(snapshot, func(issue kanban.Issue) bool {
			if pinned[issue.ID] {
				return true
			}
			return binding.SyncSinceAt == nil || !issue.CreatedAt.Before(*binding.SyncSinceAt)
		})
	case "select":
		pinned := snapshotPinnedIssueIDs(snapshot, binding.ID)
		filterSnapshotIssues(snapshot, func(issue kanban.Issue) bool {
			return pinned[issue.ID]
		})
	}
}

func snapshotPinnedIssueIDs(snapshot *kanban.ListResult, bindingID string) map[string]bool {
	pinned := map[string]bool{}
	for _, item := range snapshot.ProjectBindingIssues {
		if item.BindingID == bindingID {
			pinned[item.IssueID] = true
		}
	}
	return pinned
}

// filterSnapshotIssues 按谓词过滤快照中的 issues 及其关联集合。
func filterSnapshotIssues(snapshot *kanban.ListResult, keep func(kanban.Issue) bool) {
	issueIDs := map[string]bool{}
	filtered := make([]kanban.Issue, 0, len(snapshot.Issues))
	for _, issue := range snapshot.Issues {
		if !keep(issue) {
			continue
		}
		filtered = append(filtered, issue)
		issueIDs[issue.ID] = true
	}
	snapshot.Issues = filtered
	snapshot.ProjectIssueStats = kanban.ComputeProjectIssueStats(snapshot.Projects, snapshot.Issues)
	snapshot.IssueLabelLinks = filterSlice(snapshot.IssueLabelLinks, func(link kanban.IssueLabelLink) bool {
		return issueIDs[link.IssueID]
	})
	snapshot.IssueDependencies = filterSlice(snapshot.IssueDependencies, func(dependency kanban.IssueDependency) bool {
		return issueIDs[dependency.FromIssueID] || issueIDs[dependency.ToIssueID]
	})
	snapshot.Reviews = filterSlice(snapshot.Reviews, func(review kanban.Review) bool {
		return issueIDs[review.IssueID]
	})
	snapshot.ReviewComments = filterSlice(snapshot.ReviewComments, func(comment kanban.ReviewComment) bool {
		return issueIDs[comment.IssueID]
	})
	snapshot.AgentRuns = filterSlice(snapshot.AgentRuns, func(run kanban.AgentRun) bool {
		return issueIDs[run.IssueID]
	})
	runIDs := map[string]bool{}
	for _, run := range snapshot.AgentRuns {
		runIDs[run.ID] = true
	}
	snapshot.AgentToolCalls = filterSlice(snapshot.AgentToolCalls, func(toolCall kanban.AgentToolCall) bool {
		return runIDs[toolCall.AgentRunID]
	})
	snapshot.RecentEvents = filterSlice(snapshot.RecentEvents, func(event kanban.EventLogItem) bool {
		return event.IssueID != nil && issueIDs[*event.IssueID]
	})
}

// desktopProjectSelect 处理 desktop.project.select:免重连切换项目。
func (h *Hub) desktopProjectSelect(session *Session, env Envelope) {
	if session.role != "desktop" {
		session.respondError(env, "forbidden", "仅 desktop 会话可切换项目。")
		return
	}
	var payload struct {
		SelectedProjectID string `json:"selectedProjectId"`
	}
	if !decodeOrRespond(session, env, &payload) {
		return
	}
	selected := strings.TrimSpace(payload.SelectedProjectID)
	if selected == "" {
		selected = kanban.DefaultProjectID
	}
	session.projectID = selected
	if err := h.store.UpdateDesktopClientSelectedProject(context.Background(), session.id, selected); err != nil {
		h.logger.Warn("failed to persist desktop selected project", "error", err)
	}
	session.respond(env, map[string]any{
		"ok":                true,
		"message":           "已切换项目。",
		"selectedProjectId": selected,
	})
	snapshotEnv := Envelope{ID: randomID(), Op: "kanban.snapshot", BoardID: env.BoardID, ProjectID: selected}
	session.sendSnapshotFor(snapshotEnv)
	h.broadcastDesktopStatus()
}

// desktopIssueSync 处理 desktop.issue.sync:桌面端本地 issue 批量上行。
func (h *Hub) desktopIssueSync(session *Session, env Envelope) {
	if session.role != "desktop" {
		session.respondError(env, "forbidden", "仅 desktop 会话可上行同步。")
		return
	}
	var input kanban.DesktopIssueSyncInput
	if !decodeOrRespond(session, env, &input) {
		return
	}
	if strings.TrimSpace(input.DeviceID) == "" {
		input.DeviceID = session.deviceID
	}
	if strings.TrimSpace(input.ProjectID) == "" {
		input.ProjectID = env.ProjectID
	}
	result, err := h.service.SyncDesktopIssues(context.Background(), env.BoardID, input, session.actorID())
	if err != nil {
		session.respondError(env, "server_error", err.Error())
		return
	}
	session.respond(env, result)
	if result.OK && len(result.Results) > 0 {
		h.broadcastSnapshots(env.BoardID)
	}
}

// desktopProjectBind 处理 desktop.project.bind:先让 desktop 确认本地项目,再落库绑定。
func (h *Hub) desktopProjectBind(session *Session, env Envelope) {
	var payload struct {
		TargetDesktopSessionID string `json:"targetDesktopSessionId"`
		ProjectID              string `json:"projectId"`
		DeviceID               string `json:"deviceId"`
		LocalProjectID         string `json:"localProjectId"`
		LocalDisplayName       string `json:"localDisplayName"`
		SyncPolicy             string `json:"syncPolicy"`
		ControlMode            string `json:"controlMode"`
		CurrentUserID          string `json:"currentUserId"`
	}
	if !decodeOrRespond(session, env, &payload) {
		return
	}
	projectID := strings.TrimSpace(payload.ProjectID)
	if projectID == "" {
		projectID = env.ProjectID
	}
	response, err := h.requestDesktop("desktop.project.bind", env.BoardID, projectID, payload.TargetDesktopSessionID, json.RawMessage(env.Payload))
	if err != nil {
		session.respondError(env, "desktop_unavailable", err.Error())
		return
	}
	deviceID := strings.TrimSpace(payload.DeviceID)
	if deviceID == "" {
		deviceID = h.desktopSessionDeviceID(payload.TargetDesktopSessionID)
	}
	result, err := h.service.CreateProjectBinding(context.Background(), kanban.ProjectBindingInput{
		ProjectID:        projectID,
		DeviceID:         deviceID,
		CurrentUserID:    strings.TrimSpace(payload.CurrentUserID),
		LocalProjectID:   strings.TrimSpace(payload.LocalProjectID),
		LocalDisplayName: strings.TrimSpace(payload.LocalDisplayName),
		SyncPolicy:       strings.TrimSpace(payload.SyncPolicy),
		ControlMode:      strings.TrimSpace(payload.ControlMode),
	}, session.actorID())
	if err != nil {
		session.respondError(env, "server_error", err.Error())
		return
	}
	session.respond(env, map[string]any{
		"ok":      result.OK,
		"message": result.Message,
		"binding": result.Binding,
		"desktop": json.RawMessage(response.Payload),
	})
	if result.OK {
		h.broadcastSnapshots(env.BoardID)
	}
}

// desktopProjectUnbind 处理 desktop.project.unbind:通知 desktop 后软删绑定。
func (h *Hub) desktopProjectUnbind(session *Session, env Envelope) {
	var payload struct {
		TargetDesktopSessionID string `json:"targetDesktopSessionId"`
		BindingID              string `json:"bindingId"`
		ID                     string `json:"id"`
	}
	if !decodeOrRespond(session, env, &payload) {
		return
	}
	bindingID := strings.TrimSpace(payload.BindingID)
	if bindingID == "" {
		bindingID = strings.TrimSpace(payload.ID)
	}
	if bindingID == "" {
		session.respondError(env, "bad_payload", "缺少绑定 ID。")
		return
	}
	// desktop 离线时也允许解绑(本地状态由其下次连上后的快照对账)
	response, desktopErr := h.requestDesktop("desktop.project.unbind", env.BoardID, env.ProjectID, payload.TargetDesktopSessionID, json.RawMessage(env.Payload))
	result, err := h.service.DeleteProjectBinding(context.Background(), bindingID, session.actorID())
	if err != nil {
		session.respondError(env, "server_error", err.Error())
		return
	}
	out := map[string]any{
		"ok":      result.OK,
		"message": result.Message,
	}
	if desktopErr != nil {
		out["desktopError"] = desktopErr.Error()
	} else {
		out["desktop"] = json.RawMessage(response.Payload)
	}
	session.respond(env, out)
	if result.OK {
		h.broadcastSnapshots(env.BoardID)
	}
}

// desktopProjectCreateLocal 处理 kanban.desktop.project.createLocal:
// 编排 desktop 创建本地项目 → 服务端落库绑定 → 广播。
func (h *Hub) desktopProjectCreateLocal(session *Session, env Envelope) {
	var input kanban.DesktopCreateLocalProjectInput
	if !decodeOrRespond(session, env, &input) {
		return
	}
	projectID := strings.TrimSpace(input.ProjectID)
	if projectID == "" {
		projectID = env.ProjectID
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		session.respondError(env, "bad_payload", "缺少本地项目名称。")
		return
	}
	response, err := h.requestDesktop("desktop.project.createLocal", env.BoardID, projectID, input.TargetDesktopSessionID, map[string]any{
		"name":           name,
		"localProjectId": strings.TrimSpace(input.LocalProjectID),
		"cloudProjectId": projectID,
	})
	if err != nil {
		session.respondError(env, "desktop_unavailable", err.Error())
		return
	}
	var created struct {
		OK      bool                        `json:"ok"`
		Message string                      `json:"message"`
		Project *kanban.DesktopLocalProject `json:"project"`
	}
	if err := json.Unmarshal(response.Payload, &created); err != nil || created.Project == nil || strings.TrimSpace(created.Project.ID) == "" {
		message := "desktop 创建本地项目返回格式无效。"
		if created.Message != "" {
			message = created.Message
		}
		session.respond(env, map[string]any{"ok": false, "message": message})
		return
	}
	deviceID := h.desktopSessionDeviceID(input.TargetDesktopSessionID)
	localDisplayName := created.Project.Name
	if localDisplayName == "" {
		localDisplayName = name
	}
	result, err := h.service.CreateProjectBinding(context.Background(), kanban.ProjectBindingInput{
		ProjectID:        projectID,
		DeviceID:         deviceID,
		LocalProjectID:   strings.TrimSpace(created.Project.ID),
		LocalDisplayName: localDisplayName,
		SyncPolicy:       strings.TrimSpace(input.SyncPolicy),
		ControlMode:      strings.TrimSpace(input.ControlMode),
	}, session.actorID())
	if err != nil {
		session.respondError(env, "server_error", err.Error())
		return
	}
	session.respond(env, map[string]any{
		"ok":      result.OK,
		"message": result.Message,
		"project": created.Project,
		"binding": result.Binding,
	})
	if result.OK {
		h.broadcastSnapshots(env.BoardID)
	}
}

// desktopSessionDeviceID 取目标会话的 deviceId;未指定会话时仅在恰有一台 desktop 在线时返回。
func (h *Hub) desktopSessionDeviceID(targetDesktopSessionID string) string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	targetDesktopSessionID = strings.TrimSpace(targetDesktopSessionID)
	if targetDesktopSessionID != "" {
		if session := h.desktopSessions[targetDesktopSessionID]; session != nil {
			return session.deviceID
		}
		return ""
	}
	if len(h.desktopSessions) == 1 {
		for _, session := range h.desktopSessions {
			return session.deviceID
		}
	}
	return ""
}
