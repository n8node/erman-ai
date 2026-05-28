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
)

var (
	ErrVerificationTokenInvalid = errors.New("invalid or expired verification token")
	ErrVerificationTooSoon    = errors.New("verification email was sent recently")
)

const verificationTokenTTL = 24 * time.Hour
const resendCooldown = 2 * time.Minute

type EmailVerificationService struct {
	users    *repository.UserRepository
	tokens   *repository.EmailVerificationTokenRepository
	mail     *MailService
	telegram *TelegramService
	cfg      *config.Config
	keySalt  string
	lastSent map[string]time.Time
}

func NewEmailVerificationService(
	users *repository.UserRepository,
	tokens *repository.EmailVerificationTokenRepository,
	mail *MailService,
	telegram *TelegramService,
	cfg *config.Config,
) *EmailVerificationService {
	return &EmailVerificationService{
		users:    users,
		tokens:   tokens,
		mail:     mail,
		telegram: telegram,
		cfg:      cfg,
		keySalt:  cfg.APIKeySalt,
		lastSent: make(map[string]time.Time),
	}
}

func (s *EmailVerificationService) SendVerification(ctx context.Context, user *model.User) (string, error) {
	raw, hash, err := s.generateToken()
	if err != nil {
		return "", err
	}

	if err := s.tokens.DeleteForUser(ctx, user.ID); err != nil {
		return "", err
	}
	if err := s.tokens.Create(ctx, user.ID, hash, time.Now().Add(verificationTokenTTL)); err != nil {
		return "", err
	}

	link := s.verificationURL(raw)
	subject, body := verificationEmailContent(user.Locale, link)

	if err := s.mail.Send(ctx, user.Email, subject, body); err != nil {
		if s.cfg.Environment == "development" {
			slog.Info("email verification link (smtp unavailable)", "email", user.Email, "url", link, "err", err)
			return link, nil
		}
		return "", err
	}
	return link, nil
}

func (s *EmailVerificationService) Resend(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	user, _, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if user.EmailVerified() {
		return nil
	}
	if t, ok := s.lastSent[user.ID]; ok && time.Since(t) < resendCooldown {
		return ErrVerificationTooSoon
	}
	_, err = s.SendVerification(ctx, user)
	if err != nil {
		return err
	}
	s.lastSent[user.ID] = time.Now()
	return nil
}

func (s *EmailVerificationService) Verify(ctx context.Context, rawToken string) (*model.User, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return nil, ErrVerificationTokenInvalid
	}
	hash := s.hashToken(rawToken)
	userID, err := s.tokens.FindValidUserID(ctx, hash, time.Now())
	if err != nil {
		return nil, ErrVerificationTokenInvalid
	}
	user, err := s.users.MarkEmailVerified(ctx, userID)
	if err != nil {
		return nil, err
	}
	_ = s.tokens.DeleteByHash(ctx, hash)
	if s.telegram != nil {
		s.telegram.NotifyEmailVerified(ctx, user)
	}
	return user, nil
}

func (s *EmailVerificationService) verificationURL(rawToken string) string {
	return fmt.Sprintf("%s/dashboard/verify-email?token=%s", strings.TrimRight(s.cfg.PublicBaseURL(), "/"), rawToken)
}

func (s *EmailVerificationService) generateToken() (raw, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	raw = hex.EncodeToString(b)
	return raw, s.hashToken(raw), nil
}

func (s *EmailVerificationService) hashToken(raw string) string {
	sum := sha256.Sum256([]byte(s.keySalt + ":" + raw))
	return hex.EncodeToString(sum[:])
}

func verificationEmailContent(locale, link string) (subject, html string) {
	if locale == "en" {
		return "Confirm your Erman AI email",
			fmt.Sprintf(`<p>Please confirm your email address to access Erman AI.</p>
<p><a href="%s">Confirm email</a></p>
<p>This link expires in 24 hours. If you did not register, you can ignore this message.</p>`, link)
	}
	return "Подтвердите email в Erman AI",
		fmt.Sprintf(`<p>Подтвердите адрес email, чтобы получить доступ к Erman AI.</p>
<p><a href="%s">Подтвердить email</a></p>
<p>Ссылка действует 24 часа. Если вы не регистрировались — просто проигнорируйте письмо.</p>`, link)
}
