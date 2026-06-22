package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

// HealthChecker abstracts health checks from concrete database implementations.
type HealthChecker interface {
	Ping(ctx context.Context) error
}

// HealthHandler handles the /api/health endpoint.
type HealthHandler struct {
	checker HealthChecker // optional; may be nil in tests
}

// NewHealthHandler creates a HealthHandler. checker may be nil (health will report DBOk: false).
func NewHealthHandler(checker HealthChecker) *HealthHandler {
	return &HealthHandler{checker: checker}
}

// Register wires the health route into the given ServeMux.
func (h *HealthHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/health", h.GetHealth)
}

// GetHealth handles GET /api/health.
func (h *HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	dbOk := false
	if h.checker != nil {
		dbOk = h.checker.Ping(r.Context()) == nil
	}

	resp := map[string]any{
		"status": "ok",
		"db_ok":  dbOk,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
