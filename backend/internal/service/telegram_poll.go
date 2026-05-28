package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/erman-ai/erman-ai/internal/model"
)

type telegramUpdatesResponse struct {
	Result []telegramUpdate `json:"result"`
}

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
		select {
		case <-stopCh:
			return
		case <-s.stopCh:
			return
		default:
		}

		cfg, err := s.settings.GetEffective(context.Background())
		if err != nil || !cfg.StartEnabled || strings.TrimSpace(cfg.BotToken) == "" {
			time.Sleep(2 * time.Second)
			continue
		}

		updates, err := s.fetchUpdates(cfg)
		if err != nil {
			slog.Warn("telegram getUpdates failed", "err", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, upd := range updates {
			if upd.Message != nil && isStartCommand(upd.Message) {
				chatID := formatChatID(upd.Message.Chat.ID)
				go func(targetChat string) {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					current, err := s.settings.GetEffective(ctx)
					if err != nil {
						return
					}
					if err := s.sendStartReply(ctx, current, targetChat); err != nil {
						s.logger.Warn("telegram /start reply failed", "chat_id", targetChat, "err", err)
					}
				}(chatID)
			}
		}

		if len(updates) == 0 {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (s *TelegramService) fetchUpdates(cfg model.TelegramSettings) ([]telegramUpdate, error) {
	s.mu.Lock()
	offset := s.updateOffset
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	raw, err := s.telegramAPI(ctx, strings.TrimSpace(cfg.BotToken), "getUpdates", map[string]any{
		"offset":          offset,
		"timeout":         25,
		"allowed_updates": []string{"message"},
	})
	if err != nil {
		return nil, err
	}

	var resp telegramUpdatesResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}

	var maxID int64
	for _, u := range resp.Result {
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

	return resp.Result, nil
}

func isStartCommand(msg *telegramMessage) bool {
	text := strings.TrimSpace(msg.Text)
	if text == "/start" {
		return true
	}
	if strings.HasPrefix(text, "/start ") || strings.HasPrefix(text, "/start@") {
		return true
	}
	for _, e := range msg.Entities {
		if e.Type == "bot_command" && e.Offset == 0 {
			cmd := text
			if e.Length > 0 && e.Length <= len(text) {
				cmd = text[:e.Length]
			}
			if cmd == "/start" || strings.HasPrefix(cmd, "/start@") {
				return true
			}
		}
	}
	return false
}

func formatChatID(id int64) string {
	return strconv.FormatInt(id, 10)
}
