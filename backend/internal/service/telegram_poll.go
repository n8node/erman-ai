package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/erman-ai/erman-ai/internal/model"
)

type telegramUpdate struct {
	UpdateID int64            `json:"update_id"`
	Message  *telegramMessage `json:"message"`
}

type telegramMessage struct {
	MessageID int64              `json:"message_id"`
	Text      string             `json:"text"`
	Chat      telegramChat       `json:"chat"`
	Entities  []telegramEntity   `json:"entities"`
}

type telegramChat struct {
	ID int64 `json:"id"`
}

type telegramEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}

func (s *TelegramService) syncPolling(healthy bool, cfg model.TelegramSettings) {
	token := strings.TrimSpace(cfg.BotToken)
	shouldPoll := healthy && cfg.StartEnabled && token != ""

	s.mu.Lock()
	running := s.pollRunning
	s.mu.Unlock()

	if shouldPoll && !running {
		s.startPolling()
	} else if !shouldPoll && running {
		s.stopPolling()
	}
}

func (s *TelegramService) startPolling() {
	s.mu.Lock()
	if s.pollRunning {
		s.mu.Unlock()
		return
	}
	s.pollRunning = true
	s.pollStopCh = make(chan struct{})
	stopCh := s.pollStopCh
	s.mu.Unlock()

	cfg, err := s.settings.GetEffective(context.Background())
	if err == nil {
		token := strings.TrimSpace(cfg.BotToken)
		if token != "" {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			_, _ = s.telegramAPI(ctx, token, "deleteWebhook", map[string]any{"drop_pending_updates": false})
			cancel()
		}
	}

	s.logger.Info("telegram updates polling started")
	go s.pollLoop(stopCh)
}

func (s *TelegramService) stopPolling() {
	s.mu.Lock()
	if !s.pollRunning {
		s.mu.Unlock()
		return
	}
	close(s.pollStopCh)
	s.pollRunning = false
	s.mu.Unlock()
	s.logger.Info("telegram updates polling stopped")
}

func (s *TelegramService) pollLoop(stopCh <-chan struct{}) {
	defer func() {
		s.mu.Lock()
		s.pollRunning = false
		s.mu.Unlock()
	}()

	for {
		if s.pollShouldStop(stopCh) {
			return
		}

		cfg, err := s.settings.GetEffective(context.Background())
		if err != nil || !cfg.StartEnabled || strings.TrimSpace(cfg.BotToken) == "" {
			time.Sleep(2 * time.Second)
			continue
		}

		updates, err := s.fetchUpdates(cfg)
		if err != nil {
			s.logger.Warn("telegram getUpdates failed", "err", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, upd := range updates {
			if upd.Message != nil && isStartCommand(upd.Message) {
				chatID := formatChatID(upd.Message.Chat.ID)
				s.logger.Info("telegram /start received", "chat_id", chatID)
				go func(targetChat string) {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					current, err := s.settings.GetEffective(ctx)
					if err != nil {
						s.logger.Warn("telegram /start settings load failed", "err", err)
						return
					}
					if err := s.sendStartReply(ctx, current, targetChat); err != nil {
						s.logger.Warn("telegram /start reply failed", "chat_id", targetChat, "err", err)
						return
					}
					s.logger.Info("telegram /start reply sent", "chat_id", targetChat)
				}(chatID)
			}
		}
	}
}

func (s *TelegramService) pollShouldStop(stopCh <-chan struct{}) bool {
	select {
	case <-stopCh:
		return true
	case <-s.stopCh:
		return true
	default:
		return false
	}
}

func (s *TelegramService) fetchUpdates(cfg model.TelegramSettings) ([]telegramUpdate, error) {
	s.mu.Lock()
	offset := s.updateOffset
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	raw, err := s.telegramAPI(ctx, strings.TrimSpace(cfg.BotToken), "getUpdates", map[string]any{
		"offset":  offset,
		"timeout": 25,
	})
	if err != nil {
		return nil, err
	}

	var updates []telegramUpdate
	if len(raw) > 0 && raw[0] != '[' {
		return nil, fmt.Errorf("telegram getUpdates: unexpected payload")
	}
	if err := json.Unmarshal(raw, &updates); err != nil {
		return nil, fmt.Errorf("telegram getUpdates parse: %w", err)
	}

	var maxID int64
	for _, u := range updates {
		if u.UpdateID >= maxID {
			maxID = u.UpdateID + 1
		}
	}
	if maxID > 0 {
		s.mu.Lock()
		if maxID > s.updateOffset {
			s.updateOffset = maxID
		}
		s.mu.Unlock()
	}

	return updates, nil
}

func isStartCommand(msg *telegramMessage) bool {
	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return false
	}
	if text == "/start" {
		return true
	}
	if strings.HasPrefix(text, "/start ") {
		return true
	}
	if strings.HasPrefix(text, "/start@") {
		return true
	}
	lower := strings.ToLower(text)
	if lower == "/start" || strings.HasPrefix(lower, "/start ") || strings.HasPrefix(lower, "/start@") {
		return true
	}
	for _, e := range msg.Entities {
		if e.Type != "bot_command" || e.Offset != 0 {
			continue
		}
		// /start is ASCII; byte offset matches UTF-8 for these commands
		end := e.Length
		if end <= 0 || end > len(text) {
			end = len(text)
		}
		cmd := strings.ToLower(strings.TrimSpace(text[:end]))
		if cmd == "/start" || strings.HasPrefix(cmd, "/start@") {
			return true
		}
	}
	return false
}

func formatChatID(id int64) string {
	return strconv.FormatInt(id, 10)
}
