package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func postProject(t *testing.T, url, name string) map[string]any {
	t.Helper()
	res := mustJSONRequest(t, http.MethodPost, url, map[string]any{"name": name})
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode created project: %v", err)
	}
	return body
}

func TestIntegration_ProjectCRUDCycle(t *testing.T) {
	srv, cleanup := newTestServer(t)
	defer cleanup()

	created := postProject(t, srv.URL+"/api/projects", "Homelab")
	id := int64(created["id"].(float64))
	if created["color"] != "default" {
		t.Fatalf("expected default color, got %v", created["color"])
	}

	resGet := mustRequest(t, http.MethodGet, fmt.Sprintf("%s/api/projects/%d", srv.URL, id), nil)
	defer func() { _ = resGet.Body.Close() }()
	if resGet.StatusCode != http.StatusOK {
		t.Fatalf("expected GET 200, got %d", resGet.StatusCode)
	}

	resUpdate := mustJSONRequest(t, http.MethodPut, fmt.Sprintf("%s/api/projects/%d", srv.URL, id), map[string]any{
		"name":  "Homelab renamed",
		"color": "blue",
	})
	defer func() { _ = resUpdate.Body.Close() }()
	if resUpdate.StatusCode != http.StatusOK {
		t.Fatalf("expected PUT 200, got %d", resUpdate.StatusCode)
	}

	resDelete := mustRequest(t, http.MethodDelete, fmt.Sprintf("%s/api/projects/%d", srv.URL, id), nil)
	defer func() { _ = resDelete.Body.Close() }()
	if resDelete.StatusCode != http.StatusNoContent {
		t.Fatalf("expected DELETE 204, got %d", resDelete.StatusCode)
	}

	resNotFound := mustRequest(t, http.MethodGet, fmt.Sprintf("%s/api/projects/%d", srv.URL, id), nil)
	defer func() { _ = resNotFound.Body.Close() }()
	if resNotFound.StatusCode != http.StatusNotFound {
		t.Fatalf("expected GET after delete 404, got %d", resNotFound.StatusCode)
	}
}

func TestIntegration_ProjectCreate_BlankNameRejected(t *testing.T) {
	srv, cleanup := newTestServer(t)
	defer cleanup()

	res := mustJSONRequest(t, http.MethodPost, srv.URL+"/api/projects", map[string]any{"name": "  "})
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}

func TestIntegration_TodoFilterByProjectID(t *testing.T) {
	srv, cleanup := newTestServer(t)
	defer cleanup()

	project := postProject(t, srv.URL+"/api/projects", "Homelab")
	projectID := project["id"].(float64)

	resCreate := mustJSONRequest(t, http.MethodPost, srv.URL+"/api/todos", map[string]any{
		"title":      "In project",
		"project_id": projectID,
	})
	defer func() { _ = resCreate.Body.Close() }()
	if resCreate.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resCreate.StatusCode)
	}

	_ = postTodo(t, srv.URL+"/api/todos", "No project")

	res := mustRequest(t, http.MethodGet, fmt.Sprintf("%s/api/todos?project_id=%d", srv.URL, int64(projectID)), nil)
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	var todos []map[string]any
	if err := json.NewDecoder(res.Body).Decode(&todos); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(todos) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(todos))
	}
	if todos[0]["title"] != "In project" {
		t.Fatalf("expected 'In project', got %v", todos[0]["title"])
	}

	resNone := mustRequest(t, http.MethodGet, srv.URL+"/api/todos?project_id=none", nil)
	defer func() { _ = resNone.Body.Close() }()
	var noneTodos []map[string]any
	if err := json.NewDecoder(resNone.Body).Decode(&noneTodos); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(noneTodos) != 1 {
		t.Fatalf("expected 1 todo without project, got %d", len(noneTodos))
	}
	if noneTodos[0]["title"] != "No project" {
		t.Fatalf("expected 'No project', got %v", noneTodos[0]["title"])
	}
}

func TestIntegration_DeleteProject_ClearsTodoProjectID(t *testing.T) {
	srv, cleanup := newTestServer(t)
	defer cleanup()

	project := postProject(t, srv.URL+"/api/projects", "Homelab")
	projectID := project["id"].(float64)

	created := mustJSONRequest(t, http.MethodPost, srv.URL+"/api/todos", map[string]any{
		"title":      "In project",
		"project_id": projectID,
	})
	defer func() { _ = created.Body.Close() }()
	var todoBody map[string]any
	if err := json.NewDecoder(created.Body).Decode(&todoBody); err != nil {
		t.Fatalf("decode created todo: %v", err)
	}
	todoID := int64(todoBody["id"].(float64))

	resDelete := mustRequest(t, http.MethodDelete, fmt.Sprintf("%s/api/projects/%d", srv.URL, int64(projectID)), nil)
	defer func() { _ = resDelete.Body.Close() }()
	if resDelete.StatusCode != http.StatusNoContent {
		t.Fatalf("expected DELETE 204, got %d", resDelete.StatusCode)
	}

	resGet := mustRequest(t, http.MethodGet, fmt.Sprintf("%s/api/todos/%d", srv.URL, todoID), nil)
	defer func() { _ = resGet.Body.Close() }()
	var got map[string]any
	if err := json.NewDecoder(resGet.Body).Decode(&got); err != nil {
		t.Fatalf("decode todo: %v", err)
	}
	if got["project_id"] != nil {
		t.Fatalf("expected project_id to be cleared, got %v", got["project_id"])
	}
}
