package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/service"
)

type AuthHandler struct {
	auth    *service.AuthService
	mw      *middleware.Auth
	cfg     *config.Config
}

func NewAuthHandler(auth *service.AuthService, mw *middleware.Auth, cfg *config.Config) *AuthHandler {
	return &AuthHandler{auth: auth, mw: mw, cfg: cfg}
}

type credentialsRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	Referral   string `json:"referral"`
}

type onboardingRequest struct {
	Segment string `json:"segment"`
}

type updateMeRequest struct {
	Email  string `json:"email"`
	Locale string `json:"locale"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.auth.Register(r.Context(), req.Email, req.Password, req.Referral)
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"status":              "verification_required",
		"email":               result.User.Email,
		"email_verified":      false,
		"onboarding_completed": result.User.OnboardingCompleted,
	})
}

type verifyEmailRequest struct {
	Token string `json:"token"`
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.auth.VerifyEmail(r.Context(), req.Token)
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	h.mw.SetTokenCookie(w, result.Token, h.secureCookies())
	writeJSON(w, http.StatusOK, userResponse(result.User))
}

type resendVerificationRequest struct {
	Email string `json:"email"`
}

func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req resendVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.auth.ResendVerification(r.Context(), req.Email); err != nil {
		h.writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	h.mw.SetTokenCookie(w, result.Token, h.secureCookies())
	writeJSON(w, http.StatusOK, userResponse(result.User))
}

func (h *AuthHandler) Logout(w http.ResponseWriter, _ *http.Request) {
	h.mw.ClearTokenCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.auth.GetMe(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	writeJSON(w, http.StatusOK, userResponse(user))
}

func (h *AuthHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req updateMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.auth.UpdateMe(r.Context(), userID, req.Email, req.Locale)
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, userResponse(user))
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.auth.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		h.writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) Onboarding(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req onboardingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.auth.CompleteOnboarding(r.Context(), userID, req.Segment)
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, userResponse(user))
}

func (h *AuthHandler) writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrEmailTaken):
		writeError(w, http.StatusConflict, "email already registered")
	case errors.Is(err, service.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, service.ErrUserBlocked):
		writeError(w, http.StatusForbidden, "account blocked")
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid input")
	case errors.Is(err, service.ErrEmailNotVerified):
		writeErrorCode(w, http.StatusForbidden, "confirm your email before signing in", "email_not_verified")
	case errors.Is(err, service.ErrVerificationTokenInvalid):
		writeError(w, http.StatusBadRequest, "invalid or expired verification link")
	case errors.Is(err, service.ErrVerificationTooSoon):
		writeError(w, http.StatusTooManyRequests, "please wait before requesting another email")
	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func (h *AuthHandler) secureCookies() bool {
	return h.cfg.Environment == "production"
}

func userResponse(u *model.User) map[string]any {
	return map[string]any{
		"id":                    u.ID,
		"email":                 u.Email,
		"role":                  u.Role,
		"plan_id":               u.PlanID,
		"locale":                u.Locale,
		"account_segment":       u.AccountSegment,
		"onboarding_completed":  u.OnboardingCompleted,
		"email_verified":        u.EmailVerified(),
		"is_blocked":            u.IsBlocked,
		"created_at":            u.CreatedAt,
	}
}
