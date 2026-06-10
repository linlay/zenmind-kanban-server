package realtime

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"zenmind-kanban-server/internal/kanban"
)

func registerTestDesktop(t *testing.T, hub *Hub, sessionID string, deviceID string) *Session {
	t.Helper()
	desktop := &Session{
		id:        sessionID,
		role:      "desktop",
		board:     kanban.DefaultBoardID,
		projectID: kanban.DefaultProjectID,
		hub:       hub,
		send:      make(chan OutEnvelope, 16),
		closed:    make(chan struct{}),
	}
	hub.register(desktop)
	helloPayload, err := json.Marshal(map[string]any{
		"deviceId":          deviceID,
		"selectedProjectId": kanban.DefaultProjectID,
		"capabilities":      []string{"desktop.assistant.startRun", "desktop.issue.sync"},
		"currentUser":       map[string]string{"id": "desktop-user", "name": "Desktop User"},
	})
	if err != nil {
		t.Fatal(err)
	}
	hub.handle(desktop, Envelope{
		V:         ProtocolVersion,
		Type:      "rpc.req",
		ID:        "hello-" + sessionID,
		Op:        "desktop.hello",
		Role:      "desktop",
		BoardID:   kanban.DefaultBoardID,
		ProjectID: kanban.DefaultProjectID,
		Payload:   helloPayload,
	})
	<-desktop.send // hello 响应
	<-desktop.send // 初始快照
	drainSession(desktop)
	return desktop
}

func drainSession(session *Session) {
	for {
		select {
		case <-session.send:
		default:
			return
		}
	}
}

func createHubBinding(t *testing.T, hub *Hub, deviceID string, syncPolicy string, controlMode string) kanban.ProjectBinding {
	t.Helper()
	result, err := hub.service.CreateProjectBinding(context.Background(), kanban.ProjectBindingInput{
		ProjectID:      kanban.DefaultProjectID,
		DeviceID:       deviceID,
		LocalProjectID: "local-" + deviceID,
		SyncPolicy:     syncPolicy,
		ControlMode:    controlMode,
	}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Binding == nil {
		t.Fatalf("expected binding create to succeed: %#v", result)
	}
	return *result.Binding
}

func createHubIssue(t *testing.T, hub *Hub, title string) kanban.Issue {
	t.Helper()
	created, err := hub.service.CreateIssue(context.Background(), kanban.DefaultBoardID, kanban.IssueInput{Title: title}, "web")
	if err != nil {
		t.Fatal(err)
	}
	if !created.OK || created.Issue == nil {
		t.Fatalf("expected issue create to succeed: %#v", created)
	}
	return *created.Issue
}

func TestHubSnapshotAppliesSelectBindingScope(t *testing.T) {
	hub, closeStore := newTestHub(t)
	defer closeStore()
	issueA := createHubIssue(t, hub, "selected issue")
	createHubIssue(t, hub, "unselected issue")
	desktop := registerTestDesktop(t, hub, "desktop-select", "device-select")
	binding := createHubBinding(t, hub, "device-select", "select", "dispatch")
	if _, err := hub.service.SetProjectBindingIssues(context.Background(), kanban.ProjectBindingIssuesSetInput{
		BindingID: binding.ID,
		IssueIDs:  []string{issueA.ID},
	}, "test"); err != nil {
		t.Fatal(err)
	}

	snapshot, err := hub.snapshotForSession(desktop, kanban.DefaultBoardID, kanban.DefaultProjectID)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Issues) != 1 || snapshot.Issues[0].ID != issueA.ID {
		t.Fatalf("expected select scope to keep only pinned issue, got %#v", snapshot.Issues)
	}
	if !snapshot.Complete {
		t.Fatalf("expected select-scoped snapshot to stay complete for tombstoning")
	}
}

func TestHubSnapshotAppliesFutureBindingScope(t *testing.T) {
	hub, closeStore := newTestHub(t)
	defer closeStore()
	oldIssue := createHubIssue(t, hub, "old issue")
	desktop := registerTestDesktop(t, hub, "desktop-future", "device-future")
	createHubBinding(t, hub, "device-future", "future", "dispatch")
	// 绑定后再创建的 issue 应被同步
	time.Sleep(5 * time.Millisecond)
	newIssue := createHubIssue(t, hub, "new issue")

	snapshot, err := hub.snapshotForSession(desktop, kanban.DefaultBoardID, kanban.DefaultProjectID)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Issues) != 1 || snapshot.Issues[0].ID != newIssue.ID {
		t.Fatalf("expected future scope to keep only post-binding issue, got %#v", snapshot.Issues)
	}
	for _, issue := range snapshot.Issues {
		if issue.ID == oldIssue.ID {
			t.Fatalf("expected pre-binding issue to be filtered out")
		}
	}
}

