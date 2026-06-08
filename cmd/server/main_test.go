package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"zenmind-kanban-server/internal/config"
	"zenmind-kanban-server/internal/kanban"
	"zenmind-kanban-server/internal/store"
)

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

func newTestApp(t *testing.T) (*kanban.Service, *httptest.Server) {
	t.Helper()
	service, close := newTestService(t)
	t.Cleanup(close)

	cfg := config.Config{Addr: ":0", AllowedOrigins: []string{"*"}}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/issues", handleIssues(cfg, service))

	ts := httptest.NewServer(withCORS(cfg, mux))
	t.Cleanup(ts.Close)
	return service, ts
}

func TestIssuesHandlerReturns200(t *testing.T) {
	_, ts := newTestApp(t)

	resp, err := http.Get(ts.URL + "/api/issues?projectId=default")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestIssuesHandlerRejectsNonGET(t *testing.T) {
	_, ts := newTestApp(t)

	resp, err := http.Post(ts.URL+"/api/issues", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

func TestIssuesHandlerRequiresAuth(t *testing.T) {
	service, close := newTestService(t)
	defer close()

	cfg := config.Config{Addr: ":0", Token: "secret", AllowedOrigins: []string{"*"}}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/issues", handleIssues(cfg, service))
	ts := httptest.NewServer(withCORS(cfg, mux))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/issues?projectId=default")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/issues?projectId=default", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer secret")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected authorized request to return 200, got %d", resp.StatusCode)
	}
}

func TestIssuesHandlerHeaders(t *testing.T) {
	service, ts := newTestApp(t)

	// Create an issue so we have data
	service.CreateIssue(context.Background(), kanban.DefaultBoardID, kanban.IssueInput{Title: "Header test"}, "test")

	resp, err := http.Get(ts.URL + "/api/issues?projectId=default")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	serverTiming := resp.Header.Get("Server-Timing")
	if serverTiming == "" || !strings.HasPrefix(serverTiming, "issues;dur=") {
		t.Fatalf("expected Server-Timing header starting with 'issues;dur=', got %q", serverTiming)
	}

	issueCount := resp.Header.Get("X-Issue-Count")
	if issueCount == "" || issueCount == "0" {
		t.Fatalf("expected non-zero X-Issue-Count header, got %q", issueCount)
	}
}
