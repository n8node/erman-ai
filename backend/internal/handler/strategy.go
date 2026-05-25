package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
	"github.com/erman-ai/erman-ai/internal/service"
	"github.com/go-chi/chi/v5"
)

type StrategyHandler struct {
	strategy *service.StrategyService
	auth     *service.AuthService
	billing  *service.BillingService
}

func NewStrategyHandler(strategy *service.StrategyService, auth *service.AuthService, billing *service.BillingService) *StrategyHandler {
	return &StrategyHandler{strategy: strategy, auth: auth, billing: billing}
}

func (h *StrategyHandler) Run(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input model.StrategyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.auth.GetMe(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	run, err := h.strategy.StartRun(r.Context(), userID, input, user.Locale)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid input")
		case errors.Is(err, service.ErrToolLimitExceeded):
			writeErrorCode(w, http.StatusPaymentRequired, "strategy run limit exceeded", "tool_limit_exceeded")
		case errors.Is(err, repository.ErrNotFound):
			writeError(w, http.StatusBadRequest, "calculator run not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to start strategy run")
		}
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"run_id": run.ID,
		"status": run.Status,
	})
}

func (h *StrategyHandler) Stream(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	runID := chi.URLParam(r, "id")
	if err := h.strategy.CanStreamRun(r.Context(), runID, userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "run not found")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid run")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	hub := h.strategy.StreamHub()
	ch, replay, closed := hub.Subscribe(runID)
	defer hub.Unsubscribe(runID, ch)

	writeSSE := func(ev service.StrategyStreamEvent) {
		fmt.Fprintf(w, "event: %s\n", ev.Type)
		fmt.Fprintf(w, "data: %s\n\n", string(ev.Data))
		flusher.Flush()
	}

	for _, ev := range replay {
		writeSSE(ev)
		if ev.Type == service.StrategyEventDone || ev.Type == service.StrategyEventRunError {
			return
		}
	}
	if closed {
		return
	}

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		case ev, open := <-ch:
			if !open {
				return
			}
			writeSSE(ev)
			if ev.Type == service.StrategyEventDone || ev.Type == service.StrategyEventRunError {
				return
			}
		}
	}
}

func (h *StrategyHandler) Export(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.billing.CanExportPDF(r.Context(), userID); err != nil {
		if errors.Is(err, service.ErrFeatureNotAvailable) {
			writeError(w, http.StatusPaymentRequired, "pdf export requires pro plan")
			return
		}
		writeError(w, http.StatusInternalServerError, "billing check failed")
		return
	}

	var req exportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RunID == "" {
		writeError(w, http.StatusBadRequest, "run_id required")
		return
	}

	user, err := h.auth.GetMe(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	run, err := h.strategy.GetRunForUser(r.Context(), req.RunID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "run not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "run lookup failed")
		return
	}
	if run.ToolSlug != "strategy" || run.Status != model.RunStatusDone {
		writeError(w, http.StatusBadRequest, "strategy run not ready")
		return
	}

	in, out, err := service.ParseStrategyRun(run)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid run data")
		return
	}

	pdfBytes, err := service.GenerateStrategyPDF(in, out, user.Locale)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "pdf generation failed")
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="strategy-report.pdf"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdfBytes)
}