func TestHubSnapshotDisabledBindingSendsIncomplete(t *testing.T) {
	hub, closeStore := newTestHub(t)
	defer closeStore()
	createHubIssue(t, hub, "hidden issue")
	desktop := registerTestDesktop(t, hub, "desktop-disabled", "device-disabled")
	createHubBinding(t, hub, "device-disabled", "all", "disabled")

	snapshot, err := hub.snapshotForSession(desktop, kanban.DefaultBoardID, kanban.DefaultProjectID)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Issues) != 0 {
		t.Fatalf("expected disabled binding to hide issues, got %#v", snapshot.Issues)
	}
	if snapshot.Complete {
		t.Fatalf("expected incomplete snapshot to suppress tombstoning")
	}
}

func TestHubSnapshotWebSessionUnaffectedByBindingScope(t *testing.T) {
	hub, closeStore := newTestHub(t)
	defer closeStore()
	createHubIssue(t, hub, "issue one")
	createHubIssue(t, hub, "issue two")
	registerTestDesktop(t, hub, "desktop-x", "device-x")
	createHubBinding(t, hub, "device-x", "select", "dispatch")

	web := &Session{
		id:        "web-scope",
		role:      "web",
		board:     kanban.DefaultBoardID,
		projectID: kanban.DefaultProjectID,
		send:      make(chan OutEnvelope, 8),
	}
	snapshot, err := hub.snapshotForSession(web, kanban.DefaultBoardID, kanban.DefaultProjectID)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Issues) != 2 {
		t.Fatalf("expected web session to see all issues, got %#v", snapshot.Issues)
	}
}

func TestHubDispatchAllowsUnboundProject(t *testing.T) {
	hub, closeStore := newTestHub(t)
	defer closeStore()
	issue := createHubIssue(t, hub, "unbound dispatch")
	desktop := registerTestDesktop(t, hub, "desktop-unbound", "device-unbound")

	web := &Session{
		id:        "web-unbound",
		role:      "web",
		board:     kanban.DefaultBoardID,
		projectID: kanban.DefaultProjectID,
		send:      make(chan OutEnvelope, 8),
	}
	dispatchPayload, err := json.Marshal(map[string]any{
		"id":                     issue.ID,
		"targetDesktopSessionId": desktop.id,
	})
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		hub.handle(web, Envelope{
			V:         ProtocolVersion,
			Type:      "rpc.req",
			ID:        "dispatch-unbound",
			Op:        "kanban.issue.dispatchToDesktop",
			Role:      "web",
			BoardID:   kanban.DefaultBoardID,
			ProjectID: kanban.DefaultProjectID,
			Payload:   dispatchPayload,
		})
		close(done)
	}()
	// 模拟 desktop 应答派发请求
	var desktopRequest OutEnvelope
	select {
	case desktopRequest = <-desktop.send:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for desktop dispatch request")
	}
	if desktopRequest.Op != "desktop.kanban.issue.dispatch" {
		t.Fatalf("expected dispatch op, got %q", desktopRequest.Op)
	}
	payloadMap, ok := desktopRequest.Payload.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %T", desktopRequest.Payload)
	}
	if _, hasBinding := payloadMap["binding"]; hasBinding {
		t.Fatalf("expected unbound dispatch payload without binding context")
	}
	responsePayload, err := json.Marshal(map[string]any{"ok": true})
	if err != nil {
		t.Fatal(err)
	}
	hub.handle(desktop, Envelope{
		V:       ProtocolVersion,
		Type:    "rpc.res",
		ID:      desktopRequest.ID,
		Op:      desktopRequest.Op,
		OK:      boolPtr(true),
		Payload: responsePayload,
	})
	<-done
	webResponse := <-web.send
	if webResponse.OK == nil || !*webResponse.OK {
		t.Fatalf("expected unbound dispatch to succeed, got %#v", webResponse)
	}
}

