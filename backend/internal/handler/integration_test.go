package handler_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/mateo/homelab/backend/internal/handler"
	"github.com/mateo/homelab/backend/internal/service"
	"github.com/mateo/homelab/backend/internal/store"
)

func newTestServer(t *testing.T) (*httptest.Server, func()) {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := store.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	todoStore := store.New(db)
	todoSvc := service.NewTodoService(todoStore)

	checker := &dbHealthChecker{db: db}

	mux := http.NewServeMux()
	todoHandler := handler.NewTodoHandler(todoSvc)
	healthHandler := handler.NewHealthHandler(checker)
	todoHandler.Register(mux)
	healthHandler.Register(mux)

	chain := handler.RecoveryMiddleware(handler.LoggingMiddleware(handler.CORSMiddleware(mux)))
	srv := httptest.NewServer(chain)

	return srv, func() {
		srv.Close()
		_ = db.Close()
	}
}

type dbHealthChecker struct{ db *sql.DB }

func (c *dbHealthChecker) Ping(ctx context.Context) error { return c.db.PingContext(ctx) }

func TestIntegration_CRUDCycle(t *testing.T) {
	srv, cleanup := newTestServer(t)
	defer cleanup()

	created := postTodo(t, srv.URL+"/api/todos", "first")
	id := int64(created["id"].(float64))

	resGet := mustRequest(t, http.MethodGet, fmt.Sprintf("%s/api/todos/%d", srv.URL, id), nil)
	defer resGet.Body.Close()
	if resGet.StatusCode != http.StatusOK {
		t.Fatalf("expected GET 200, got %d", resGet.StatusCode)
	}

	resUpdate := mustJSONRequest(t, http.MethodPut, fmt.Sprintf("%s/api/todos/%d", srv.URL, id), map[string]any{
		"title": "first updated",
		"done":  true,
	})
	defer resUpdate.Body.Close()
	if resUpdate.StatusCode != http.StatusOK {
		t.Fatalf("expected PUT 200, got %d", resUpdate.StatusCode)
	}

	resDelete := mustRequest(t, http.MethodDelete, fmt.Sprintf("%s/api/todos/%d", srv.URL, id), nil)
	defer resDelete.Body.Close()
	if resDelete.StatusCode != http.StatusNoContent {
		t.Fatalf("expected DELETE 204, got %d", resDelete.StatusCode)
	}

	resNotFound := mustRequest(t, http.MethodGet, fmt.Sprintf("%s/api/todos/%d", srv.URL, id), nil)
	defer resNotFound.Body.Close()
	if resNotFound.StatusCode != http.StatusNotFound {
		t.Fatalf("expected GET after delete 404, got %d", resNotFound.StatusCode)
	}
}

func TestIntegration_GetNotFound(t *testing.T) {
	srv, cleanup := newTestServer(t)
	defer cleanup()

	res := mustRequest(t, http.MethodGet, srv.URL+"/api/todos/999", nil)
	defer res.Body.Close()

	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["error"] != "not found" {
		t.Fatalf("expected error 'not found', got %q", body["error"])
	}
}

func TestIntegration_DoneFilter(t *testing.T) {
	srv, cleanup := newTestServer(t)
	defer cleanup()

	a := postTodo(t, srv.URL+"/api/todos", "a")
	b := postTodo(t, srv.URL+"/api/todos", "b")
	_ = postTodo(t, srv.URL+"/api/todos", "c")

	mustJSONStatus(t, http.MethodPut, fmt.Sprintf("%s/api/todos/%d", srv.URL, int64(a["id"].(float64))), map[string]any{"title": "a", "done": true}, http.StatusOK)
	mustJSONStatus(t, http.MethodPut, fmt.Sprintf("%s/api/todos/%d", srv.URL, int64(b["id"].(float64))), map[string]any{"title": "b", "done": true}, http.StatusOK)

	res := mustRequest(t, http.MethodGet, srv.URL+"/api/todos?done=true", nil)
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	var todos []map[string]any
	if err := json.NewDecoder(res.Body).Decode(&todos); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(todos) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(todos))
	}
	for i, todo := range todos {
		done, ok := todo["done"].(bool)
		if !ok || !done {
			t.Fatalf("todo[%d] expected done=true, got %v", i, todo["done"])
		}
	}
}

func TestIntegration_PUTBlankTitle(t *testing.T) {
	srv, cleanup := newTestServer(t)
	defer cleanup()

	created := postTodo(t, srv.URL+"/api/todos", "first")
	id := int64(created["id"].(float64))

	res := mustJSONRequest(t, http.MethodPut, fmt.Sprintf("%s/api/todos/%d", srv.URL, id), map[string]any{
		"title": "  ",
		"done":  true,
	})
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["error"] != "title is required" {
		t.Fatalf("expected title is required, got %q", body["error"])
	}
}

func TestIntegration_CORSPreflight(t *testing.T) {
	srv, cleanup := newTestServer(t)
	defer cleanup()

	res := mustRequest(t, http.MethodOptions, srv.URL+"/api/todos", nil)
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", res.StatusCode)
	}
	if got := res.Header.Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin '*', got %q", got)
	}
	if got := res.Header.Get("Access-Control-Allow-Methods"); got != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Fatalf("unexpected Access-Control-Allow-Methods: %q", got)
	}
	if got := res.Header.Get("Access-Control-Allow-Headers"); got != "Content-Type" {
		t.Fatalf("unexpected Access-Control-Allow-Headers: %q", got)
	}
}

func postTodo(t *testing.T, url, title string) map[string]any {
	t.Helper()
	res := mustJSONRequest(t, http.MethodPost, url, map[string]any{"title": title})
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode created todo: %v", err)
	}
	return body
}

func mustJSONStatus(t *testing.T, method, url string, payload map[string]any, want int) {
	t.Helper()
	res := mustJSONRequest(t, method, url, payload)
	defer res.Body.Close()
	if res.StatusCode != want {
		t.Fatalf("%s %s expected %d, got %d", method, url, want, res.StatusCode)
	}
}

func mustJSONRequest(t *testing.T, method, url string, payload map[string]any) *http.Response {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return res
}

func mustRequest(t *testing.T, method, url string, body *bytes.Reader) *http.Response {
	t.Helper()
	var reqBody io.Reader
	if body != nil {
		reqBody = body
	}
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return res
}
