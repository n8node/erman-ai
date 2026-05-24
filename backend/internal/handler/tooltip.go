package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
	"github.com/erman-ai/erman-ai/internal/service"
	"github.com/go-chi/chi/v5"
)

type TooltipHandler struct {
	tooltips *service.TooltipService
	auth     *service.AuthService
}

func NewTooltipHandler(tooltips *service.TooltipService, auth *service.AuthService) *TooltipHandler {
	return &TooltipHandler{tooltips: tooltips, auth: auth}
}

func (h *TooltipHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	prefix := r.URL.Query().Get("prefix")
	if prefix == "" {
		prefix = "calculator"
	}

	user, err := h.auth.GetMe(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tooltips, err := h.tooltips.ListForLocale(r.Context(), prefix, user.Locale)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load tooltips")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"tooltips": tooltips})
}

func (h *TooltipHandler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	items, err := h.tooltips.ListAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load tooltips")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

type updateTooltipRequest struct {
	TextRU string `json:"text_ru"`
	TextEN string `json:"text_en"`
}

func (h *TooltipHandler) UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	var req updateTooltipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.tooltips.Update(r.Context(), key, req.TextRU, req.TextEN)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "tooltip not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update tooltip")
		return
	}

	writeJSON(w, http.StatusOK, item)
}

type bulkUpdateTooltipsRequest struct {
	Items []struct {
		Key    string `json:"key"`
		TextRU string `json:"text_ru"`
		TextEN string `json:"text_en"`
	} `json:"items"`
}

func (h *TooltipHandler) BulkUpdateAdmin(w http.ResponseWriter, r *http.Request) {
	var req bulkUpdateTooltipsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updated := make([]model.UITooltip, 0, len(req.Items))
	for _, item := range req.Items {
		t, err := h.tooltips.Update(r.Context(), item.Key, item.TextRU, item.TextEN)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				writeError(w, http.StatusNotFound, "tooltip not found: "+item.Key)
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to update tooltips")
			return
		}
		updated = append(updated, *t)
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": updated})
}
