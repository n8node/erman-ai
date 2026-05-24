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

type AdminUserHandler struct {
	users *service.AdminUserService
	mw    *middleware.Auth
	cfg   *config.Config
}

func NewAdminUserHandler(users *service.AdminUserService, mw *middleware.Auth, cfg *config.Config) *AdminUserHandler {
	return &AdminUserHandler{users: users, mw: mw, cfg: cfg}
}

type updateUserPlanRequest struct {
	PlanID string `json:"plan_id"`
}

func (h *AdminUserHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	result, err := h.users.List(r.Context(), q, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load users")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *AdminUserHandler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	var req updateUserPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.PlanID == "" {
		writeError(w, http.StatusBadRequest, "plan_id required")
		return
	}

	user, err := h.users.UpdatePlan(r.Context(), actorID, id, req.PlanID)
	if err != nil {
		h.writeUserError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *AdminUserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.users.Delete(r.Context(), actorID, id); err != nil {
		h.writeUserError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *AdminUserHandler) Impersonate(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")

	result, err := h.users.Impersonate(r.Context(), actorID, id)
	if err != nil {
		h.writeUserError(w, err)
		return
	}

	h.mw.SetTokenCookie(w, result.Token, h.cfg.Environment == "production")
	writeJSON(w, http.StatusOK, userResponse(result.User))
}

func (h *AdminUserHandler) writeUserError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "user not found")
	case errors.Is(err, service.ErrCannotDeleteSelf):
		writeError(w, http.StatusBadRequest, "cannot delete your own account")
	case errors.Is(err, service.ErrCannotModifySuperadmin):
		writeError(w, http.StatusForbidden, "cannot modify superadmin account")
	case errors.Is(err, service.ErrCannotImpersonateAdmin):
		writeError(w, http.StatusForbidden, "cannot impersonate superadmin")
	case errors.Is(err, service.ErrUserBlocked):
		writeError(w, http.StatusForbidden, "account blocked")
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid input")
	default:
		writeError(w, http.StatusInternalServerError, "user operation failed")
	}
}
