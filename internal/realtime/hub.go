package realtime

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"zenmind-kanban-server/internal/config"
	"zenmind-kanban-server/internal/kanban"
	"zenmind-kanban-server/internal/store"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 70 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 1 << 20
	requestTimeout = 60 * time.Second
)

type Hub struct {
	cfg     config.Config
	service *kanban.Service
	store   *store.Store
	logger  *slog.Logger

	mu              sync.RWMutex
	sessions        map[string]*Session
	desktopSessions map[string]*Session
	pending         map[string]chan Envelope
	idempotent      map[string]*idempotentRequest
}

type idempotentRequest struct {
	done   chan struct{}
	result any
	err    error
}

type desktopAgentLoadResult struct {
	agents []kanban.DesktopAgentOption
	err    error
}

type desktopAgentListPayload struct {
	Items  []kanban.DesktopAgentOption `json:"items"`
	Agents []kanban.DesktopAgentOption `json:"agents"`
}

func NewHub(cfg config.Config, service *kanban.Service, store *store.Store, logger *slog.Logger) *Hub {
	return &Hub{
		cfg:             cfg,
		service:         service,
		store:           store,
		logger:          logger,
		sessions:        map[string]*Session{},
		desktopSessions: map[string]*Session{},
		pending:         map[string]chan Envelope{},
		idempotent:      map[string]*idempotentRequest{},
	}
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     h.checkOrigin,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Warn("websocket upgrade failed", "error", err)
		return
	}
	role := strings.TrimSpace(r.URL.Query().Get("role"))
	if role != "desktop" {
		role = "web"
	}
	session := &Session{
		id:        randomID(),
		role:      role,
		conn:      conn,
		hub:       h,
		send:      make(chan OutEnvelope, 32),
		board:     kanban.DefaultBoardID,
		projectID: kanban.DefaultProjectID,
		closed:    make(chan struct{}),
	}
	h.register(session)
	go session.writePump()
	go session.readPump()
	if role != "desktop" {
		session.sendSnapshot()
	}
}

func (h *Hub) register(session *Session) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessions[session.id] = session
	if session.role == "desktop" {
		h.desktopSessions[session.id] = session
	}
}

func (h *Hub) unregister(session *Session) {
	h.mu.Lock()
	if _, ok := h.sessions[session.id]; !ok {
		h.mu.Unlock()
		return
	}
	delete(h.sessions, session.id)
	_, wasDesktop := h.desktopSessions[session.id]
	delete(h.desktopSessions, session.id)
	close(session.send)
	h.mu.Unlock()

	if wasDesktop {
		if err := h.store.RemoveDesktopClient(context.Background(), session.id); err != nil {
			h.logger.Warn("failed to remove desktop client", "error", err)
		}
		h.broadcastDesktopStatus()
	}
}

