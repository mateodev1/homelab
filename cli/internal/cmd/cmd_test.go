package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// captureHandler records the last request and writes a canned JSON response.
type captureHandler struct {
	lastMethod string
	lastPath   string
	lastBody   string
	lastAuth   string
	status     int
	resp       []byte
}

func (h *captureHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.lastMethod = r.Method
	h.lastPath = r.URL.RequestURI() // path + query, so filter assertions see params
	h.lastAuth = r.Header.Get("Authorization")
	if r.Body != nil {
		if b, err := io.ReadAll(r.Body); err == nil {
			h.lastBody = string(b)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(h.status)
	if h.resp != nil {
		_, _ = w.Write(h.resp)
	}
}

// runRoot wires a root command against a test server, executes args, and
// returns captured stdout, stderr, and any execution error.
func runRoot(t *testing.T, srv *httptest.Server, apiKey string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	root := NewRootCommand("test")
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs(append([]string{"--base-url", srv.URL, "--api-key", apiKey}, args...))
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	// keep cobra from calling os.Exit
	err = root.Execute()
	return outBuf.String(), errBuf.String(), err
}

func TestTodoList_ParsesJSON(t *testing.T) {
	t.Parallel()
	srvResp := []byte(`[{"id":1,"title":"a","body":"","status":"todo","priority":0,"due_date":null,"kind":"note","issue_type":null,"project_id":null,"created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"}]`)
	h := &captureHandler{status: http.StatusOK, resp: srvResp}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	out, _, err := runRoot(t, srv, "k", "todo", "list")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if h.lastPath != "/api/todos" {
		t.Errorf("path = %q", h.lastPath)
	}
	if h.lastAuth != "Bearer k" {
		t.Errorf("auth = %q", h.lastAuth)
	}
	var got []map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("stdout not JSON: %v\n%q", err, out)
	}
	if len(got) != 1 || got[0]["title"] != "a" {
		t.Fatalf("unexpected list output: %s", out)
	}
}

func TestTodoList_FiltersInQuery(t *testing.T) {
	t.Parallel()
	h := &captureHandler{status: http.StatusOK, resp: []byte("[]")}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	_, _, err := runRoot(t, srv, "k", "todo", "list", "--kind=issue", "--issue-type=bug", "--project-id=none")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(h.lastPath, "kind=issue") {
		t.Errorf("missing kind filter: %q", h.lastPath)
	}
	if !strings.Contains(h.lastPath, "issue_type=bug") {
		t.Errorf("missing issue_type filter: %q", h.lastPath)
	}
	if !strings.Contains(h.lastPath, "project_id=none") {
		t.Errorf("missing project_id filter: %q", h.lastPath)
	}
}

func TestTodoDone_SendsStatusDoneBody(t *testing.T) {
	t.Parallel()
	h := &captureHandler{status: http.StatusOK, resp: []byte(`{"id":3,"title":"x","body":"","status":"done","priority":0,"due_date":null,"kind":"note","issue_type":null,"project_id":null,"created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"}`)}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	_, _, err := runRoot(t, srv, "k", "todo", "done", "3")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if h.lastMethod != http.MethodPut {
		t.Errorf("method = %q", h.lastMethod)
	}
	if h.lastPath != "/api/todos/3" {
		t.Errorf("path = %q", h.lastPath)
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(h.lastBody), &body); err != nil {
		t.Fatalf("body not JSON: %v (%q)", err, h.lastBody)
	}
	if body["status"] != "done" {
		t.Errorf("body = %q, want status=done", h.lastBody)
	}
	if len(body) != 1 {
		t.Errorf("done should only send status key, got %q", h.lastBody)
	}
}

func TestTodoUpdate_ClearsDueDateWithLiteralNull(t *testing.T) {
	t.Parallel()
	h := &captureHandler{status: http.StatusOK, resp: []byte(`{"id":1,"title":"x","body":"","status":"todo","priority":0,"due_date":null,"kind":"note","issue_type":null,"project_id":null,"created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"}`)}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	_, _, err := runRoot(t, srv, "k", "todo", "update", "1", "--due-date=null")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(h.lastBody, `"due_date":null`) {
		t.Errorf("expected literal null for due_date clear, got %q", h.lastBody)
	}
}

func TestTodoUpdate_OmitsUnsetKeys(t *testing.T) {
	t.Parallel()
	h := &captureHandler{status: http.StatusOK, resp: []byte(`{"id":1,"title":"x","body":"","status":"todo","priority":0,"due_date":null,"kind":"note","issue_type":null,"project_id":null,"created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"}`)}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	_, _, err := runRoot(t, srv, "k", "todo", "update", "1", "--status=done")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(h.lastBody), &body); err != nil {
		t.Fatalf("body not JSON: %v (%q)", err, h.lastBody)
	}
	if len(body) != 1 || body["status"] != "done" {
		t.Errorf("expected only status key, got %q", h.lastBody)
	}
}

func TestHealth_NoAPIKeyRequired(t *testing.T) {
	t.Parallel()
	h := &captureHandler{status: http.StatusOK, resp: []byte(`{"status":"ok","db_ok":true}`)}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	root := NewRootCommand("test")
	root.SetArgs([]string{"--base-url", srv.URL, "health"})
	outBuf := &bytes.Buffer{}
	root.SetOut(outBuf)
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(outBuf.String(), `"ok"`) {
		t.Errorf("health output: %s", outBuf)
	}
	if h.lastAuth != "" {
		t.Errorf("health must not send auth, got %q", h.lastAuth)
	}
}

func TestRequiresAPIKey_MissingKeyErrors(t *testing.T) {
	t.Parallel()
	h := &captureHandler{status: http.StatusOK, resp: []byte("[]")}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	root := NewRootCommand("test")
	root.SetArgs([]string{"--base-url", srv.URL, "todo", "list"})
	errBuf := &bytes.Buffer{}
	root.SetErr(errBuf)
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when api key missing for todo list")
	}
	if !strings.Contains(err.Error(), "api key required") {
		t.Errorf("error %q missing hint", err)
	}
}

func TestTodoAdd_InvalidPriorityRejected(t *testing.T) {
	t.Parallel()
	h := &captureHandler{status: http.StatusCreated}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	root := NewRootCommand("test")
	root.SetArgs([]string{"--base-url", srv.URL, "--api-key", "k", "todo", "add", "--title=x", "--priority=9"})
	root.SetErr(&bytes.Buffer{})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected validation error for priority 9")
	}
	if !strings.Contains(err.Error(), "priority") {
		t.Errorf("error %q should mention priority", err)
	}
}

func TestProjectDelete_WarnsBeforeDelete(t *testing.T) {
	t.Parallel()
	h := &captureHandler{status: http.StatusNoContent}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	root := NewRootCommand("test")
	root.SetArgs([]string{"--base-url", srv.URL, "--api-key", "k", "project", "delete", "5", "--yes"})
	errBuf := &bytes.Buffer{}
	root.SetErr(errBuf)
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(errBuf.String(), "warning") {
		t.Errorf("expected warning in stderr, got %q", errBuf.String())
	}
	if h.lastMethod != http.MethodDelete || h.lastPath != "/api/projects/5" {
		t.Errorf("delete request wrong: %s %s", h.lastMethod, h.lastPath)
	}
}

