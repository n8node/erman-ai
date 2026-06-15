package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/service"
)

type YandexMetrikaSettingsHandler struct {
	settings *service.YandexMetrikaSettingsService
}

func NewYandexMetrikaSettingsHandler(settings *service.YandexMetrikaSettingsService) *YandexMetrikaSettingsHandler {
	return &YandexMetrikaSettingsHandler{settings: settings}
}

func (h *YandexMetrikaSettingsHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	view, err := h.settings.GetPublicView(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load yandex metrika settings")
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (h *YandexMetrikaSettingsHandler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	view, err := h.settings.GetAdminView(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load yandex metrika settings")
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (h *YandexMetrikaSettingsHandler) UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	var req model.YandexMetrikaAdminUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	view, err := h.settings.Update(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidYandexMetrikaSettings) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update yandex metrika settings")
		return
	}
	writeJSON(w, http.StatusOK, view)
}
