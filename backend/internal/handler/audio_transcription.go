package handler

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
	"github.com/erman-ai/erman-ai/internal/service"
)

type AudioTranscriptionHandler struct {
	svc *service.AudioTranscriptionService
}

func NewAudioTranscriptionHandler(svc *service.AudioTranscriptionService) *AudioTranscriptionHandler {
	return &AudioTranscriptionHandler{svc: svc}
}

func audioTranscriptionIdentity(r *http.Request) (string, string, bool) {
	id, ok := middleware.UserIDFromContext(r.Context())
	role, _ := middleware.UserRoleFromContext(r.Context())
	return id, role, ok
}

func (h *AudioTranscriptionHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := audioTranscriptionIdentity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	data, name, contentType, err := readAudioMultipartFile(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "valid audio file up to 500 MiB required (MP3, WAV, OGG, M4A, WebM)")
		return
	}
	file, run, err := h.svc.UploadAndTranscribe(r.Context(), userID, role, name, contentType, data)
	if err != nil {
		h.writeServiceError(w, err, "failed to upload audio")
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"file": file, "run_id": run.ID})
}

func (h *AudioTranscriptionHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := audioTranscriptionIdentity(r)
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

func (h *AudioTranscriptionHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	userID, role, _ := audioTranscriptionIdentity(r)
	w.Header().Set("Cache-Control", "no-store")
	item, err := h.svc.GetFile(r.Context(), r.PathValue("id"), userID, role)
	if err != nil {
		h.writeServiceError(w, err, "failed to load file")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *AudioTranscriptionHandler) DownloadTranscript(w http.ResponseWriter, r *http.Request) {
	userID, role, _ := audioTranscriptionIdentity(r)
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

func (h *AudioTranscriptionHandler) TranscribeFile(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := audioTranscriptionIdentity(r)
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

func (h *AudioTranscriptionHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := audioTranscriptionIdentity(r)
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

func (h *AudioTranscriptionHandler) ListAccess(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListAccessUsers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list access")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *AudioTranscriptionHandler) PutAccess(w http.ResponseWriter, r *http.Request) {
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

func (h *AudioTranscriptionHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	rec, err := h.svc.GetSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

func (h *AudioTranscriptionHandler) PutSettings(w http.ResponseWriter, r *http.Request) {
	var settings model.AudioTranscriptionSettings
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

func readAudioMultipartFile(w http.ResponseWriter, r *http.Request) ([]byte, string, string, error) {
	const maxBytes = service.AudioTranscriptionMaxUploadBytes + (1 << 20)
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		return nil, "", "", err
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		file, header, err = r.FormFile("audio")
	}
	if err != nil {
		return nil, "", "", err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, "", "", err
	}
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if err := service.ValidateAudioTranscriptionFile(data, contentType); err != nil {
		return nil, "", "", err
	}
	name := header.Filename
	if strings.TrimSpace(name) == "" {
		name = "audio"
	}
	return data, name, contentType, nil
}

func (h *AudioTranscriptionHandler) writeServiceError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, service.ErrAudioTranscriptionForbidden):
		writeError(w, http.StatusForbidden, "audio transcription access denied")
	case errors.Is(err, service.ErrAudioTranscriptionInvalidAudio):
		writeError(w, http.StatusBadRequest, "invalid audio file")
	case errors.Is(err, service.ErrToolLimitExceeded):
		writeError(w, http.StatusPaymentRequired, "tool limit exceeded")
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, service.ErrAudioTranscriptionAlreadyProcessing):
		writeError(w, http.StatusConflict, "audio transcription already in progress")
	case errors.Is(err, service.ErrYandexSpeechKitNotConfigured):
		writeError(w, http.StatusServiceUnavailable, "yandex speechkit not configured")
	default:
		writeError(w, http.StatusInternalServerError, fallback)
	}
}
