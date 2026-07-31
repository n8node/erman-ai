package model

import (
	"regexp"
	"strings"
	"time"
)

var yandexCloudFolderIDPattern = regexp.MustCompile(`^b1[a-z0-9]{17,25}$`)

type LLMProvider string

const (
	LLMProviderOpenRouter LLMProvider = "openrouter"
	LLMProviderDeepSeek   LLMProvider = "deepseek"
	LLMProviderYandex     LLMProvider = "yandex"
)

// LLMProviderPricing — admin-configured token cost (per 1000 tokens).
type LLMProviderPricing struct {
	InputPer1K  float64 `json:"input_per_1k"`
	OutputPer1K float64 `json:"output_per_1k"`
	Currency    string  `json:"currency"` // RUB or USD
}

// StrategyLLMSettings — public settings (no API keys).
type StrategyLLMSettings struct {
	Provider             LLMProvider `json:"provider"`
	OpenRouterModel      string      `json:"openrouter_model"`
	DeepSeekModel        string      `json:"deepseek_model"`
	YandexModel          string      `json:"yandex_model"`
	OpenRouterProxy      LLMHTTPProxySettings `json:"openrouter_proxy"`
	DeepSeekProxy        LLMHTTPProxySettings `json:"deepseek_proxy"`
	SystemPrompt         string      `json:"system_prompt"`
	ProposalSystemPrompt string      `json:"proposal_system_prompt"`
	Temperature          float64     `json:"temperature"`
	MaxTokens            int         `json:"max_tokens"`
}

// StrategyLLMStoredConfig — full row persisted in JSONB.
type StrategyLLMStoredConfig struct {
	StrategyLLMSettings
	OpenRouterAPIKey string                          `json:"openrouter_api_key"`
	DeepSeekAPIKey   string                          `json:"deepseek_api_key"`
	YandexAPIKey     string                          `json:"yandex_api_key"`
	YandexFolderID   string                          `json:"yandex_folder_id"`
	OpenRouterModels []string                        `json:"openrouter_models"`
	DeepSeekModels   []string                        `json:"deepseek_models"`
	YandexModels     []string                        `json:"yandex_models"`
	Pricing          map[LLMProvider]LLMProviderPricing `json:"pricing"`
	ModelPricing     map[string]LLMProviderPricing   `json:"model_pricing"`
}

type StrategyLLMSettingsRecord struct {
	Config    StrategyLLMStoredConfig `json:"config"`
	UpdatedAt time.Time               `json:"updated_at"`
}

type LLMProviderStatus struct {
	ID           LLMProvider `json:"id"`
	Configured   bool        `json:"configured"`
	KeyHint      string      `json:"key_hint,omitempty"`
	FolderHint   string      `json:"folder_hint,omitempty"`
	DefaultModel string      `json:"default_model"`
	Models       []string    `json:"models"`
}

type StrategyLLMAdminView struct {
	Settings                    StrategyLLMSettings                `json:"settings"`
	Providers                   []LLMProviderStatus                `json:"providers"`
	Pricing                     map[LLMProvider]LLMProviderPricing `json:"pricing"`
	ModelPricing                map[string]LLMProviderPricing      `json:"model_pricing"`
	DefaultSystemPrompt         string                             `json:"default_system_prompt"`
	DefaultProposalSystemPrompt string                             `json:"default_proposal_system_prompt"`
	UpdatedAt                   time.Time                          `json:"updated_at"`
}

type StrategyLLMAdminUpdateRequest struct {
	Settings         StrategyLLMSettings                `json:"settings"`
	OpenRouterAPIKey string                             `json:"openrouter_api_key,omitempty"`
	DeepSeekAPIKey   string                             `json:"deepseek_api_key,omitempty"`
	YandexAPIKey     string                             `json:"yandex_api_key,omitempty"`
	YandexFolderID   string                             `json:"yandex_folder_id,omitempty"`
	Pricing          map[LLMProvider]LLMProviderPricing `json:"pricing,omitempty"`
	ModelPricing     map[string]LLMProviderPricing      `json:"model_pricing,omitempty"`
}

type StrategyLLMTestConnectionResult struct {
	Provider LLMProvider `json:"provider"`
	OK       bool        `json:"ok"`
	Message  string      `json:"message"`
	Models   []string    `json:"models"`
}

func DefaultStrategyLLMStoredConfig() StrategyLLMStoredConfig {
	return StrategyLLMStoredConfig{
		StrategyLLMSettings: StrategyLLMSettings{
			Provider:        LLMProviderYandex,
			OpenRouterModel: "anthropic/claude-sonnet-4-5",
			DeepSeekModel:   "deepseek-chat",
			YandexModel:     "yandexgpt/latest",
			OpenRouterProxy: DefaultLLMHTTPProxySettings(),
			DeepSeekProxy:   DefaultLLMHTTPProxySettings(),
			SystemPrompt:    "",
			Temperature:     0.7,
			MaxTokens:       32000,
		},
		OpenRouterModels: []string{},
		DeepSeekModels:   []string{},
		YandexModels:     []string{},
		Pricing:          map[LLMProvider]LLMProviderPricing{},
		ModelPricing:     map[string]LLMProviderPricing{},
	}
}

func (s StrategyLLMSettings) ActiveModel() string {
	switch s.Provider {
	case LLMProviderDeepSeek:
		return s.DeepSeekModel
	case LLMProviderYandex:
		return s.YandexModel
	default:
		return s.OpenRouterModel
	}
}

func MaskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "••••••••"
	}
	return key[:4] + "…" + key[len(key)-4:]
}

func MaskFolderID(id string) string {
	if id == "" {
		return ""
	}
	if len(id) <= 6 {
		return "••••••"
	}
	return id[:3] + "…" + id[len(id)-3:]
}

// IsValidYandexCloudFolderID checks Yandex Cloud catalog ID (b1…), not account email.
func IsValidYandexCloudFolderID(id string) bool {
	id = strings.TrimSpace(id)
	if id == "" || strings.Contains(id, "@") {
		return false
	}
	return yandexCloudFolderIDPattern.MatchString(id)
}
