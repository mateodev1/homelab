package handler_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func postJSON(t *testing.T, url string, payload map[string]any) map[string]any {
	t.Helper()
	res := mustJSONRequest(t, http.MethodPost, url, payload)
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("POST %s: expected 201, got %d", url, res.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return body
}

func TestIntegration_SecretHierarchy_ProductToEnvironment(t *testing.T) {
	srv, cleanup := newTestServer(t)
	defer cleanup()

	product := postJSON(t, srv.URL+"/api/products", map[string]any{"name": "Homelab"})
	productID := int64(product["id"].(float64))

	project := postJSON(t, srv.URL+"/api/products/"+fmt.Sprint(productID)+"/projects", map[string]any{"name": "api"})
	envsRaw, ok := project["environments"].([]any)
	if !ok || len(envsRaw) != 3 {
		t.Fatalf("expected 3 environments in create response, got %#v", project["environments"])
	}

	resList := mustRequest(t, http.MethodGet, fmt.Sprintf("%s/api/products/%d/projects", srv.URL, productID), nil)
	defer func() { _ = resList.Body.Close() }()
	if resList.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resList.StatusCode)
	}

	projectID := int64(project["id"].(float64))
	resEnvs := mustRequest(t, http.MethodGet, fmt.Sprintf("%s/api/products/%d/projects/%d/environments", srv.URL, productID, projectID), nil)
	defer func() { _ = resEnvs.Body.Close() }()
	var envs []map[string]any
	if err := json.NewDecoder(resEnvs.Body).Decode(&envs); err != nil {
		t.Fatalf("decode environments: %v", err)
	}
	if len(envs) != 3 {
		t.Fatalf("expected 3 environments, got %d", len(envs))
	}
	names := map[string]bool{}
	for _, e := range envs {
		names[e["name"].(string)] = true
	}
	for _, want := range []string{"development", "staging", "production"} {
		if !names[want] {
			t.Fatalf("expected environment %q to exist", want)
		}
	}
}

func TestIntegration_SecretHierarchy_ProductNotFound(t *testing.T) {
	srv, cleanup := newTestServer(t)
	defer cleanup()

	res := mustRequest(t, http.MethodGet, srv.URL+"/api/products/999/projects", nil)
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.StatusCode)
	}
}

func setupSecretProject(t *testing.T, srv string) (productID, projectID int64) {
	t.Helper()
	product := postJSON(t, srv+"/api/products", map[string]any{"name": "Homelab"})
	productID = int64(product["id"].(float64))
	project := postJSON(t, srv+"/api/products/"+fmt.Sprint(productID)+"/projects", map[string]any{"name": "api"})
	projectID = int64(project["id"].(float64))
	return
}

func secretsBaseURL(srv string, productID, projectID int64, env string) string {
	return fmt.Sprintf("%s/api/products/%d/projects/%d/environments/%s/secrets", srv, productID, projectID, env)
}

