package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
	"github.com/erman-ai/erman-ai/internal/service"
	"github.com/go-chi/chi/v5"
)

type RunsHandler struct {
	runs *service.RunService
}

func NewRunsHandler(runs *service.RunService) *RunsHandler {
	return &RunsHandler{runs: runs}
}

func (h *RunsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	toolSlug := r.URL.Query().Get("tool_slug")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	items, total, err := h.runs.List(r.Context(), userID, toolSlug, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list runs")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": items,
		"total": total,
	})
}

func (h *RunsHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	runID := chi.URLParam(r, "id")
	run, err := h.runs.Get(r.Context(), runID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "run not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load run")
		return
	}

	resp := map[string]any{
		"id":           run.ID,
		"tool_slug":    run.ToolSlug,
		"status":       run.Status,
		"created_at":   run.CreatedAt,
		"updated_at":   run.UpdatedAt,
		"artifact_url": run.ArtifactURL,
	}
	if run.ErrorMsg != nil {
		resp["error_msg"] = *run.ErrorMsg
	}
	if run.ModelUsed != nil {
		resp["model_used"] = *run.ModelUsed
	}
	if run.CompletedAt != nil {
		resp["completed_at"] = run.CompletedAt
	}

	if run.ToolSlug == "calculator" {
		var input model.CalculatorInput
		var output model.CalculatorOutput
		if err := json.Unmarshal(run.Input, &input); err == nil {
			resp["input"] = input
		}
		if err := json.Unmarshal(run.Output, &output); err == nil {
			resp["output"] = output
		}
	} else if run.ToolSlug == "strategy" {
		var input model.StrategyInput
		var output model.StrategyOutput
		if err := json.Unmarshal(run.Input, &input); err == nil {
			resp["input"] = input
		}
		if len(run.Output) > 0 {
			if err := json.Unmarshal(run.Output, &output); err == nil {
				resp["output"] = output
			}
		}
	} else {
		resp["input"] = json.RawMessage(run.Input)
		if len(run.Output) > 0 {
			resp["output"] = json.RawMessage(run.Output)
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *RunsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	runID := chi.URLParam(r, "id")
	if err := h.runs.Delete(r.Context(), runID, userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "run not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete run")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
