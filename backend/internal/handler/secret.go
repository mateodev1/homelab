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
	"github.com/mateo/homelab/backend/internal/service"
)

// SecretServicer is the service interface the handler depends on.
type SecretServicer interface {
	CreateProduct(ctx context.Context, name string, createdAt time.Time) (*domain.Product, error)
	ListProducts(ctx context.Context) ([]*domain.Product, error)

	CreateSecretProject(ctx context.Context, productID int64, name string, createdAt time.Time) (*domain.SecretProject, []*domain.SecretEnvironment, error)
	ListSecretProjects(ctx context.Context, productID int64) ([]*domain.SecretProject, error)
	ListEnvironments(ctx context.Context, projectID int64) ([]*domain.SecretEnvironment, error)

	CreateSecret(ctx context.Context, projectID int64, envName, key, value, actor string) (service.SecretMeta, error)
	ListSecrets(ctx context.Context, projectID int64, envName string) ([]service.SecretMeta, error)
	GetSecret(ctx context.Context, projectID int64, envName, key string) (service.SecretMeta, error)
	RevealSecret(ctx context.Context, projectID int64, envName, key, actor string) (string, error)
	UpdateSecret(ctx context.Context, projectID int64, envName, key, value, actor string) (service.SecretMeta, error)
	DeleteSecret(ctx context.Context, projectID int64, envName, key, actor string) error
	ExportEnvironment(ctx context.Context, projectID int64, envName, actor string) (string, error)
	ImportEnvironment(ctx context.Context, projectID int64, envName, content, actor string) (int, error)
}

// SecretHandler handles HTTP requests for the secret manager hierarchy
// (Product -> SecretProject -> SecretEnvironment -> Secret).
type SecretHandler struct {
	svc    SecretServicer
	apiKey string
}

// NewSecretHandler creates a new SecretHandler. apiKey is used only to
// classify the actor label recorded on audit log entries (api-key vs jwt);
// it performs no authentication itself — that's AuthMiddleware's job.
func NewSecretHandler(svc SecretServicer, apiKey string) *SecretHandler {
	return &SecretHandler{svc: svc, apiKey: apiKey}
}

// Register wires all secret-manager routes into the given ServeMux using
// Go 1.22+ method+pattern routing.
func (h *SecretHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/products", h.ListProducts)
	mux.HandleFunc("POST /api/products", h.CreateProduct)

	mux.HandleFunc("GET /api/products/{productID}/projects", h.ListProjects)
	mux.HandleFunc("POST /api/products/{productID}/projects", h.CreateProject)

	mux.HandleFunc("GET /api/products/{productID}/projects/{projectID}/environments", h.ListEnvironments)

	mux.HandleFunc("GET /api/products/{productID}/projects/{projectID}/environments/{envName}/secrets", h.ListSecrets)
	mux.HandleFunc("POST /api/products/{productID}/projects/{projectID}/environments/{envName}/secrets", h.CreateSecret)
	mux.HandleFunc("GET /api/products/{productID}/projects/{projectID}/environments/{envName}/secrets/{key}", h.GetSecret)
	mux.HandleFunc("PUT /api/products/{productID}/projects/{projectID}/environments/{envName}/secrets/{key}", h.UpdateSecret)
	mux.HandleFunc("DELETE /api/products/{productID}/projects/{projectID}/environments/{envName}/secrets/{key}", h.DeleteSecret)
	mux.HandleFunc("GET /api/products/{productID}/projects/{projectID}/environments/{envName}/secrets/{key}/reveal", h.RevealSecret)

	mux.HandleFunc("GET /api/products/{productID}/projects/{projectID}/environments/{envName}/export", h.Export)
	mux.HandleFunc("POST /api/products/{productID}/projects/{projectID}/environments/{envName}/import", h.Import)
}

// --- Products ---

func (h *SecretHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.svc.ListProducts(r.Context())
	if err != nil {
		jsonError(w, "failed to list products", http.StatusInternalServerError)
		return
	}
	out := make([]map[string]any, len(products))
	for i, p := range products {
		out[i] = productResponse(p)
	}
	jsonOK(w, out)
}