func TestIntegration_SecretCRUDCycle_MasksListAndGet(t *testing.T) {
	srv, cleanup := newTestServer(t)
	defer cleanup()

	productID, projectID := setupSecretProject(t, srv.URL)
	base := secretsBaseURL(srv.URL, productID, projectID, "development")

	created := postJSON(t, base, map[string]any{"key": "DB_PASSWORD", "value": "s3cr3t"})
	if created["value"] == "s3cr3t" {
		t.Fatalf("create response must not return plaintext")
	}

	resList := mustRequest(t, http.MethodGet, base, nil)
	defer func() { _ = resList.Body.Close() }()
	var list []map[string]any
	if err := json.NewDecoder(resList.Body).Decode(&list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 1 || list[0]["value"] == "s3cr3t" {
		t.Fatalf("list must return masked metadata only, got %#v", list)
	}

	resGet := mustRequest(t, http.MethodGet, base+"/DB_PASSWORD", nil)
	defer func() { _ = resGet.Body.Close() }()
	var got map[string]any
	if err := json.NewDecoder(resGet.Body).Decode(&got); err != nil {
		t.Fatalf("decode get: %v", err)
	}
	if got["value"] == "s3cr3t" {
		t.Fatalf("get must return masked value, got %v", got["value"])
	}

	resReveal := mustRequest(t, http.MethodGet, base+"/DB_PASSWORD/reveal", nil)
	defer func() { _ = resReveal.Body.Close() }()
	var revealed map[string]any
	if err := json.NewDecoder(resReveal.Body).Decode(&revealed); err != nil {
		t.Fatalf("decode reveal: %v", err)
	}
	if revealed["value"] != "s3cr3t" {
		t.Fatalf("reveal must return plaintext, got %v", revealed["value"])
	}

	resUpdate := mustJSONRequest(t, http.MethodPut, base+"/DB_PASSWORD", map[string]any{"value": "new-secret"})
	defer func() { _ = resUpdate.Body.Close() }()
	if resUpdate.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on update, got %d", resUpdate.StatusCode)
	}

	resRevealAfterUpdate := mustRequest(t, http.MethodGet, base+"/DB_PASSWORD/reveal", nil)
	defer func() { _ = resRevealAfterUpdate.Body.Close() }()
	var revealedAfter map[string]any
	_ = json.NewDecoder(resRevealAfterUpdate.Body).Decode(&revealedAfter)
	if revealedAfter["value"] != "new-secret" {
		t.Fatalf("expected updated value 'new-secret', got %v", revealedAfter["value"])
	}

	resDelete := mustRequest(t, http.MethodDelete, base+"/DB_PASSWORD", nil)
	defer func() { _ = resDelete.Body.Close() }()
	if resDelete.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 on delete, got %d", resDelete.StatusCode)
	}

	resGetAfterDelete := mustRequest(t, http.MethodGet, base+"/DB_PASSWORD", nil)
	defer func() { _ = resGetAfterDelete.Body.Close() }()
	if resGetAfterDelete.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", resGetAfterDelete.StatusCode)
	}
}

func TestIntegration_SecretExport_DotenvFormat(t *testing.T) {
	srv, cleanup := newTestServer(t)
	defer cleanup()

	productID, projectID := setupSecretProject(t, srv.URL)
	base := secretsBaseURL(srv.URL, productID, projectID, "production")

	_ = postJSON(t, base, map[string]any{"key": "ZEBRA", "value": "simple"})
	_ = postJSON(t, base, map[string]any{"key": "ALPHA", "value": "has spaces"})

	exportURL := fmt.Sprintf("%s/api/products/%d/projects/%d/environments/production/export", srv.URL, productID, projectID)
	res := mustRequest(t, http.MethodGet, exportURL, nil)
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Fatalf("expected text/plain content type, got %q", ct)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, `ALPHA="has spaces"`) {
		t.Fatalf("expected quoted ALPHA line, got %q", text)
	}
	if !strings.Contains(text, "ZEBRA=simple") {
		t.Fatalf("expected unquoted ZEBRA line, got %q", text)
	}
	if strings.Index(text, "ALPHA") > strings.Index(text, "ZEBRA") {
		t.Fatalf("expected sorted keys (ALPHA before ZEBRA), got %q", text)
	}
}

func TestIntegration_SecretCreate_ValidationError(t *testing.T) {
	srv, cleanup := newTestServer(t)
	defer cleanup()

	productID, projectID := setupSecretProject(t, srv.URL)
	base := secretsBaseURL(srv.URL, productID, projectID, "development")

	res := mustJSONRequest(t, http.MethodPost, base, map[string]any{"key": "", "value": "v"})
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}

func TestIntegration_SecretGet_NotFound(t *testing.T) {
	srv, cleanup := newTestServer(t)
	defer cleanup()

	productID, projectID := setupSecretProject(t, srv.URL)
	base := secretsBaseURL(srv.URL, productID, projectID, "development")

	res := mustRequest(t, http.MethodGet, base+"/MISSING", nil)
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.StatusCode)
	}
}
