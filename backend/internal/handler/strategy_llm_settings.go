package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/service"
)

type StrategyLLMSettingsHandler struct {
	svc *service.StrategyLLMSettingsService
}

func NewStrategyLLMSettingsHandler(svc *service.StrategyLLMSettingsService) *StrategyLLMSettingsHandler {
	return &StrategyLLMSettingsHandler{svc: svc}
}

func (h *StrategyLLMSettingsHandler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	view, err := h.svc.GetAdminView(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load strategy llm settings")
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (h *StrategyLLMSettingsHandler) UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	var req model.StrategyLLMAdminUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	view, err := h.svc.Update(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidStrategyLLMSettings) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update strategy llm settings")
		return
	}
	writeJSON(w, http.StatusOK, view)
}

type testConnectionRequest struct {
	Provider model.LLMProvider `json:"provider"`
}

func (h *StrategyLLMSettingsHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	var req testConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.svc.TestConnection(r.Context(), req.Provider)
	if err != nil {
		if errors.Is(err, service.ErrInvalidStrategyLLMSettings) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "connection test failed")
		return
	}
	writeJSON(w, http.StatusOK, result)
}
