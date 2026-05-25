package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/service"
)

type AuditHandler struct {
	audit *service.AuditService
	auth  *service.AuthService
}

func NewAuditHandler(audit *service.AuditService, auth *service.AuthService) *AuditHandler {
	return &AuditHandler{audit: audit, auth: auth}
}

func (h *AuditHandler) Run(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input model.AuditInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.auth.GetMe(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	run, err := h.audit.StartRun(r.Context(), userID, input, user.Locale)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid input")
		case errors.Is(err, service.ErrToolLimitExceeded):
			writeErrorCode(w, http.StatusPaymentRequired, "audit run limit exceeded", "tool_limit_exceeded")
		default:
			writeError(w, http.StatusInternalServerError, "failed to start audit run")
		}
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"run_id": run.ID,
		"status": run.Status,
	})
}
