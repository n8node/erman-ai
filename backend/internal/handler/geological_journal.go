package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
	"github.com/erman-ai/erman-ai/internal/service"
)

type GeologicalJournalHandler struct {
	svc *service.GeologicalJournalService
}

func NewGeologicalJournalHandler(svc *service.GeologicalJournalService) *GeologicalJournalHandler {
	return &GeologicalJournalHandler{svc: svc}
}

func journalIdentity(r *http.Request) (string, string, bool) {
	id, ok := middleware.UserIDFromContext(r.Context())
	role, _ := middleware.UserRoleFromContext(r.Context())
	return id, role, ok
}

func (h *GeologicalJournalHandler) CreatePage(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := journalIdentity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	image, name, err := readJournalMultipartImage(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "valid JPEG, PNG, WebP, or GIF image up to 10 MiB required")
		return
	}
	page, run, err := h.svc.CreatePage(r.Context(), userID, role, name, image)
	if err != nil {
		h.writeServiceError(w, err, "failed to create page")
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"page": page, "run_id": run.ID})
}

func (h *GeologicalJournalHandler) ListPages(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := journalIdentity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	items, err := h.svc.ListPages(r.Context(), userID, role)
	if err != nil {
		h.writeServiceError(w, err, "failed to list pages")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *GeologicalJournalHandler) GetPage(w http.ResponseWriter, r *http.Request) {
	userID, role, _ := journalIdentity(r)
	item, err := h.svc.GetPage(r.Context(), r.PathValue("id"), userID, role)
	if err != nil {
		h.writeServiceError(w, err, "failed to load page")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *GeologicalJournalHandler) PageImage(w http.ResponseWriter, r *http.Request) {
	userID, role, _ := journalIdentity(r)
	path, contentType, err := h.svc.PageImage(r.Context(), r.PathValue("id"), userID, role)
	if err != nil {
		h.writeServiceError(w, err, "failed to load image")
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, path)
}

func (h *GeologicalJournalHandler) Analyze(w http.ResponseWriter, r *http.Request) {
	userID, role, _ := journalIdentity(r)
	run, err := h.svc.StartAnalysis(r.Context(), r.PathValue("id"), userID, role)
	if err != nil {
		h.writeServiceError(w, err, "failed to start analysis")
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"run_id": run.ID, "status": run.Status})
}

func (h *GeologicalJournalHandler) SaveResult(w http.ResponseWriter, r *http.Request) {
	userID, role, _ := journalIdentity(r)
	raw, err := io.ReadAll(io.LimitReader(r.Body, 2<<20))
	if err != nil || len(raw) == 0 {
		writeError(w, http.StatusBadRequest, "invalid result")
		return
	}
	version, err := h.svc.SaveCorrectedResult(r.Context(), r.PathValue("id"), userID, role, raw)
	if err != nil {
		h.writeServiceError(w, err, "failed to save result")
		return
	}
	writeJSON(w, http.StatusOK, version)
}

func (h *GeologicalJournalHandler) DeletePage(w http.ResponseWriter, r *http.Request) {
	userID, role, _ := journalIdentity(r)
	if err := h.svc.DeletePage(r.Context(), r.PathValue("id"), userID, role); err != nil {
		h.writeServiceError(w, err, "failed to delete page")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *GeologicalJournalHandler) ListExamples(w http.ResponseWriter, r *http.Request) {
	userID, role, _ := journalIdentity(r)
	items, err := h.svc.ListExamples(r.Context(), userID, role, false)
	if err != nil {
		h.writeServiceError(w, err, "failed to list examples")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *GeologicalJournalHandler) ExampleImage(w http.ResponseWriter, r *http.Request) {
	userID, role, _ := journalIdentity(r)
	path, contentType, err := h.svc.ExampleImage(r.Context(), r.PathValue("id"), userID, role)
	if err != nil {
		h.writeServiceError(w, err, "failed to load example")
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, path)
}

func (h *GeologicalJournalHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	item, err := h.svc.GetSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *GeologicalJournalHandler) PutSettings(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Settings model.GeologicalJournalSettings `json:"settings"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := h.svc.UpdateSettings(r.Context(), req.Settings)
	if err != nil {
		h.writeServiceError(w, err, "failed to update settings")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *GeologicalJournalHandler) RefreshModels(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Provider model.LLMProvider `json:"provider"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := h.svc.RefreshModels(r.Context(), req.Provider)
	if err != nil {
		h.writeServiceError(w, err, "failed to refresh models")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *GeologicalJournalHandler) ListAccess(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListAccessUsers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list access")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *GeologicalJournalHandler) PutAccess(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.SetAccess(r.Context(), r.PathValue("user_id"), req.Enabled); err != nil {
		h.writeServiceError(w, err, "failed to update access")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user_id": r.PathValue("user_id"), "enabled": req.Enabled})
}

func (h *GeologicalJournalHandler) AdminListExamples(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListExamples(r.Context(), "", "superadmin", true)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list examples")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *GeologicalJournalHandler) AdminCreateExample(w http.ResponseWriter, r *http.Request) {
	image, _, err := readJournalMultipartImage(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "valid JPEG, PNG, WebP, or GIF image up to 10 MiB required")
		return
	}
	sortOrder, _ := strconv.Atoi(r.FormValue("sort_order"))
	published := true
	if raw := strings.TrimSpace(r.FormValue("is_published")); raw != "" {
		published, _ = strconv.ParseBool(raw)
	}
	item, err := h.svc.CreateExample(r.Context(), model.GeologicalJournalExampleMetadata{
		Title: r.FormValue("title"), Description: r.FormValue("description"),
		SortOrder: sortOrder, IsPublished: published,
	}, image)
	if err != nil {
		h.writeServiceError(w, err, "failed to create example")
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *GeologicalJournalHandler) AdminUpdateExample(w http.ResponseWriter, r *http.Request) {
	var req model.GeologicalJournalExampleMetadata
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := h.svc.UpdateExample(r.Context(), r.PathValue("id"), req)
	if err != nil {
		h.writeServiceError(w, err, "failed to update example")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *GeologicalJournalHandler) AdminDeleteExample(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteExample(r.Context(), r.PathValue("id")); err != nil {
		h.writeServiceError(w, err, "failed to delete example")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *GeologicalJournalHandler) writeServiceError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, service.ErrGeologicalJournalForbidden):
		writeError(w, http.StatusForbidden, "geological journal access denied")
	case errors.Is(err, service.ErrGeologicalJournalInvalidImage), errors.Is(err, service.ErrInvalidInput), errors.Is(err, service.ErrGeologicalJournalSettings):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, repository.ErrNotFound), errors.Is(err, os.ErrNotExist):
		writeError(w, http.StatusNotFound, "not found")
	default:
		writeError(w, http.StatusInternalServerError, fallback)
	}
}

func readJournalMultipartImage(w http.ResponseWriter, r *http.Request) (*service.ValidatedJournalImage, string, error) {
	r.Body = http.MaxBytesReader(w, r.Body, service.GeologicalJournalMaxImageBytes+(1<<20))
	if err := r.ParseMultipartForm(service.GeologicalJournalMaxImageBytes + (1 << 20)); err != nil {
		return nil, "", err
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		return nil, "", err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, service.GeologicalJournalMaxImageBytes+1))
	if err != nil || int64(len(data)) > service.GeologicalJournalMaxImageBytes {
		return nil, "", service.ErrGeologicalJournalInvalidImage
	}
	image, err := service.ValidateGeologicalJournalImage(data)
	if err != nil {
		return nil, "", err
	}
	return image, header.Filename, nil
}
