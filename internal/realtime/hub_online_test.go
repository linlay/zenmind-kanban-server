package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"zenmind-kanban-server/internal/config"
	"zenmind-kanban-server/internal/kanban"
	"zenmind-kanban-server/internal/store"
)

func TestBuildDesktopOnlineListGroupsSessionsByDeviceID(t *testing.T) {
	now := time.Date(2026, 6, 8, 9, 0, 0, 0, time.UTC)
	status := kanban.DesktopStatus{
		Online: true,
		Sessions: []kanban.DesktopSessionStatus{
			{
				SessionID:         "session-a",
				DeviceID:          "device-1",
				CurrentUserID:     "user-1",
				CurrentUserName:   "Alice",
				SelectedProjectID: "project-a",
				Capabilities:      []string{"desktop.assistant.listAgents", "kanban.issue.dispatch"},
				LastSeenAt:        now,
			},
			{
				SessionID:         "session-b",
				DeviceID:          "device-1",
				CurrentUserID:     "user-1",
				CurrentUserName:   "Alice",
				SelectedProjectID: "project-a",
				Capabilities:      []string{"desktop.assistant.listAgents", "desktop.automation.sync"},
				LastSeenAt:        now.Add(2 * time.Minute),
			},
			{
				SessionID:         "session-c",
				CurrentUserName:   "Fallback",
				SelectedProjectID: "default",
				Capabilities:      []string{"desktop.assistant.listAgents"},
				LastSeenAt:        now.Add(time.Minute),
			},
		},
	}
	agents := map[string]desktopAgentLoadResult{
		"session-a": {
			agents: []kanban.DesktopAgentOption{
				{AgentKey: "zenmi", DisplayName: "小宅", Role: "平台总管"},
				{AgentKey: "webOperator", DisplayName: "网驭", Role: "Desktop 内嵌网站操作助手"},
			},
		},
		"session-b": {
			agents: []kanban.DesktopAgentOption{
				{AgentKey: "zenmi", DisplayName: "小宅", Role: "平台总管"},
				{AgentKey: "dailyOfficeProAssistant", DisplayName: "文衡", Role: "办公助手Pro"},
			},
		},
		"session-c": {
			err: errors.New("agent-platform unavailable"),
		},
	}

	result := buildDesktopOnlineList(status, agents)

	if !result.OK || !result.Online {
		t.Fatalf("expected online ok result, got %#v", result)
	}
	if result.SessionCount != 3 {
		t.Fatalf("expected 3 sessions, got %d", result.SessionCount)
	}
	if result.AgentCount != 3 {
		t.Fatalf("expected 3 unique agents, got %d", result.AgentCount)
	}
	if len(result.Devices) != 2 {
		t.Fatalf("expected 2 devices, got %#v", result.Devices)
	}

	device := result.Devices[0]
	if device.DeviceID != "device-1" {
		t.Fatalf("expected first device to be device-1, got %q", device.DeviceID)
	}
	if len(device.Sessions) != 2 {
		t.Fatalf("expected device-1 to contain 2 sessions, got %#v", device.Sessions)
	}
	if device.LastSeenAt != now.Add(2*time.Minute) {
		t.Fatalf("expected latest lastSeenAt, got %s", device.LastSeenAt)
	}
	if len(device.Capabilities) != 3 {
		t.Fatalf("expected merged capabilities, got %#v", device.Capabilities)
	}
	if len(device.Agents) != 3 {
		t.Fatalf("expected deduped agents, got %#v", device.Agents)
	}

	fallback := result.Devices[1]
	if fallback.DeviceID != "session:session-c" {
		t.Fatalf("expected fallback device id, got %q", fallback.DeviceID)
	}
	if fallback.AgentError != "agent-platform unavailable" {
		t.Fatalf("expected per-device agent error, got %q", fallback.AgentError)
	}
}

func TestHubAcceptsNewProtocolIssueCreateOp(t *testing.T) {
	hub, closeStore := newTestHub(t)
	defer closeStore()
	session := &Session{
		id:        "web-test",
		role:      "web",
		board:     kanban.DefaultBoardID,
		projectID: kanban.DefaultProjectID,
		send:      make(chan OutEnvelope, 4),
	}
	payload, err := json.Marshal(kanban.IssueInput{Title: "New protocol task"})
	if err != nil {
		t.Fatal(err)
	}

	hub.handle(session, Envelope{
		V:         ProtocolVersion,
		Type:      "rpc.req",
		ID:        "req-1",
		Op:        "issue.create",
		Role:      "web",
		BoardID:   kanban.DefaultBoardID,
		ProjectID: kanban.DefaultProjectID,
		Payload:   payload,
	})

	response := <-session.send
	if response.Op != "issue.create" {
		t.Fatalf("expected response op issue.create, got %q", response.Op)
	}
	if response.OK == nil || !*response.OK {
		t.Fatalf("expected ok response, got %#v", response)
	}
	result, ok := response.Payload.(kanban.ChangeResult)
	if !ok {
		t.Fatalf("expected ChangeResult payload, got %T", response.Payload)
	}
	if !result.OK || result.Issue == nil || result.Issue.Title != "New protocol task" {
		t.Fatalf("expected created issue result, got %#v", result)
	}
}

