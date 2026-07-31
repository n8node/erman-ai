package model

import "time"

type LegalScanLLMSettings struct {
	Provider        LLMProvider `json:"provider"`
	OpenRouterModel string      `json:"openrouter_model"`
	DeepSeekModel   string      `json:"deepseek_model"`
	YandexModel     string      `json:"yandex_model"`
	OpenRouterProxy LLMHTTPProxySettings `json:"openrouter_proxy"`
	DeepSeekProxy   LLMHTTPProxySettings `json:"deepseek_proxy"`
	SystemPrompt    string      `json:"system_prompt"`
	Temperature     float64     `json:"temperature"`
	MaxTokens       int         `json:"max_tokens"`
}

type LegalScanLLMStoredConfig struct {
	LegalScanLLMSettings
}

type LegalScanLLMSettingsRecord struct {
	Config    LegalScanLLMStoredConfig `json:"config"`
	UpdatedAt time.Time                `json:"updated_at"`
}

type LegalScanLLMAdminView struct {
	Settings            LegalScanLLMSettings `json:"settings"`
	Providers           []LLMProviderStatus  `json:"providers"`
	DefaultSystemPrompt string               `json:"default_system_prompt"`
	UpdatedAt           time.Time            `json:"updated_at"`
}

type LegalScanLLMAdminUpdateRequest struct {
	Settings LegalScanLLMSettings `json:"settings"`
}

func DefaultLegalScanLLMStoredConfig() LegalScanLLMStoredConfig {
	return LegalScanLLMStoredConfig{
		LegalScanLLMSettings: LegalScanLLMSettings{
			Provider:        LLMProviderYandex,
			OpenRouterModel: "google/gemini-flash-1.5-8b",
			DeepSeekModel:   "deepseek-chat",
			YandexModel:     "yandexgpt-lite/latest",
			OpenRouterProxy: DefaultLLMHTTPProxySettings(),
			DeepSeekProxy:   DefaultLLMHTTPProxySettings(),
			SystemPrompt:    "",
			Temperature:     0.3,
			MaxTokens:       4096,
		},
	}
}

func (s LegalScanLLMSettings) ActiveModel() string {
	switch s.Provider {
	case LLMProviderDeepSeek:
		return s.DeepSeekModel
	case LLMProviderYandex:
		return s.YandexModel
	default:
		return s.OpenRouterModel
	}
}
