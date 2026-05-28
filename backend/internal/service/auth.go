package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/erman-ai/erman-ai/internal/i18n"
	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/model"
	pwdpolicy "github.com/erman-ai/erman-ai/internal/password"
	"github.com/erman-ai/erman-ai/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken         = errors.New("email already registered")
	ErrUserBlocked        = errors.New("account blocked")
	ErrInvalidInput       = errors.New("invalid input")
	ErrEmailNotVerified   = errors.New("email not verified")
)

const bcryptCost = 12
const tokenTTL = 7 * 24 * time.Hour

type AuthService struct {
	users    *repository.UserRepository
	auth     *middleware.Auth
	verify   *EmailVerificationService
	telegram *TelegramService
}

func NewAuthService(users *repository.UserRepository, auth *middleware.Auth, verify *EmailVerificationService, telegram *TelegramService) *AuthService {
	return &AuthService{users: users, auth: auth, verify: verify, telegram: telegram}
}

type AuthResult struct {
	Token string
	User  *model.User
}

type RegisterResult struct {
	RequiresVerification bool
	User                 *model.User
}

func (s *AuthService) Register(ctx context.Context, email, password, referral string) (*RegisterResult, error) {
	email = normalizeEmail(email)
	if err := validateCredentials(email, password); err != nil {
		return nil, err
	}

	exists, err := s.users.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailTaken
	}

	planID, err := s.users.GetPlanIDBySlug(ctx, "free")
	if err != nil {
		return nil, fmt.Errorf("resolve free plan: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, err
	}

	segment := model.AccountSegmentPartner
	onboardingDone := false
	if referral == "erman" {
		segment = model.AccountSegmentDirectLead
		onboardingDone = true
	}

	user, err := s.users.Create(ctx, email, string(hash), planID, "user", segment, onboardingDone, false)
	if err != nil {
		return nil, err
	}

	if s.telegram != nil {
		s.telegram.NotifyRegistration(ctx, user, referral)
	}

	if _, err := s.verify.SendVerification(ctx, user); err != nil {
		return nil, fmt.Errorf("send verification email: %w", err)
	}

	return &RegisterResult{RequiresVerification: true, User: user}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	email = normalizeEmail(email)
	if email == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	user, hash, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if user.IsBlocked {
		return nil, ErrUserBlocked
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	if !user.EmailVerified() && user.Role != "superadmin" {
		return nil, ErrEmailNotVerified
	}

	s.users.TouchLastActive(ctx, user.ID)

	token, err := s.auth.IssueToken(user.ID, user.Role, tokenTTL)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, User: user}, nil
}

func (s *AuthService) GetMe(ctx context.Context, userID string) (*model.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrInvalidCredentials
	}
	return user, err
}

func (s *AuthService) UpdateMe(ctx context.Context, userID, email, locale string) (*model.User, error) {
	email = normalizeEmail(email)
	if email == "" {
		return nil, ErrInvalidInput
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, ErrInvalidInput
	}
	locale = i18n.NormalizeLocale(locale)
	return s.users.UpdateProfile(ctx, userID, email, locale)
}

func (s *AuthService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	hash, err := s.users.GetPasswordHashByID(ctx, userID)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(currentPassword)); err != nil {
		return ErrInvalidCredentials
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return err
	}
	return s.users.UpdatePassword(ctx, userID, string(newHash))
}

func (s *AuthService) CompleteOnboarding(ctx context.Context, userID, segment string) (*model.User, error) {
	if segment != model.AccountSegmentPartner && segment != model.AccountSegmentDirectLead {
		return nil, ErrInvalidInput
	}
	return s.users.CompleteOnboarding(ctx, userID, segment)
}

func (s *AuthService) SeedAdmin(ctx context.Context, email, password string) (*model.User, error) {
	email = normalizeEmail(email)
	if err := validateCredentials(email, password); err != nil {
		return nil, err
	}

	planID, err := s.users.GetPlanIDBySlug(ctx, "business")
	if err != nil {
		planID, _ = s.users.GetPlanIDBySlug(ctx, "free")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, err
	}

	exists, err := s.users.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		u, _, err := s.users.GetByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		if err := s.users.SetRole(ctx, u.ID, "superadmin"); err != nil {
			return nil, err
		}
		if err := s.users.UpdatePassword(ctx, u.ID, string(hash)); err != nil {
			return nil, err
		}
		return s.users.GetByID(ctx, u.ID)
	}

	return s.users.Create(ctx, email, string(hash), planID, "superadmin", model.AccountSegmentPartner, true, true)
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) (*AuthResult, error) {
	user, err := s.verify.Verify(ctx, token)
	if err != nil {
		return nil, err
	}
	jwt, err := s.auth.IssueToken(user.ID, user.Role, tokenTTL)
	if err != nil {
		return nil, err
	}
	return &AuthResult{Token: jwt, User: user}, nil
}

func (s *AuthService) ResendVerification(ctx context.Context, email string) error {
	return s.verify.Resend(ctx, email)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validateCredentials(email, password string) error {
	if email == "" {
		return ErrInvalidInput
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidInput
	}
	return validatePassword(password)
}

func validatePassword(password string) error {
	if err := pwdpolicy.Validate(password); err != nil {
		return ErrInvalidInput
	}
	return nil
}
