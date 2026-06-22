package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mateo/homelab/backend/internal/domain"
)

// TodoServicer is the service interface the handler depends on.
// Defined here to satisfy the dependency-inversion principle —
// the handler package owns the interface it needs.
type TodoServicer interface {
	CreateTodo(ctx context.Context, title string, createdAt time.Time) (*domain.Todo, error)
	ListTodos(ctx context.Context, done *bool) ([]*domain.Todo, error)
	GetTodo(ctx context.Context, id int64) (*domain.Todo, error)
	UpdateTodo(ctx context.Context, id int64, title string, done bool) (*domain.Todo, error)
	DeleteTodo(ctx context.Context, id int64) error
}

// TodoHandler handles HTTP requests for Todo resources.
type TodoHandler struct {
	svc TodoServicer
}

// NewTodoHandler creates a new TodoHandler.
func NewTodoHandler(svc TodoServicer) *TodoHandler {
	return &TodoHandler{svc: svc}
}

// Register wires all Todo routes into the given ServeMux.
func (h *TodoHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/todos", h.collection)
	mux.HandleFunc("/api/todos/", h.item)
}

// collection handles requests to /api/todos (no trailing ID).
func (h *TodoHandler) collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListTodos(w, r)
	case http.MethodPost:
		h.CreateTodo(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// item handles requests to /api/todos/{id}.
func (h *TodoHandler) item(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetTodo(w, r)
	case http.MethodPut:
		h.UpdateTodo(w, r)
	case http.MethodDelete:
		h.DeleteTodo(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// ListTodos handles GET /api/todos.
func (h *TodoHandler) ListTodos(w http.ResponseWriter, r *http.Request) {
	var done *bool
	if raw := r.URL.Query().Get("done"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			jsonError(w, "invalid done param", http.StatusBadRequest)
			return
		}
		done = &parsed
	}

	todos, err := h.svc.ListTodos(r.Context(), done)
	if err != nil {
		jsonError(w, "failed to list todos", http.StatusInternalServerError)
		return
	}
	// Ensure we encode [] not null for an empty slice.
	if todos == nil {
		todos = []*domain.Todo{}
	}
	jsonOK(w, todos)
}

// CreateTodo handles POST /api/todos.
func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Title == "" {
		jsonError(w, "title is required", http.StatusBadRequest)
		return
	}

	todo, err := h.svc.CreateTodo(r.Context(), req.Title, time.Now())
	if err != nil {
		jsonError(w, "failed to create todo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(todoResponse(todo))
}

// GetTodo handles GET /api/todos/{id}.
func (h *TodoHandler) GetTodo(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r.URL.Path)
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}

	todo, err := h.svc.GetTodo(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	jsonOK(w, todoResponse(todo))
}

// UpdateTodo handles PUT /api/todos/{id}.
func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r.URL.Path)
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req struct {
		Title string `json:"title"`
		Done  bool   `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		jsonError(w, "title is required", http.StatusBadRequest)
		return
	}

	todo, err := h.svc.UpdateTodo(r.Context(), id, req.Title, req.Done)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	jsonOK(w, todoResponse(todo))
}

// DeleteTodo handles DELETE /api/todos/{id}.
func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r.URL.Path)
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteTodo(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- helpers ---

// idFromPath extracts the numeric ID from a path like /api/todos/42.
func idFromPath(path string) (int64, error) {
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	raw := parts[len(parts)-1]
	return strconv.ParseInt(raw, 10, 64)
}

// todoResponse converts a domain.Todo to a JSON-serialisable map.
func todoResponse(t *domain.Todo) map[string]any {
	return map[string]any{
		"id":         t.ID,
		"title":      t.Title,
		"done":       t.Done,
		"created_at": t.CreatedAt.Format(time.RFC3339),
	}
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
