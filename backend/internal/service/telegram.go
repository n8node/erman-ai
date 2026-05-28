package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/erman-ai/erman-ai/internal/model"
)

var (
	ErrTelegramDisabled      = errors.New("telegram notifications disabled")
	ErrTelegramNotConfigured = errors.New("telegram bot not configured")
)

func (s *TelegramService) SendTest(ctx context.Context) (bool, string) {
	cfg, err := s.settings.GetEffective(ctx)
	if err != nil {
		return false, err.Error()
	}
	if !cfg.Enabled {
		return false, "Включите Telegram-уведомления в настройках"
	}
	if strings.TrimSpace(cfg.BotToken) == "" || strings.TrimSpace(cfg.ChatID) == "" {
		return false, "Укажите токен бота и ID чата"
	}
	text := "✅ Тестовое сообщение Erman AI\nБот подключён и может отправлять уведомления в этот чат."
	if err := s.send(ctx, cfg, text); err != nil {
		return false, err.Error()
	}
	s.triggerHealthCheck()
	return true, "Тестовое сообщение отправлено"
}

func (s *TelegramService) NotifyRegistration(ctx context.Context, user *model.User, referral string) {
	cfg, err := s.settings.GetEffective(ctx)
	if err != nil || !cfg.Enabled || !cfg.NotifyRegistration {
		return
	}
	vars := map[string]string{
		"email":          user.Email,
		"name":           displayName(user.Email),
		"referral":       strings.TrimSpace(referral),
		"accountSegment": user.AccountSegment,
		"inviteCode":     strings.TrimSpace(referral),
		"inviteScope":    user.AccountSegment,
		"inviteOwner":    "",
	}
	text := applyTemplate(cfg.RegistrationTemplate, vars)
	s.sendAsync(cfg, text, "registration")
}

func (s *TelegramService) NotifyEmailVerified(ctx context.Context, user *model.User) {
	cfg, err := s.settings.GetEffective(ctx)
	if err != nil || !cfg.Enabled || !cfg.NotifyEmailVerified {
		return
	}
	vars := map[string]string{
		"email": user.Email,
		"name":  displayName(user.Email),
	}
	text := applyTemplate(cfg.EmailVerifiedTemplate, vars)
	s.sendAsync(cfg, text, "email_verified")
}

func (s *TelegramService) NotifyPayment(ctx context.Context, user *model.User, plan *model.Plan, amountRub int) {
	cfg, err := s.settings.GetEffective(ctx)
	if err != nil || !cfg.Enabled || !cfg.NotifyPayment {
		return
	}
	if plan == nil {
		return
	}
	vars := map[string]string{
		"userEmail": user.Email,
		"userName":  displayName(user.Email),
		"planName":  plan.Name,
		"amount":    fmt.Sprintf("%d", amountRub),
		"currency":  "₽",
	}
	text := applyTemplate(cfg.PaymentTemplate, vars)
	s.sendAsync(cfg, text, "payment")
}

func (s *TelegramService) sendAsync(cfg model.TelegramSettings, text, kind string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := s.send(ctx, cfg, text); err != nil {
			slog.Warn("telegram notification failed", "kind", kind, "err", err)
		}
	}()
}

func (s *TelegramService) send(ctx context.Context, cfg model.TelegramSettings, text string) error {
	token := strings.TrimSpace(cfg.BotToken)
	chatID := strings.TrimSpace(cfg.ChatID)
	if token == "" || chatID == "" {
		return ErrTelegramNotConfigured
	}
	if strings.TrimSpace(text) == "" {
		return errors.New("empty message")
	}

	_, err := s.telegramAPI(ctx, token, "sendMessage", map[string]string{
		"chat_id": chatID,
		"text":    text,
	})
	return err
}

func applyTemplate(tpl string, vars map[string]string) string {
	out := tpl
	for k, v := range vars {
		out = strings.ReplaceAll(out, "{"+k+"}", v)
	}
	lines := strings.Split(out, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.Contains(line, "{") && strings.Contains(line, "}") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.TrimSpace(strings.Join(filtered, "\n"))
}

func displayName(email string) string {
	email = strings.TrimSpace(email)
	if at := strings.Index(email, "@"); at > 0 {
		return email[:at]
	}
	return email
}