func TestHubDispatchRejectsObserveBinding(t *testing.T) {
	hub, closeStore := newTestHub(t)
	defer closeStore()
	issue := createHubIssue(t, hub, "observe dispatch")
	desktop := registerTestDesktop(t, hub, "desktop-observe", "device-observe")
	createHubBinding(t, hub, "device-observe", "all", "observe")

	web := &Session{
		id:        "web-observe",
		role:      "web",
		board:     kanban.DefaultBoardID,
		projectID: kanban.DefaultProjectID,
		send:      make(chan OutEnvelope, 8),
	}
	dispatchPayload, err := json.Marshal(map[string]any{
		"id":                     issue.ID,
		"targetDesktopSessionId": desktop.id,
	})
	if err != nil {
		t.Fatal(err)
	}
	hub.handle(web, Envelope{
		V:         ProtocolVersion,
		Type:      "rpc.req",
		ID:        "dispatch-observe",
		Op:        "kanban.issue.dispatchToDesktop",
		Role:      "web",
		BoardID:   kanban.DefaultBoardID,
		ProjectID: kanban.DefaultProjectID,
		Payload:   dispatchPayload,
	})
	webResponse := <-web.send
	payloadMap, ok := webResponse.Payload.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %T", webResponse.Payload)
	}
	if okValue, _ := payloadMap["ok"].(bool); okValue {
		t.Fatalf("expected observe binding to reject dispatch, got %#v", payloadMap)
	}
	if code, _ := payloadMap["code"].(string); code != "binding_forbidden" {
		t.Fatalf("expected binding_forbidden code, got %#v", payloadMap)
	}
}

func TestHubDispatchInjectsBindingContextAndPinsIssue(t *testing.T) {
	hub, closeStore := newTestHub(t)
	defer closeStore()
	issue := createHubIssue(t, hub, "bound dispatch")
	desktop := registerTestDesktop(t, hub, "desktop-bound", "device-bound")
	binding := createHubBinding(t, hub, "device-bound", "select", "dispatch")

	web := &Session{
		id:        "web-bound",
		role:      "web",
		board:     kanban.DefaultBoardID,
		projectID: kanban.DefaultProjectID,
		send:      make(chan OutEnvelope, 8),
	}
	dispatchPayload, err := json.Marshal(map[string]any{
		"id":                     issue.ID,
		"targetDesktopSessionId": desktop.id,
	})
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		hub.handle(web, Envelope{
			V:         ProtocolVersion,
			Type:      "rpc.req",
			ID:        "dispatch-bound",
			Op:        "kanban.issue.dispatchToDesktop",
			Role:      "web",
			BoardID:   kanban.DefaultBoardID,
			ProjectID: kanban.DefaultProjectID,
			Payload:   dispatchPayload,
		})
		close(done)
	}()
	var desktopRequest OutEnvelope
	select {
	case desktopRequest = <-desktop.send:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for desktop dispatch request")
	}
	payloadMap, ok := desktopRequest.Payload.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %T", desktopRequest.Payload)
	}
	bindingContext, ok := payloadMap["binding"].(map[string]any)
	if !ok {
		t.Fatalf("expected binding context in dispatch payload, got %#v", payloadMap)
	}
	if bindingContext["localProjectId"] != binding.LocalProjectID {
		t.Fatalf("expected localProjectId %q, got %#v", binding.LocalProjectID, bindingContext)
	}
	responsePayload, err := json.Marshal(map[string]any{"ok": true})
	if err != nil {
		t.Fatal(err)
	}
	hub.handle(desktop, Envelope{
		V:       ProtocolVersion,
		Type:    "rpc.res",
		ID:      desktopRequest.ID,
		Op:      desktopRequest.Op,
		OK:      boolPtr(true),
		Payload: responsePayload,
	})
	<-done
	<-web.send

	// 派发后 issue 应被 pin,select 策略快照对该设备可见
	snapshot, err := hub.snapshotForSession(desktop, kanban.DefaultBoardID, kanban.DefaultProjectID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range snapshot.Issues {
		if item.ID == issue.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected dispatched issue pinned into select scope, got %#v", snapshot.Issues)
	}
}

