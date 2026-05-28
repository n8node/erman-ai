package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/service"
)

type PaymentSettingsHandler struct {
	svc *service.PaymentSettingsService
}

func NewPaymentSettingsHandler(svc *service.PaymentSettingsService) *PaymentSettingsHandler {
	return &PaymentSettingsHandler{svc: svc}
}

func (h *PaymentSettingsHandler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	view, err := h.svc.GetAdminView(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load payment settings")
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (h *PaymentSettingsHandler) UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	var req model.PaymentAdminUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	view, err := h.svc.Update(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPaymentSettings) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update payment settings")
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (h *PaymentSettingsHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	var req model.PaymentTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.svc.TestConnection(r.Context(), req.Provider)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPaymentSettings) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "connection test failed")
		return
	}
	writeJSON(w, http.StatusOK, result)
}
