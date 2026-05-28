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

type ExternalProjectHandler struct {
	projects *service.ExternalProjectService
}

func NewExternalProjectHandler(projects *service.ExternalProjectService) *ExternalProjectHandler {
	return &ExternalProjectHandler{projects: projects}
}

type externalProjectRequest struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	SortOrder int    `json:"sort_order"`
	IsEnabled bool   `json:"is_enabled"`
}

func (h *ExternalProjectHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	items, err := h.projects.ListPublic(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load projects")
		return
	}
	if items == nil {
		items = []model.ExternalProject{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ExternalProjectHandler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	items, err := h.projects.ListAdmin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load projects")
		return
	}
	if items == nil {
		items = []model.ExternalProject{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ExternalProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req externalProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.projects.Create(r.Context(), externalProjectRequestToInput(req))
	if err != nil {
		h.writeProjectError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *ExternalProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req externalProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.projects.Update(r.Context(), id, externalProjectRequestToInput(req))
	if err != nil {
		h.writeProjectError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ExternalProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.projects.Delete(r.Context(), id); err != nil {
		h.writeProjectError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func externalProjectRequestToInput(req externalProjectRequest) service.ExternalProjectInput {
	return service.ExternalProjectInput{
		Title:     req.Title,
		URL:       req.URL,
		SortOrder: req.SortOrder,
		IsEnabled: req.IsEnabled,
	}
}

func (h *ExternalProjectHandler) writeProjectError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid input")
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "project not found")
	default:
		writeError(w, http.StatusInternalServerError, "failed to process project")
	}
}
