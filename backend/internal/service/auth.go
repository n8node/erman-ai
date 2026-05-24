package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken         = errors.New("email already registered")
	ErrUserBlocked        = errors.New("account blocked")
	ErrInvalidInput       = errors.New("invalid input")
)

const bcryptCost = 12
const tokenTTL = 7 * 24 * time.Hour

type AuthService struct {
	users *repository.UserRepository
	auth  *middleware.Auth
}

func NewAuthService(users *repository.UserRepository, auth *middleware.Auth) *AuthService {
	return &AuthService{users: users, auth: auth}
}

type AuthResult struct {
	Token string
	User  *model.User
}

func (s *AuthService) Register(ctx context.Context, email, password, referral string) (*AuthResult, error) {
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

	user, err := s.users.Create(ctx, email, string(hash), planID, "user", segment, onboardingDone)
	if err != nil {
		return nil, err
	}

	token, err := s.auth.IssueToken(user.ID, user.Role, tokenTTL)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, User: user}, nil
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
	if locale != "ru" && locale != "en" {
		locale = "ru"
	}
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

	return s.users.Create(ctx, email, string(hash), planID, "superadmin", model.AccountSegmentPartner, true)
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
	if utf8.RuneCountInString(password) < 8 {
		return ErrInvalidInput
	}
	return nil
}
