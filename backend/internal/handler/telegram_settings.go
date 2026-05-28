package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/service"
)

type TelegramSettingsHandler struct {
	settings *service.TelegramSettingsService
	telegram *service.TelegramService
}

func NewTelegramSettingsHandler(settings *service.TelegramSettingsService, telegram *service.TelegramService) *TelegramSettingsHandler {
	return &TelegramSettingsHandler{settings: settings, telegram: telegram}
}

func (h *TelegramSettingsHandler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	view, err := h.settings.GetAdminView(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load telegram settings")
		return
	}
	view.Runtime = h.telegram.GetRuntimeStatus()
	writeJSON(w, http.StatusOK, view)
}

func (h *TelegramSettingsHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.telegram.GetRuntimeStatus())
}

func (h *TelegramSettingsHandler) Restart(w http.ResponseWriter, r *http.Request) {
	st := h.telegram.Restart(r.Context())
	writeJSON(w, http.StatusOK, st)
}

func (h *TelegramSettingsHandler) UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	var req model.TelegramAdminUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	view, err := h.settings.Update(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidTelegramSettings) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update telegram settings")
		return
	}
	view.Runtime = h.telegram.Restart(r.Context())
	writeJSON(w, http.StatusOK, view)
}

func (h *TelegramSettingsHandler) SendTest(w http.ResponseWriter, r *http.Request) {
	ok, msg := h.telegram.SendTest(r.Context())
	result := model.TelegramTestResult{OK: ok, Message: msg}
	if ok {
		st := h.telegram.GetRuntimeStatus()
		result.Runtime = &st
	}
	writeJSON(w, http.StatusOK, result)
}