func TestHubHandlesProjectBindingRPCs(t *testing.T) {
	hub, closeStore := newTestHub(t)
	defer closeStore()
	session := &Session{
		id:        "web-test",
		role:      "web",
		board:     kanban.DefaultBoardID,
		projectID: kanban.DefaultProjectID,
		send:      make(chan OutEnvelope, 8),
	}
	createPayload, err := json.Marshal(kanban.ProjectBindingInput{
		ProjectID:        kanban.DefaultProjectID,
		DeviceID:         "device-1",
		CurrentUserID:    "user-1",
		LocalProjectID:   "local-alpha",
		LocalDisplayName: "/Users/jialin/Desktop/local-alpha",
	})
	if err != nil {
		t.Fatal(err)
	}

	hub.handle(session, Envelope{
		V:         ProtocolVersion,
		Type:      "rpc.req",
		ID:        "bind-create",
		Op:        "projectBinding.create",
		Role:      "web",
		BoardID:   kanban.DefaultBoardID,
		ProjectID: kanban.DefaultProjectID,
		Payload:   createPayload,
	})
	createResponse := <-session.send
	if createResponse.OK == nil || !*createResponse.OK {
		t.Fatalf("expected binding create ok response, got %#v", createResponse)
	}
	createResult, ok := createResponse.Payload.(kanban.ProjectBindingResult)
	if !ok || !createResult.OK || createResult.Binding == nil {
		t.Fatalf("expected ProjectBindingResult payload, got %#v", createResponse.Payload)
	}
	if createResult.Binding.LocalDisplayName != "local-alpha" {
		t.Fatalf("expected sanitized local display name, got %q", createResult.Binding.LocalDisplayName)
	}

	listPayload, err := json.Marshal(map[string]string{"projectId": kanban.DefaultProjectID})
	if err != nil {
		t.Fatal(err)
	}
	hub.handle(session, Envelope{
		V:         ProtocolVersion,
		Type:      "rpc.req",
		ID:        "bind-list",
		Op:        "projectBinding.list",
		Role:      "web",
		BoardID:   kanban.DefaultBoardID,
		ProjectID: kanban.DefaultProjectID,
		Payload:   listPayload,
	})
	listResponse := <-session.send
	listResult, ok := listResponse.Payload.(kanban.ProjectBindingResult)
	if listResponse.OK == nil || !*listResponse.OK || !ok || len(listResult.Bindings) != 1 {
		t.Fatalf("expected one listed binding, got %#v", listResponse)
	}

	deletePayload, err := json.Marshal(map[string]string{"id": createResult.Binding.ID})
	if err != nil {
		t.Fatal(err)
	}
	hub.handle(session, Envelope{
		V:         ProtocolVersion,
		Type:      "rpc.req",
		ID:        "bind-delete",
		Op:        "projectBinding.delete",
		Role:      "web",
		BoardID:   kanban.DefaultBoardID,
		ProjectID: kanban.DefaultProjectID,
		Payload:   deletePayload,
	})
	deleteResponse := <-session.send
	deleteResult, ok := deleteResponse.Payload.(kanban.ProjectBindingResult)
	if deleteResponse.OK == nil || !*deleteResponse.OK || !ok || !deleteResult.OK {
		t.Fatalf("expected binding delete ok response, got %#v", deleteResponse)
	}
}

func TestHubOriginCheckAllowsNativeDesktopWithoutOrigin(t *testing.T) {
	hub := &Hub{cfg: config.Config{AllowedOrigins: []string{"http://127.0.0.1:5174"}}}

	desktopRequest := httptest.NewRequest(http.MethodGet, "/ws?role=desktop", nil)
	if !hub.checkOrigin(desktopRequest) {
		t.Fatal("expected native desktop websocket without Origin to be allowed")
	}

	webRequest := httptest.NewRequest(http.MethodGet, "/ws?role=web", nil)
	if hub.checkOrigin(webRequest) {
		t.Fatal("expected web websocket without Origin to stay blocked")
	}

	browserRequest := httptest.NewRequest(http.MethodGet, "/ws?role=web", nil)
	browserRequest.Header.Set("Origin", "http://127.0.0.1:5174")
	if !hub.checkOrigin(browserRequest) {
		t.Fatal("expected allowed browser Origin to be accepted")
	}
}

func newTestHub(t *testing.T) (*Hub, func()) {
	t.Helper()
	sqliteStore, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "kanban.db"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := sqliteStore.SeedWorkflowCatalog(ctx); err != nil {
		t.Fatal(err)
	}
	if err := sqliteStore.EnsureDefaultProject(ctx); err != nil {
		t.Fatal(err)
	}
	if err := sqliteStore.EnsureDefaultBoard(ctx); err != nil {
		t.Fatal(err)
	}
	service := kanban.NewService(sqliteStore)
	hub := NewHub(config.Config{AllowedOrigins: []string{"*"}}, service, sqliteStore, slog.New(slog.NewTextHandler(io.Discard, nil)))
	return hub, func() {
		_ = sqliteStore.Close()
	}
}
