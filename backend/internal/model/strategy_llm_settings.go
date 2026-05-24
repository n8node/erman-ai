package model

import "time"

type LLMProvider string

const (
	LLMProviderOpenRouter LLMProvider = "openrouter"
	LLMProviderDeepSeek   LLMProvider = "deepseek"
)

type StrategyLLMSettings struct {
	Provider        LLMProvider `json:"provider"`
	OpenRouterModel string      `json:"openrouter_model"`
	DeepSeekModel   string      `json:"deepseek_model"`
	SystemPrompt    string      `json:"system_prompt"`
	Temperature     float64     `json:"temperature"`
	MaxTokens       int         `json:"max_tokens"`
}

type StrategyLLMSettingsRecord struct {
	Settings  StrategyLLMSettings `json:"settings"`
	UpdatedAt time.Time           `json:"updated_at"`
}

type LLMProviderStatus struct {
	ID              LLMProvider `json:"id"`
	Configured      bool        `json:"configured"`
	DefaultModel    string      `json:"default_model"`
	SuggestedModels []string    `json:"suggested_models"`
}

type StrategyLLMAdminView struct {
	Settings            StrategyLLMSettings `json:"settings"`
	Providers           []LLMProviderStatus `json:"providers"`
	DefaultSystemPrompt string              `json:"default_system_prompt"`
	UpdatedAt           time.Time           `json:"updated_at"`
}

func DefaultStrategyLLMSettings() StrategyLLMSettings {
	return StrategyLLMSettings{
		Provider:        LLMProviderOpenRouter,
		OpenRouterModel: "anthropic/claude-sonnet-4-5",
		DeepSeekModel:   "deepseek-chat",
		SystemPrompt:    "",
		Temperature:     0.7,
		MaxTokens:       8192,
	}
}

func (s StrategyLLMSettings) ActiveModel() string {
	if s.Provider == LLMProviderDeepSeek {
		return s.DeepSeekModel
	}
	return s.OpenRouterModel
}
