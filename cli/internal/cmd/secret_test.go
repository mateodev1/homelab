package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSecretProductCreate_SendsNameBody(t *testing.T) {
	h := &captureHandler{status: http.StatusOK, resp: []byte(`{"id":1,"name":"acme","created_at":"2024-01-01T00:00:00Z"}`)}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	out, _, err := runRoot(t, srv, "k", "secret", "product-create", "--name=acme")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if h.lastMethod != http.MethodPost || h.lastPath != "/api/products" {
		t.Fatalf("request wrong: %s %s", h.lastMethod, h.lastPath)
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(h.lastBody), &body); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if body["name"] != "acme" {
		t.Errorf("body = %v", body)
	}
	if !strings.Contains(out, "acme") {
		t.Errorf("output missing name: %s", out)
	}
}

func TestSecretProjectList_RequiresProductID(t *testing.T) {
	h := &captureHandler{status: http.StatusOK, resp: []byte("[]")}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	_, _, err := runRoot(t, srv, "k", "secret", "project-list")
	if err == nil {
		t.Fatal("expected error when --product-id is missing")
	}
	if !strings.Contains(err.Error(), "product-id") {
		t.Errorf("error %q should mention product-id", err)
	}
}

func TestSecretProjectCreate_BuildsNestedPath(t *testing.T) {
	h := &captureHandler{status: http.StatusOK, resp: []byte(`{"id":2,"product_id":1,"name":"webapp","created_at":"2024-01-01T00:00:00Z","environments":[{"id":1,"project_id":2,"name":"development","created_at":"2024-01-01T00:00:00Z"}]}`)}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	out, _, err := runRoot(t, srv, "k", "secret", "project-create", "--product-id=1", "--name=webapp")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if h.lastPath != "/api/products/1/projects" {
		t.Errorf("path = %q", h.lastPath)
	}
	if !strings.Contains(out, "development") {
		t.Errorf("expected environments in output: %s", out)
	}
}

func TestSecretEnvironmentList_BuildsPath(t *testing.T) {
	h := &captureHandler{status: http.StatusOK, resp: []byte("[]")}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	_, _, err := runRoot(t, srv, "k", "secret", "environment-list", "--product-id=1", "--project-id=2")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if h.lastPath != "/api/products/1/projects/2/environments" {
		t.Errorf("path = %q", h.lastPath)
	}
}

func TestSecretList_BuildsScopedPath(t *testing.T) {
	h := &captureHandler{status: http.StatusOK, resp: []byte("[]")}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	_, _, err := runRoot(t, srv, "k", "secret", "list", "--product-id=1", "--project-id=2", "--environment=staging")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if h.lastPath != "/api/products/1/projects/2/environments/staging/secrets" {
		t.Errorf("path = %q", h.lastPath)
	}
}

func TestSecretCreate_SendsKeyValueBody(t *testing.T) {
	h := &captureHandler{status: http.StatusOK, resp: []byte(`{"environment_id":1,"key":"API_KEY","value":"***","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"}`)}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	_, _, err := runRoot(t, srv, "k", "secret", "create",
		"--product-id=1", "--project-id=2", "--environment=production",
		"--key=API_KEY", "--value=secretvalue")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if h.lastMethod != http.MethodPost {
		t.Errorf("method = %q", h.lastMethod)
	}
	if h.lastPath != "/api/products/1/projects/2/environments/production/secrets" {
		t.Errorf("path = %q", h.lastPath)
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(h.lastBody), &body); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if body["key"] != "API_KEY" || body["value"] != "secretvalue" {
		t.Errorf("body = %v", body)
	}
}

func TestSecretReveal_BuildsRevealPath(t *testing.T) {
	h := &captureHandler{status: http.StatusOK, resp: []byte(`{"key":"API_KEY","value":"plaintext"}`)}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	out, _, err := runRoot(t, srv, "k", "secret", "reveal",
		"--product-id=1", "--project-id=2", "--environment=development", "--key=API_KEY")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if h.lastPath != "/api/products/1/projects/2/environments/development/secrets/API_KEY/reveal" {
		t.Errorf("path = %q", h.lastPath)
	}
	if !strings.Contains(out, "plaintext") {
		t.Errorf("expected plaintext value in output: %s", out)
	}
}

func TestSecretUpdate_SendsValueBody(t *testing.T) {
	h := &captureHandler{status: http.StatusOK, resp: []byte(`{"environment_id":1,"key":"API_KEY","value":"***","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"}`)}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	_, _, err := runRoot(t, srv, "k", "secret", "update",
		"--product-id=1", "--project-id=2", "--environment=development",
		"--key=API_KEY", "--value=newvalue")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if h.lastMethod != http.MethodPut {
		t.Errorf("method = %q", h.lastMethod)
	}
	if h.lastPath != "/api/products/1/projects/2/environments/development/secrets/API_KEY" {
		t.Errorf("path = %q", h.lastPath)
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(h.lastBody), &body); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if body["value"] != "newvalue" {
		t.Errorf("body = %v", body)
	}
}

func TestSecretDelete_RequiresConfirmationUnlessYes(t *testing.T) {
	h := &captureHandler{status: http.StatusNoContent}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	_, _, err := runRoot(t, srv, "k", "secret", "delete",
		"--product-id=1", "--project-id=2", "--environment=development", "--key=API_KEY", "--yes")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if h.lastMethod != http.MethodDelete {
		t.Errorf("method = %q", h.lastMethod)
	}
	if h.lastPath != "/api/products/1/projects/2/environments/development/secrets/API_KEY" {
		t.Errorf("path = %q", h.lastPath)
	}
}

func TestSecretExport_PrintsDotenvToStdout(t *testing.T) {
	dotenv := "API_KEY=abc\nDB_URL=postgres://x\n"
	h := &captureHandler{status: http.StatusOK, resp: []byte(dotenv)}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	out, _, err := runRoot(t, srv, "k", "secret", "export",
		"--product-id=1", "--project-id=2", "--environment=development")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if h.lastPath != "/api/products/1/projects/2/environments/development/export" {
		t.Errorf("path = %q", h.lastPath)
	}
	if out != dotenv {
		t.Errorf("stdout = %q, want %q", out, dotenv)
	}
}

func TestSecretExport_WritesToOutputFile(t *testing.T) {
	dotenv := "API_KEY=abc\n"
	h := &captureHandler{status: http.StatusOK, resp: []byte(dotenv)}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	outPath := filepath.Join(dir, ".env")

	_, _, err := runRoot(t, srv, "k", "secret", "export",
		"--product-id=1", "--project-id=2", "--environment=development", "--output="+outPath)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if string(got) != dotenv {
		t.Errorf("file content = %q, want %q", got, dotenv)
	}
}
