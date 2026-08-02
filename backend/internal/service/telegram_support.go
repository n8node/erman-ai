package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

func (s *TelegramService) supportConfigured(cfg model.TelegramSettings) bool {
	return cfg.SupportEnabled &&
		strings.TrimSpace(cfg.BotToken) != "" &&
		strings.TrimSpace(cfg.SupportForumChatID) != ""
}

func (s *TelegramService) handleCallbackQuery(ctx context.Context, cfg model.TelegramSettings, cq *telegramCallbackQuery) {
	token := strings.TrimSpace(cfg.BotToken)
	if cq == nil || token == "" {
		return
	}
	ack := ""
	defer func() { _ = s.telegramAnswerCallbackQuery(ctx, token, cq.ID, ack) }()

	if cq.Data == telegramCallbackUrgent {
		var err error
		ack, err = s.handleUrgentCallback(ctx, cfg, cq)
		if err != nil {
			return
		}
		if ack == "" {
			ack = "Напишите сообщение ниже"
		}
		return
	}

	if cq.Data != telegramCallbackConsultation {
		return
	}
	if !s.supportConfigured(cfg) {
		ack = "Консультация временно недоступна"
		return
	}

	userChatID := formatChatID(cq.From.ID)
	if err := s.beginConsultation(ctx, cfg, userChatID, &cq.From); err != nil {
		s.logger.Warn("telegram consultation start failed", "chat_id", userChatID, "err", err)
		ack = "Не удалось открыть консультацию"
		return
	}
	ack = "Напишите ваш вопрос — ответим здесь"
}

func (s *TelegramService) beginConsultation(ctx context.Context, cfg model.TelegramSettings, userChatID string, from *telegramUser) error {
	if s.threads == nil {
		return errors.New("support threads repository not configured")
	}

	forumID := strings.TrimSpace(cfg.SupportForumChatID)
	display := formatUserDisplay(from)

	existing, err := s.threads.GetByUserChatID(ctx, userChatID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return err
	}
	if existing != nil {
		if s.supportThreadCurrent(existing, forumID) {
			_ = s.threads.Touch(ctx, userChatID)
			return s.telegramSendMessageOpts(ctx, cfg.BotToken, userChatID,
				"Консультация уже открыта. Напишите сообщение — мы ответим в этом чате.", 0, nil)
		}
		s.logger.Info("telegram consultation thread stale, recreating",
			"chat_id", userChatID,
			"old_forum", existing.ForumChatID,
			"new_forum", forumID,
		)
		if err := s.threads.DeleteByUserChatID(ctx, userChatID); err != nil {
			return err
		}
	}

	topicName := truncateRunes("👤 "+display, 128)
	topicID, err := s.telegramCreateForumTopic(ctx, cfg.BotToken, forumID, topicName)
	if err != nil {
		return err
	}

	thread := model.TelegramSupportThread{
		UserChatID:  userChatID,
		ForumChatID: forumID,
		TopicID:     topicID,
		DisplayName: display,
	}
	if err := s.threads.Upsert(ctx, thread); err != nil {
		return err
	}

	intro := fmt.Sprintf("🆕 Консультация\nКлиент: %s\nChat ID: %s\n\nОтветьте в этом топике — сообщение уйдёт клиенту в бота.", display, userChatID)
	if err := s.telegramSendMessageOpts(ctx, cfg.BotToken, forumID, intro, topicID, nil); err != nil {
		return err
	}

	return s.telegramSendMessageOpts(ctx, cfg.BotToken, userChatID,
		"Вы на связи с консультантом. Напишите сообщение — ответ придёт сюда.", 0, nil)
}

