package service

import (
	"bufio"
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
		// Explicit Proxy=nil: LLM egress uses admin proxy settings only, not HTTP_PROXY env.
		client: &http.Client{
			Timeout: 0,
			Transport: &http.Transport{
				Proxy: nil,
			},
		},
	}
}

const llmUserAgent = "ErmanAI/1.0 (+https://erman.ai)"

func (s *LLMService) applyLLMRequestHeaders(req *http.Request, provider model.LLMProvider, creds LLMCredentials) {
	s.applyLLMHeaders(req, s.httpHeaders(provider, creds))
}

func (s *LLMService) applyLLMHeaders(req *http.Request, headers llmHTTPHeaders) {
	req.Header.Set("User-Agent", llmUserAgent)
	if headers.openRouter {
		req.Header.Set("HTTP-Referer", s.cfg.PublicBaseURL())
		req.Header.Set("X-Title", "Erman AI")
	}
	if headers.yandex {
		req.Header.Set("x-folder-id", headers.folderID)
		req.Header.Set("x-data-logging-enabled", "false")
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
	FolderID     string
	Proxy        *model.LLMHTTPProxySettings
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

func (s *LLMService) ResolveYandexKey(stored string) string {
	if strings.TrimSpace(stored) != "" {
		return strings.TrimSpace(stored)
	}
	return strings.TrimSpace(s.cfg.YandexAPIKey)
}

func (s *LLMService) ResolveYandexFolderID(stored string) string {
	if strings.TrimSpace(stored) != "" {
		return strings.TrimSpace(stored)
	}
	return strings.TrimSpace(s.cfg.YandexFolderID)
}

func (s *LLMService) CredentialsFromStored(stored model.StrategyLLMStoredConfig) LLMCredentials {
	return LLMCredentials{
		OpenRouterKey:  s.ResolveOpenRouterKey(stored.OpenRouterAPIKey),
		DeepSeekKey:    s.ResolveDeepSeekKey(stored.DeepSeekAPIKey),
		YandexKey:      s.ResolveYandexKey(stored.YandexAPIKey),
		YandexFolderID: s.ResolveYandexFolderID(stored.YandexFolderID),
	}
}

func (s *LLMService) ResolveKey(provider model.LLMProvider, creds LLMCredentials) string {
	switch provider {
	case model.LLMProviderDeepSeek:
		return creds.DeepSeekKey
	case model.LLMProviderYandex:
		return creds.YandexKey
	default:
		return creds.OpenRouterKey
	}
}

func (s *LLMService) BaseURL(provider model.LLMProvider) string {
	switch provider {
	case model.LLMProviderDeepSeek:
		return strings.TrimRight(s.cfg.DeepSeekBaseURL, "/")
	case model.LLMProviderYandex:
		return strings.TrimRight(s.cfg.YandexBaseURL, "/")
	default:
		return strings.TrimRight(s.cfg.OpenRouterBaseURL, "/")
	}
}

type llmHTTPHeaders struct {
	openRouter bool
	yandex     bool
	folderID   string
}

func (s *LLMService) httpHeaders(provider model.LLMProvider, creds LLMCredentials) llmHTTPHeaders {
	switch provider {
	case model.LLMProviderDeepSeek:
		return llmHTTPHeaders{}
	case model.LLMProviderYandex:
		folderID := creds.YandexFolderID
		if folderID == "" {
			folderID = s.cfg.YandexFolderID
		}
		return llmHTTPHeaders{yandex: true, folderID: folderID}
	default:
		return llmHTTPHeaders{openRouter: true}
	}
}

func (s *LLMService) resolveModel(provider model.LLMProvider, modelID string, creds LLMCredentials) string {
	if provider != model.LLMProviderYandex {
		return modelID
	}
	folderID := creds.YandexFolderID
	if folderID == "" {
		folderID = s.cfg.YandexFolderID
	}
	return YandexModelURI(folderID, modelID)
}

func (s *LLMService) Complete(ctx context.Context, req LLMCompletionRequest) (*LLMCompletionResult, error) {
	creds := LLMCredentials{
		YandexFolderID: req.FolderID,
	}
	if req.APIKey != "" {
		switch req.Provider {
		case model.LLMProviderDeepSeek:
			creds.DeepSeekKey = req.APIKey
		case model.LLMProviderYandex:
			creds.YandexKey = req.APIKey
		default:
			creds.OpenRouterKey = req.APIKey
		}
	} else {
		creds.YandexFolderID = req.FolderID
	}

	apiKey := req.APIKey
	if apiKey == "" {
		apiKey = s.ResolveKey(req.Provider, creds)
	}
	baseURL := req.BaseURL
	if baseURL == "" {
		baseURL = s.BaseURL(req.Provider)
	}
	if apiKey == "" {
		return nil, errors.New("api key not configured")
	}
	if req.Provider == model.LLMProviderYandex {
		folderID := req.FolderID
		if folderID == "" {
			folderID = s.cfg.YandexFolderID
		}
		if folderID == "" {
			return nil, errors.New("yandex folder id not configured")
		}
		creds.YandexFolderID = folderID
	}

	req.Model = s.resolveModel(req.Provider, req.Model, creds)
	headers := s.httpHeaders(req.Provider, creds)
	return s.postChatCompletion(ctx, baseURL+"/chat/completions", apiKey, req, req.Provider, headers, req.Proxy)
}

// StreamComplete calls the chat API with stream=true and invokes onDelta for each content token.
func (s *LLMService) StreamComplete(ctx context.Context, req LLMCompletionRequest, onDelta func(string)) (*LLMCompletionResult, error) {
	creds := LLMCredentials{YandexFolderID: req.FolderID}
	apiKey := req.APIKey
	if apiKey == "" {
		apiKey = s.ResolveKey(req.Provider, creds)
	}
	baseURL := req.BaseURL
	if baseURL == "" {
		baseURL = s.BaseURL(req.Provider)
	}
	if apiKey == "" {
		return nil, errors.New("api key not configured")
	}
	if req.Provider == model.LLMProviderYandex {
		folderID := req.FolderID
		if folderID == "" {
			folderID = s.cfg.YandexFolderID
		}
		if folderID == "" {
			return nil, errors.New("yandex folder id not configured")
		}
		creds.YandexFolderID = folderID
	}

	req.Model = s.resolveModel(req.Provider, req.Model, creds)
	headers := s.httpHeaders(req.Provider, creds)
	return s.postChatCompletionStream(ctx, baseURL+"/chat/completions", apiKey, req, req.Provider, headers, onDelta, req.Proxy)
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
	Stream      bool          `json:"stream,omitempty"`
}

type streamChunkResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Model string `json:"model"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
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

func (s *LLMService) ListModels(ctx context.Context, provider model.LLMProvider, creds LLMCredentials, proxy *model.LLMHTTPProxySettings) ([]string, error) {
	apiKey := s.ResolveKey(provider, creds)
	if apiKey == "" {
		return nil, errors.New("api key not configured")
	}
	if provider == model.LLMProviderYandex {
		folderID := creds.YandexFolderID
		if folderID == "" {
			return nil, errors.New("yandex folder id not configured")
		}
	}

	baseURL := s.BaseURL(provider)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	s.applyLLMRequestHeaders(req, provider, creds)

	resp, err := doHTTPWithProxy(ctx, s.client, proxy, func(client *http.Client) (*http.Response, error) {
		return client.Do(req)
	})
	if err != nil {
		if provider == model.LLMProviderYandex {
			return DefaultYandexModels(creds.YandexFolderID), nil
		}
		return nil, fmt.Errorf("models request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		if provider == model.LLMProviderYandex {
			return DefaultYandexModels(creds.YandexFolderID), nil
		}
		return nil, fmt.Errorf("models http %d: %s", resp.StatusCode, formatLLMModelsError(provider, resp.StatusCode, string(raw)))
	}

	var parsed modelsListResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		if provider == model.LLMProviderYandex {
			return DefaultYandexModels(creds.YandexFolderID), nil
		}
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
	if provider == model.LLMProviderYandex {
		ids = MergeYandexChatModels(ids, creds.YandexFolderID)
		if len(ids) == 0 {
			return DefaultYandexModels(creds.YandexFolderID), nil
		}
		return ids, nil
	}
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
	headers llmHTTPHeaders,
	proxy *model.LLMHTTPProxySettings,
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
	s.applyLLMHeaders(httpReq, headers)

	resp, err := doHTTPWithProxy(ctx, s.client, proxy, func(client *http.Client) (*http.Response, error) {
		return client.Do(httpReq)
	})
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

func (s *LLMService) postChatCompletionStream(
	ctx context.Context,
	url, apiKey string,
	req LLMCompletionRequest,
	provider model.LLMProvider,
	headers llmHTTPHeaders,
	onDelta func(string),
	proxy *model.LLMHTTPProxySettings,
) (*LLMCompletionResult, error) {
	body, err := json.Marshal(chatCompletionRequest{
		Model: req.Model,
		Messages: []chatMessage{
			{Role: "system", Content: req.SystemPrompt},
			{Role: "user", Content: req.UserPrompt},
		},
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      true,
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
	httpReq.Header.Set("Accept", "text/event-stream")
	s.applyLLMHeaders(httpReq, headers)

	resp, err := doHTTPWithProxy(ctx, s.client, proxy, func(client *http.Client) (*http.Response, error) {
		return client.Do(httpReq)
	})
	if err != nil {
		return nil, fmt.Errorf("llm stream request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("llm http %d: %s", resp.StatusCode, string(raw))
	}

	var content strings.Builder
	modelUsed := req.Model
	var promptTokens, completionTokens, totalTokens int

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")
		if payload == "[DONE]" {
			break
		}
		var chunk streamChunkResponse
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}
		if chunk.Error != nil {
			return nil, fmt.Errorf("llm error: %s", chunk.Error.Message)
		}
		if chunk.Model != "" {
			modelUsed = chunk.Model
		}
		if chunk.Usage != nil {
			promptTokens = chunk.Usage.PromptTokens
			completionTokens = chunk.Usage.CompletionTokens
			totalTokens = chunk.Usage.TotalTokens
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta.Content
		if delta == "" {
			continue
		}
		content.WriteString(delta)
		if onDelta != nil {
			onDelta(delta)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("llm stream read error: %w", err)
	}

	full := content.String()
	if strings.TrimSpace(full) == "" {
		return nil, errors.New("llm returned empty stream")
	}

	if totalTokens == 0 {
		totalTokens = len(full) / 4
		completionTokens = totalTokens
	}

	return &LLMCompletionResult{
		Content:          full,
		Model:            modelUsed,
		Provider:         provider,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
	}, nil
}

func formatLLMModelsError(provider model.LLMProvider, status int, body string) string {
	body = strings.TrimSpace(body)
	if provider == model.LLMProviderOpenRouter && status == 403 && strings.Contains(body, "security policy") {
		return body + " — enable OpenRouter proxy, save settings, then test again; verify API key at openrouter.ai/keys"
	}
	return body
}