func (h *SecretHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	product, err := h.svc.CreateProduct(r.Context(), req.Name, time.Now())
	if err != nil {
		if isValidationErr(err) {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonError(w, "failed to create product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(productResponse(product))
}

// --- Secret Projects ---

func (h *SecretHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	productID, err := pathInt64(r, "productID")
	if err != nil {
		jsonError(w, "invalid product id", http.StatusBadRequest)
		return
	}

	projects, err := h.svc.ListSecretProjects(r.Context(), productID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		jsonError(w, "failed to list projects", http.StatusInternalServerError)
		return
	}

	out := make([]map[string]any, len(projects))
	for i, p := range projects {
		out[i] = secretProjectResponse(p)
	}
	jsonOK(w, out)
}

func (h *SecretHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	productID, err := pathInt64(r, "productID")
	if err != nil {
		jsonError(w, "invalid product id", http.StatusBadRequest)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	project, envs, err := h.svc.CreateSecretProject(r.Context(), productID, req.Name, time.Now())
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, "product not found", http.StatusNotFound)
			return
		}
		if isValidationErr(err) {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonError(w, "failed to create project", http.StatusInternalServerError)
		return
	}

	envOut := make([]map[string]any, len(envs))
	for i, e := range envs {
		envOut[i] = environmentResponse(e)
	}

	resp := secretProjectResponse(project)
	resp["environments"] = envOut

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// --- Environments ---

func (h *SecretHandler) ListEnvironments(w http.ResponseWriter, r *http.Request) {
	projectID, err := pathInt64(r, "projectID")
	if err != nil {
		jsonError(w, "invalid project id", http.StatusBadRequest)
		return
	}

	envs, err := h.svc.ListEnvironments(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		jsonError(w, "failed to list environments", http.StatusInternalServerError)
		return
	}

	out := make([]map[string]any, len(envs))
	for i, e := range envs {
		out[i] = environmentResponse(e)
	}
	jsonOK(w, out)
}

// --- Secrets ---

func (h *SecretHandler) ListSecrets(w http.ResponseWriter, r *http.Request) {
	projectID, err := pathInt64(r, "projectID")
	if err != nil {
		jsonError(w, "invalid project id", http.StatusBadRequest)
		return
	}
	envName := r.PathValue("envName")

	secrets, err := h.svc.ListSecrets(r.Context(), projectID, envName)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		jsonError(w, "failed to list secrets", http.StatusInternalServerError)
		return
	}

	out := make([]map[string]any, len(secrets))
	for i, s := range secrets {
		out[i] = secretMetaResponse(s)
	}
	jsonOK(w, out)
}

func (h *SecretHandler) CreateSecret(w http.ResponseWriter, r *http.Request) {
	projectID, err := pathInt64(r, "projectID")
	if err != nil {
		jsonError(w, "invalid project id", http.StatusBadRequest)
		return
	}
	envName := r.PathValue("envName")

	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	meta, err := h.svc.CreateSecret(r.Context(), projectID, envName, req.Key, req.Value, h.actor(r))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		if isValidationErr(err) {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonError(w, "failed to create secret", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(secretMetaResponse(meta))
}

func (h *SecretHandler) GetSecret(w http.ResponseWriter, r *http.Request) {
	projectID, err := pathInt64(r, "projectID")
	if err != nil {
		jsonError(w, "invalid project id", http.StatusBadRequest)
		return
	}
	envName := r.PathValue("envName")
	key := r.PathValue("key")

	meta, err := h.svc.GetSecret(r.Context(), projectID, envName, key)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	jsonOK(w, secretMetaResponse(meta))
}

func (h *SecretHandler) UpdateSecret(w http.ResponseWriter, r *http.Request) {
	projectID, err := pathInt64(r, "projectID")
	if err != nil {
		jsonError(w, "invalid project id", http.StatusBadRequest)
		return
	}
	envName := r.PathValue("envName")
	key := r.PathValue("key")

	var req struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	meta, err := h.svc.UpdateSecret(r.Context(), projectID, envName, key, req.Value, h.actor(r))
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
	jsonOK(w, secretMetaResponse(meta))
}

func (h *SecretHandler) DeleteSecret(w http.ResponseWriter, r *http.Request) {
	projectID, err := pathInt64(r, "projectID")
	if err != nil {
		jsonError(w, "invalid project id", http.StatusBadRequest)
		return
	}
	envName := r.PathValue("envName")
	key := r.PathValue("key")

	if err := h.svc.DeleteSecret(r.Context(), projectID, envName, key, h.actor(r)); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SecretHandler) RevealSecret(w http.ResponseWriter, r *http.Request) {
	projectID, err := pathInt64(r, "projectID")
	if err != nil {
		jsonError(w, "invalid project id", http.StatusBadRequest)
		return
	}
	envName := r.PathValue("envName")
	key := r.PathValue("key")

	plaintext, err := h.svc.RevealSecret(r.Context(), projectID, envName, key, h.actor(r))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]string{"key": key, "value": plaintext})
}

// Export handles GET .../export, returning a text/plain dotenv document with
// every secret in the environment.
func (h *SecretHandler) Export(w http.ResponseWriter, r *http.Request) {
	projectID, err := pathInt64(r, "projectID")
	if err != nil {
		jsonError(w, "invalid project id", http.StatusBadRequest)
		return
	}
	envName := r.PathValue("envName")

	body, err := h.svc.ExportEnvironment(r.Context(), projectID, envName, h.actor(r))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}

// Import parses and upserts a complete dotenv document into the environment.
// Keys missing from the document are intentionally preserved.
func (h *SecretHandler) Import(w http.ResponseWriter, r *http.Request) {
	projectID, err := pathInt64(r, "projectID")
	if err != nil {
		jsonError(w, "invalid project id", http.StatusBadRequest)
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	count, err := h.svc.ImportEnvironment(r.Context(), projectID, r.PathValue("envName"), req.Content, h.actor(r))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		if strings.HasPrefix(err.Error(), "invalid dotenv") {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonError(w, "failed to import secrets", http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]int{"imported": count})
}

// --- helpers ---

// actor classifies the caller for audit logging. There is no user identity
// model in the current auth layer (see AuthMiddleware), so the caller is
// labelled by credential type only.
func (h *SecretHandler) actor(r *http.Request) string {
	const prefix = "Bearer "
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, prefix) {
		return "anonymous"
	}
	token := strings.TrimPrefix(auth, prefix)
	if h.apiKey != "" && token == h.apiKey {
		return "api-key"
	}
	return "jwt"
}

// pathInt64 parses a numeric path value set by the enhanced ServeMux pattern.
func pathInt64(r *http.Request, name string) (int64, error) {
	return strconv.ParseInt(r.PathValue(name), 10, 64)
}

func productResponse(p *domain.Product) map[string]any {
	return map[string]any{
		"id":         p.ID,
		"name":       p.Name,
		"created_at": p.CreatedAt.Format(time.RFC3339),
	}
}

func secretProjectResponse(p *domain.SecretProject) map[string]any {
	return map[string]any{
		"id":         p.ID,
		"product_id": p.ProductID,
		"name":       p.Name,
		"created_at": p.CreatedAt.Format(time.RFC3339),
	}
}

func environmentResponse(e *domain.SecretEnvironment) map[string]any {
	return map[string]any{
		"id":         e.ID,
		"project_id": e.ProjectID,
		"name":       e.Name,
		"created_at": e.CreatedAt.Format(time.RFC3339),
	}
}

func secretMetaResponse(s service.SecretMeta) map[string]any {
	return map[string]any{
		"environment_id": s.EnvironmentID,
		"key":            s.Key,
		"value":          s.MaskedValue,
		"created_at":     s.CreatedAt.Format(time.RFC3339),
		"updated_at":     s.UpdatedAt.Format(time.RFC3339),
	}
}
