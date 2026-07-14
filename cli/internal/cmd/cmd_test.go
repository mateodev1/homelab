package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mateo/homelab/cli/internal/config"
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

// pinConfigDir swaps the config.ConfigDirFn seam to a fresh temp dir for this
// test and restores it on cleanup (REQ-CFG-007 / SC-CFG-008). Every cmd test
// that invokes Execute — and therefore resolveClient, which calls
// config.Load(config.ConfigDirFn()) — must call this so it never reads or
// writes the developer's real ~/.config/homelab/config.json.
//
// The seam is package-level shared state, so tests that pin it MUST NOT run in
// parallel with other pinning tests (a parallel swap/restore/read would race
// under -race). The cmd test suite is therefore serial.
func pinConfigDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	prev := config.ConfigDirFn
	config.ConfigDirFn = func() (string, error) { return dir, nil }
	t.Cleanup(func() { config.ConfigDirFn = prev })
	return dir
}

// runRoot wires a root command against a test server, executes args, and
// returns captured stdout, stderr, and any execution error. It pins the config
// directory to a fresh temp dir (REQ-CFG-007) so the test never touches the
// real config file; the explicit --base-url/--api-key flags still win over any
// (empty) file value.
func runRoot(t *testing.T, srv *httptest.Server, apiKey string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	pinConfigDir(t)
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
	srvResp := []byte(`[{"id":1,"title":"a","body":"","status":"todo","priority":0,"due_date":null,"kind":"note","issue_type":null,"project_id":null,"created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"}]`)
	h := &captureHandler{status: http.StatusOK, resp: srvResp}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	out, _, err := runRoot(t, srv, "k", "todo", "list", "--env=production")
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
	h := &captureHandler{status: http.StatusOK, resp: []byte(`{"status":"ok","db_ok":true}`)}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	root := NewRootCommand("test")
	pinConfigDir(t)
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

func TestRequiresAPIKey_MissingKeyErrorsInProd(t *testing.T) {
	h := &captureHandler{status: http.StatusOK, resp: []byte("[]")}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	root := NewRootCommand("test")
	pinConfigDir(t)
	root.SetArgs([]string{"--base-url", srv.URL, "--env=production", "todo", "list"})
	errBuf := &bytes.Buffer{}
	root.SetErr(errBuf)
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when api key missing in production for todo list")
	}
	if !strings.Contains(err.Error(), "api key required") {
		t.Errorf("error %q missing hint", err)
	}
}

func TestTodoList_DevWithoutAPIKeyDoesNotRequireKey(t *testing.T) {
	h := &captureHandler{status: http.StatusOK, resp: []byte("[]")}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	root := NewRootCommand("test")
	pinConfigDir(t)
	root.SetArgs([]string{"--base-url", srv.URL, "todo", "list"})
	if err := root.Execute(); err != nil {
		t.Fatalf("dev todo list must not require an api key: %v", err)
	}
	// Dev mode must NOT send an Authorization header (dev = no auth).
	if h.lastAuth != "" {
		t.Errorf("dev must not send Authorization, got %q", h.lastAuth)
	}
}

func TestTodoAdd_InvalidPriorityRejected(t *testing.T) {
	h := &captureHandler{status: http.StatusCreated}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	root := NewRootCommand("test")
	pinConfigDir(t)
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
	h := &captureHandler{status: http.StatusNoContent}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	root := NewRootCommand("test")
	pinConfigDir(t)
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

// TestHealth_ReadsBaseURLFromFile covers SC-CFG-009 / REQ-CFG-008: an existing
// command (health) consumes the config file's base_url when no --base-url flag
// is passed, and an explicit --base-url flag still overrides the file.
func TestHealth_ReadsBaseURLFromFile(t *testing.T) {
	dir := pinConfigDir(t)

	// file server — the config file's base_url points here.
	fileH := &captureHandler{status: http.StatusOK, resp: []byte(`{"status":"ok","db_ok":true}`)}
	fileSrv := httptest.NewServer(fileH)
	t.Cleanup(fileSrv.Close)

	if err := config.Save(dir, config.Config{BaseURL: fileSrv.URL, Env: "development"}); err != nil {
		t.Fatalf("seed config file: %v", err)
	}

	// Without --base-url: resolveClient must Load the file and use fileSrv.URL.
	root := NewRootCommand("test")
	root.SetArgs([]string{"health"})
	outBuf := &bytes.Buffer{}
	root.SetOut(outBuf)
	if err := root.Execute(); err != nil {
		t.Fatalf("health without --base-url: %v", err)
	}
	if fileH.lastPath != "/api/health" {
		t.Fatalf("file base_url not used by health: lastPath=%q (want /api/health)", fileH.lastPath)
	}
	if !strings.Contains(outBuf.String(), `"ok"`) {
		t.Errorf("health output: %s", outBuf.String())
	}

	// With --base-url=override: the flag must win over the file, hitting the
	// override server and NOT the file server.
	flagH := &captureHandler{status: http.StatusOK, resp: []byte(`{"status":"ok","db_ok":true}`)}
	flagSrv := httptest.NewServer(flagH)
	t.Cleanup(flagSrv.Close)
	fileH.lastPath = "" // reset to confirm file server is not hit again

	root2 := NewRootCommand("test")
	root2.SetArgs([]string{"--base-url", flagSrv.URL, "health"})
	if err := root2.Execute(); err != nil {
		t.Fatalf("health with --base-url override: %v", err)
	}
	if flagH.lastPath != "/api/health" {
		t.Fatalf("override base_url not used: lastPath=%q (want /api/health)", flagH.lastPath)
	}
	if fileH.lastPath != "" {
		t.Fatalf("file server should NOT be hit when --base-url is set, but got %q", fileH.lastPath)
	}
}

