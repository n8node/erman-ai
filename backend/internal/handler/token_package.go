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

type TokenPackageHandler struct {
	svc *service.TokenPackageService
}

func NewTokenPackageHandler(svc *service.TokenPackageService) *TokenPackageHandler {
	return &TokenPackageHandler{svc: svc}
}

func (h *TokenPackageHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListPublic(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load token packages")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *TokenPackageHandler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListAdmin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load token packages")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *TokenPackageHandler) CreateAdmin(w http.ResponseWriter, r *http.Request) {
	var req model.TokenPackageUpsertInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := h.svc.CreateAdmin(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidTokenPackage) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create token package")
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *TokenPackageHandler) UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req model.TokenPackageUpsertInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := h.svc.UpdateAdmin(r.Context(), id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidTokenPackage):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, repository.ErrNotFound):
			writeError(w, http.StatusNotFound, "token package not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to update token package")
		}
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *TokenPackageHandler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		TokenPackageID string `json:"token_package_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.TokenPackageID == "" {
		writeError(w, http.StatusBadRequest, "token_package_id required")
		return
	}
	result, err := h.svc.CreateCheckout(r.Context(), userID, req.TokenPackageID)
	if err != nil {
		writeTokenCheckoutError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func writeTokenCheckoutError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrCheckoutUnavailable):
		writeError(w, http.StatusServiceUnavailable, "checkout unavailable")
	case errors.Is(err, service.ErrUserBlocked):
		writeError(w, http.StatusForbidden, "account blocked")
	case errors.Is(err, service.ErrInvalidTokenPackage), errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid token package")
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "token package not found")
	default:
		writeError(w, http.StatusInternalServerError, "checkout failed")
	}
}