func TestHubDesktopProjectSelectUpdatesSessionAndBroadcasts(t *testing.T) {
	hub, closeStore := newTestHub(t)
	defer closeStore()
	desktop := registerTestDesktop(t, hub, "desktop-switch", "device-switch")

	// 创建一个新项目供切换
	projectResult, err := hub.service.CreateProject(context.Background(), kanban.ProjectInput{Name: "Sub Project"}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !projectResult.OK || projectResult.Project == nil {
		t.Fatalf("expected project create to succeed: %#v", projectResult)
	}
	newProjectID := projectResult.Project.ID
	drainSession(desktop)

	selectPayload, err := json.Marshal(map[string]string{"selectedProjectId": newProjectID})
	if err != nil {
		t.Fatal(err)
	}
	hub.handle(desktop, Envelope{
		V:         ProtocolVersion,
		Type:      "rpc.req",
		ID:        "project-select",
		Op:        "desktop.project.select",
		Role:      "desktop",
		BoardID:   kanban.DefaultBoardID,
		ProjectID: kanban.DefaultProjectID,
		Payload:   selectPayload,
	})
	response := <-desktop.send
	if response.OK == nil || !*response.OK {
		t.Fatalf("expected project select ok, got %#v", response)
	}
	if desktop.projectID != newProjectID {
		t.Fatalf("expected session projectID updated to %q, got %q", newProjectID, desktop.projectID)
	}
	// 后续应收到新 scope 快照
	snapshotEnv := <-desktop.send
	if snapshotEnv.Op != "kanban.snapshot" {
		t.Fatalf("expected snapshot after project select, got %q", snapshotEnv.Op)
	}
	if snapshotEnv.ProjectID != newProjectID {
		t.Fatalf("expected snapshot scoped to %q, got %q", newProjectID, snapshotEnv.ProjectID)
	}
}

func TestHubDesktopProjectSelectRejectsWebSession(t *testing.T) {
	hub, closeStore := newTestHub(t)
	defer closeStore()
	web := &Session{
		id:        "web-no-select",
		role:      "web",
		board:     kanban.DefaultBoardID,
		projectID: kanban.DefaultProjectID,
		send:      make(chan OutEnvelope, 4),
	}
	selectPayload, err := json.Marshal(map[string]string{"selectedProjectId": "any"})
	if err != nil {
		t.Fatal(err)
	}
	hub.handle(web, Envelope{
		V:       ProtocolVersion,
		Type:    "rpc.req",
		ID:      "web-select",
		Op:      "desktop.project.select",
		Role:    "web",
		BoardID: kanban.DefaultBoardID,
		Payload: selectPayload,
	})
	response := <-web.send
	if response.OK == nil || *response.OK {
		t.Fatalf("expected web project select to be rejected, got %#v", response)
	}
}

func TestHubDesktopIssueSyncCreatesIssuesAndBroadcastsOnce(t *testing.T) {
	hub, closeStore := newTestHub(t)
	defer closeStore()
	desktop := registerTestDesktop(t, hub, "desktop-sync", "device-sync")
	createHubBinding(t, hub, "device-sync", "all", "dispatch")
	drainSession(desktop)

	syncPayload, err := json.Marshal(kanban.DesktopIssueSyncInput{
		ProjectID:      kanban.DefaultProjectID,
		LocalProjectID: "local-device-sync",
		Upserts: []kanban.DesktopIssueSyncUpsert{
			{LocalIssueID: "local-1", Input: kanban.IssueInput{Title: "上行任务一"}},
			{LocalIssueID: "local-2", Input: kanban.IssueInput{Title: "上行任务二"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	hub.handle(desktop, Envelope{
		V:         ProtocolVersion,
		Type:      "rpc.req",
		ID:        "issue-sync",
		Op:        "desktop.issue.sync",
		Role:      "desktop",
		BoardID:   kanban.DefaultBoardID,
		ProjectID: kanban.DefaultProjectID,
		Payload:   syncPayload,
	})
	response := <-desktop.send
	if response.OK == nil || !*response.OK {
		t.Fatalf("expected issue sync ok, got %#v", response)
	}
	result, ok := response.Payload.(kanban.DesktopIssueSyncResult)
	if !ok {
		t.Fatalf("expected DesktopIssueSyncResult payload, got %T", response.Payload)
	}
	if len(result.Results) != 2 {
		t.Fatalf("expected 2 sync results, got %#v", result.Results)
	}
	for _, item := range result.Results {
		if item.Status != "created" || item.RemoteIssueID == "" {
			t.Fatalf("expected created mapping, got %#v", item)
		}
	}
	// 同步后应广播快照(desktop 自己也会收到一次)
	broadcastCount := 0
	timeout := time.After(time.Second)
	for broadcastCount == 0 {
		select {
		case env := <-desktop.send:
			if env.Op == "kanban.snapshot" {
				broadcastCount++
			}
		case <-timeout:
			t.Fatal("timed out waiting for snapshot broadcast after issue sync")
		}
	}
}

func TestHubResolveDispatchTargetPrefersBoundDevice(t *testing.T) {
	hub, closeStore := newTestHub(t)
	defer closeStore()
	desktopA := registerTestDesktop(t, hub, "desktop-a", "device-a")
	registerTestDesktop(t, hub, "desktop-b", "device-b")
	createHubBinding(t, hub, "device-a", "all", "dispatch")

	target := hub.resolveDispatchTarget(kanban.DefaultProjectID, "")
	if target != desktopA.id {
		t.Fatalf("expected bound device session %q auto-selected, got %q", desktopA.id, target)
	}

	// 两台都有 dispatch 绑定时不自动选择
	createHubBinding(t, hub, "device-b", "all", "dispatch")
	target = hub.resolveDispatchTarget(kanban.DefaultProjectID, "")
	if target != "" {
		t.Fatalf("expected ambiguous bindings to keep empty target, got %q", target)
	}
}
