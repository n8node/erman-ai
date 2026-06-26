package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
	"github.com/erman-ai/erman-ai/internal/service"
)

type AuthHandler struct {
	auth         *service.AuthService
	mw           *middleware.Auth
	cfg          *config.Config
	inquiries    *repository.ProjectInquiryRepository
	proposalReqs *repository.ProposalRequestRepository
}

func NewAuthHandler(
	auth *service.AuthService,
	mw *middleware.Auth,
	cfg *config.Config,
	inquiries *repository.ProjectInquiryRepository,
	proposalReqs *repository.ProposalRequestRepository,
) *AuthHandler {
	return &AuthHandler{
		auth:         auth,
		mw:           mw,
		cfg:          cfg,
		inquiries:    inquiries,
		proposalReqs: proposalReqs,
	}
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

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type emailStatusRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
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
	writeJSON(w, http.StatusOK, h.userResponseWithFlags(r.Context(), result.User))
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
	writeJSON(w, http.StatusOK, h.userResponseWithFlags(r.Context(), result.User))
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

	hasInquiry, hasProposal := h.submissionFlags(r.Context(), user)
	writeJSON(w, http.StatusOK, userResponse(user, hasInquiry, hasProposal))
}

func (h *AuthHandler) EmailStatus(w http.ResponseWriter, r *http.Request) {
	var req emailStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	exists, err := h.auth.EmailExists(r.Context(), req.Email)
	if err != nil {
		h.writeAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"exists": exists})
}

func (h *AuthHandler) submissionFlags(ctx context.Context, user *model.User) (bool, bool) {
	hasInquiry, _ := h.inquiries.ExistsByUserID(ctx, user.ID)
	if !hasInquiry {
		hasInquiry, _ = h.inquiries.ExistsByEmail(ctx, user.Email)
	}
	hasProposal, _ := h.proposalReqs.ExistsByUserID(ctx, user.ID)
	return hasInquiry, hasProposal
}

func (h *AuthHandler) userResponseWithFlags(ctx context.Context, user *model.User) map[string]any {
	hasInquiry, hasProposal := h.submissionFlags(ctx, user)
	return userResponse(user, hasInquiry, hasProposal)
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

	writeJSON(w, http.StatusOK, h.userResponseWithFlags(r.Context(), user))
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.auth.RequestPasswordReset(r.Context(), req.Email); err != nil {
		h.writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.auth.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		h.writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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

	writeJSON(w, http.StatusOK, h.userResponseWithFlags(r.Context(), user))
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
	case errors.Is(err, service.ErrResetTokenInvalid):
		writeError(w, http.StatusBadRequest, "invalid or expired reset link")
	case errors.Is(err, service.ErrResetTooSoon):
		writeError(w, http.StatusTooManyRequests, "please wait before requesting another email")
	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func (h *AuthHandler) secureCookies() bool {
	return h.cfg.Environment == "production"
}

func userResponse(u *model.User, hasProjectInquiry, hasProposalRequest bool) map[string]any {
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
		"has_project_inquiry":   hasProjectInquiry,
		"has_proposal_request":  hasProposalRequest,
	}
}
