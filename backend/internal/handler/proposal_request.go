package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/repository"
	"github.com/erman-ai/erman-ai/internal/service"
	"github.com/go-chi/chi/v5"
)

type ProposalRequestHandler struct {
	requests *service.ProposalRequestService
}

func NewProposalRequestHandler(requests *service.ProposalRequestService) *ProposalRequestHandler {
	return &ProposalRequestHandler{requests: requests}
}

type proposalRequestCreateBody struct {
	RunID          string `json:"run_id"`
	RequesterName  string `json:"requester_name"`
	Telegram       string `json:"telegram"`
	BusinessNote   string `json:"business_note"`
}

func (h *ProposalRequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req proposalRequestCreateBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RunID == "" {
		writeError(w, http.StatusBadRequest, "run_id required")
		return
	}

	created, err := h.requests.Create(
		r.Context(),
		userID,
		req.RunID,
		req.RequesterName,
		req.Telegram,
		req.BusinessNote,
	)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *ProposalRequestHandler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	result, err := h.requests.ListAdmin(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load proposal requests")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *ProposalRequestHandler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	detail, err := h.requests.GetAdminDetail(r.Context(), id)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

type proposalRequestStatusBody struct {
	Status string `json:"status"`
}

func (h *ProposalRequestHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req proposalRequestStatusBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Status == "" {
		writeError(w, http.StatusBadRequest, "status required")
		return
	}

	updated, err := h.requests.UpdateStatus(r.Context(), id, req.Status)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *ProposalRequestHandler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrProposalRequestPartnerOnly):
		writeError(w, http.StatusForbidden, "proposal requests for partners only")
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid input")
	default:
		writeError(w, http.StatusInternalServerError, "request failed")
	}
}