func (h *Hub) handle(session *Session, env Envelope) {
	if env.Type == "rpc.res" {
		h.completePending(env)
		return
	}
	if env.ID == "" {
		env.ID = randomID()
	}
	if env.BoardID == "" {
		env.BoardID = kanban.DefaultBoardID
	}
	if env.ProjectID == "" {
		env.ProjectID = session.projectID
	}

	switch env.Op {
	case "kanban.snapshot.get":
		session.sendSnapshotFor(env)
	case "kanban.issue.create":
		var input kanban.IssueInput
		if !decodeOrRespond(session, env, &input) {
			return
		}
		if input.ProjectID == nil {
			projectID := env.ProjectID
			input.ProjectID = &projectID
		}
		result, err := h.service.CreateIssue(context.Background(), env.BoardID, input, session.actorID())
		h.respondChange(session, env, result, err, "kanban.issue.created")
	case "kanban.issue.update":
		var payload struct {
			ID    string                  `json:"id"`
			Input kanban.IssueUpdateInput `json:"input"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		result, err := h.service.UpdateIssue(context.Background(), env.BoardID, env.ProjectID, payload.ID, payload.Input, session.actorID())
		h.respondChange(session, env, result, err, "kanban.issue.updated")
	case "kanban.issue.delete":
		var payload struct {
			ID                string `json:"id"`
			BaseIssueRevision *int64 `json:"baseIssueRevision"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		result, err := h.service.DeleteIssue(context.Background(), env.BoardID, env.ProjectID, payload.ID, payload.BaseIssueRevision, session.actorID())
		h.respondChange(session, env, result, err, "kanban.issue.deleted")
	case "kanban.issue.move":
		var input kanban.MoveInput
		if !decodeOrRespond(session, env, &input) {
			return
		}
		result, err := h.service.MoveIssue(context.Background(), env.BoardID, env.ProjectID, input, session.actorID())
		h.respondChange(session, env, result, err, "kanban.issue.updated")
	case "kanban.issue.transition":
		var input kanban.TransitionInput
		if !decodeOrRespond(session, env, &input) {
			return
		}
		result, err := h.service.TransitionIssue(context.Background(), env.BoardID, env.ProjectID, input, session.actorID())
		h.respondChange(session, env, result, err, "kanban.issue.transitioned")
	case "kanban.issue.assignAndRun":
		h.assignAndRun(session, env)
	case "kanban.issue.dispatchToDesktop":
		h.dispatchToDesktop(session, env)
	case "kanban.project.create":
		var input kanban.ProjectInput
		if !decodeOrRespond(session, env, &input) {
			return
		}
		result, err := h.service.CreateProject(context.Background(), input, session.actorID())
		h.respondProjectChange(session, env, result, err, "kanban.project.created")
	case "kanban.project.update":
		var payload struct {
			ID    string                    `json:"id"`
			Input kanban.ProjectUpdateInput `json:"input"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		result, err := h.service.UpdateProject(context.Background(), payload.ID, payload.Input, session.actorID())
		h.respondProjectChange(session, env, result, err, "kanban.project.updated")
	case "kanban.project.move":
		var input kanban.ProjectMoveInput
		if !decodeOrRespond(session, env, &input) {
			return
		}
		result, err := h.service.MoveProject(context.Background(), input, session.actorID())
		h.respondProjectChange(session, env, result, err, "kanban.project.moved")
	case "kanban.user.create":
		var input kanban.UserInput
		if !decodeOrRespond(session, env, &input) {
			return
		}
		result, err := h.service.CreateUser(context.Background(), input, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.user.update":
		var payload struct {
			ID    string                 `json:"id"`
			Input kanban.UserUpdateInput `json:"input"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		result, err := h.service.UpdateUser(context.Background(), payload.ID, payload.Input, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.user.delete":
		var payload struct {
			ID string `json:"id"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		result, err := h.service.DeleteUser(context.Background(), payload.ID, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.team.create":
		var input kanban.TeamInput
		if !decodeOrRespond(session, env, &input) {
			return
		}
		result, err := h.service.CreateTeam(context.Background(), input, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.team.update":
		var payload struct {
			ID    string                 `json:"id"`
			Input kanban.TeamUpdateInput `json:"input"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		result, err := h.service.UpdateTeam(context.Background(), payload.ID, payload.Input, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.team.delete":
		var payload struct {
			ID string `json:"id"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		result, err := h.service.DeleteTeam(context.Background(), payload.ID, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.teamMember.add":
		var input kanban.TeamMemberInput
		if !decodeOrRespond(session, env, &input) {
			return
		}
		result, err := h.service.AddTeamMember(context.Background(), input, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.teamMember.update":
		var payload struct {
			TeamID string                       `json:"teamId"`
			UserID string                       `json:"userId"`
			Input  kanban.TeamMemberUpdateInput `json:"input"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		result, err := h.service.UpdateTeamMember(context.Background(), payload.TeamID, payload.UserID, payload.Input, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.teamMember.remove":
		var payload struct {
			TeamID string `json:"teamId"`
			UserID string `json:"userId"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		result, err := h.service.RemoveTeamMember(context.Background(), payload.TeamID, payload.UserID, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.agent.create":
		var input kanban.AgentInput
		if !decodeOrRespond(session, env, &input) {
			return
		}
		result, err := h.service.CreateAgent(context.Background(), input, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.agent.update":
		var payload struct {
			ID    string                  `json:"id"`
			Input kanban.AgentUpdateInput `json:"input"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		result, err := h.service.UpdateAgent(context.Background(), payload.ID, payload.Input, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.agent.delete":
		var payload struct {
			ID string `json:"id"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		result, err := h.service.DeleteAgent(context.Background(), payload.ID, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.projectPermission.grant":
		var input kanban.ProjectPermissionInput
		if !decodeOrRespond(session, env, &input) {
			return
		}
		result, err := h.service.GrantProjectPermission(context.Background(), input, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.projectPermission.update":
		var payload struct {
			ID    string                              `json:"id"`
			Input kanban.ProjectPermissionUpdateInput `json:"input"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		result, err := h.service.UpdateProjectPermission(context.Background(), payload.ID, payload.Input, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.projectPermission.revoke":
		var payload struct {
			ID string `json:"id"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		result, err := h.service.RevokeProjectPermission(context.Background(), payload.ID, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.issueLabel.create":
		var input kanban.IssueLabelInput
		if !decodeOrRespond(session, env, &input) {
			return
		}
		result, err := h.service.CreateIssueLabel(context.Background(), input, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.issueLabel.update":
		var payload struct {
			ID    string                       `json:"id"`
			Input kanban.IssueLabelUpdateInput `json:"input"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		result, err := h.service.UpdateIssueLabel(context.Background(), payload.ID, payload.Input, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.issueLabel.delete":
		var payload struct {
			ID string `json:"id"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		result, err := h.service.DeleteIssueLabel(context.Background(), payload.ID, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.issue.labels.set":
		var input kanban.IssueLabelsSetInput
		if !decodeOrRespond(session, env, &input) {
			return
		}
		result, err := h.service.SetIssueLabels(context.Background(), input, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.issueDependency.create":
		var input kanban.IssueDependencyInput
		if !decodeOrRespond(session, env, &input) {
			return
		}
		result, err := h.service.CreateIssueDependency(context.Background(), input, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.issueDependency.delete":
		var payload struct {
			ID string `json:"id"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		result, err := h.service.DeleteIssueDependency(context.Background(), payload.ID, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.review.create":
		var input kanban.ReviewInput
		if !decodeOrRespond(session, env, &input) {
			return
		}
		result, err := h.service.CreateReview(context.Background(), input, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.review.update":
		var payload struct {
			ID    string                   `json:"id"`
			Input kanban.ReviewUpdateInput `json:"input"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		result, err := h.service.UpdateReview(context.Background(), payload.ID, payload.Input, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.review.delete":
		var payload struct {
			ID string `json:"id"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		result, err := h.service.DeleteReview(context.Background(), payload.ID, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.reviewComment.create":
		var input kanban.ReviewCommentInput
		if !decodeOrRespond(session, env, &input) {
			return
		}
		result, err := h.service.CreateReviewComment(context.Background(), input, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.reviewComment.update":
		var payload struct {
			ID    string                          `json:"id"`
			Input kanban.ReviewCommentUpdateInput `json:"input"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		result, err := h.service.UpdateReviewComment(context.Background(), payload.ID, payload.Input, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.reviewComment.delete":
		var payload struct {
			ID string `json:"id"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		result, err := h.service.DeleteReviewComment(context.Background(), payload.ID, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.activity.list":
		var payload struct {
			Limit int `json:"limit"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		items, err := h.service.ListActivity(context.Background(), env.BoardID, payload.Limit)
		if err != nil {
			session.respondError(env, "server_error", err.Error())
			return
		}
		session.respond(env, map[string]any{"ok": true, "items": items})
	case "kanban.agentRun.list":
		var payload struct {
			IssueID string `json:"issueId"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		items, err := h.service.ListAgentRuns(context.Background(), payload.IssueID)
		if err != nil {
			session.respondError(env, "server_error", err.Error())
			return
		}
		session.respond(env, map[string]any{"ok": true, "items": items})
	case "kanban.agentToolCall.list":
		var payload struct {
			AgentRunID string `json:"agentRunId"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		items, err := h.service.ListAgentToolCalls(context.Background(), payload.AgentRunID)
		if err != nil {
			session.respondError(env, "server_error", err.Error())
			return
		}
		session.respond(env, map[string]any{"ok": true, "items": items})
	case "desktop.online.list":
		h.desktopOnlineList(session, env)
	case "desktop.assistant.listAgents":
		h.forwardDesktop(session, env, "desktop.assistant.listAgents")
	case "kanban.workflow.create":
		var input kanban.WorkflowInput
		if !decodeOrRespond(session, env, &input) {
			return
		}
		result, err := h.service.CreateWorkflow(context.Background(), input, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.workflow.update":
		var payload struct {
			ID    string                     `json:"id"`
			Input kanban.WorkflowUpdateInput `json:"input"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		result, err := h.service.UpdateWorkflow(context.Background(), payload.ID, payload.Input, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.workflow.delete":
		var payload struct {
			ID string `json:"id"`
		}
		if !decodeOrRespond(session, env, &payload) {
			return
		}
		result, err := h.service.DeleteWorkflow(context.Background(), payload.ID, session.actorID())
		h.respondMutation(session, env, result, err)
	case "kanban.automation.sync":
		h.forwardDesktop(session, env, "desktop.automation.sync")
	case "desktop.hello":
		h.desktopHello(session, env)
	case "desktop.assistant.event":
		h.desktopAssistantEvent(session, env)
	default:
		session.respondError(env, "unknown_op", "未知 WebSocket 操作。")
	}
}

func (h *Hub) snapshotForSession(session *Session, boardID string, projectID string) (kanban.ListResult, error) {
	snapshot, err := h.service.Snapshot(context.Background(), boardID, projectID, h.DesktopStatus())
	if err != nil {
		return kanban.ListResult{}, err
	}
	h.applyDesktopSnapshotScope(session, &snapshot)
	return snapshot, nil
}

func (h *Hub) applyDesktopSnapshotScope(session *Session, snapshot *kanban.ListResult) {
	if session.role != "desktop" || session.scope != "current_user" {
		return
	}
	snapshot.Scope = "current_user"
	snapshot.Complete = false
	currentUserID := strings.TrimSpace(session.currentUser.ID)
	if currentUserID == "" {
		snapshot.Issues = []kanban.Issue{}
		snapshot.ProjectIssueStats = kanban.ComputeProjectIssueStats(snapshot.Projects, snapshot.Issues)
		snapshot.IssueLabelLinks = []kanban.IssueLabelLink{}
		snapshot.IssueDependencies = []kanban.IssueDependency{}
		snapshot.Reviews = []kanban.Review{}
		snapshot.ReviewComments = []kanban.ReviewComment{}
		snapshot.AgentRuns = []kanban.AgentRun{}
		snapshot.AgentToolCalls = []kanban.AgentToolCall{}
		snapshot.RecentEvents = []kanban.EventLogItem{}
		return
	}

	issueIDs := map[string]bool{}
	filteredIssues := make([]kanban.Issue, 0, len(snapshot.Issues))
	for _, issue := range snapshot.Issues {
		if !issueVisibleToDesktopUser(issue, currentUserID) {
			continue
		}
		filteredIssues = append(filteredIssues, issue)
		issueIDs[issue.ID] = true
	}
	snapshot.Issues = filteredIssues
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

func issueVisibleToDesktopUser(issue kanban.Issue, userID string) bool {
	return ptrMatches(issue.CreatedBy, userID) ||
		ptrMatches(issue.AssigneeID, userID) ||
		ptrMatches(issue.WorkerID, userID) ||
		ptrMatches(issue.ReviewerID, userID)
}

func ptrMatches(value *string, expected string) bool {
	return value != nil && strings.TrimSpace(*value) == expected
}

func filterSlice[T any](items []T, keep func(T) bool) []T {
	filtered := make([]T, 0, len(items))
	for _, item := range items {
		if keep(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func (h *Hub) respondChange(session *Session, env Envelope, result kanban.ChangeResult, err error, _ string) {
	if err != nil {
		session.respondError(env, "server_error", err.Error())
		return
	}
	if result.Scope == "" {
		result.Scope = "project"
	}
	result.Complete = true
	session.respond(env, result)
	if result.OK {
		h.broadcastSnapshots(result.BoardID)
	}
}

func (h *Hub) respondProjectChange(session *Session, env Envelope, result kanban.ProjectChangeResult, err error, _ string) {
	if err != nil {
		session.respondError(env, "server_error", err.Error())
		return
	}
	session.respond(env, result)
	if result.OK {
		h.broadcastSnapshots(result.BoardID)
	}
}

func (h *Hub) respondMutation(session *Session, env Envelope, result kanban.MutationResult, err error) {
	if err != nil {
		session.respondError(env, "server_error", err.Error())
		return
	}
	session.respond(env, result)
	if result.OK {
		h.broadcastSnapshots(result.BoardID)
	}
}

func (h *Hub) assignAndRun(session *Session, env Envelope) {
	var input kanban.AssignAndRunInput
	if !decodeOrRespond(session, env, &input) {
		return
	}
	snapshot, err := h.service.Snapshot(context.Background(), env.BoardID, env.ProjectID, h.DesktopStatus())
	if err != nil {
		session.respondError(env, "server_error", err.Error())
		return
	}
	var issue *kanban.Issue
	for i := range snapshot.Issues {
		if snapshot.Issues[i].ID == input.ID {
			issue = &snapshot.Issues[i]
			break
		}
	}
	if issue == nil {
		session.respond(env, kanban.ChangeResult{OK: false, Message: "任务不存在。", BoardID: env.BoardID, ProjectID: env.ProjectID, Revision: snapshot.Revision, Complete: true, Scope: "project", Issues: snapshot.Issues, ProjectIssueStats: snapshot.ProjectIssueStats})
		return
	}
	if input.BaseIssueRevision != nil && *input.BaseIssueRevision > 0 && *input.BaseIssueRevision != issue.Revision {
		session.respond(env, kanban.ChangeResult{OK: false, Code: "conflict", Message: "任务已被其他端更新，请刷新后重试。", BoardID: env.BoardID, ProjectID: env.ProjectID, Revision: snapshot.Revision, Complete: true, Scope: "project", Issue: issue, Issues: kanban.SortIssues(snapshot.Issues), ProjectIssueStats: snapshot.ProjectIssueStats})
		return
	}
	if issue.RunID != nil && strings.TrimSpace(*issue.RunID) != "" {
		session.respond(env, kanban.ChangeResult{OK: true, Message: "任务已在运行中。", BoardID: env.BoardID, ProjectID: env.ProjectID, Revision: snapshot.Revision, Complete: true, Scope: "project", Issue: issue, Issues: kanban.SortIssues(snapshot.Issues), ProjectIssueStats: snapshot.ProjectIssueStats})
		return
	}
	agentKey := input.AgentKey
	if agentKey == nil {
		agentKey = issue.AssigneeAgentKey
	}
	resultPayload, err := h.runIdempotent(input.IdempotencyKey, func() (any, error) {
		payload := map[string]any{
			"issue":    issue,
			"agentKey": agentKey,
			"message":  buildAssistantPrompt(*issue),
		}
		response, err := h.requestDesktop("desktop.assistant.startRun", env.BoardID, env.ProjectID, input.TargetDesktopSessionID, payload)
		if err != nil {
			return kanban.ChangeResult{OK: false, Message: err.Error(), BoardID: env.BoardID, ProjectID: env.ProjectID, Revision: snapshot.Revision, Complete: true, Scope: "project", Issues: snapshot.Issues, ProjectIssueStats: snapshot.ProjectIssueStats}, nil
		}
		var result kanban.StartRunResult
		if err := json.Unmarshal(response.Payload, &result); err != nil {
			return kanban.ChangeResult{OK: false, Message: "desktop 返回格式无效。", BoardID: env.BoardID, ProjectID: env.ProjectID, Revision: snapshot.Revision, Complete: true, Scope: "project", Issues: snapshot.Issues, ProjectIssueStats: snapshot.ProjectIssueStats}, nil
		}
		change, err := h.service.StartRun(context.Background(), env.BoardID, env.ProjectID, input.ID, agentKey, result, session.actorID())
		if err != nil {
			return nil, err
		}
		return change, nil
	})
	if err != nil {
		session.respondError(env, "server_error", err.Error())
		return
	}
	result, ok := resultPayload.(kanban.ChangeResult)
	if !ok {
		session.respondError(env, "server_error", "启动 task 返回格式无效。")
		return
	}
	if !result.OK {
		session.respond(env, result)
		return
	}
	h.respondChange(session, env, result, nil, "kanban.issue.updated")
}

func (h *Hub) dispatchToDesktop(session *Session, env Envelope) {
	var payload struct {
		ID                     string        `json:"id"`
		Issue                  *kanban.Issue `json:"issue"`
		IdempotencyKey         string        `json:"idempotencyKey"`
		TargetDesktopSessionID string        `json:"targetDesktopSessionId"`
	}
	if !decodeOrRespond(session, env, &payload) {
		return
	}
	issue := payload.Issue
	if issue == nil {
		issueID := strings.TrimSpace(payload.ID)
		if issueID == "" {
			session.respondError(env, "bad_payload", "缺少要派发的 issue。")
			return
		}
		snapshot, err := h.service.Snapshot(context.Background(), env.BoardID, env.ProjectID, h.DesktopStatus())
		if err != nil {
			session.respondError(env, "server_error", err.Error())
			return
		}
		for i := range snapshot.Issues {
			if snapshot.Issues[i].ID == issueID {
				issue = &snapshot.Issues[i]
				break
			}
		}
		if issue == nil {
			session.respond(env, kanban.ChangeResult{OK: false, Message: "任务不存在。", BoardID: env.BoardID, ProjectID: env.ProjectID, Revision: snapshot.Revision, Complete: true, Scope: "project", Issues: snapshot.Issues, ProjectIssueStats: snapshot.ProjectIssueStats})
			return
		}
	}
	resultPayload, err := h.runIdempotent(payload.IdempotencyKey, func() (any, error) {
		response, err := h.requestDesktop("desktop.kanban.issue.dispatch", env.BoardID, env.ProjectID, payload.TargetDesktopSessionID, map[string]any{
			"issue": issue,
		})
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"ok":        true,
			"message":   "任务已派发到 Desktop。",
			"issue":     issue,
			"desktop":   json.RawMessage(response.Payload),
			"boardId":   env.BoardID,
			"projectId": env.ProjectID,
		}, nil
	})
	if err != nil {
		session.respondError(env, "desktop_unavailable", err.Error())
		return
	}
	session.respond(env, resultPayload)
}

func (h *Hub) forwardDesktop(session *Session, env Envelope, op string) {
	targetDesktopSessionID := ""
	var payload struct {
		TargetDesktopSessionID string `json:"targetDesktopSessionId"`
	}
	if len(env.Payload) > 0 {
		_ = json.Unmarshal(env.Payload, &payload)
		targetDesktopSessionID = strings.TrimSpace(payload.TargetDesktopSessionID)
	}
	response, err := h.requestDesktop(op, env.BoardID, env.ProjectID, targetDesktopSessionID, json.RawMessage(env.Payload))
	if err != nil {
		session.respondError(env, "desktop_unavailable", err.Error())
		return
	}
	session.send <- OutEnvelope{
		V:         ProtocolVersion,
		Type:      "rpc.res",
		ID:        env.ID,
		Op:        env.Op,
		BoardID:   env.BoardID,
		ProjectID: env.ProjectID,
		Revision:  env.Revision,
		OK:        response.OK,
		Error:     response.Error,
		Payload:   json.RawMessage(response.Payload),
	}
}

func (h *Hub) desktopHello(session *Session, env Envelope) {
	session.role = "desktop"
	var payload struct {
		Capabilities      []string           `json:"capabilities"`
		SelectedProjectID string             `json:"selectedProjectId"`
		CurrentUser       DesktopCurrentUser `json:"currentUser"`
		Scope             string             `json:"scope"`
		DeviceID          string             `json:"deviceId"`
	}
	_ = json.Unmarshal(env.Payload, &payload)
	session.capabilities = payload.Capabilities
	session.currentUser = payload.CurrentUser
	session.deviceID = strings.TrimSpace(payload.DeviceID)
	session.scope = strings.TrimSpace(payload.Scope)
	session.lastSeenAt = time.Now().UTC()
	if strings.TrimSpace(payload.SelectedProjectID) != "" {
		session.projectID = strings.TrimSpace(payload.SelectedProjectID)
	}
	h.mu.Lock()
	h.desktopSessions[session.id] = session
	h.mu.Unlock()
	if err := h.store.SaveDesktopClient(context.Background(), session.id, session.deviceID, session.currentUser.ID, session.currentUser.Name, payload.Capabilities, session.projectID); err != nil {
		h.logger.Warn("failed to save desktop client", "error", err)
	}
	session.respond(env, map[string]any{
		"ok":                true,
		"message":           "desktop 已连接。",
		"sessionId":         session.id,
		"capabilities":      payload.Capabilities,
		"selectedProjectId": session.projectID,
		"currentUser":       session.currentUser,
		"scope":             session.scope,
		"desktopStatus":     h.DesktopStatus(),
	})
	session.sendSnapshot()
	h.broadcastDesktopStatus()
}

func (h *Hub) desktopAssistantEvent(session *Session, env Envelope) {
	var event kanban.AssistantEvent
	if !decodeOrRespond(session, env, &event) {
		return
	}
	result, err := h.service.SyncAssistantEvent(context.Background(), env.BoardID, env.ProjectID, event, "desktop")
	h.respondChange(session, env, result, err, "kanban.issue.updated")
}

func (h *Hub) desktopOnlineList(session *Session, env Envelope) {
	status := h.DesktopStatus()
	agentResults := make(map[string]desktopAgentLoadResult, len(status.Sessions))
	for _, desktopSession := range status.Sessions {
		response, err := h.requestDesktop("desktop.assistant.listAgents", env.BoardID, env.ProjectID, desktopSession.SessionID, map[string]any{})
		if err != nil {
			agentResults[desktopSession.SessionID] = desktopAgentLoadResult{err: err}
			continue
		}
		var payload desktopAgentListPayload
		if err := json.Unmarshal(response.Payload, &payload); err != nil {
			agentResults[desktopSession.SessionID] = desktopAgentLoadResult{err: errors.New("desktop agent 列表返回格式无效。")}
			continue
		}
		agents := payload.Items
		if len(agents) == 0 {
			agents = payload.Agents
		}
		agentResults[desktopSession.SessionID] = desktopAgentLoadResult{agents: agents}
	}
	session.respond(env, buildDesktopOnlineList(status, agentResults))
}

func (h *Hub) requestDesktop(op string, boardID string, projectID string, targetDesktopSessionID string, payload any) (Envelope, error) {
	desktop, err := h.pickDesktop(targetDesktopSessionID)
	if err != nil {
		return Envelope{}, err
	}
	id := randomID()
	ch := make(chan Envelope, 1)
	h.mu.Lock()
	h.pending[id] = ch
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		delete(h.pending, id)
		h.mu.Unlock()
	}()
	desktop.send <- OutEnvelope{
		V:         ProtocolVersion,
		Type:      "rpc.req",
		ID:        id,
		Op:        op,
		Role:      "server",
		BoardID:   boardID,
		ProjectID: projectID,
		Payload:   payload,
	}
	select {
	case response := <-ch:
		if response.OK != nil && !*response.OK {
			if response.Error != nil && response.Error.Message != "" {
				return response, errors.New(response.Error.Message)
			}
			return response, errors.New("desktop 操作失败。")
		}
		return response, nil
	case <-time.After(requestTimeout):
		return Envelope{}, errors.New("desktop 操作超时。")
	}
}

func (h *Hub) completePending(env Envelope) {
	h.mu.RLock()
	ch := h.pending[env.ID]
	h.mu.RUnlock()
	if ch == nil {
		return
	}
	select {
	case ch <- env:
	default:
	}
}

func (h *Hub) runIdempotent(key string, fn func() (any, error)) (any, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return fn()
	}

	h.mu.Lock()
	if existing := h.idempotent[key]; existing != nil {
		h.mu.Unlock()
		<-existing.done
		return existing.result, existing.err
	}
	item := &idempotentRequest{done: make(chan struct{})}
	h.idempotent[key] = item
	h.mu.Unlock()

	item.result, item.err = fn()
	close(item.done)

	go func() {
		time.Sleep(2 * time.Minute)
		h.mu.Lock()
		if h.idempotent[key] == item {
			delete(h.idempotent, key)
		}
		h.mu.Unlock()
	}()

	return item.result, item.err
}

func (h *Hub) broadcast(env OutEnvelope) {
	h.mu.RLock()
	sessions := make([]*Session, 0, len(h.sessions))
	for _, session := range h.sessions {
		sessions = append(sessions, session)
	}
	h.mu.RUnlock()
	for _, session := range sessions {
		select {
		case session.send <- env:
		default:
			h.unregister(session)
		}
	}
}

func (h *Hub) broadcastSnapshots(boardID string) {
	h.mu.RLock()
	sessions := make([]*Session, 0, len(h.sessions))
	for _, session := range h.sessions {
		sessions = append(sessions, session)
	}
	h.mu.RUnlock()
	for _, session := range sessions {
		snapshot, err := h.snapshotForSession(session, boardID, session.projectID)
		if err != nil {
			h.logger.Warn("failed to build kanban snapshot", "error", err, "session", session.id)
			continue
		}
		session.projectID = snapshot.ProjectID
		select {
		case session.send <- OutEnvelope{
			V:         ProtocolVersion,
			Type:      "event",
			Op:        "kanban.snapshot",
			Role:      "server",
			BoardID:   snapshot.BoardID,
			ProjectID: snapshot.ProjectID,
			Revision:  snapshot.Revision,
			OK:        boolPtr(true),
			Payload:   snapshot,
		}:
		default:
			h.unregister(session)
		}
	}
}

func (h *Hub) broadcastDesktopStatus() {
	h.broadcast(OutEnvelope{
		V:       ProtocolVersion,
		Type:    "event",
		Op:      "kanban.desktop.status",
		BoardID: kanban.DefaultBoardID,
		Payload: h.DesktopStatus(),
	})
}

func (h *Hub) DesktopStatus() kanban.DesktopStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	sessions := make([]kanban.DesktopSessionStatus, 0, len(h.desktopSessions))
	for _, session := range h.desktopSessions {
		lastSeenAt := session.lastSeenAt
		if lastSeenAt.IsZero() {
			lastSeenAt = time.Now().UTC()
		}
		sessions = append(sessions, kanban.DesktopSessionStatus{
			SessionID:         session.id,
			DeviceID:          session.deviceID,
			CurrentUserID:     strings.TrimSpace(session.currentUser.ID),
			CurrentUserName:   strings.TrimSpace(session.currentUser.Name),
			SelectedProjectID: session.projectID,
			Capabilities:      append([]string(nil), session.capabilities...),
			LastSeenAt:        lastSeenAt,
		})
	}
	status := kanban.DesktopStatus{Online: len(sessions) > 0, Capabilities: []string{}, Sessions: sessions}
	if len(sessions) > 0 {
		first := sessions[0]
		status.SessionID = first.SessionID
		status.Capabilities = append([]string(nil), first.Capabilities...)
		status.SelectedProjectID = first.SelectedProjectID
	}
	return status
}

func buildDesktopOnlineList(status kanban.DesktopStatus, agentResults map[string]desktopAgentLoadResult) kanban.DesktopOnlineListResult {
	type onlineDeviceAccumulator struct {
		device       kanban.DesktopOnlineDevice
		capabilities map[string]bool
		agents       map[string]bool
	}

	order := []string{}
	devices := map[string]*onlineDeviceAccumulator{}
	for _, desktopSession := range status.Sessions {
		deviceID := strings.TrimSpace(desktopSession.DeviceID)
		if deviceID == "" {
			deviceID = "session:" + strings.TrimSpace(desktopSession.SessionID)
		}
		item := devices[deviceID]
		if item == nil {
			item = &onlineDeviceAccumulator{
				device: kanban.DesktopOnlineDevice{
					DeviceID:          deviceID,
					CurrentUserID:     strings.TrimSpace(desktopSession.CurrentUserID),
					CurrentUserName:   strings.TrimSpace(desktopSession.CurrentUserName),
					SelectedProjectID: strings.TrimSpace(desktopSession.SelectedProjectID),
					Capabilities:      []string{},
					LastSeenAt:        desktopSession.LastSeenAt,
					Sessions:          []kanban.DesktopSessionStatus{},
					Agents:            []kanban.DesktopAgentOption{},
				},
				capabilities: map[string]bool{},
				agents:       map[string]bool{},
			}
			devices[deviceID] = item
			order = append(order, deviceID)
		}
		item.device.Sessions = append(item.device.Sessions, desktopSession)
		if desktopSession.LastSeenAt.After(item.device.LastSeenAt) {
			item.device.LastSeenAt = desktopSession.LastSeenAt
		}
		for _, capability := range desktopSession.Capabilities {
			capability = strings.TrimSpace(capability)
			if capability == "" || item.capabilities[capability] {
				continue
			}
			item.capabilities[capability] = true
			item.device.Capabilities = append(item.device.Capabilities, capability)
		}
		result := agentResults[desktopSession.SessionID]
		if result.err != nil && item.device.AgentError == "" {
			item.device.AgentError = result.err.Error()
		}
		for _, agent := range result.agents {
			agent.AgentKey = strings.TrimSpace(agent.AgentKey)
			agent.DisplayName = strings.TrimSpace(agent.DisplayName)
			agent.Role = strings.TrimSpace(agent.Role)
			if agent.AgentKey == "" || item.agents[agent.AgentKey] {
				continue
			}
			if agent.DisplayName == "" {
				agent.DisplayName = agent.AgentKey
			}
			item.agents[agent.AgentKey] = true
			item.device.Agents = append(item.device.Agents, agent)
		}
	}

	devicesList := make([]kanban.DesktopOnlineDevice, 0, len(order))
	agentCount := 0
	sessionCount := 0
	for _, deviceID := range order {
		device := devices[deviceID].device
		sessionCount += len(device.Sessions)
		agentCount += len(device.Agents)
		devicesList = append(devicesList, device)
	}
	return kanban.DesktopOnlineListResult{
		OK:           true,
		Online:       status.Online,
		DeviceCount:  len(devicesList),
		SessionCount: sessionCount,
		AgentCount:   agentCount,
		Devices:      devicesList,
	}
}

func (h *Hub) pickDesktop(targetDesktopSessionID string) (*Session, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	targetDesktopSessionID = strings.TrimSpace(targetDesktopSessionID)
	if targetDesktopSessionID != "" {
		session := h.desktopSessions[targetDesktopSessionID]
		if session == nil {
			return nil, errors.New("目标 Desktop 未在线。")
		}
		return session, nil
	}
	if len(h.desktopSessions) == 0 {
		return nil, errors.New("desktop 未在线。")
	}
	if len(h.desktopSessions) > 1 {
		return nil, errors.New("多个 Desktop 在线，请选择目标 Desktop。")
	}
	for _, session := range h.desktopSessions {
		return session, nil
	}
	return nil, errors.New("desktop 未在线。")
}

func (h *Hub) authorized(r *http.Request) bool {
	if h.cfg.Token == "" {
		return true
	}
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		value := strings.TrimSpace(r.Header.Get("Authorization"))
		token = strings.TrimPrefix(value, "Bearer ")
	}
	return token == h.cfg.Token
}

func (h *Hub) checkOrigin(r *http.Request) bool {
	if len(h.cfg.AllowedOrigins) == 0 {
		return true
	}
	origin := r.Header.Get("Origin")
	for _, allowed := range h.cfg.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

func decodeOrRespond[T any](session *Session, env Envelope, target *T) bool {
	if len(env.Payload) == 0 {
		env.Payload = []byte("{}")
	}
	if err := json.Unmarshal(env.Payload, target); err != nil {
		session.respondError(env, "bad_payload", "请求 payload 格式无效。")
		return false
	}
	return true
}

func buildAssistantPrompt(issue kanban.Issue) string {
	parts := []string{
		"请根据以下 ZenMind Kanban issue 执行工作，并在需要用户确认时明确说明。",
		"不要修改 issue 编号，完成后由系统回写 issue 状态。",
		"issue 编号：" + issue.ID,
		"标题：" + issue.Title,
		"状态：" + string(issue.Status),
		"优先级：" + string(issue.Priority),
	}
	if strings.TrimSpace(issue.ProjectPath) != "" {
		parts = append(parts, "项目："+strings.TrimSpace(issue.ProjectPath))
	}
	if strings.TrimSpace(issue.Description) != "" {
		parts = append(parts, "描述："+strings.TrimSpace(issue.Description))
	}
	return strings.Join(parts, "\n")
}

func randomID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(bytes[:])
}
