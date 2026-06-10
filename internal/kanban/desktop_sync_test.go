package kanban_test

import (
	"context"
	"testing"

	"zenmind-kanban-server/internal/kanban"
)

func createTestBinding(t *testing.T, service *kanban.Service, syncPolicy string, controlMode string) kanban.ProjectBinding {
	t.Helper()
	result, err := service.CreateProjectBinding(context.Background(), kanban.ProjectBindingInput{
		ProjectID:      kanban.DefaultProjectID,
		DeviceID:       "device-sync",
		LocalProjectID: "local-sync",
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

func TestServiceUpdateProjectBindingChangesPolicyAndMode(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	binding := createTestBinding(t, service, "future", "dispatch")

	syncPolicy := "select"
	controlMode := "observe"
	result, err := service.UpdateProjectBinding(context.Background(), binding.ID, kanban.ProjectBindingUpdateInput{
		SyncPolicy:  &syncPolicy,
		ControlMode: &controlMode,
	}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Binding == nil {
		t.Fatalf("expected binding update to succeed: %#v", result)
	}
	if result.Binding.SyncPolicy != "select" || result.Binding.ControlMode != "observe" {
		t.Fatalf("expected updated policy/mode, got %#v", result.Binding)
	}

	missing, err := service.UpdateProjectBinding(context.Background(), "missing-id", kanban.ProjectBindingUpdateInput{}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if missing.OK {
		t.Fatalf("expected update of missing binding to fail")
	}
}

func TestServiceProjectBindingHasSyncSinceAt(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	binding := createTestBinding(t, service, "future", "dispatch")
	if binding.SyncSinceAt == nil {
		t.Fatalf("expected new binding to carry syncSinceAt anchor")
	}
}

func TestServiceSetProjectBindingIssuesReplacesSelection(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	binding := createTestBinding(t, service, "select", "dispatch")
	issueA := createIssue(t, service, "selected A")
	issueB := createIssue(t, service, "selected B")

	result, err := service.SetProjectBindingIssues(context.Background(), kanban.ProjectBindingIssuesSetInput{
		BindingID: binding.ID,
		IssueIDs:  []string{issueA.ID, issueB.ID},
	}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK {
		t.Fatalf("expected set issues to succeed: %#v", result)
	}
	snapshot, err := service.Snapshot(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, kanban.DesktopStatus{})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.ProjectBindingIssues) != 2 {
		t.Fatalf("expected 2 binding issues in snapshot, got %#v", snapshot.ProjectBindingIssues)
	}

	// 整体替换为只剩 issueB
	result, err = service.SetProjectBindingIssues(context.Background(), kanban.ProjectBindingIssuesSetInput{
		BindingID: binding.ID,
		IssueIDs:  []string{issueB.ID},
	}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK {
		t.Fatalf("expected set issues to succeed: %#v", result)
	}
	snapshot, err = service.Snapshot(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, kanban.DesktopStatus{})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.ProjectBindingIssues) != 1 || snapshot.ProjectBindingIssues[0].IssueID != issueB.ID {
		t.Fatalf("expected selection replaced with issueB, got %#v", snapshot.ProjectBindingIssues)
	}
}

func TestServiceSyncDesktopIssuesCreatesAndMapsLocalIDs(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	createTestBinding(t, service, "all", "dispatch")

	result, err := service.SyncDesktopIssues(context.Background(), kanban.DefaultBoardID, kanban.DesktopIssueSyncInput{
		DeviceID:       "device-sync",
		ProjectID:      kanban.DefaultProjectID,
		LocalProjectID: "local-sync",
		Upserts: []kanban.DesktopIssueSyncUpsert{
			{LocalIssueID: "local-1", Input: kanban.IssueInput{Title: "本地任务一"}},
			{LocalIssueID: "local-2", Input: kanban.IssueInput{Title: "本地任务二"}},
		},
	}, "desktop-user")
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || len(result.Results) != 2 {
		t.Fatalf("expected 2 sync results, got %#v", result)
	}
	for _, item := range result.Results {
		if item.Status != "created" || item.RemoteIssueID == "" || item.Issue == nil {
			t.Fatalf("expected created result with remote id, got %#v", item)
		}
	}
	if result.Results[0].LocalIssueID != "local-1" {
		t.Fatalf("expected local id mapping preserved, got %#v", result.Results[0])
	}
}

func TestServiceSyncDesktopIssuesConflictReturnsCloudIssue(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	createTestBinding(t, service, "all", "dispatch")
	issue := createIssue(t, service, "cloud title")

	staleRevision := issue.Revision - 1
	if staleRevision <= 0 {
		staleRevision = 1
	}
	// 先把云端 issue 更新一版,制造修订差
	newTitle := "cloud title v2"
	updated, err := service.UpdateIssue(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, issue.ID, kanban.IssueUpdateInput{Title: &newTitle}, "web")
	if err != nil {
		t.Fatal(err)
	}
	if !updated.OK {
		t.Fatalf("expected cloud update to succeed: %#v", updated)
	}

	result, err := service.SyncDesktopIssues(context.Background(), kanban.DefaultBoardID, kanban.DesktopIssueSyncInput{
		DeviceID:       "device-sync",
		ProjectID:      kanban.DefaultProjectID,
		LocalProjectID: "local-sync",
		Upserts: []kanban.DesktopIssueSyncUpsert{
			{
				LocalIssueID:      "local-9",
				RemoteIssueID:     issue.ID,
				BaseIssueRevision: issue.Revision, // 桌面端持有旧修订
				Input:             kanban.IssueInput{Title: "desktop title"},
			},
		},
	}, "desktop-user")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Results) != 1 {
		t.Fatalf("expected 1 result, got %#v", result)
	}
	item := result.Results[0]
	if item.Status != "conflict" {
		t.Fatalf("expected conflict, got %#v", item)
	}
	if item.Issue == nil || item.Issue.Title != "cloud title v2" {
		t.Fatalf("expected cloud authoritative issue returned, got %#v", item.Issue)
	}
}

func TestServiceSyncDesktopIssuesRejectsDisabledBinding(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	createTestBinding(t, service, "all", "disabled")

	result, err := service.SyncDesktopIssues(context.Background(), kanban.DefaultBoardID, kanban.DesktopIssueSyncInput{
		DeviceID:       "device-sync",
		ProjectID:      kanban.DefaultProjectID,
		LocalProjectID: "local-sync",
		Upserts: []kanban.DesktopIssueSyncUpsert{
			{LocalIssueID: "local-1", Input: kanban.IssueInput{Title: "blocked"}},
		},
	}, "desktop-user")
	if err != nil {
		t.Fatal(err)
	}
	if result.OK {
		t.Fatalf("expected disabled binding to reject sync, got %#v", result)
	}
}

func TestServiceSyncDesktopIssuesSelectPolicyPinsCreatedIssue(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	binding := createTestBinding(t, service, "select", "dispatch")

	result, err := service.SyncDesktopIssues(context.Background(), kanban.DefaultBoardID, kanban.DesktopIssueSyncInput{
		DeviceID:       "device-sync",
		ProjectID:      kanban.DefaultProjectID,
		LocalProjectID: "local-sync",
		Upserts: []kanban.DesktopIssueSyncUpsert{
			{LocalIssueID: "local-1", Input: kanban.IssueInput{Title: "select pinned"}},
		},
	}, "desktop-user")
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || len(result.Results) != 1 || result.Results[0].Status != "created" {
		t.Fatalf("expected created result, got %#v", result)
	}
	snapshot, err := service.Snapshot(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, kanban.DesktopStatus{})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range snapshot.ProjectBindingIssues {
		if item.BindingID == binding.ID && item.IssueID == result.Results[0].RemoteIssueID {
			found = true
			if item.Source != "desktop_sync" {
				t.Fatalf("expected desktop_sync source, got %q", item.Source)
			}
		}
	}
	if !found {
		t.Fatalf("expected created issue pinned to select binding, got %#v", snapshot.ProjectBindingIssues)
	}
}

func TestServiceSyncDesktopIssuesDeleteFlow(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	createTestBinding(t, service, "all", "dispatch")
	issue := createIssue(t, service, "to delete")

	result, err := service.SyncDesktopIssues(context.Background(), kanban.DefaultBoardID, kanban.DesktopIssueSyncInput{
		DeviceID:       "device-sync",
		ProjectID:      kanban.DefaultProjectID,
		LocalProjectID: "local-sync",
		Deletes: []kanban.DesktopIssueSyncDelete{
			{LocalIssueID: "local-d", RemoteIssueID: issue.ID},
		},
	}, "desktop-user")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Results) != 1 || result.Results[0].Status != "deleted" {
		t.Fatalf("expected deleted result, got %#v", result)
	}
}
