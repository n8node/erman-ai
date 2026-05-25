package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/erman-ai/erman-ai/internal/i18n"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
	"github.com/erman-ai/erman-ai/internal/service"
	"github.com/go-chi/chi/v5"
)

type TranslationHandler struct {
	translations *service.TranslationService
}

func NewTranslationHandler(translations *service.TranslationService) *TranslationHandler {
	return &TranslationHandler{translations: translations}
}

func (h *TranslationHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	locale := i18n.NormalizeLocale(r.URL.Query().Get("locale"))
	items, err := h.translations.ListPublic(r.Context(), locale)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load translations")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"translations": items})
}

func (h *TranslationHandler) SearchAdmin(w http.ResponseWriter, r *http.Request) {
	locale := r.URL.Query().Get("locale")
	search := r.URL.Query().Get("search")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	items, err := h.translations.SearchAdmin(r.Context(), locale, search, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to search translations")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

type translationUpsertRequest struct {
	Key    string `json:"key"`
	Locale string `json:"locale"`
	Value  string `json:"value"`
}

func (h *TranslationHandler) UpsertAdmin(w http.ResponseWriter, r *http.Request) {
	var req translationUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := h.translations.Upsert(r.Context(), req.Key, req.Locale, req.Value)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "invalid input")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to save translation")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *TranslationHandler) BulkUpsertAdmin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Items []model.UITranslation `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	items, err := h.translations.BulkUpsert(r.Context(), req.Items)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save translations")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *TranslationHandler) DeleteAdmin(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	locale := r.URL.Query().Get("locale")
	if key == "" || locale == "" {
		writeError(w, http.StatusBadRequest, "key and locale required")
		return
	}
	if err := h.translations.Delete(r.Context(), key, locale); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete translation")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
