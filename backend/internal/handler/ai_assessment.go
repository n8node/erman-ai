package handler

import (
	"encoding/json"
	"errors"
	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/service"
	"net/http"
)

type AIAssessmentHandler struct {
	svc  *service.AIAssessmentService
	auth *service.AuthService
}

func NewAIAssessmentHandler(svc *service.AIAssessmentService, auth *service.AuthService) *AIAssessmentHandler {
	return &AIAssessmentHandler{svc: svc, auth: auth}
}

func (h *AIAssessmentHandler) Run(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var in model.AIAssessmentInput
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	u, err := h.auth.GetMe(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	run, err := h.svc.StartRun(r.Context(), userID, in, u.Locale)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "description is required")
		} else if errors.Is(err, service.ErrToolLimitExceeded) {
			writeErrorCode(w, http.StatusPaymentRequired, "AI assessment limit exceeded", "tool_limit_exceeded")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to start AI assessment")
		}
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"run_id": run.ID, "status": run.Status})
}

func (h *AIAssessmentHandler) Preview(w http.ResponseWriter, r *http.Request) {
	var in model.AIAssessmentInput
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	// Preview deliberately remains deterministic and does not call the LLM.
	out := model.AIAssessmentOutput{ProcessName: "Предварительная оценка", Financial: service.CalculateAssessmentFinancial(in, in.AutomationPct)}
	writeJSON(w, http.StatusOK, out)
}
