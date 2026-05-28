package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type telegramUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
}

type telegramAPIResponse struct {
	OK          bool            `json:"ok"`
	Description string          `json:"description"`
	Result      json.RawMessage `json:"result"`
}

func (s *TelegramService) telegramAPI(ctx context.Context, token, method string, payload any) (json.RawMessage, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", strings.TrimSpace(token), method)

	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	} else {
		body = []byte("{}")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	var parsed telegramAPIResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("telegram api: invalid response")
	}
	if !parsed.OK {
		if parsed.Description != "" {
			return nil, fmt.Errorf("telegram api: %s", parsed.Description)
		}
		return nil, fmt.Errorf("telegram api: request failed")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("telegram api: status %d", resp.StatusCode)
	}
	return parsed.Result, nil
}

func (s *TelegramService) telegramGetMe(ctx context.Context, token string) (*telegramUser, error) {
	raw, err := s.telegramAPI(ctx, token, "getMe", nil)
	if err != nil {
		return nil, err
	}
	var user telegramUser
	if err := json.Unmarshal(raw, &user); err != nil {
		return nil, errors.New("telegram api: invalid getMe result")
	}
	return &user, nil
}

func (s *TelegramService) telegramGetChat(ctx context.Context, token, chatID string) error {
	_, err := s.telegramAPI(ctx, token, "getChat", map[string]string{
		"chat_id": strings.TrimSpace(chatID),
	})
	return err
}
