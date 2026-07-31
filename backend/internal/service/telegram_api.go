package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

type telegramUser struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type telegramForumTopic struct {
	MessageThreadID int    `json:"message_thread_id"`
	Name            string `json:"name"`
}

type telegramAPIResponse struct {
	OK          bool            `json:"ok"`
	Description string          `json:"description"`
	Result      json.RawMessage `json:"result"`
}

func (s *TelegramService) doTelegramRequest(
	ctx context.Context,
	method string,
	endpoint string,
	contentType string,
	body []byte,
) (*http.Response, error) {
	reqBody := body
	if reqBody == nil {
		reqBody = []byte{}
	}
	makeRequest := func(client *http.Client) (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(reqBody))
		if err != nil {
			return nil, err
		}
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		return client.Do(req)
	}

	client := s.client
	cfg, err := s.settings.GetEffective(ctx)
	if err != nil || !cfg.ProxyEnabled || len(cfg.ProxyURLs) == 0 {
		return makeRequest(client)
	}

	proxies := telegramProxyOrder(cfg)
	var lastErr error
	for idx, proxyURL := range proxies {
		proxyClient, err := httpClientForProxy(client, proxyURL)
		if err != nil {
			lastErr = fmt.Errorf("proxy %q: %w", proxyURL, err)
			if !cfg.ProxyAutoFailover {
				return nil, lastErr
			}
			continue
		}
		resp, reqErr := makeRequest(proxyClient)
		if reqErr == nil {
			return resp, nil
		}
		lastErr = fmt.Errorf("proxy %q: %w", proxyURL, reqErr)
		if !cfg.ProxyAutoFailover || idx == len(proxies)-1 {
			return nil, lastErr
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("proxy request failed")
}

func telegramProxyOrder(cfg model.TelegramSettings) []string {
	urls := normalizeProxyURLs(cfg.ProxyURLs)
	if len(urls) == 0 {
		return nil
	}
	active := strings.TrimSpace(cfg.ProxyActiveURL)
	if active == "" || !slices.Contains(urls, active) {
		return urls
	}
	out := make([]string, 0, len(urls))
	out = append(out, active)
	for _, raw := range urls {
		if raw == active {
			continue
		}
		out = append(out, raw)
	}
	return out
}

func httpClientForProxy(base *http.Client, proxyURL string) (*http.Client, error) {
	parsed, err := url.Parse(strings.TrimSpace(proxyURL))
	if err != nil {
		return nil, err
	}
	if base == nil {
		base = &http.Client{}
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(parsed),
	}
	return &http.Client{
		Timeout:   base.Timeout,
		Transport: transport,
	}, nil
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

	resp, err := s.doTelegramRequest(ctx, http.MethodPost, url, "application/json", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	readLimit := int64(8192)
	if method == "getUpdates" {
		readLimit = 512 * 1024
	}
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, readLimit))
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

func (s *TelegramService) telegramSendMessage(ctx context.Context, token, chatID, text string) error {
	return s.telegramSendMessageOpts(ctx, token, chatID, text, 0, nil)
}

func (s *TelegramService) telegramSendMessageOpts(ctx context.Context, token, chatID, text string, threadID int, replyMarkup []byte) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	payload := map[string]any{
		"chat_id": strings.TrimSpace(chatID),
		"text":    truncateRunes(text, model.TelegramMessageMaxRunes),
	}
	if threadID > 0 {
		payload["message_thread_id"] = threadID
	}
	if len(replyMarkup) > 0 {
		payload["reply_markup"] = json.RawMessage(replyMarkup)
	}
	_, err := s.telegramAPI(ctx, token, "sendMessage", payload)
	return err
}

func (s *TelegramService) telegramSendChatAction(ctx context.Context, token, chatID, action string) error {
	_, err := s.telegramAPI(ctx, token, "sendChatAction", map[string]string{
		"chat_id": strings.TrimSpace(chatID),
		"action":  action,
	})
	return err
}

func (s *TelegramService) telegramAnswerCallbackQuery(ctx context.Context, token, callbackID, text string) error {
	payload := map[string]string{"callback_query_id": callbackID}
	if text != "" {
		payload["text"] = text
		payload["show_alert"] = "false"
	}
	_, err := s.telegramAPI(ctx, token, "answerCallbackQuery", payload)
	return err
}

func (s *TelegramService) telegramCreateForumTopic(ctx context.Context, token, forumChatID, name string) (int, error) {
	name = truncateRunes(strings.TrimSpace(name), 128)
	raw, err := s.telegramAPI(ctx, token, "createForumTopic", map[string]any{
		"chat_id": forumChatID,
		"name":    name,
	})
	if err != nil {
		return 0, err
	}
	var topic telegramForumTopic
	if err := json.Unmarshal(raw, &topic); err != nil {
		return 0, err
	}
	if topic.MessageThreadID <= 0 {
		return 0, errors.New("telegram api: invalid topic id")
	}
	return topic.MessageThreadID, nil
}

func (s *TelegramService) telegramCopyMessage(ctx context.Context, token string, dest map[string]any) error {
	_, err := s.telegramAPI(ctx, token, "copyMessage", dest)
	return err
}

func (s *TelegramService) telegramSendPhotoFile(ctx context.Context, token, chatID, filePath, caption string, replyMarkup []byte) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	_ = w.WriteField("chat_id", strings.TrimSpace(chatID))
	if c := strings.TrimSpace(caption); c != "" {
		_ = w.WriteField("caption", truncateRunes(c, model.TelegramCaptionMaxRunes))
	}
	if len(replyMarkup) > 0 {
		_ = w.WriteField("reply_markup", string(replyMarkup))
	}

	part, err := w.CreateFormFile("photo", filepath.Base(filePath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", strings.TrimSpace(token))
	resp, err := s.doTelegramRequest(ctx, http.MethodPost, url, w.FormDataContentType(), body.Bytes())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	var parsed telegramAPIResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return fmt.Errorf("telegram api: invalid response")
	}
	if !parsed.OK {
		if parsed.Description != "" {
			return fmt.Errorf("telegram api: %s", parsed.Description)
		}
		return fmt.Errorf("telegram api: sendPhoto failed")
	}
	return nil
}
