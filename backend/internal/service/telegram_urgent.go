package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

const telegramCallbackUrgent = "urgent_contact"

func (s *TelegramService) urgentConfigured(cfg model.TelegramSettings) bool {
	return cfg.UrgentEnabled && strings.TrimSpace(cfg.BotToken) != ""
}

func (s *TelegramService) beginUrgentWait(ctx context.Context, cfg model.TelegramSettings, userChatID string) error {
	if s.userState == nil {
		return errors.New("user state repository not configured")
	}
	if err := s.userState.SetMode(ctx, userChatID, model.TelegramUserModeUrgentWait); err != nil {
		return err
	}
	text := strings.TrimSpace(cfg.UrgentInstruction)
	if text == "" {
		text = model.DefaultUrgentInstruction()
	}
	return s.telegramSendMessageOpts(ctx, cfg.BotToken, userChatID, text, 0, nil)
}

func (s *TelegramService) handleUrgentCallback(ctx context.Context, cfg model.TelegramSettings, cq *telegramCallbackQuery) (ack string, err error) {
	if !s.urgentConfigured(cfg) {
		return "Срочная связь временно недоступна", nil
	}
	userChatID := formatChatID(cq.From.ID)
	if err := s.beginUrgentWait(ctx, cfg, userChatID); err != nil {
		s.logger.Warn("telegram urgent wait failed", "chat_id", userChatID, "err", err)
		return "Не удалось начать срочное обращение", err
	}
	return "", nil
}

func (s *TelegramService) handleUrgentUserMessage(ctx context.Context, cfg model.TelegramSettings, msg *telegramMessage) {
	userChatID := formatChatID(msg.Chat.ID)
	token := strings.TrimSpace(cfg.BotToken)

	text := strings.TrimSpace(msg.Text)
	if text == "" {
		_ = s.telegramSendMessageOpts(ctx, token, userChatID,
			"Пожалуйста, отправьте текстовое сообщение для срочного обращения.", 0, nil)
		return
	}

	if s.urgentSends != nil {
		count, err := s.urgentSends.CountTodayUTC(ctx, userChatID)
		if err != nil {
			s.logger.Warn("telegram urgent rate check failed", "err", err)
		} else if count >= model.TelegramUrgentDailyLimit {
			_ = s.userState.SetMode(ctx, userChatID, model.TelegramUserModeIdle)
			_ = s.telegramSendMessageOpts(ctx, token, userChatID,
				"Срочные обращения ограничены: не более 2 сообщений в сутки. Попробуйте завтра.", 0, nil)
			return
		}
	}

	alertText := formatUrgentAlert(msg, userChatID, text)
	delivered := s.dispatchUrgentAlert(ctx, cfg, alertText)

	if delivered == 0 {
		_ = s.telegramSendMessageOpts(ctx, token, userChatID,
			"Не удалось доставить сообщение. Попробуйте позже или напишите напрямую в Telegram.", 0, nil)
		return
	}

	if s.urgentSends != nil {
		preview := truncateRunes(text, 500)
		_ = s.urgentSends.Insert(ctx, userChatID, preview)
	}
	_ = s.userState.SetMode(ctx, userChatID, model.TelegramUserModeIdle)

	_ = s.telegramSendMessageOpts(ctx, token, userChatID,
		"Сообщение отправлено. Ответим в Telegram или по контактам, которые вы указали.", 0, nil)
}

func formatUrgentAlert(msg *telegramMessage, userChatID, text string) string {
	display := formatUserDisplay(msg.From)
	return fmt.Sprintf(
		"🚨 Срочное обращение (Telegram)\n\nКлиент: %s\nChat ID: %s\n\n%s",
		display, userChatID, text,
	)
}

func (s *TelegramService) dispatchUrgentAlert(ctx context.Context, cfg model.TelegramSettings, alertText string) int {
	delivered := 0

	if err := s.sendUrgentTelegram(ctx, cfg, alertText); err != nil {
		s.logger.Warn("urgent telegram delivery failed", "err", err)
	} else {
		delivered++
	}

	email := strings.TrimSpace(cfg.UrgentEmail)
	if email == "" {
		email = "erman.ai@yandex.ru"
	}
	if s.mail != nil {
		subject := "🚨 Срочное обращение — Erman AI"
		body := fmt.Sprintf("<p>%s</p>", strings.ReplaceAll(alertText, "\n", "<br>"))
		if err := s.mail.Send(ctx, email, subject, body); err != nil {
			s.logger.Warn("urgent email delivery failed", "err", err)
		} else {
			delivered++
		}
	}

	if s.max != nil {
		if err := s.max.SendUrgentAlert(ctx, alertText); err != nil {
			if !errors.Is(err, ErrMaxDisabled) {
				s.logger.Warn("urgent max delivery failed", "err", err)
			}
		} else {
			delivered++
		}
	}

	return delivered
}

func (s *TelegramService) sendUrgentTelegram(ctx context.Context, cfg model.TelegramSettings, text string) error {
	chatID := strings.TrimSpace(cfg.ChatID)
	token := strings.TrimSpace(cfg.BotToken)
	if chatID == "" {
		return ErrTelegramNotConfigured
	}
	return s.telegramSendMessageOpts(ctx, token, chatID, text, 0, nil)
}
