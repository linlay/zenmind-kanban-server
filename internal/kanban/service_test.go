package kanban_test

import (
	"context"
	"path/filepath"
	"testing"

	"zenmind-kanban-server/internal/kanban"
	"zenmind-kanban-server/internal/store"
)

func TestServiceRejectsEmptyTitle(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()

	result, err := service.CreateIssue(context.Background(), kanban.DefaultBoardID, kanban.IssueInput{Title: "   "}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if result.OK {
		t.Fatalf("expected empty title to be rejected")
	}
}

func TestServicePreventsDirectCompletedUpdate(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	issue := createIssue(t, service, "Review plan")

	completed := string(kanban.StatusCompleted)
	result, err := service.UpdateIssue(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, issue.ID, kanban.IssueUpdateInput{Status: &completed}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if result.OK {
		t.Fatalf("expected direct completed update to be rejected")
	}
}

func TestServiceAllowsMoveToCompleted(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	issue := createIssue(t, service, "Ship task")

	result, err := service.MoveIssue(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, kanban.MoveInput{
		ID:       issue.ID,
		Status:   string(kanban.StatusCompleted),
		Position: 1,
	}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Issue == nil || result.Issue.Status != kanban.StatusCompleted {
		t.Fatalf("expected move to completed, got %#v", result)
	}
}

func TestServiceSnapshotMarksCompleteProjectScope(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	createIssue(t, service, "Scoped task")

	snapshot, err := service.Snapshot(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, kanban.DesktopStatus{})
	if err != nil {
		t.Fatal(err)
	}
	if !snapshot.Complete {
		t.Fatalf("expected complete project snapshot")
	}
	if snapshot.Scope != "project" {
		t.Fatalf("expected project scope, got %q", snapshot.Scope)
	}
	if snapshot.ProjectID != kanban.DefaultProjectID {
		t.Fatalf("expected projectId=%q, got %q", kanban.DefaultProjectID, snapshot.ProjectID)
	}
}

func TestServiceRejectsStaleUpdateRevision(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	issue := createIssue(t, service, "Concurrent update")
	staleRevision := issue.Revision
	updatedTitle := "Updated elsewhere"
	if result, err := service.UpdateIssue(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, issue.ID, kanban.IssueUpdateInput{
		Title: &updatedTitle,
	}, "first"); err != nil || !result.OK {
		t.Fatalf("expected first update to succeed: %#v, %v", result, err)
	}

	nextTitle := "Stale update"
	result, err := service.UpdateIssue(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, issue.ID, kanban.IssueUpdateInput{
		Title:             &nextTitle,
		BaseIssueRevision: &staleRevision,
	}, "second")
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Code != "conflict" {
		t.Fatalf("expected conflict for stale update, got %#v", result)
	}
	if len(result.Issues) == 0 || !result.Complete || result.Scope != "project" {
		t.Fatalf("expected conflict response to include project issues, got %#v", result)
	}
}

func TestServiceRejectsStaleMoveRevision(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	issue := createIssue(t, service, "Concurrent move")
	staleRevision := issue.Revision
	updatedTitle := "Moved elsewhere"
	if result, err := service.UpdateIssue(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, issue.ID, kanban.IssueUpdateInput{
		Title: &updatedTitle,
	}, "first"); err != nil || !result.OK {
		t.Fatalf("expected first update to succeed: %#v, %v", result, err)
	}

	result, err := service.MoveIssue(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, kanban.MoveInput{
		ID:                issue.ID,
		Status:            string(kanban.StatusTodo),
		Position:          1,
		BaseIssueRevision: &staleRevision,
	}, "second")
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Code != "conflict" {
		t.Fatalf("expected conflict for stale move, got %#v", result)
	}
}

func TestServiceRejectsStaleDeleteRevision(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	issue := createIssue(t, service, "Concurrent delete")
	staleRevision := issue.Revision
	updatedTitle := "Delete elsewhere"
	if result, err := service.UpdateIssue(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, issue.ID, kanban.IssueUpdateInput{
		Title: &updatedTitle,
	}, "first"); err != nil || !result.OK {
		t.Fatalf("expected first update to succeed: %#v, %v", result, err)
	}

	result, err := service.DeleteIssue(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, issue.ID, &staleRevision, "second")
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Code != "conflict" {
		t.Fatalf("expected conflict for stale delete, got %#v", result)
	}
}

func TestServiceLocksMoveWhileRunActive(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	issue := createIssue(t, service, "Run task")
	chatID := "chat-1"
	runID := "run-1"
	startResult, err := service.StartRun(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, issue.ID, strPtr("agent"), kanban.StartRunResult{
		OK:      true,
		Message: "started",
		ChatID:  &chatID,
		RunID:   &runID,
	}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !startResult.OK {
		t.Fatalf("expected run start to succeed: %#v", startResult)
	}

	moveResult, err := service.MoveIssue(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, kanban.MoveInput{
		ID:       issue.ID,
		Status:   string(kanban.StatusTodo),
		Position: 1,
	}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if moveResult.OK {
		t.Fatalf("expected active run move to be rejected")
	}
}

func TestServiceMovesAssistantCompletionToCompletedWhenReviewNotRequired(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	issue := createIssue(t, service, "Assistant task")
	chatID := "chat-2"
	runID := "run-2"
	_, err := service.StartRun(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, issue.ID, strPtr("agent"), kanban.StartRunResult{
		OK:      true,
		Message: "started",
		ChatID:  &chatID,
		RunID:   &runID,
	}, "test")
	if err != nil {
		t.Fatal(err)
	}

	result, err := service.SyncAssistantEvent(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, kanban.AssistantEvent{
		Type:   "run.complete",
		ChatID: &chatID,
		RunID:  &runID,
	}, "desktop")
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Issue == nil || result.Issue.Status != kanban.StatusCompleted {
		t.Fatalf("expected assistant event to move issue to completed: %#v", result)
	}
	if result.Issue.RunID != nil {
		t.Fatalf("expected run id to be cleared")
	}
}

func TestServiceMovesAssistantCompletionToReviewWhenReviewRequired(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	issue := createIssue(t, service, "Assistant task")
	reviewRequired := true
	_, err := service.UpdateIssue(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, issue.ID, kanban.IssueUpdateInput{
		ReviewRequired: &reviewRequired,
	}, "test")
	if err != nil {
		t.Fatal(err)
	}
	chatID := "chat-review"
	runID := "run-review"
	_, err = service.StartRun(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, issue.ID, strPtr("agent"), kanban.StartRunResult{
		OK:      true,
		Message: "started",
		ChatID:  &chatID,
		RunID:   &runID,
	}, "test")
	if err != nil {
		t.Fatal(err)
	}

	result, err := service.SyncAssistantEvent(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, kanban.AssistantEvent{
		Type:   "run.complete",
		ChatID: &chatID,
		RunID:  &runID,
	}, "desktop")
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Issue == nil || result.Issue.Status != kanban.StatusInReview {
		t.Fatalf("expected assistant event to move review-required issue to review: %#v", result)
	}
	if !result.Issue.ReviewRequired {
		t.Fatalf("expected reviewRequired to remain true")
	}
}

func TestServiceMovesAssistantFailureBackToTodo(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	issue := createIssue(t, service, "Failing assistant task")
	chatID := "chat-fail"
	runID := "run-fail"
	_, err := service.StartRun(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, issue.ID, strPtr("agent"), kanban.StartRunResult{
		OK:      true,
		Message: "started",
		ChatID:  &chatID,
		RunID:   &runID,
	}, "test")
	if err != nil {
		t.Fatal(err)
	}

	result, err := service.SyncAssistantEvent(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, kanban.AssistantEvent{
		Type:   "run.failed",
		ChatID: &chatID,
		RunID:  &runID,
	}, "desktop")
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Issue == nil || result.Issue.Status != kanban.StatusTodo || result.Issue.RunState == nil || *result.Issue.RunState != kanban.RunStateFailed {
		t.Fatalf("expected failed run to return issue to todo: %#v", result)
	}
	if result.Issue.RunID != nil {
		t.Fatalf("expected failed run id to be cleared")
	}
}

func TestServiceMovesAssistantCancelBackToTodo(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	issue := createIssue(t, service, "Cancelled assistant task")
	chatID := "chat-cancel"
	runID := "run-cancel"
	_, err := service.StartRun(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, issue.ID, strPtr("agent"), kanban.StartRunResult{
		OK:      true,
		Message: "started",
		ChatID:  &chatID,
		RunID:   &runID,
	}, "test")
	if err != nil {
		t.Fatal(err)
	}

	result, err := service.SyncAssistantEvent(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, kanban.AssistantEvent{
		Type:   "run.cancelled",
		ChatID: &chatID,
		RunID:  &runID,
	}, "desktop")
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Issue == nil || result.Issue.Status != kanban.StatusTodo || result.Issue.RunState == nil || *result.Issue.RunState != kanban.RunStateCancelled {
		t.Fatalf("expected cancelled run to return issue to todo: %#v", result)
	}
	if result.Issue.RunID != nil {
		t.Fatalf("expected cancelled run id to be cleared")
	}
}

func TestServiceSeedsWorkflowCatalog(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()

	snapshot, err := service.Snapshot(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, kanban.DesktopStatus{})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Workflows) != 5 {
		t.Fatalf("expected 5 workflow presets, got %d", len(snapshot.Workflows))
	}
	expectedWorkflows := map[string]string{
		"standard_requirement":   "标准需求",
		"bug_fix":                "BUG 修复",
		"optimization_iteration": "优化迭代",
		"graphic_publish":        "图文制作发布",
	}
	for key, name := range expectedWorkflows {
		if !workflowExists(snapshot.Workflows, key, name) {
			t.Fatalf("expected workflow %s/%s in %#v", key, name, snapshot.Workflows)
		}
	}
	statusCounts := map[string]int{}
	for _, status := range snapshot.WorkflowStatuses {
		statusCounts[status.WorkflowID]++
	}
	for _, workflow := range snapshot.Workflows {
		if statusCounts[workflow.ID] != 5 {
			t.Fatalf("expected workflow %s to have 5 statuses, got %d", workflow.Key, statusCounts[workflow.ID])
		}
	}
	if !transitionExists(snapshot.WorkflowTransitions, "workflow-standard-requirement", "agent_done") {
		t.Fatalf("expected standard requirement workflow to include agent_done transition")
	}
	if !transitionExists(snapshot.WorkflowTransitions, "workflow-bug-fix", "approve") {
		t.Fatalf("expected bug fix workflow to include approve transition")
	}
}

func TestServiceManagesExpandedCatalog(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	issue := createIssue(t, service, "Catalog issue")
	dependencyTarget := createIssue(t, service, "Dependency target")

	mutation, err := service.CreateUser(context.Background(), kanban.UserInput{Email: "reviewer@example.com", DisplayName: "Reviewer"}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !mutation.OK {
		t.Fatalf("expected user create to succeed: %#v", mutation)
	}
	mutation, err = service.CreateTeam(context.Background(), kanban.TeamInput{Name: "Platform"}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !mutation.OK {
		t.Fatalf("expected team create to succeed: %#v", mutation)
	}
	mutation, err = service.CreateAgent(context.Background(), kanban.AgentInput{AgentKey: "codex", Name: "Codex"}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !mutation.OK {
		t.Fatalf("expected agent create to succeed: %#v", mutation)
	}

	snapshot, err := service.Snapshot(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, kanban.DesktopStatus{})
	if err != nil {
		t.Fatal(err)
	}
	userID := snapshot.Users[0].ID
	teamID := snapshot.Teams[0].ID

	mutation, err = service.AddTeamMember(context.Background(), kanban.TeamMemberInput{TeamID: teamID, UserID: userID, Role: "member"}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !mutation.OK {
		t.Fatalf("expected team member add to succeed: %#v", mutation)
	}
	mutation, err = service.GrantProjectPermission(context.Background(), kanban.ProjectPermissionInput{
		ProjectID:     kanban.DefaultProjectID,
		PrincipalType: "user",
		PrincipalID:   userID,
		Role:          "reviewer",
	}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !mutation.OK {
		t.Fatalf("expected permission grant to succeed: %#v", mutation)
	}
	mutation, err = service.CreateIssueLabel(context.Background(), kanban.IssueLabelInput{Name: "Needs review"}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !mutation.OK {
		t.Fatalf("expected label create to succeed: %#v", mutation)
	}

	snapshot, err = service.Snapshot(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, kanban.DesktopStatus{})
	if err != nil {
		t.Fatal(err)
	}
	labelID := snapshot.IssueLabels[0].ID
	mutation, err = service.SetIssueLabels(context.Background(), kanban.IssueLabelsSetInput{IssueID: issue.ID, LabelIDs: []string{labelID}}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !mutation.OK {
		t.Fatalf("expected issue label set to succeed: %#v", mutation)
	}
	mutation, err = service.CreateIssueDependency(context.Background(), kanban.IssueDependencyInput{FromIssueID: issue.ID, ToIssueID: dependencyTarget.ID}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !mutation.OK {
		t.Fatalf("expected dependency create to succeed: %#v", mutation)
	}
	mutation, err = service.CreateReview(context.Background(), kanban.ReviewInput{IssueID: issue.ID, ReviewerID: &userID, Summary: strPtr("Please review")}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !mutation.OK {
		t.Fatalf("expected review create to succeed: %#v", mutation)
	}

	snapshot, err = service.Snapshot(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, kanban.DesktopStatus{})
	if err != nil {
		t.Fatal(err)
	}
	reviewID := snapshot.Reviews[0].ID
	mutation, err = service.CreateReviewComment(context.Background(), kanban.ReviewCommentInput{ReviewID: reviewID, IssueID: issue.ID, Body: "Looks good"}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !mutation.OK {
		t.Fatalf("expected review comment create to succeed: %#v", mutation)
	}
	approved := "approved"
	mutation, err = service.UpdateReview(context.Background(), reviewID, kanban.ReviewUpdateInput{Status: &approved}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !mutation.OK {
		t.Fatalf("expected review update to succeed: %#v", mutation)
	}

	snapshot, err = service.Snapshot(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, kanban.DesktopStatus{})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Users) != 1 || len(snapshot.Teams) != 1 || len(snapshot.TeamMembers) != 1 || len(snapshot.Agents) != 1 {
		t.Fatalf("expected people/admin catalog to be populated: %#v", snapshot)
	}
	if len(snapshot.ProjectPermissions) != 1 || len(snapshot.IssueLabels) != 1 || len(snapshot.IssueLabelLinks) != 1 || len(snapshot.IssueDependencies) != 1 {
		t.Fatalf("expected issue/project metadata to be populated: %#v", snapshot)
	}
	if len(snapshot.Reviews) != 1 || len(snapshot.ReviewComments) != 1 || len(snapshot.RecentEvents) == 0 {
		t.Fatalf("expected reviews and activity to be populated: %#v", snapshot)
	}
}

func TestServiceTransitionsIssueByWorkflowAction(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	issue := createIssue(t, service, "Transition task")

	action := "plan"
	result, err := service.TransitionIssue(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, kanban.TransitionInput{
		ID:        issue.ID,
		ActionKey: &action,
	}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Issue == nil || result.Issue.Status != kanban.StatusTodo {
		t.Fatalf("expected plan transition to move issue to todo: %#v", result)
	}
}

func TestServiceSnapshotsProjectSubtree(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	a := createProject(t, service, kanban.DefaultProjectID, "a")
	a1 := createProject(t, service, a.ID, "a1")
	a11 := createProject(t, service, a1.ID, "a1.1")
	a12 := createProject(t, service, a1.ID, "a1.2")

	createIssueInProject(t, service, a.ID, "root issue")
	a1Issue := createIssueInProject(t, service, a1.ID, "a1 issue")
	a11Issue := createIssueInProject(t, service, a11.ID, "a1.1 issue")
	a12Issue := createIssueInProject(t, service, a12.ID, "a1.2 issue")

	a1Snapshot, err := service.Snapshot(context.Background(), kanban.DefaultBoardID, a1.ID, kanban.DesktopStatus{})
	if err != nil {
		t.Fatal(err)
	}
	if got := issueIDSet(a1Snapshot.Issues); !got[a1Issue.ID] || !got[a11Issue.ID] || !got[a12Issue.ID] || len(got) != 3 {
		t.Fatalf("expected a1 subtree issues only, got %#v", got)
	}

	a11Snapshot, err := service.Snapshot(context.Background(), kanban.DefaultBoardID, a11.ID, kanban.DesktopStatus{})
	if err != nil {
		t.Fatal(err)
	}
	if got := issueIDSet(a11Snapshot.Issues); !got[a11Issue.ID] || len(got) != 1 {
		t.Fatalf("expected a1.1 issues only, got %#v", got)
	}
}

func TestServiceKeepsParentProjectOwnIssuesSeparate(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	a := createProject(t, service, kanban.DefaultProjectID, "a")
	a1 := createProject(t, service, a.ID, "a1")
	a11 := createProject(t, service, a1.ID, "a1.1")
	parentIssue := createIssueInProject(t, service, a1.ID, "parent issue")
	childIssue := createIssueInProject(t, service, a11.ID, "child issue")

	parentSnapshot, err := service.Snapshot(context.Background(), kanban.DefaultBoardID, a1.ID, kanban.DesktopStatus{})
	if err != nil {
		t.Fatal(err)
	}
	if got := issueIDSet(parentSnapshot.Issues); !got[parentIssue.ID] || !got[childIssue.ID] || len(got) != 2 {
		t.Fatalf("expected parent plus child issues, got %#v", got)
	}

	childSnapshot, err := service.Snapshot(context.Background(), kanban.DefaultBoardID, a11.ID, kanban.DesktopStatus{})
	if err != nil {
		t.Fatal(err)
	}
	if got := issueIDSet(childSnapshot.Issues); !got[childIssue.ID] || got[parentIssue.ID] || len(got) != 1 {
		t.Fatalf("expected child issue without parent issue, got %#v", got)
	}
}

func TestServiceScopesSnapshotIssueMetadataToSelectedSubtree(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	a := createProject(t, service, kanban.DefaultProjectID, "a")
	b := createProject(t, service, kanban.DefaultProjectID, "b")
	aIssue := createIssueInProject(t, service, a.ID, "a issue")
	bIssue := createIssueInProject(t, service, b.ID, "b issue")
	bOtherIssue := createIssueInProject(t, service, b.ID, "b other issue")

	createLabel(t, service, a.ID, "A label")
	createLabel(t, service, b.ID, "B label")
	snapshot, err := service.Snapshot(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID, kanban.DesktopStatus{})
	if err != nil {
		t.Fatal(err)
	}
	aLabelID := labelIDByName(t, snapshot.IssueLabels, "A label")
	bLabelID := labelIDByName(t, snapshot.IssueLabels, "B label")
	if mutation, err := service.SetIssueLabels(context.Background(), kanban.IssueLabelsSetInput{IssueID: aIssue.ID, LabelIDs: []string{aLabelID}}, "test"); err != nil || !mutation.OK {
		t.Fatalf("expected a label set to succeed: %#v, %v", mutation, err)
	}
	if mutation, err := service.SetIssueLabels(context.Background(), kanban.IssueLabelsSetInput{IssueID: bIssue.ID, LabelIDs: []string{bLabelID}}, "test"); err != nil || !mutation.OK {
		t.Fatalf("expected b label set to succeed: %#v, %v", mutation, err)
	}
	if mutation, err := service.CreateIssueDependency(context.Background(), kanban.IssueDependencyInput{FromIssueID: bIssue.ID, ToIssueID: bOtherIssue.ID}, "test"); err != nil || !mutation.OK {
		t.Fatalf("expected dependency create to succeed: %#v, %v", mutation, err)
	}
	if mutation, err := service.CreateReview(context.Background(), kanban.ReviewInput{IssueID: aIssue.ID, Summary: strPtr("a review")}, "test"); err != nil || !mutation.OK {
		t.Fatalf("expected a review create to succeed: %#v, %v", mutation, err)
	}
	if mutation, err := service.CreateReview(context.Background(), kanban.ReviewInput{IssueID: bIssue.ID, Summary: strPtr("b review")}, "test"); err != nil || !mutation.OK {
		t.Fatalf("expected b review create to succeed: %#v, %v", mutation, err)
	}
	aSnapshot, err := service.Snapshot(context.Background(), kanban.DefaultBoardID, a.ID, kanban.DesktopStatus{})
	if err != nil {
		t.Fatal(err)
	}
	if got := issueIDSet(aSnapshot.Issues); !got[aIssue.ID] || got[bIssue.ID] || len(got) != 1 {
		t.Fatalf("expected only a issue, got %#v", got)
	}
	if len(aSnapshot.IssueLabels) != 1 || aSnapshot.IssueLabels[0].ID != aLabelID {
		t.Fatalf("expected only a project label, got %#v", aSnapshot.IssueLabels)
	}
	if len(aSnapshot.IssueLabelLinks) != 1 || aSnapshot.IssueLabelLinks[0].IssueID != aIssue.ID {
		t.Fatalf("expected only a issue label link, got %#v", aSnapshot.IssueLabelLinks)
	}
	if len(aSnapshot.IssueDependencies) != 0 {
		t.Fatalf("expected sibling dependencies to be excluded, got %#v", aSnapshot.IssueDependencies)
	}
	if len(aSnapshot.Reviews) != 1 || aSnapshot.Reviews[0].IssueID != aIssue.ID {
		t.Fatalf("expected only a review, got %#v", aSnapshot.Reviews)
	}
}

func TestServiceUpdatesProjectPathsAfterRenameAndMove(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	a := createProject(t, service, kanban.DefaultProjectID, "a")
	a1 := createProject(t, service, a.ID, "a1")
	a11 := createProject(t, service, a1.ID, "a1.1")
	b := createProject(t, service, kanban.DefaultProjectID, "b")
	createIssueInProject(t, service, a11.ID, "nested issue")

	renamedSlug := "plan"
	renameResult, err := service.UpdateProject(context.Background(), a1.ID, kanban.ProjectUpdateInput{Slug: &renamedSlug, Name: strPtr("plan")}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !renameResult.OK {
		t.Fatalf("expected rename to succeed: %#v", renameResult)
	}
	a11 = projectByID(t, renameResult.Projects, a11.ID)
	if a11.Path != "a/plan/a1.1" {
		t.Fatalf("expected renamed child path, got %s", a11.Path)
	}

	moveResult, err := service.MoveProject(context.Background(), kanban.ProjectMoveInput{ID: a1.ID, ParentID: &b.ID}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !moveResult.OK {
		t.Fatalf("expected move to succeed: %#v", moveResult)
	}
	a1 = projectByID(t, moveResult.Projects, a1.ID)
	a11 = projectByID(t, moveResult.Projects, a11.ID)
	if a1.Path != "b/plan" || a11.Path != "b/plan/a1.1" {
		t.Fatalf("expected moved paths, got %s and %s", a1.Path, a11.Path)
	}
	bSnapshot, err := service.Snapshot(context.Background(), kanban.DefaultBoardID, b.ID, kanban.DesktopStatus{})
	if err != nil {
		t.Fatal(err)
	}
	if len(bSnapshot.Issues) != 1 || bSnapshot.Issues[0].ProjectPath != "b/plan/a1.1" {
		t.Fatalf("expected moved subtree issue path, got %#v", bSnapshot.Issues)
	}
}

func newTestService(t *testing.T) (*kanban.Service, func()) {
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
	return kanban.NewService(sqliteStore), func() {
		_ = sqliteStore.Close()
	}
}

func createIssue(t *testing.T, service *kanban.Service, title string) kanban.Issue {
	t.Helper()
	result, err := service.CreateIssue(context.Background(), kanban.DefaultBoardID, kanban.IssueInput{Title: title}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Issue == nil {
		t.Fatalf("expected issue create to succeed: %#v", result)
	}
	return *result.Issue
}

func createIssueInProject(t *testing.T, service *kanban.Service, projectID string, title string) kanban.Issue {
	t.Helper()
	result, err := service.CreateIssue(context.Background(), kanban.DefaultBoardID, kanban.IssueInput{Title: title, ProjectID: &projectID}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Issue == nil {
		t.Fatalf("expected issue create to succeed: %#v", result)
	}
	return *result.Issue
}

func createProject(t *testing.T, service *kanban.Service, parentID string, name string) kanban.Project {
	t.Helper()
	result, err := service.CreateProject(context.Background(), kanban.ProjectInput{ParentID: &parentID, Name: name}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Project == nil {
		t.Fatalf("expected project create to succeed: %#v", result)
	}
	return *result.Project
}

func createLabel(t *testing.T, service *kanban.Service, projectID string, name string) {
	t.Helper()
	result, err := service.CreateIssueLabel(context.Background(), kanban.IssueLabelInput{ProjectID: &projectID, Name: name}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK {
		t.Fatalf("expected label create to succeed: %#v", result)
	}
}

func labelIDByName(t *testing.T, labels []kanban.IssueLabel, name string) string {
	t.Helper()
	for _, label := range labels {
		if label.Name == name {
			return label.ID
		}
	}
	t.Fatalf("label %s not found in %#v", name, labels)
	return ""
}

func issueIDSet(issues []kanban.Issue) map[string]bool {
	result := map[string]bool{}
	for _, issue := range issues {
		result[issue.ID] = true
	}
	return result
}

func projectByID(t *testing.T, projects []kanban.Project, projectID string) kanban.Project {
	t.Helper()
	for _, project := range projects {
		if project.ID == projectID {
			return project
		}
	}
	t.Fatalf("project %s not found in %#v", projectID, projects)
	return kanban.Project{}
}

func workflowExists(workflows []kanban.Workflow, key string, name string) bool {
	for _, workflow := range workflows {
		if workflow.Key == key && workflow.Name == name {
			return true
		}
	}
	return false
}

func transitionExists(transitions []kanban.WorkflowTransition, workflowID string, actionKey string) bool {
	for _, transition := range transitions {
		if transition.WorkflowID == workflowID && transition.ActionKey == actionKey {
			return true
		}
	}
	return false
}

func strPtr(value string) *string {
	return &value
}

func TestListProjectIssuesReturnsIssues(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	createIssue(t, service, "Issue one")
	createIssue(t, service, "Issue two")

	result, err := service.ListProjectIssues(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID)
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK {
		t.Fatalf("expected ok=true, got %#v", result)
	}
	if result.Count != len(result.Issues) {
		t.Fatalf("expected count=%d to match len(issues)=%d", result.Count, len(result.Issues))
	}
	if result.Count < 2 {
		t.Fatalf("expected at least 2 issues, got %d", result.Count)
	}
	if result.BoardID != kanban.DefaultBoardID {
		t.Fatalf("expected boardId=%q, got %q", kanban.DefaultBoardID, result.BoardID)
	}
	if result.ProjectID != kanban.DefaultProjectID {
		t.Fatalf("expected projectId=%q, got %q", kanban.DefaultProjectID, result.ProjectID)
	}
	if result.Message != "issue 列表已加载。" {
		t.Fatalf("expected issue message, got %q", result.Message)
	}
}

func TestListProjectIssuesIncludesSubProjectIssues(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	a := createProject(t, service, kanban.DefaultProjectID, "a")
	a1 := createProject(t, service, a.ID, "a1")
	parentIssue := createIssueInProject(t, service, a.ID, "parent")
	childIssue := createIssueInProject(t, service, a1.ID, "child")

	result, err := service.ListProjectIssues(context.Background(), kanban.DefaultBoardID, a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK {
		t.Fatalf("expected ok=true, got %#v", result)
	}
	ids := issueIDSet(result.Issues)
	if !ids[parentIssue.ID] {
		t.Fatal("expected parent issue to be included")
	}
	if !ids[childIssue.ID] {
		t.Fatal("expected child project issue to be included")
	}
	if result.Count != 2 {
		t.Fatalf("expected 2 issues (parent+child), got %d", result.Count)
	}
}

func TestListProjectIssuesExcludesSiblingProjectIssues(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	a := createProject(t, service, kanban.DefaultProjectID, "a")
	b := createProject(t, service, kanban.DefaultProjectID, "b")
	aIssue := createIssueInProject(t, service, a.ID, "a issue")
	createIssueInProject(t, service, b.ID, "b issue")

	result, err := service.ListProjectIssues(context.Background(), kanban.DefaultBoardID, a.ID)
	if err != nil {
		t.Fatal(err)
	}
	ids := issueIDSet(result.Issues)
	if !ids[aIssue.ID] {
		t.Fatal("expected a's issue to be included")
	}
	if len(ids) != 1 {
		t.Fatalf("expected only a's issue, got %d issues", len(ids))
	}
}

func TestListProjectIssuesResponseShape(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()
	createIssue(t, service, "Shape test")

	result, err := service.ListProjectIssues(context.Background(), kanban.DefaultBoardID, kanban.DefaultProjectID)
	if err != nil {
		t.Fatal(err)
	}
	// IssuesResult 类型不包含 workflows/labels/reviews/agentRuns 等 snapshot 字段
	if result.Count == 0 {
		t.Fatal("expected at least 1 issue")
	}
	if len(result.Issues) == 0 {
		t.Fatal("expected non-empty issues")
	}
}

func TestListProjectIssuesFallsBackToDefault(t *testing.T) {
	service, closeStore := newTestService(t)
	defer closeStore()

	createIssue(t, service, "Default project issue")

	result, err := service.ListProjectIssues(context.Background(), kanban.DefaultBoardID, "nonexistent-project")
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK {
		t.Fatalf("expected ok=true, got %#v", result)
	}
	if result.ProjectID != kanban.DefaultProjectID {
		t.Fatalf("expected fallback to default project, got %q", result.ProjectID)
	}
	if result.Count < 1 {
		t.Fatalf("expected at least 1 issue from default project, got %d", result.Count)
	}
}
