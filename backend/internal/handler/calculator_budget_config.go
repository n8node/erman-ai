package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/service"
)

type CalculatorBudgetConfigHandler struct {
	svc *service.CalculatorBudgetConfigService
}

func NewCalculatorBudgetConfigHandler(svc *service.CalculatorBudgetConfigService) *CalculatorBudgetConfigHandler {
	return &CalculatorBudgetConfigHandler{svc: svc}
}

func (h *CalculatorBudgetConfigHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	rec, err := h.svc.Get(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load budget config")
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

func (h *CalculatorBudgetConfigHandler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	rec, err := h.svc.Get(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load budget config")
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

func (h *CalculatorBudgetConfigHandler) UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	var cfg model.CalculatorBudgetConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rec, err := h.svc.Update(r.Context(), cfg)
	if err != nil {
		if errors.Is(err, service.ErrInvalidBudgetConfig) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update budget config")
		return
	}
	writeJSON(w, http.StatusOK, rec)
}
