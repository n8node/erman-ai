package handler

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
	"github.com/erman-ai/erman-ai/internal/service"
)

type VideoTranscriptionHandler struct {
	svc *service.VideoTranscriptionService
}

func NewVideoTranscriptionHandler(svc *service.VideoTranscriptionService) *VideoTranscriptionHandler {
	return &VideoTranscriptionHandler{svc: svc}
}

func videoTranscriptionIdentity(r *http.Request) (string, string, bool) {
	id, ok := middleware.UserIDFromContext(r.Context())
	role, _ := middleware.UserRoleFromContext(r.Context())
	return id, role, ok
}

func (h *VideoTranscriptionHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := videoTranscriptionIdentity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	part, err := openStreamingMultipartFile(w, r, service.VideoTranscriptionMaxUploadBytes+(1<<20), "file", "video")
	if err != nil {
		writeError(w, http.StatusBadRequest, "valid video file up to 500 MiB required (MP4, WebM, MOV, AVI, MKV)")
		return
	}
	defer part.Reader.Close()
	file, run, err := h.svc.UploadAndTranscribe(r.Context(), userID, role, part.Filename, part.ContentType, part.Reader)
	if err != nil {
		h.writeServiceError(w, err, "failed to upload video")
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"file": file, "run_id": run.ID})
}

func (h *VideoTranscriptionHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := videoTranscriptionIdentity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	items, err := h.svc.ListFiles(r.Context(), userID, role)
	if err != nil {
		h.writeServiceError(w, err, "failed to list files")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *VideoTranscriptionHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	userID, role, _ := videoTranscriptionIdentity(r)
	w.Header().Set("Cache-Control", "no-store")
	item, err := h.svc.GetFile(r.Context(), r.PathValue("id"), userID, role)
	if err != nil {
		h.writeServiceError(w, err, "failed to load file")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *VideoTranscriptionHandler) DownloadTranscript(w http.ResponseWriter, r *http.Request) {
	userID, role, _ := videoTranscriptionIdentity(r)
	path, originalName, err := h.svc.TranscriptDownload(r.Context(), r.PathValue("id"), userID, role)
	if err != nil {
		h.writeServiceError(w, err, "transcript not available")
		return
	}
	base := strings.TrimSuffix(filepath.Base(originalName), filepath.Ext(originalName))
	if base == "" {
		base = "transcript"
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": base + ".txt"}))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-store")
	http.ServeFile(w, r, path)
}

func (h *VideoTranscriptionHandler) TranscribeFile(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := videoTranscriptionIdentity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	run, err := h.svc.RetranscribeFile(r.Context(), r.PathValue("id"), userID, role)
	if err != nil {
		h.writeServiceError(w, err, "failed to start transcription")
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"run_id": run.ID, "status": run.Status})
}

func (h *VideoTranscriptionHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := videoTranscriptionIdentity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.svc.DeleteFile(r.Context(), r.PathValue("id"), userID, role); err != nil {
		h.writeServiceError(w, err, "failed to delete file")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *VideoTranscriptionHandler) ListAccess(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListAccessUsers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list access")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *VideoTranscriptionHandler) PutAccess(w http.ResponseWriter, r *http.Request) {
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

func (h *VideoTranscriptionHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	rec, err := h.svc.GetSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

func (h *VideoTranscriptionHandler) PutSettings(w http.ResponseWriter, r *http.Request) {
	var settings model.VideoTranscriptionSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	rec, err := h.svc.UpdateSettings(r.Context(), settings)
	if err != nil {
		h.writeServiceError(w, err, "failed to update settings")
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

func (h *VideoTranscriptionHandler) writeServiceError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, service.ErrVideoTranscriptionForbidden):
		writeError(w, http.StatusForbidden, "video transcription access denied")
	case errors.Is(err, service.ErrVideoTranscriptionInvalidVideo):
		writeError(w, http.StatusBadRequest, "invalid video file")
	case errors.Is(err, service.ErrVideoTranscriptionNoAudio):
		writeError(w, http.StatusUnprocessableEntity, "video has no audio track")
	case errors.Is(err, service.ErrToolLimitExceeded):
		writeError(w, http.StatusPaymentRequired, "tool limit exceeded")
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, service.ErrVideoTranscriptionAlreadyProcessing):
		writeError(w, http.StatusConflict, "video transcription already in progress")
	case errors.Is(err, service.ErrYandexSpeechKitNotConfigured):
		writeError(w, http.StatusServiceUnavailable, "yandex speechkit not configured")
	default:
		writeError(w, http.StatusInternalServerError, fallback)
	}
}
