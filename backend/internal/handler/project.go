package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/mateo/homelab/backend/internal/domain"
	"github.com/mateo/homelab/backend/internal/service"
)

// ProjectServicer is the service interface the handler depends on.
type ProjectServicer interface {
	CreateProject(ctx context.Context, name, color string, createdAt time.Time) (*domain.Project, error)
	ListProjects(ctx context.Context) ([]*domain.Project, error)
	GetProject(ctx context.Context, id int64) (*domain.Project, error)
	UpdateProject(ctx context.Context, id int64, patch service.ProjectPatch) (*domain.Project, error)
	DeleteProject(ctx context.Context, id int64) error
}

// ProjectHandler handles HTTP requests for Project resources.
type ProjectHandler struct {
	svc ProjectServicer
}

// NewProjectHandler creates a new ProjectHandler.
func NewProjectHandler(svc ProjectServicer) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

// Register wires all Project routes into the given ServeMux.
func (h *ProjectHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/projects", h.collection)
	mux.HandleFunc("/api/projects/", h.item)
}

func (h *ProjectHandler) collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListProjects(w, r)
	case http.MethodPost:
		h.CreateProject(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ProjectHandler) item(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetProject(w, r)
	case http.MethodPut:
		h.UpdateProject(w, r)
	case http.MethodDelete:
		h.DeleteProject(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// ListProjects handles GET /api/projects.
func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.svc.ListProjects(r.Context())
	if err != nil {
		jsonError(w, "failed to list projects", http.StatusInternalServerError)
		return
	}

	out := make([]map[string]any, len(projects))
	for i, p := range projects {
		out[i] = projectResponse(p)
	}
	jsonOK(w, out)
}

// CreateProject handles POST /api/projects.
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	project, err := h.svc.CreateProject(r.Context(), req.Name, req.Color, time.Now())
	if err != nil {
		if isValidationErr(err) {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonError(w, "failed to create project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(projectResponse(project))
}

// GetProject handles GET /api/projects/{id}.
func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r.URL.Path)
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}

	project, err := h.svc.GetProject(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	jsonOK(w, projectResponse(project))
}

// UpdateProject handles PUT /api/projects/{id}.
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r.URL.Path)
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req struct {
		Name  *string `json:"name"`
		Color *string `json:"color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		jsonError(w, "name is required", http.StatusBadRequest)
		return
	}

	project, err := h.svc.UpdateProject(r.Context(), id, service.ProjectPatch{
		Name:  req.Name,
		Color: req.Color,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		if isValidationErr(err) {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	jsonOK(w, projectResponse(project))
}

// DeleteProject handles DELETE /api/projects/{id}.
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r.URL.Path)
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteProject(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// projectResponse converts a domain.Project to a JSON-serialisable map.
func projectResponse(p *domain.Project) map[string]any {
	return map[string]any{
		"id":         p.ID,
		"name":       p.Name,
		"color":      p.Color,
		"created_at": p.CreatedAt.Format(time.RFC3339),
	}
}
