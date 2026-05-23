package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/repository"
)

type HealthHandler struct {
	cfg  *config.Config
	db   *repository.Postgres
	start time.Time
}

func NewHealthHandler(cfg *config.Config, db *repository.Postgres) *HealthHandler {
	return &HealthHandler{cfg: cfg, db: db, start: time.Now()}
}

type healthResponse struct {
	Status        string            `json:"status"`
	Version       string            `json:"version"`
	UptimeSeconds int64             `json:"uptime_seconds"`
	Postgres      string            `json:"postgres"`
	Workers       map[string]int    `json:"workers"`
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	postgresStatus := "ok"
	if err := h.db.Ping(r.Context()); err != nil {
		postgresStatus = "error"
	}

	status := "ok"
	code := http.StatusOK
	if postgresStatus != "ok" {
		status = "degraded"
		code = http.StatusServiceUnavailable
	}

	resp := healthResponse{
		Status:        status,
		Version:       config.Version,
		UptimeSeconds: int64(time.Since(h.start).Seconds()),
		Postgres:      postgresStatus,
		Workers: map[string]int{
			"strategy_queue": 0,
			"proposal_queue": 0,
			"errors_24h":     0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(resp)
}
