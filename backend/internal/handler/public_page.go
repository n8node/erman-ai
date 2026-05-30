package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
	"github.com/erman-ai/erman-ai/internal/service"
	"github.com/go-chi/chi/v5"
)

type PublicPageHandler struct {
	pages *service.PublicPageService
}

func NewPublicPageHandler(pages *service.PublicPageService) *PublicPageHandler {
	return &PublicPageHandler{pages: pages}
}

func (h *PublicPageHandler) ListSlugs(w http.ResponseWriter, r *http.Request) {
	slugs, err := h.pages.ListPublishedSlugs(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load public page slugs")
		return
	}
	if slugs == nil {
		slugs = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"slugs": slugs})
}

func (h *PublicPageHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	page, err := h.pages.GetPublishedBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "page not found")
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "invalid slug")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load page")
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *PublicPageHandler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	items, err := h.pages.ListAdmin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load public pages")
		return
	}
	if items == nil {
		items = []model.PublicPage{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *PublicPageHandler) CreateAdmin(w http.ResponseWriter, r *http.Request) {
	var req publicPageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.pages.Create(r.Context(), req.toInput())
	if err != nil {
		h.writePageError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *PublicPageHandler) UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req publicPageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.pages.Update(r.Context(), id, req.toInput())
	if err != nil {
		h.writePageError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *PublicPageHandler) DeleteAdmin(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.pages.Delete(r.Context(), id); err != nil {
		h.writePageError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

type publicPageRequest struct {
	Slug            string `json:"slug"`
	Title           string `json:"title"`
	ContentHTML     string `json:"content_html"`
	MetaDescription string `json:"meta_description"`
	IsPublished     bool   `json:"is_published"`
	SortOrder       int    `json:"sort_order"`
}

func (req publicPageRequest) toInput() service.PublicPageInput {
	return service.PublicPageInput{
		Slug:            req.Slug,
		Title:           req.Title,
		ContentHTML:     req.ContentHTML,
		MetaDescription: req.MetaDescription,
		IsPublished:     req.IsPublished,
		SortOrder:       req.SortOrder,
	}
}

func (h *PublicPageHandler) writePageError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid input")
	case errors.Is(err, service.ErrPublicPageSlugInvalid):
		writeError(w, http.StatusBadRequest, "invalid slug")
	case errors.Is(err, service.ErrPublicPageSlugTaken):
		writeError(w, http.StatusConflict, "slug already taken")
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "page not found")
	default:
		writeError(w, http.StatusInternalServerError, "failed to process public page")
	}
}
