package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mateo/homelab/backend/internal/domain"
	"github.com/mateo/homelab/backend/internal/handler"
)

// --- in-test mock service ---

type mockTodoService struct {
	todos []*domain.Todo
	err   error
}

func (m *mockTodoService) CreateTodo(_ context.Context, title, body, color string, createdAt time.Time) (*domain.Todo, error) {
	if m.err != nil {
		return nil, m.err
	}
	t := &domain.Todo{ID: int64(len(m.todos) + 1), Title: title, Body: body, Color: color, Done: false, CreatedAt: createdAt}
	m.todos = append(m.todos, t)
	return t, nil
}

func (m *mockTodoService) ListTodos(_ context.Context, done *bool) ([]*domain.Todo, error) {
	if m.err != nil {
		return nil, m.err
	}
	if done == nil {
		return m.todos, nil
	}
	filtered := make([]*domain.Todo, 0, len(m.todos))
	for _, t := range m.todos {
		if t.Done == *done {
			filtered = append(filtered, t)
		}
	}
	return filtered, nil
}

func (m *mockTodoService) GetTodo(_ context.Context, id int64) (*domain.Todo, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, t := range m.todos {
		if t.ID == id {
			cp := *t
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockTodoService) UpdateTodo(_ context.Context, id int64, title, body, color string, pinned, done bool) (*domain.Todo, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, t := range m.todos {
		if t.ID == id {
			t.Title = title
			t.Body = body
			t.Color = color
			t.Pinned = pinned
			t.Done = done
			cp := *t
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockTodoService) DeleteTodo(_ context.Context, id int64) error {
	if m.err != nil {
		return m.err
	}
	for i, t := range m.todos {
		if t.ID == id {
			m.todos = append(m.todos[:i], m.todos[i+1:]...)
			return nil
		}
	}
	return domain.ErrNotFound
}

// Compile-time assertion: mockTodoService satisfies handler.TodoServicer.
var _ handler.TodoServicer = (*mockTodoService)(nil)

// buildMux wires a handler into a test ServeMux.
func buildMux(svc handler.TodoServicer) http.Handler {
	h := handler.NewTodoHandler(svc)
	hh := handler.NewHealthHandler(nil)
	mux := http.NewServeMux()
	h.Register(mux)
	hh.Register(mux)
	return mux
}

// --- tests ---

func TestListTodos_OK(t *testing.T) {
	t.Parallel()

	ms := &mockTodoService{todos: []*domain.Todo{
		{ID: 1, Title: "First", Done: false, CreatedAt: time.Now()},
		{ID: 2, Title: "Second", Done: true, CreatedAt: time.Now()},
	}}
	mux := buildMux(ms)

	req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(body) != 2 {
		t.Errorf("expected 2 todos in response, got %d", len(body))
	}
}

func TestListTodos_EmptyReturnsArray(t *testing.T) {
	t.Parallel()

	mux := buildMux(&mockTodoService{todos: []*domain.Todo{}})

	req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	// Must return JSON array, not null.
	var body []any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body) != 0 {
		t.Errorf("expected empty array, got %d items", len(body))
	}
}

func TestListTodos_DoneFilter_True(t *testing.T) {
	t.Parallel()

	ms := &mockTodoService{todos: []*domain.Todo{
		{ID: 1, Title: "A", Done: true, CreatedAt: time.Now()},
		{ID: 2, Title: "B", Done: false, CreatedAt: time.Now()},
		{ID: 3, Title: "C", Done: true, CreatedAt: time.Now()},
	}}
	mux := buildMux(ms)

	req := httptest.NewRequest(http.MethodGet, "/api/todos?done=true", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(body) != 2 {
		t.Fatalf("expected 2 todos in response, got %d", len(body))
	}
	for i, todo := range body {
		if done, ok := todo["done"].(bool); !ok || !done {
			t.Fatalf("todo[%d] expected done=true, got %v", i, todo["done"])
		}
	}
}

func TestListTodos_DoneFilter_InvalidParam(t *testing.T) {
	t.Parallel()

	mux := buildMux(&mockTodoService{})

	req := httptest.NewRequest(http.MethodGet, "/api/todos?done=banana", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["error"] != "invalid done param" {
		t.Fatalf("expected invalid done param error, got %q", body["error"])
	}
}

func TestCreateTodo_Created(t *testing.T) {
	t.Parallel()

	ms := &mockTodoService{}
	mux := buildMux(ms)

	body := `{"title":"New task"}`
	req := httptest.NewRequest(http.MethodPost, "/api/todos", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d — body: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["title"] != "New task" {
		t.Errorf("expected title \"New task\", got %v", resp["title"])
	}
	if _, ok := resp["id"]; !ok {
		t.Error("expected id field in response")
	}
}

func TestCreateTodo_BadJSON(t *testing.T) {
	t.Parallel()

	mux := buildMux(&mockTodoService{})

	req := httptest.NewRequest(http.MethodPost, "/api/todos", strings.NewReader("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestGetTodo_Found(t *testing.T) {
	t.Parallel()

	ms := &mockTodoService{todos: []*domain.Todo{
		{ID: 7, Title: "Find me", Done: false, CreatedAt: time.Now()},
	}}
	mux := buildMux(ms)

	req := httptest.NewRequest(http.MethodGet, "/api/todos/7", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["title"] != "Find me" {
		t.Errorf("expected title \"Find me\", got %v", resp["title"])
	}
}

func TestGetTodo_NotFound(t *testing.T) {
	t.Parallel()

	mux := buildMux(&mockTodoService{})

	req := httptest.NewRequest(http.MethodGet, "/api/todos/999", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["error"] != "not found" {
		t.Fatalf("expected not found error, got %q", body["error"])
	}
}

func TestGetTodo_DBError(t *testing.T) {
	t.Parallel()

	mux := buildMux(&mockTodoService{err: errors.New("db down")})

	req := httptest.NewRequest(http.MethodGet, "/api/todos/999", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestUpdateTodo_OK(t *testing.T) {
	t.Parallel()

	ms := &mockTodoService{todos: []*domain.Todo{
		{ID: 3, Title: "Before", Done: false, CreatedAt: time.Now()},
	}}
	mux := buildMux(ms)

	body := `{"title":"After","done":true}`
	req := httptest.NewRequest(http.MethodPut, "/api/todos/3", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["title"] != "After" {
		t.Errorf("expected \"After\", got %v", resp["title"])
	}
	if resp["done"] != true {
		t.Errorf("expected done=true, got %v", resp["done"])
	}
}

func TestUpdateTodo_BlankTitle(t *testing.T) {
	t.Parallel()

	mux := buildMux(&mockTodoService{todos: []*domain.Todo{{ID: 1, Title: "Before", Done: false, CreatedAt: time.Now()}}})

	body := `{"title":"  ","done":true}`
	req := httptest.NewRequest(http.MethodPut, "/api/todos/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["error"] != "title is required" {
		t.Fatalf("expected title is required error, got %q", resp["error"])
	}
}

func TestUpdateTodo_NotFound(t *testing.T) {
	t.Parallel()

	mux := buildMux(&mockTodoService{})

	body := `{"title":"After","done":true}`
	req := httptest.NewRequest(http.MethodPut, "/api/todos/999", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDeleteTodo_NoContent(t *testing.T) {
	t.Parallel()

	ms := &mockTodoService{todos: []*domain.Todo{
		{ID: 5, Title: "Delete me", CreatedAt: time.Now()},
	}}
	mux := buildMux(ms)

	req := httptest.NewRequest(http.MethodDelete, "/api/todos/5", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestDeleteTodo_NotFound(t *testing.T) {
	t.Parallel()

	mux := buildMux(&mockTodoService{})

	req := httptest.NewRequest(http.MethodDelete, "/api/todos/9999", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestGetHealth_OK(t *testing.T) {
	t.Parallel()

	mux := buildMux(&mockTodoService{})

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status \"ok\", got %v", resp["status"])
	}
}
