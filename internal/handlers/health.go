package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type healthHandler struct {
	pool *pgxpool.Pool
}

type healthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}

func (h *healthHandler) readyz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := h.pool.Ping(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(healthResponse{Status: "degraded", DB: err.Error()}) //nolint:errcheck
		return
	}

	json.NewEncoder(w).Encode(healthResponse{Status: "ok", DB: "ok"}) //nolint:errcheck
}
