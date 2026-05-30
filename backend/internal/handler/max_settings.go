package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/service"
)

type MaxSettingsHandler struct {
	settings *service.MaxSettingsService
	max      *service.MaxService
}

func NewMaxSettingsHandler(settings *service.MaxSettingsService, max *service.MaxService) *MaxSettingsHandler {
	return &MaxSettingsHandler{settings: settings, max: max}
}

func (h *MaxSettingsHandler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	view, err := h.settings.GetAdminView(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load max settings")
		return
	}
	view.Runtime = h.max.CheckHealth(r.Context())
	writeJSON(w, http.StatusOK, view)
}

func (h *MaxSettingsHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.max.CheckHealth(r.Context()))
}

func (h *MaxSettingsHandler) UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	var req model.MaxAdminUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	view, err := h.settings.Update(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidMaxSettings) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update max settings")
		return
	}
	view.Runtime = h.max.CheckHealth(r.Context())
	writeJSON(w, http.StatusOK, view)
}

func (h *MaxSettingsHandler) SendTest(w http.ResponseWriter, r *http.Request) {
	ok, msg := h.max.SendTest(r.Context())
	result := model.MaxTestResult{OK: ok, Message: msg}
	if ok {
		st := h.max.CheckHealth(r.Context())
		result.Runtime = &st
	}
	writeJSON(w, http.StatusOK, result)
}
