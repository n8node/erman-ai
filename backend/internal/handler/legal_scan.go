package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/service"
)

type LegalScanHandler struct {
	svc  *service.LegalScanService
	auth *service.AuthService
}

func NewLegalScanHandler(svc *service.LegalScanService, auth *service.AuthService) *LegalScanHandler {
	return &LegalScanHandler{svc: svc, auth: auth}
}

func (h *LegalScanHandler) Run(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input model.LegalScanInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.auth.GetMe(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	run, err := h.svc.StartRun(r.Context(), userID, input, user.Locale)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid input")
		case errors.Is(err, service.ErrToolLimitExceeded):
			writeErrorCode(w, http.StatusPaymentRequired, "legal scan limit exceeded", "tool_limit_exceeded")
		default:
			writeError(w, http.StatusInternalServerError, "failed to start legal scan")
		}
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"run_id": run.ID,
		"status": run.Status,
	})
}

type LegalRiskAdminHandler struct {
	svc *service.LegalRiskService
}

func NewLegalRiskAdminHandler(svc *service.LegalRiskService) *LegalRiskAdminHandler {
	return &LegalRiskAdminHandler{svc: svc}
}

func (h *LegalRiskAdminHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load legal risks")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *LegalRiskAdminHandler) Update(w http.ResponseWriter, r *http.Request) {
	riskID := r.PathValue("risk_id")
	var req model.LegalRiskUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := h.svc.Update(r.Context(), riskID, req)
	if err != nil {
		if errors.Is(err, service.ErrLegalRiskNotFound) {
			writeError(w, http.StatusNotFound, "risk not found")
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update risk")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

type legalRiskCreateRequest struct {
	RiskID string `json:"risk_id"`
	model.LegalRiskUpdateRequest
}

func (h *LegalRiskAdminHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req legalRiskCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := h.svc.Create(r.Context(), req.RiskID, req.LegalRiskUpdateRequest)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create risk")
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

type LegalScanLLMSettingsHandler struct {
	svc *service.LegalScanLLMSettingsService
}

func NewLegalScanLLMSettingsHandler(svc *service.LegalScanLLMSettingsService) *LegalScanLLMSettingsHandler {
	return &LegalScanLLMSettingsHandler{svc: svc}
}

func (h *LegalScanLLMSettingsHandler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	view, err := h.svc.GetAdminView(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load legal scan llm settings")
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (h *LegalScanLLMSettingsHandler) UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	var req model.LegalScanLLMAdminUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	view, err := h.svc.Update(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidLegalScanLLMSettings) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update legal scan llm settings")
		return
	}
	writeJSON(w, http.StatusOK, view)
}
