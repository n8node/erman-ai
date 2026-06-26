package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
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
	if err != nil {
		return
	}
	if !cfg.NotifyRegistration && s.max == nil {
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
	s.dispatchAdminNotification(ctx, text, "registration", cfg.Enabled && cfg.NotifyRegistration)
}

func (s *TelegramService) NotifyEmailVerified(ctx context.Context, user *model.User) {
	cfg, err := s.settings.GetEffective(ctx)
	if err != nil {
		return
	}
	if !cfg.NotifyEmailVerified && s.max == nil {
		return
	}
	vars := map[string]string{
		"email": user.Email,
		"name":  displayName(user.Email),
	}
	text := applyTemplate(cfg.EmailVerifiedTemplate, vars)
	s.dispatchAdminNotification(ctx, text, "email_verified", cfg.Enabled && cfg.NotifyEmailVerified)
}

func (s *TelegramService) NotifyPayment(ctx context.Context, user *model.User, plan *model.Plan, amountRub int) {
	if plan == nil {
		return
	}
	cfg, err := s.settings.GetEffective(ctx)
	if err != nil {
		return
	}
	if !cfg.NotifyPayment && s.max == nil {
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
	s.dispatchAdminNotification(ctx, text, "payment", cfg.Enabled && cfg.NotifyPayment)
}

func (s *TelegramService) NotifyConsultationBooking(ctx context.Context, booking *model.ConsultationBookingDetail, baseURL string) {
	if booking == nil {
		return
	}
	cfg, err := s.settings.GetEffective(ctx)
	if err != nil {
		return
	}
	if !cfg.NotifyPayment && s.max == nil {
		return
	}
	when := booking.StartsAt.Format("02.01.2006 15:04")
	adminURL := strings.TrimRight(baseURL, "/") + "/dashboard/admin/consultations"
	lines := []string{
		"Новая оплаченная консультация",
		"",
		"Услуга: " + booking.ServiceName,
		"Дата и время: " + when + " " + booking.Timezone,
		"Клиент: " + booking.CustomerName,
		"Email: " + booking.CustomerEmail,
	}
	if booking.CustomerPhone != "" {
		lines = append(lines, "Телефон: "+booking.CustomerPhone)
	}
	if booking.CustomerTelegram != "" {
		lines = append(lines, "Telegram: "+booking.CustomerTelegram)
	}
	lines = append(lines, fmt.Sprintf("Сумма: %d ₽", booking.AmountRUB))
	if booking.CustomerNote != "" {
		lines = append(lines, "", "Комментарий: "+booking.CustomerNote)
	}
	lines = append(lines, "", "Админка: "+adminURL)
	s.dispatchAdminNotification(ctx, strings.Join(lines, "\n"), "consultation_booking", cfg.Enabled && cfg.NotifyPayment)
}

func (s *TelegramService) dispatchAdminNotification(ctx context.Context, text, kind string, sendTelegram bool) {
	if strings.TrimSpace(text) == "" {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if sendTelegram {
			cfg, err := s.settings.GetEffective(ctx)
			if err == nil && cfg.Enabled {
				if err := s.send(ctx, cfg, text); err != nil {
					slog.Warn("telegram notification failed", "kind", kind, "err", err)
				}
			}
		}
		if s.max != nil {
			if err := s.max.SendNotification(ctx, text); err != nil {
				if !errors.Is(err, ErrMaxDisabled) && !errors.Is(err, ErrMaxNotConfigured) {
					slog.Warn("max notification failed", "kind", kind, "err", err)
				}
			}
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

func (s *TelegramService) NotifyProjectInquiry(ctx context.Context, detail *repository.AdminProjectInquiryDetail, publicBaseURL string) {
	cfg, err := s.settings.GetEffective(ctx)
	sendTelegram := err == nil && cfg.Enabled

	calcLine := "Без расчёта"
	if detail.ProcessName != "" {
		calcLine = detail.ProcessName
		if detail.NetBenefitMonthly != nil {
			calcLine += fmt.Sprintf(" · %.0f ₽/мес", *detail.NetBenefitMonthly)
		}
		if detail.PaybackMonths != nil && *detail.PaybackMonths > 0 && *detail.PaybackMonths < 1e6 {
			calcLine += fmt.Sprintf(" · окупаемость %.1f мес", *detail.PaybackMonths)
		}
	}

	title := detail.ProjectTitle
	if title == "" {
		title = "—"
	}

	adminURL := strings.TrimRight(publicBaseURL, "/") + "/dashboard/admin/project-inquiries/" + detail.ID
	shareURL := ""
	if detail.ShareToken != nil && *detail.ShareToken != "" {
		shareURL = strings.TrimRight(publicBaseURL, "/") + "/dashboard/share/" + *detail.ShareToken
	}

	text := fmt.Sprintf(
		"📋 Новая заявка «Обсудить проект»\n\n"+
			"Имя: %s\nEmail: %s\nTelegram: %s\n\n"+
			"Проект: %s\n\n"+
			"Описание:\n%s\n\n"+
			"Расчёт: %s\n\n"+
			"Открыть: %s",
		detail.Name, detail.Email, detail.Telegram,
		title,
		truncateRunes(detail.ProjectDescription, 800),
		calcLine,
		adminURL,
	)
	if shareURL != "" {
		text += fmt.Sprintf("\nРасчёт (как у клиента): %s", shareURL)
	}
	s.dispatchAdminNotification(ctx, text, "project_inquiry", sendTelegram)
}

func (s *TelegramService) NotifyProposalRequest(ctx context.Context, detail *repository.AdminProposalRequestDetail, publicBaseURL string) {
	cfg, err := s.settings.GetEffective(ctx)
	sendTelegram := err == nil && cfg.Enabled

	calcLine := detail.ProcessName
	if calcLine == "" {
		calcLine = "—"
	}
	if detail.NetBenefitMonthly != nil {
		calcLine += fmt.Sprintf(" · %.0f ₽/мес", *detail.NetBenefitMonthly)
	}
	if detail.PaybackMonths != nil && *detail.PaybackMonths > 0 && *detail.PaybackMonths < 1e6 {
		calcLine += fmt.Sprintf(" · окупаемость %.1f мес", *detail.PaybackMonths)
	}

	adminURL := strings.TrimRight(publicBaseURL, "/") + "/dashboard/admin/proposal-requests/" + detail.ID

	text := fmt.Sprintf(
		"📄 Новая заявка «КП от Erman AI»\n\n"+
			"Имя: %s\nEmail: %s\nTelegram: %s\n\n"+
			"Расчёт: %s\n\n"+
			"Комментарий:\n%s\n\n"+
			"Открыть: %s",
		detail.RequesterName, detail.UserEmail, detail.Telegram,
		calcLine,
		truncateRunes(detail.BusinessNote, 800),
		adminURL,
	)
	s.dispatchAdminNotification(ctx, text, "proposal_request", sendTelegram)
}
