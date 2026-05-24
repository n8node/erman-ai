package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/model"
)

type LLMService struct {
	cfg    *config.Config
	client *http.Client
}

func NewLLMService(cfg *config.Config) *LLMService {
	return &LLMService{
		cfg: cfg,
		// No fixed timeout — callers pass context.WithTimeout (strategy: 120s).
		client: &http.Client{Timeout: 0},
	}
}

type LLMCompletionRequest struct {
	Provider     model.LLMProvider
	Model        string
	SystemPrompt string
	UserPrompt   string
	Temperature  float64
	MaxTokens    int
	APIKey       string
	BaseURL      string
}

type LLMCompletionResult struct {
	Content          string
	Model            string
	Provider         model.LLMProvider
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

func (s *LLMService) ResolveOpenRouterKey(stored string) string {
	if strings.TrimSpace(stored) != "" {
		return strings.TrimSpace(stored)
	}
	return strings.TrimSpace(s.cfg.OpenRouterAPIKey)
}

func (s *LLMService) ResolveDeepSeekKey(stored string) string {
	if strings.TrimSpace(stored) != "" {
		return strings.TrimSpace(stored)
	}
	return strings.TrimSpace(s.cfg.DeepSeekAPIKey)
}

func (s *LLMService) ResolveKey(provider model.LLMProvider, storedOpenRouter, storedDeepSeek string) string {
	if provider == model.LLMProviderDeepSeek {
		return s.ResolveDeepSeekKey(storedDeepSeek)
	}
	return s.ResolveOpenRouterKey(storedOpenRouter)
}

func (s *LLMService) BaseURL(provider model.LLMProvider) string {
	if provider == model.LLMProviderDeepSeek {
		return strings.TrimRight(s.cfg.DeepSeekBaseURL, "/")
	}
	return strings.TrimRight(s.cfg.OpenRouterBaseURL, "/")
}

func (s *LLMService) Complete(ctx context.Context, req LLMCompletionRequest) (*LLMCompletionResult, error) {
	apiKey := req.APIKey
	baseURL := req.BaseURL
	if apiKey == "" {
		apiKey = s.ResolveKey(req.Provider, "", "")
	}
	if baseURL == "" {
		baseURL = s.BaseURL(req.Provider)
	}
	if apiKey == "" {
		return nil, errors.New("api key not configured")
	}
	openRouter := req.Provider != model.LLMProviderDeepSeek
	return s.postChatCompletion(ctx, baseURL+"/chat/completions", apiKey, req, req.Provider, openRouter)
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type modelsListResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (s *LLMService) ListModels(ctx context.Context, provider model.LLMProvider, apiKey string) ([]string, error) {
	if apiKey == "" {
		return nil, errors.New("api key not configured")
	}
	baseURL := s.BaseURL(provider)
	openRouter := provider != model.LLMProviderDeepSeek

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	if openRouter {
		req.Header.Set("HTTP-Referer", s.cfg.PublicBaseURL())
		req.Header.Set("X-Title", "Erman AI")
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("models request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("models http %d: %s", resp.StatusCode, string(raw))
	}

	var parsed modelsListResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("models parse error: %w", err)
	}

	ids := make([]string, 0, len(parsed.Data))
	seen := make(map[string]struct{})
	for _, item := range parsed.Data {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	if len(ids) == 0 {
		return nil, errors.New("provider returned no models")
	}
	return ids, nil
}

func (s *LLMService) postChatCompletion(
	ctx context.Context,
	url, apiKey string,
	req LLMCompletionRequest,
	provider model.LLMProvider,
	openRouterHeaders bool,
) (*LLMCompletionResult, error) {
	body, err := json.Marshal(chatCompletionRequest{
		Model: req.Model,
		Messages: []chatMessage{
			{Role: "system", Content: req.SystemPrompt},
			{Role: "user", Content: req.UserPrompt},
		},
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	})
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	if openRouterHeaders {
		httpReq.Header.Set("HTTP-Referer", s.cfg.PublicBaseURL())
		httpReq.Header.Set("X-Title", "Erman AI")
	}

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("llm request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed chatCompletionResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("llm response parse error: %w", err)
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("llm error: %s", parsed.Error.Message)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("llm http %d: %s", resp.StatusCode, string(raw))
	}
	if len(parsed.Choices) == 0 {
		return nil, errors.New("llm returned empty choices")
	}

	modelUsed := parsed.Model
	if modelUsed == "" {
		modelUsed = req.Model
	}

	return &LLMCompletionResult{
		Content:          parsed.Choices[0].Message.Content,
		Model:            modelUsed,
		Provider:         provider,
		PromptTokens:     parsed.Usage.PromptTokens,
		CompletionTokens: parsed.Usage.CompletionTokens,
		TotalTokens:      parsed.Usage.TotalTokens,
	}, nil
}
