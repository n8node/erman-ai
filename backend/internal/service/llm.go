package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

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
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type LLMCompletionRequest struct {
	Provider     model.LLMProvider
	Model        string
	SystemPrompt string
	UserPrompt   string
	Temperature  float64
	MaxTokens    int
}

type LLMCompletionResult struct {
	Content          string
	Model            string
	Provider         model.LLMProvider
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

func (s *LLMService) Complete(ctx context.Context, req LLMCompletionRequest) (*LLMCompletionResult, error) {
	switch req.Provider {
	case model.LLMProviderDeepSeek:
		return s.completeDeepSeek(ctx, req)
	case model.LLMProviderOpenRouter:
		fallthrough
	default:
		return s.completeOpenRouter(ctx, req)
	}
}

func (s *LLMService) ProviderConfigured(provider model.LLMProvider) bool {
	switch provider {
	case model.LLMProviderDeepSeek:
		return s.cfg.DeepSeekAPIKey != ""
	default:
		return s.cfg.OpenRouterAPIKey != ""
	}
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

func (s *LLMService) completeOpenRouter(ctx context.Context, req LLMCompletionRequest) (*LLMCompletionResult, error) {
	if s.cfg.OpenRouterAPIKey == "" {
		return nil, errors.New("openrouter api key not configured")
	}
	return s.postChatCompletion(ctx, s.cfg.OpenRouterBaseURL+"/chat/completions", s.cfg.OpenRouterAPIKey, req, model.LLMProviderOpenRouter, true)
}

func (s *LLMService) completeDeepSeek(ctx context.Context, req LLMCompletionRequest) (*LLMCompletionResult, error) {
	if s.cfg.DeepSeekAPIKey == "" {
		return nil, errors.New("deepseek api key not configured")
	}
	return s.postChatCompletion(ctx, s.cfg.DeepSeekBaseURL+"/chat/completions", s.cfg.DeepSeekAPIKey, req, model.LLMProviderDeepSeek, false)
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
