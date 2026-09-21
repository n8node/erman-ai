package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/service"
)

type GeologicalJournalDocumentHandler struct {
	svc *service.GeologicalJournalDocumentService
}

func NewGeologicalJournalDocumentHandler(svc *service.GeologicalJournalDocumentService) *GeologicalJournalDocumentHandler {
	return &GeologicalJournalDocumentHandler{svc: svc}
}

func (h *GeologicalJournalDocumentHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := journalIdentity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, service.GeologicalJournalMaxPDFBytes+1)
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "pdf file required")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "pdf file required")
		return
	}
	defer file.Close()
	if header.Size <= 0 {
		writeError(w, http.StatusBadRequest, "empty pdf")
		return
	}
	doc, err := h.svc.Create(r.Context(), userID, role, header.Filename, file, header.Size)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, doc)
}

func (h *GeologicalJournalDocumentHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := journalIdentity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	items, err := h.svc.List(r.Context(), userID, role)
	if err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *GeologicalJournalDocumentHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := journalIdentity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	item, err := h.svc.Get(r.Context(), r.PathValue("id"), userID, role)
	if err != nil {
		writeError(w, http.StatusNotFound, "document not found")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *GeologicalJournalDocumentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := journalIdentity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.svc.Delete(r.Context(), r.PathValue("id"), userID, role); err != nil {
		writeError(w, http.StatusNotFound, "document not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *GeologicalJournalDocumentHandler) StartAnalysis(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := journalIdentity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	go func() {
		if err := h.svc.StartAnalysis(context.Background(), r.PathValue("id"), userID, role); err != nil {
			// The service records the error status; the client observes it on the next poll.
		}
	}()
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "queued"})
}

func (h *GeologicalJournalDocumentHandler) SetShared(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := journalIdentity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Shared bool `json:"shared"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := h.svc.SetShared(r.Context(), r.PathValue("id"), userID, role, req.Shared); err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"shared": req.Shared})
}

func (h *GeologicalJournalDocumentHandler) LLMProcess(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := journalIdentity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req model.GeologicalJournalDocumentLLMRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 64<<10)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	results, err := h.svc.SummarizeSelectedPages(r.Context(), r.PathValue("id"), userID, role, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": results})
}

func (h *GeologicalJournalDocumentHandler) SavePageTableResult(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := journalIdentity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, 2<<20))
	if err != nil || len(raw) == 0 {
		writeError(w, http.StatusBadRequest, "invalid result")
		return
	}
	result, err := h.svc.SavePageTableResult(r.Context(), r.PathValue("document_id"), r.PathValue("page_id"), userID, role, raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *GeologicalJournalDocumentHandler) AnalyzeSelectedPageLocally(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := journalIdentity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.svc.AnalyzeSelectedPageLocally(r.Context(), r.PathValue("document_id"), r.PathValue("page_id"), userID, role); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "done"})
}

func (h *GeologicalJournalDocumentHandler) Chat(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := journalIdentity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req model.GeologicalJournalDocumentChatRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 64<<10)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	message, err := h.svc.Chat(r.Context(), r.PathValue("id"), userID, role, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, message)
}

func (h *GeologicalJournalDocumentHandler) PageAsset(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := journalIdentity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.svc.CheckAccess(r.Context(), r.PathValue("document_id"), userID, role); err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}
	path, err := h.svc.PageAsset(r.Context(), r.PathValue("document_id"), r.PathValue("page_id"), r.PathValue("kind"))
	if err != nil {
		writeError(w, http.StatusNotFound, "asset not found")
		return
	}
	http.ServeFile(w, r, path)
}

func parsePageNumber(raw string) int { value, _ := strconv.Atoi(raw); return value }
