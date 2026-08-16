package mcpserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mateo/homelab/cli/internal/client"
	"github.com/mateo/homelab/cli/internal/config"
	"github.com/mateo/homelab/cli/internal/dto"
)

type capture struct {
	method string
	path   string
	query  string
	auth   string
	body   string
}

func newTestBackend(t *testing.T, status int, resp []byte, onReq func(*capture, *http.Request)) (*httptest.Server, *capture) {
	t.Helper()
	cap := &capture{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cap.method = r.Method
		cap.path = r.URL.Path
		cap.query = r.URL.RawQuery
		cap.auth = r.Header.Get("Authorization")
		if r.Body != nil {
			b, _ := io.ReadAll(r.Body)
			cap.body = string(b)
		}
		if onReq != nil {
			onReq(cap, r)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if resp != nil {
			_, _ = w.Write(resp)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, cap
}

func testServer(t *testing.T, baseURL string, requireAuth bool, key string) *Server {
	t.Helper()
	return NewForTest(client.New(baseURL, key, requireAuth))
}

func TestHealthTool(t *testing.T) {
	t.Parallel()
	srv, cap := newTestBackend(t, http.StatusOK, []byte(`{"status":"ok","db_ok":true}`), nil)
	s := testServer(t, srv.URL, true, "secret")

	_, out, err := s.health(context.Background(), nil, emptyInput{})
	if err != nil {
		t.Fatalf("health: %v", err)
	}
	if out.Status != "ok" || !out.DBOk {
		t.Fatalf("unexpected output: %+v", out)
	}
	if cap.path != "/api/health" {
		t.Fatalf("path = %q", cap.path)
	}
	// health must never send Authorization even in prod
	if cap.auth != "" {
		t.Fatalf("health sent Authorization %q", cap.auth)
	}
}

func TestTodoListFilters(t *testing.T) {
	t.Parallel()
	srv, cap := newTestBackend(t, http.StatusOK, []byte(`[]`), nil)
	s := testServer(t, srv.URL, false, "")

	_, out, err := s.todoList(context.Background(), nil, todoListInput{
		Kind: "issue", IssueType: "bug", ProjectID: "2",
	})
	if err != nil {
		t.Fatalf("todoList: %v", err)
	}
	if out.Todos == nil {
		t.Fatal("expected non-nil todos slice")
	}
	if cap.path != "/api/todos" {
		t.Fatalf("path = %q", cap.path)
	}
	if !strings.Contains(cap.query, "kind=issue") || !strings.Contains(cap.query, "issue_type=bug") || !strings.Contains(cap.query, "project_id=2") {
		t.Fatalf("query = %q", cap.query)
	}
}

func TestTodoCreateAndUpdateClear(t *testing.T) {
	t.Parallel()

	todoJSON := []byte(`{
		"id": 1,
		"title": "Buy milk",
		"body": "",
		"status": "todo",
		"priority": 1,
		"due_date": null,
		"kind": "issue",
		"issue_type": null,
		"project_id": null,
		"created_at": "2026-08-15T00:00:00Z",
		"updated_at": "2026-08-15T00:00:00Z"
	}`)

	t.Run("create", func(t *testing.T) {
		t.Parallel()
		srv, cap := newTestBackend(t, http.StatusCreated, todoJSON, nil)
		s := testServer(t, srv.URL, false, "")
		it := "feature"
		pid := int64(9)
		_, got, err := s.todoCreate(context.Background(), nil, todoCreateInput{
			Title: "Buy milk", Priority: 1, Kind: "issue", IssueType: &it, ProjectID: &pid,
		})
		if err != nil {
			t.Fatalf("todoCreate: %v", err)
		}
		if got.ID != 1 || got.Title != "Buy milk" {
			t.Fatalf("got %+v", got)
		}
		if cap.method != http.MethodPost || cap.path != "/api/todos" {
			t.Fatalf("request %s %s", cap.method, cap.path)
		}
		var body dto.CreateTodo
		if err := json.Unmarshal([]byte(cap.body), &body); err != nil {
			t.Fatal(err)
		}
		if body.Title != "Buy milk" || body.Kind != "issue" || body.IssueType == nil || *body.IssueType != "feature" {
			t.Fatalf("body %+v", body)
		}
	})

	t.Run("update clear flags", func(t *testing.T) {
		t.Parallel()
		srv, cap := newTestBackend(t, http.StatusOK, todoJSON, nil)
		s := testServer(t, srv.URL, false, "")
		status := "in_progress"
		_, _, err := s.todoUpdate(context.Background(), nil, todoUpdateInput{
			ID: 1, Status: &status, ClearDueDate: true, ClearIssueType: true, ClearProjectID: true,
		})
		if err != nil {
			t.Fatalf("todoUpdate: %v", err)
		}
		if cap.method != http.MethodPut || cap.path != "/api/todos/1" {
			t.Fatalf("request %s %s", cap.method, cap.path)
		}
		// backend three-state: null keys present
		if !strings.Contains(cap.body, `"due_date":null`) {
			t.Fatalf("body missing due_date null: %s", cap.body)
		}
		if !strings.Contains(cap.body, `"issue_type":null`) {
			t.Fatalf("body missing issue_type null: %s", cap.body)
		}
		if !strings.Contains(cap.body, `"project_id":null`) {
			t.Fatalf("body missing project_id null: %s", cap.body)
		}
		if !strings.Contains(cap.body, `"status":"in_progress"`) {
			t.Fatalf("body missing status: %s", cap.body)
		}
	})

	t.Run("create validation", func(t *testing.T) {
		t.Parallel()
		srv, _ := newTestBackend(t, http.StatusOK, todoJSON, nil)
		s := testServer(t, srv.URL, false, "")
		_, _, err := s.todoCreate(context.Background(), nil, todoCreateInput{})
		if err == nil || !strings.Contains(err.Error(), "title") {
			t.Fatalf("expected title error, got %v", err)
		}
	})
}

func TestTodoDoneDelete(t *testing.T) {
	t.Parallel()
	todoJSON := []byte(`{"id":5,"title":"x","body":"","status":"done","priority":0,"kind":"note","created_at":"t","updated_at":"t"}`)

	t.Run("done", func(t *testing.T) {
		t.Parallel()
		srv, cap := newTestBackend(t, http.StatusOK, todoJSON, nil)
		s := testServer(t, srv.URL, false, "")
		_, got, err := s.todoDone(context.Background(), nil, todoIDInput{ID: 5})
		if err != nil {
			t.Fatal(err)
		}
		if got.Status != "done" {
			t.Fatalf("status = %q", got.Status)
		}
		if cap.path != "/api/todos/5" || !strings.Contains(cap.body, `"status":"done"`) {
			t.Fatalf("request %s %s body %s", cap.method, cap.path, cap.body)
		}
	})

	t.Run("delete", func(t *testing.T) {
		t.Parallel()
		srv, cap := newTestBackend(t, http.StatusNoContent, nil, nil)
		s := testServer(t, srv.URL, false, "")
		_, got, err := s.todoDelete(context.Background(), nil, todoIDInput{ID: 5})
		if err != nil {
			t.Fatal(err)
		}
		if !got.Deleted || got.ID != 5 {
			t.Fatalf("got %+v", got)
		}
		if cap.method != http.MethodDelete || cap.path != "/api/todos/5" {
			t.Fatalf("request %s %s", cap.method, cap.path)
		}
	})
}

func TestProjectTools(t *testing.T) {
	t.Parallel()
	projJSON := []byte(`{"id":3,"name":"Alpha","color":"#fff","created_at":"t"}`)

	t.Run("list", func(t *testing.T) {
		t.Parallel()
		srv, _ := newTestBackend(t, http.StatusOK, []byte(`[]`), nil)
		s := testServer(t, srv.URL, false, "")
		_, out, err := s.projectList(context.Background(), nil, projectListInput{})
		if err != nil {
			t.Fatal(err)
		}
		if out.Projects == nil {
			t.Fatal("expected non-nil projects")
		}
	})

	t.Run("create", func(t *testing.T) {
		t.Parallel()
		srv, cap := newTestBackend(t, http.StatusCreated, projJSON, nil)
		s := testServer(t, srv.URL, false, "")
		_, got, err := s.projectCreate(context.Background(), nil, projectCreateInput{Name: "Alpha", Color: "#fff"})
		if err != nil {
			t.Fatal(err)
		}
		if got.Name != "Alpha" {
			t.Fatalf("got %+v", got)
		}
		if cap.method != http.MethodPost || cap.path != "/api/projects" {
			t.Fatalf("request %s %s", cap.method, cap.path)
		}
	})

	t.Run("update requires field", func(t *testing.T) {
		t.Parallel()
		srv, _ := newTestBackend(t, http.StatusOK, projJSON, nil)
		s := testServer(t, srv.URL, false, "")
		_, _, err := s.projectUpdate(context.Background(), nil, projectUpdateInput{ID: 3})
		if err == nil || !strings.Contains(err.Error(), "no fields") {
			t.Fatalf("expected no fields error, got %v", err)
		}
	})

	t.Run("delete", func(t *testing.T) {
		t.Parallel()
		srv, cap := newTestBackend(t, http.StatusNoContent, nil, nil)
		s := testServer(t, srv.URL, false, "")
		_, got, err := s.projectDelete(context.Background(), nil, projectIDInput{ID: 3})
		if err != nil {
			t.Fatal(err)
		}
		if !got.Deleted || got.ID != 3 {
			t.Fatalf("got %+v", got)
		}
		if cap.method != http.MethodDelete {
			t.Fatalf("method %s", cap.method)
		}
	})
}

func TestResolveClient_ProdMissingKey(t *testing.T) {
	// Serial: swaps ConfigDirFn and env — must not run parallel with other seam tests.
	dir := t.TempDir()
	prev := config.ConfigDirFn
	config.ConfigDirFn = func() (string, error) { return dir, nil }
	t.Cleanup(func() { config.ConfigDirFn = prev })

	t.Setenv(config.EnvEnv, "production")
	t.Setenv(config.EnvAPIKey, "")
	t.Setenv(config.EnvBaseURL, "")

	_, err := resolveClient()
	if err == nil {
		t.Fatal("expected missing key error")
	}
	if !strings.Contains(err.Error(), "homelab login") {
		t.Fatalf("error should mention login: %v", err)
	}
}

func TestResolveClient_DevOpen(t *testing.T) {
	dir := t.TempDir()
	prev := config.ConfigDirFn
	config.ConfigDirFn = func() (string, error) { return dir, nil }
	t.Cleanup(func() { config.ConfigDirFn = prev })

	t.Setenv(config.EnvEnv, "development")
	t.Setenv(config.EnvAPIKey, "")
	t.Setenv(config.EnvBaseURL, "http://example.test")

	c, err := resolveClient()
	if err != nil {
		t.Fatalf("resolveClient: %v", err)
	}
	if c == nil {
		t.Fatal("nil client")
	}
}

func TestNewMCPServerRegistersTools(t *testing.T) {
	t.Parallel()
	srv, _ := newTestBackend(t, http.StatusOK, []byte(`{}`), nil)
	s := testServer(t, srv.URL, false, "")
	mcpSrv := s.NewMCPServer("test")
	if mcpSrv == nil {
		t.Fatal("nil mcp server")
	}
}

func TestBackendErrorSurfaced(t *testing.T) {
	t.Parallel()
	srv, _ := newTestBackend(t, http.StatusUnauthorized, []byte(`{"error":"unauthorized"}`), nil)
	s := testServer(t, srv.URL, true, "bad")
	_, _, err := s.todoList(context.Background(), nil, todoListInput{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Fatalf("error %q", err)
	}
}
