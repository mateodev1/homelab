package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// HealthHandler handles the /api/health endpoint.
type HealthHandler struct {
	db *sql.DB // optional; may be nil in tests
}

// NewHealthHandler creates a HealthHandler. db may be nil (health will report DBOk: false).
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Register wires the health route into the given ServeMux.
func (h *HealthHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/health", h.GetHealth)
}

// GetHealth handles GET /api/health.
func (h *HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	dbOk := false
	if h.db != nil {
		dbOk = h.db.PingContext(r.Context()) == nil
	}

	resp := map[string]any{
		"status": "ok",
		"db_ok":  dbOk,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
