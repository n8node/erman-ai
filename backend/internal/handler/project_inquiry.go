package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/repository"
	"github.com/erman-ai/erman-ai/internal/service"
	"github.com/go-chi/chi/v5"
)

type ProjectInquiryHandler struct {
	inquiries *service.ProjectInquiryService
	auth      *service.AuthService
	cfg       *config.Config
}

func NewProjectInquiryHandler(inquiries *service.ProjectInquiryService, auth *service.AuthService, cfg *config.Config) *ProjectInquiryHandler {
	return &ProjectInquiryHandler{inquiries: inquiries, auth: auth, cfg: cfg}
}

type projectInquiryCreateBody struct {
	Name               string          `json:"name"`
	Email              string          `json:"email"`
	Telegram           string          `json:"telegram"`
	ProjectTitle       string          `json:"project_title"`
	ProjectDescription string          `json:"project_description"`
	CalculatorRunID    string          `json:"calculator_run_id"`
	CalculatorSnapshot json.RawMessage `json:"calculator_snapshot,omitempty"`
	Locale             string          `json:"locale"`
	Website            string          `json:"website"`
}

func (h *ProjectInquiryHandler) bodyToInput(req projectInquiryCreateBody) service.ProjectInquiryInput {
	return service.ProjectInquiryInput{
		Name:               req.Name,
		Email:              req.Email,
		Telegram:           req.Telegram,
		ProjectTitle:       req.ProjectTitle,
		ProjectDescription: req.ProjectDescription,
		CalculatorRunID:    req.CalculatorRunID,
		CalculatorSnapshot: req.CalculatorSnapshot,
		Locale:             req.Locale,
		Honeypot:           req.Website,
	}
}

func (h *ProjectInquiryHandler) CreatePublic(w http.ResponseWriter, r *http.Request) {
	var req projectInquiryCreateBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Website != "" {
		writeJSON(w, http.StatusAccepted, map[string]any{
			"status":  "pending_email",
			"message": "verification_required",
		})
		return
	}

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	ipHash := service.HashIP(ip, h.cfg.APIKeySalt)

	created, err := h.inquiries.CreatePublic(r.Context(), h.bodyToInput(req), ipHash)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{
		"id":      created.ID,
		"status":  created.Status,
		"message": "verification_required",
	})
}

func (h *ProjectInquiryHandler) CreateAuthenticated(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req projectInquiryCreateBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Website != "" {
		writeJSON(w, http.StatusCreated, map[string]any{"status": "new"})
		return
	}

	user, err := h.auth.GetMe(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	input := h.bodyToInput(req)
	input.Locale = user.Locale

	created, err := h.inquiries.CreateAuthenticated(r.Context(), userID, input)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *ProjectInquiryHandler) Verify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "token required")
		return
	}

	inq, err := h.inquiries.Verify(r.Context(), token)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid or expired token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":     inq.ID,
		"status": inq.Status,
	})
}

func (h *ProjectInquiryHandler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	result, err := h.inquiries.ListAdmin(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load project inquiries")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *ProjectInquiryHandler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	detail, err := h.inquiries.GetAdminDetail(r.Context(), id)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

type projectInquiryStatusBody struct {
	Status string `json:"status"`
}

func (h *ProjectInquiryHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req projectInquiryStatusBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Status == "" {
		writeError(w, http.StatusBadRequest, "status required")
		return
	}

	updated, err := h.inquiries.UpdateStatus(r.Context(), id, req.Status)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *ProjectInquiryHandler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid input")
	case errors.Is(err, service.ErrInquiryMissingFields):
		writeError(w, http.StatusBadRequest, "missing required fields")
	case errors.Is(err, service.ErrInquiryInvalidTelegram):
		writeError(w, http.StatusBadRequest, "invalid telegram")
	case errors.Is(err, service.ErrProjectInquiryAlreadySent):
		writeError(w, http.StatusConflict, "project inquiry already submitted")
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, service.ErrInquiryVerificationInvalid):
		writeError(w, http.StatusBadRequest, "invalid or expired token")
	default:
		if err != nil && err.Error() == "email not verified" {
			writeError(w, http.StatusForbidden, "email not verified")
			return
		}
		writeError(w, http.StatusInternalServerError, "request failed")
	}
}
