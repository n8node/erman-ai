package model

import "time"

type LLMProvider string

const (
	LLMProviderOpenRouter LLMProvider = "openrouter"
	LLMProviderDeepSeek   LLMProvider = "deepseek"
)

// StrategyLLMSettings — public settings (no API keys).
type StrategyLLMSettings struct {
	Provider        LLMProvider `json:"provider"`
	OpenRouterModel string      `json:"openrouter_model"`
	DeepSeekModel   string      `json:"deepseek_model"`
	SystemPrompt    string      `json:"system_prompt"`
	Temperature     float64     `json:"temperature"`
	MaxTokens       int         `json:"max_tokens"`
}

// StrategyLLMStoredConfig — full row persisted in JSONB.
type StrategyLLMStoredConfig struct {
	StrategyLLMSettings
	OpenRouterAPIKey string   `json:"openrouter_api_key"`
	DeepSeekAPIKey   string   `json:"deepseek_api_key"`
	OpenRouterModels []string `json:"openrouter_models"`
	DeepSeekModels   []string `json:"deepseek_models"`
}

type StrategyLLMSettingsRecord struct {
	Config    StrategyLLMStoredConfig `json:"config"`
	UpdatedAt time.Time               `json:"updated_at"`
}

type LLMProviderStatus struct {
	ID           LLMProvider `json:"id"`
	Configured   bool        `json:"configured"`
	KeyHint      string      `json:"key_hint,omitempty"`
	DefaultModel string      `json:"default_model"`
	Models       []string    `json:"models"`
}

type StrategyLLMAdminView struct {
	Settings            StrategyLLMSettings `json:"settings"`
	Providers           []LLMProviderStatus `json:"providers"`
	DefaultSystemPrompt string              `json:"default_system_prompt"`
	UpdatedAt           time.Time           `json:"updated_at"`
}

type StrategyLLMAdminUpdateRequest struct {
	Settings         StrategyLLMSettings `json:"settings"`
	OpenRouterAPIKey string              `json:"openrouter_api_key,omitempty"`
	DeepSeekAPIKey   string              `json:"deepseek_api_key,omitempty"`
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
			Provider:        LLMProviderOpenRouter,
			OpenRouterModel: "anthropic/claude-sonnet-4-5",
			DeepSeekModel:   "deepseek-chat",
			SystemPrompt:    "",
			Temperature:     0.7,
			MaxTokens:       32000,
		},
		OpenRouterModels: []string{},
		DeepSeekModels:   []string{},
	}
}

func (s StrategyLLMSettings) ActiveModel() string {
	if s.Provider == LLMProviderDeepSeek {
		return s.DeepSeekModel
	}
	return s.OpenRouterModel
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
