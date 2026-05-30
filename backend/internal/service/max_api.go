package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const maxAPIBase = "https://platform-api.max.ru"

type maxAPIUser struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	FirstName string `json:"first_name"`
	IsBot    bool   `json:"is_bot"`
}

func (s *MaxService) maxRequest(ctx context.Context, token, method, path string, query url.Values, body any) ([]byte, int, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, 0, ErrMaxNotConfigured
	}

	u, err := url.Parse(maxAPIBase + path)
	if err != nil {
		return nil, 0, err
	}
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	var reqBody []byte
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
	}

	var httpReq *http.Request
	if reqBody != nil {
		httpReq, err = http.NewRequestWithContext(ctx, method, u.String(), bytes.NewReader(reqBody))
	} else {
		httpReq, err = http.NewRequestWithContext(ctx, method, u.String(), nil)
	}
	if err != nil {
		return nil, 0, err
	}
	httpReq.Header.Set("Authorization", token)
	if reqBody != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	return raw, resp.StatusCode, nil
}

func (s *MaxService) maxGetMe(ctx context.Context, token string) (*maxAPIUser, error) {
	raw, status, err := s.maxRequest(ctx, token, http.MethodGet, "/me", nil, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("max api /me: status %d: %s", status, truncateRunes(string(raw), 200))
	}
	var user maxAPIUser
	if err := json.Unmarshal(raw, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *MaxService) maxSendMessage(ctx context.Context, cfg maxSendConfig, text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if len([]rune(text)) > 4000 {
		text = truncateRunes(text, 4000)
	}

	query := url.Values{}
	if cfg.NotifyUserID > 0 {
		query.Set("user_id", strconv.FormatInt(cfg.NotifyUserID, 10))
	} else if cfg.NotifyChatID > 0 {
		query.Set("chat_id", strconv.FormatInt(cfg.NotifyChatID, 10))
	} else {
		return fmt.Errorf("%w: notify user_id or chat_id required", ErrMaxNotConfigured)
	}

	body := map[string]any{
		"text":   text,
		"notify": true,
	}
	raw, status, err := s.maxRequest(ctx, cfg.Token, http.MethodPost, "/messages", query, body)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("max api send: status %d: %s", status, truncateRunes(string(raw), 200))
	}
	return nil
}

type maxSendConfig struct {
	Token        string
	NotifyUserID int64
	NotifyChatID int64
}
