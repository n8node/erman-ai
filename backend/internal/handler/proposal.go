package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
	"github.com/erman-ai/erman-ai/internal/service"
)

type ProposalHandler struct {
	proposal *service.ProposalService
	auth     *service.AuthService
	billing  *service.BillingService
}

func NewProposalHandler(proposal *service.ProposalService, auth *service.AuthService, billing *service.BillingService) *ProposalHandler {
	return &ProposalHandler{proposal: proposal, auth: auth, billing: billing}
}

func (h *ProposalHandler) Run(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input model.ProposalInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.auth.GetMe(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	run, err := h.proposal.StartRun(r.Context(), userID, input, user.Locale)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid input")
		case errors.Is(err, service.ErrToolLimitExceeded):
			writeErrorCode(w, http.StatusPaymentRequired, "proposal run limit exceeded", "tool_limit_exceeded")
		case errors.Is(err, repository.ErrNotFound):
			writeError(w, http.StatusBadRequest, "calculator run not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to start proposal run")
		}
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"run_id": run.ID,
		"status": run.Status,
	})
}

func (h *ProposalHandler) Export(w http.ResponseWriter, r *http.Request) {
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

	run, err := h.proposal.GetRunForUser(r.Context(), req.RunID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "run not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "run lookup failed")
		return
	}
	if run.Status != model.RunStatusDone {
		writeError(w, http.StatusBadRequest, "proposal run not ready")
		return
	}

	in, out, err := service.ParseProposalRun(run)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid run data")
		return
	}

	pdfBytes, err := service.GenerateProposalPDF(in, out, user.Locale)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "pdf generation failed")
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="proposal.pdf"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdfBytes)
}
