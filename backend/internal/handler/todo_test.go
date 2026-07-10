package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mateo/homelab/backend/internal/domain"
	"github.com/mateo/homelab/backend/internal/handler"
	"github.com/mateo/homelab/backend/internal/service"
)

type mockTodoService struct {
	todos []*domain.Todo
	err   error
}

func (m *mockTodoService) CreateTodo(_ context.Context, title, body string, priority int, dueDate *string, kind string, issueType *string, projectID *int64, createdAt time.Time) (*domain.Todo, error) {
	if m.err != nil {
		return nil, m.err
	}
	if kind == "" {
		kind = domain.TodoKindNote
	}
	t := &domain.Todo{
		ID:        int64(len(m.todos) + 1),
		Title:     title,
		Body:      body,
		Status:    domain.TodoStatusTodo,
		Priority:  priority,
		DueDate:   dueDate,
		Kind:      kind,
		IssueType: issueType,
		ProjectID: projectID,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}
	m.todos = append(m.todos, t)
	return t, nil
}

func (m *mockTodoService) ListTodos(_ context.Context) ([]*domain.Todo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.todos, nil
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

func (m *mockTodoService) UpdateTodo(_ context.Context, id int64, patch service.TodoPatch) (*domain.Todo, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, t := range m.todos {
		if t.ID != id {
			continue
		}
		if patch.Title != nil {
			t.Title = *patch.Title
		}
		if patch.Body != nil {
			t.Body = *patch.Body
		}
		if patch.Status != nil {
			t.Status = *patch.Status
		}
		if patch.Priority != nil {
			t.Priority = *patch.Priority
		}
		if patch.DueDate != nil {
			t.DueDate = *patch.DueDate
		}
		if patch.Kind != nil {
			t.Kind = *patch.Kind
		}
		if patch.IssueType != nil {
			t.IssueType = *patch.IssueType
		}
		if patch.ProjectID != nil {
			t.ProjectID = *patch.ProjectID
		}
		if t.IssueType != nil && t.Kind != domain.TodoKindIssue {
			return nil, fmt.Errorf("issue_type can only be set when kind is issue: %w", service.ErrValidation)
		}
		cp := *t
		return &cp, nil
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

var _ handler.TodoServicer = (*mockTodoService)(nil)

func buildMux(svc handler.TodoServicer) http.Handler {
	h := handler.NewTodoHandler(svc)
	hh := handler.NewHealthHandler(nil)
	mux := http.NewServeMux()
	h.Register(mux)
	hh.Register(mux)
	return mux
}

func TestListTodos_OK(t *testing.T) {
	t.Parallel()

	ms := &mockTodoService{todos: []*domain.Todo{{
		ID:        1,
		Title:     "First",
		Status:    domain.TodoStatusTodo,
		Priority:  1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}}}
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
	if len(body) != 1 {
		t.Errorf("expected 1 todo in response, got %d", len(body))
	}
}

func TestCreateTodo_DefaultStatusAndShape(t *testing.T) {
	t.Parallel()

	mux := buildMux(&mockTodoService{})

	body := `{"title":"New task","priority":3}`
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
	if resp["status"] != domain.TodoStatusTodo {
		t.Errorf("expected status=todo, got %v", resp["status"])
	}
	if resp["priority"] != float64(3) {
		t.Errorf("expected priority=3, got %v", resp["priority"])
	}
	if _, ok := resp["done"]; ok {
		t.Fatal("response should not include done")
	}
	if _, ok := resp["color"]; ok {
		t.Fatal("response should not include color")
	}
	if _, ok := resp["pinned"]; ok {
		t.Fatal("response should not include pinned")
	}
}

func TestCreateTodo_InvalidPriority(t *testing.T) {
	t.Parallel()

	mux := buildMux(&mockTodoService{})

	body := `{"title":"bad","priority":7}`
	req := httptest.NewRequest(http.MethodPost, "/api/todos", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateTodo_DefaultKindNote(t *testing.T) {
	t.Parallel()

	mux := buildMux(&mockTodoService{})

	body := `{"title":"New note"}`
	req := httptest.NewRequest(http.MethodPost, "/api/todos", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["kind"] != domain.TodoKindNote {
		t.Errorf("expected kind=note, got %v", resp["kind"])
	}
	if resp["issue_type"] != nil {
		t.Errorf("expected issue_type=nil, got %v", resp["issue_type"])
	}
}

func TestCreateTodo_IssueWithType(t *testing.T) {
	t.Parallel()

	mux := buildMux(&mockTodoService{})

	body := `{"title":"Fix bug","kind":"issue","issue_type":"bug"}`
	req := httptest.NewRequest(http.MethodPost, "/api/todos", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["kind"] != domain.TodoKindIssue {
		t.Errorf("expected kind=issue, got %v", resp["kind"])
	}
	if resp["issue_type"] != domain.IssueTypeBug {
		t.Errorf("expected issue_type=bug, got %v", resp["issue_type"])
	}
}

func TestCreateTodo_InvalidKind(t *testing.T) {
	t.Parallel()

	mux := buildMux(&mockTodoService{})

	body := `{"title":"bad","kind":"backlog"}`
	req := httptest.NewRequest(http.MethodPost, "/api/todos", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateTodo_IssueTypeRequiresIssueKind(t *testing.T) {
	t.Parallel()

	mux := buildMux(&mockTodoService{})

	body := `{"title":"bad","issue_type":"bug"}`
	req := httptest.NewRequest(http.MethodPost, "/api/todos", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestListTodos_FilterByKind(t *testing.T) {
	t.Parallel()

	issueType := domain.IssueTypeBug
	ms := &mockTodoService{todos: []*domain.Todo{
		{ID: 1, Title: "Note", Kind: domain.TodoKindNote, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: 2, Title: "Issue", Kind: domain.TodoKindIssue, IssueType: &issueType, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}}
	mux := buildMux(ms)

	req := httptest.NewRequest(http.MethodGet, "/api/todos?kind=issue", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(body) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(body))
	}
	if body[0]["kind"] != domain.TodoKindIssue {
		t.Errorf("expected kind=issue, got %v", body[0]["kind"])
	}
}

func TestListTodos_FilterByIssueType(t *testing.T) {
	t.Parallel()

	bugType := domain.IssueTypeBug
	featureType := domain.IssueTypeFeature
	ms := &mockTodoService{todos: []*domain.Todo{
		{ID: 1, Title: "Bug", Kind: domain.TodoKindIssue, IssueType: &bugType, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: 2, Title: "Feature", Kind: domain.TodoKindIssue, IssueType: &featureType, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}}
	mux := buildMux(ms)

	req := httptest.NewRequest(http.MethodGet, "/api/todos?issue_type=bug", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(body) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(body))
	}
	if body[0]["issue_type"] != domain.IssueTypeBug {
		t.Errorf("expected issue_type=bug, got %v", body[0]["issue_type"])
	}
}

func TestUpdateTodo_SetKindAndIssueType(t *testing.T) {
	t.Parallel()

	ms := &mockTodoService{todos: []*domain.Todo{{
		ID:        7,
		Title:     "Before",
		Status:    domain.TodoStatusTodo,
		Kind:      domain.TodoKindNote,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}}}
	mux := buildMux(ms)

	body := `{"kind":"issue","issue_type":"improvement"}`
	req := httptest.NewRequest(http.MethodPut, "/api/todos/7", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["kind"] != domain.TodoKindIssue {
		t.Errorf("expected kind=issue, got %v", resp["kind"])
	}
	if resp["issue_type"] != domain.IssueTypeImprovement {
		t.Errorf("expected issue_type=improvement, got %v", resp["issue_type"])
	}
}

func TestUpdateTodo_InvalidKind(t *testing.T) {
	t.Parallel()

	ms := &mockTodoService{todos: []*domain.Todo{{
		ID: 1, Title: "Before", Status: domain.TodoStatusTodo, Kind: domain.TodoKindNote,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}}}
	mux := buildMux(ms)

	body := `{"kind":"backlog"}`
	req := httptest.NewRequest(http.MethodPut, "/api/todos/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateTodo_IssueTypeRequiresIssueKind(t *testing.T) {
	t.Parallel()

	ms := &mockTodoService{todos: []*domain.Todo{{
		ID: 1, Title: "Before", Status: domain.TodoStatusTodo, Kind: domain.TodoKindNote,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}}}
	mux := buildMux(ms)

	body := `{"issue_type":"bug"}`
	req := httptest.NewRequest(http.MethodPut, "/api/todos/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d — body: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateTodo_OK(t *testing.T) {
	t.Parallel()

	ms := &mockTodoService{todos: []*domain.Todo{{
		ID:        3,
		Title:     "Before",
		Status:    domain.TodoStatusTodo,
		Priority:  1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}}}
	mux := buildMux(ms)

	body := `{"title":"After","status":"done","priority":3}`
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
		t.Errorf("expected title After, got %v", resp["title"])
	}
	if resp["status"] != domain.TodoStatusDone {
		t.Errorf("expected status done, got %v", resp["status"])
	}
}

func TestUpdateTodo_InvalidStatus(t *testing.T) {
	t.Parallel()

	ms := &mockTodoService{todos: []*domain.Todo{{
		ID:        1,
		Title:     "Before",
		Status:    domain.TodoStatusTodo,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}}}
	mux := buildMux(ms)

	body := `{"status":"blocked"}`
	req := httptest.NewRequest(http.MethodPut, "/api/todos/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateTodo_InvalidPriority(t *testing.T) {
	t.Parallel()

	ms := &mockTodoService{todos: []*domain.Todo{{
		ID:        1,
		Title:     "Before",
		Status:    domain.TodoStatusTodo,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}}}
	mux := buildMux(ms)

	body := `{"priority":9}`
	req := httptest.NewRequest(http.MethodPut, "/api/todos/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDeleteTodo_NoContent(t *testing.T) {
	t.Parallel()

	ms := &mockTodoService{todos: []*domain.Todo{{ID: 5, Title: "Delete me", CreatedAt: time.Now(), UpdatedAt: time.Now()}}}
	mux := buildMux(ms)

	req := httptest.NewRequest(http.MethodDelete, "/api/todos/5", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
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