func (s *TelegramService) handleUserMessage(ctx context.Context, cfg model.TelegramSettings, msg *telegramMessage) {
	if msg == nil || msg.From == nil || msg.From.IsBot {
		return
	}
	if !isPrivateChat(msg.Chat.ID) {
		return
	}
	if isStartCommand(msg) {
		return
	}

	userChatID := formatChatID(msg.Chat.ID)
	token := strings.TrimSpace(cfg.BotToken)

	if s.userState != nil {
		mode, err := s.userState.GetMode(ctx, userChatID)
		if err != nil {
			s.logger.Warn("telegram user mode lookup failed", "err", err)
		} else if mode == model.TelegramUserModeUrgentWait {
			s.handleUrgentUserMessage(ctx, cfg, msg)
			return
		}
	}

	if s.threads == nil {
		return
	}

	thread, err := s.threads.GetByUserChatID(ctx, userChatID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		s.logger.Warn("telegram support thread lookup failed", "err", err)
		return
	}
	if thread == nil {
		if s.supportConfigured(cfg) {
			_ = s.telegramSendMessageOpts(ctx, token, userChatID,
				"Чтобы связаться с консультантом, нажмите кнопку «Консультация» под приветствием или отправьте /start.", 0, nil)
		}
		return
	}

	if !s.supportConfigured(cfg) {
		return
	}

	if !s.supportThreadCurrent(thread, strings.TrimSpace(cfg.SupportForumChatID)) {
		s.logger.Info("telegram consultation thread stale on message, recreating",
			"chat_id", userChatID,
			"old_forum", thread.ForumChatID,
			"new_forum", cfg.SupportForumChatID,
		)
		_ = s.threads.DeleteByUserChatID(ctx, userChatID)
		if err := s.beginConsultation(ctx, cfg, userChatID, msg.From); err != nil {
			s.logger.Warn("telegram consultation recreate failed", "chat_id", userChatID, "err", err)
			_ = s.telegramSendMessageOpts(ctx, token, userChatID,
				"Не удалось открыть консультацию. Нажмите «Консультация» ещё раз.", 0, nil)
			return
		}
		thread, err = s.threads.GetByUserChatID(ctx, userChatID)
		if err != nil || thread == nil {
			return
		}
	}

	dest := map[string]any{
		"chat_id":           thread.ForumChatID,
		"message_thread_id": thread.TopicID,
		"from_chat_id":      userChatID,
		"message_id":        msg.MessageID,
	}
	if err := s.telegramCopyMessage(ctx, token, dest); err != nil {
		s.logger.Warn("telegram forward user->forum failed", "chat_id", userChatID, "err", err)
		if s.healSupportThread(ctx, cfg, userChatID, msg.From) {
			thread, lookupErr := s.threads.GetByUserChatID(ctx, userChatID)
			if lookupErr == nil && thread != nil {
				dest["message_thread_id"] = thread.TopicID
				dest["chat_id"] = thread.ForumChatID
				if retryErr := s.telegramCopyMessage(ctx, token, dest); retryErr == nil {
					_ = s.threads.Touch(ctx, userChatID)
					return
				}
			}
		}
		_ = s.telegramSendMessageOpts(ctx, token, userChatID,
			"Не удалось передать сообщение консультанту. Нажмите «Консультация» и попробуйте снова.", 0, nil)
		return
	}
	_ = s.threads.Touch(ctx, userChatID)
}

func (s *TelegramService) supportThreadCurrent(thread *model.TelegramSupportThread, forumChatID string) bool {
	if thread == nil {
		return false
	}
	return chatIDEqual(parseChatID(forumChatID), thread.ForumChatID) ||
		strings.TrimSpace(thread.ForumChatID) == strings.TrimSpace(forumChatID)
}

func parseChatID(raw string) int64 {
	parsed, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func (s *TelegramService) healSupportThread(ctx context.Context, cfg model.TelegramSettings, userChatID string, from *telegramUser) bool {
	if s.threads == nil {
		return false
	}
	if err := s.threads.DeleteByUserChatID(ctx, userChatID); err != nil {
		s.logger.Warn("telegram support thread delete failed", "chat_id", userChatID, "err", err)
		return false
	}
	if err := s.beginConsultation(ctx, cfg, userChatID, from); err != nil {
		s.logger.Warn("telegram support thread heal failed", "chat_id", userChatID, "err", err)
		return false
	}
	return true
}

func (s *TelegramService) handleForumMessage(ctx context.Context, cfg model.TelegramSettings, msg *telegramMessage) {
	if msg == nil || !s.supportConfigured(cfg) || s.threads == nil {
		return
	}
	if msg.From != nil && msg.From.IsBot {
		return
	}

	forumID := strings.TrimSpace(cfg.SupportForumChatID)
	if !chatIDEqual(msg.Chat.ID, forumID) {
		return
	}
	if msg.MessageThreadID <= 0 {
		return
	}

	thread, err := s.threads.GetByForumTopic(ctx, forumID, msg.MessageThreadID)
	if err != nil || thread == nil {
		return
	}

	token := strings.TrimSpace(cfg.BotToken)
	userChatID := thread.UserChatID

	_ = s.telegramSendChatAction(ctx, token, userChatID, "typing")

	dest := map[string]any{
		"chat_id":      userChatID,
		"from_chat_id": forumID,
		"message_id":   msg.MessageID,
	}
	if err := s.telegramCopyMessage(ctx, token, dest); err != nil {
		s.logger.Warn("telegram forward forum->user failed", "err", err)
	}
}

func formatUserDisplay(from *telegramUser) string {
	if from == nil {
		return "Клиент"
	}
	name := strings.TrimSpace(strings.TrimSpace(from.FirstName + " " + from.LastName))
	if from.Username != "" {
		if name != "" {
			return fmt.Sprintf("%s (@%s)", name, from.Username)
		}
		return "@" + from.Username
	}
	if name != "" {
		return name
	}
	return "ID " + formatChatID(from.ID)
}

func isPrivateChat(chatID int64) bool {
	return chatID > 0
}

func chatIDEqual(chatID int64, chatIDStr string) bool {
	parsed, err := strconv.ParseInt(strings.TrimSpace(chatIDStr), 10, 64)
	if err != nil {
		return false
	}
	return parsed == chatID
}
