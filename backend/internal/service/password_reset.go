package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrResetTokenInvalid = errors.New("invalid or expired reset token")
	ErrResetTooSoon      = errors.New("reset email was sent recently")
)

const resetTokenTTL = time.Hour
const resetResendCooldown = 2 * time.Minute

type PasswordResetService struct {
	users    *repository.UserRepository
	tokens   *repository.PasswordResetTokenRepository
	mail     *MailService
	cfg      *config.Config
	keySalt  string
	lastSent map[string]time.Time
}

func NewPasswordResetService(
	users *repository.UserRepository,
	tokens *repository.PasswordResetTokenRepository,
	mail *MailService,
	cfg *config.Config,
) *PasswordResetService {
	return &PasswordResetService{
		users:    users,
		tokens:   tokens,
		mail:     mail,
		cfg:      cfg,
		keySalt:  cfg.APIKeySalt,
		lastSent: make(map[string]time.Time),
	}
}

func (s *PasswordResetService) RequestReset(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	user, _, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if user.IsBlocked {
		return nil
	}
	if t, ok := s.lastSent[user.ID]; ok && time.Since(t) < resetResendCooldown {
		return ErrResetTooSoon
	}

	if _, err := s.sendReset(ctx, user); err != nil {
		return err
	}
	s.lastSent[user.ID] = time.Now()
	return nil
}

func (s *PasswordResetService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return ErrResetTokenInvalid
	}
	hash := s.hashToken(rawToken)
	userID, err := s.tokens.FindValidUserID(ctx, hash, time.Now())
	if err != nil {
		return ErrResetTokenInvalid
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return err
	}
	if err := s.users.UpdatePassword(ctx, userID, string(newHash)); err != nil {
		return err
	}
	return s.tokens.DeleteByHash(ctx, hash)
}

func (s *PasswordResetService) sendReset(ctx context.Context, user *model.User) (string, error) {
	raw, hash, err := s.generateToken()
	if err != nil {
		return "", err
	}

	if err := s.tokens.DeleteForUser(ctx, user.ID); err != nil {
		return "", err
	}
	if err := s.tokens.Create(ctx, user.ID, hash, time.Now().Add(resetTokenTTL)); err != nil {
		return "", err
	}

	link := s.resetURL(raw)
	subject, body := resetEmailContent(user.Locale, link)

	if err := s.mail.Send(ctx, user.Email, subject, body); err != nil {
		if s.cfg.Environment == "development" {
			slog.Info("password reset link (smtp unavailable)", "email", user.Email, "url", link, "err", err)
			return link, nil
		}
		return "", err
	}
	return link, nil
}

func (s *PasswordResetService) resetURL(rawToken string) string {
	return fmt.Sprintf("%s/dashboard/reset-password?token=%s", strings.TrimRight(s.cfg.PublicBaseURL(), "/"), rawToken)
}

func (s *PasswordResetService) generateToken() (raw, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	raw = hex.EncodeToString(b)
	return raw, s.hashToken(raw), nil
}

func (s *PasswordResetService) hashToken(raw string) string {
	sum := sha256.Sum256([]byte(s.keySalt + ":" + raw))
	return hex.EncodeToString(sum[:])
}

func resetEmailContent(locale, link string) (subject, html string) {
	if locale == "en" {
		return "Reset your Erman AI password",
			fmt.Sprintf(`<p>We received a request to reset your password.</p>
<p><a href="%s">Set a new password</a></p>
<p>This link expires in 1 hour. If you did not request a reset, you can ignore this message.</p>`, link)
	}
	return "Сброс пароля Erman AI",
		fmt.Sprintf(`<p>Мы получили запрос на сброс пароля.</p>
<p><a href="%s">Задать новый пароль</a></p>
<p>Ссылка действует 1 час. Если вы не запрашивали сброс — просто проигнорируйте письмо.</p>`, link)
}
