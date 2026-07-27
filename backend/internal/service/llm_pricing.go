package service

import (
	"strings"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/model"
)

// LLMCredentials holds resolved or stored API credentials for all providers.
type LLMCredentials struct {
	OpenRouterKey  string
	DeepSeekKey    string
	YandexKey      string
	YandexFolderID string
}

func CalculateTokenCost(p model.LLMProviderPricing, promptTokens, completionTokens int) (costUSD, costRUB float64) {
	if p.InputPer1K <= 0 && p.OutputPer1K <= 0 {
		return 0, 0
	}
	total := (float64(promptTokens)/1000)*p.InputPer1K +
		(float64(completionTokens)/1000)*p.OutputPer1K
	switch strings.ToUpper(strings.TrimSpace(p.Currency)) {
	case "USD":
		return total, 0
	default:
		return 0, total
	}
}

func ResolveProviderPricing(
	provider model.LLMProvider,
	stored map[model.LLMProvider]model.LLMProviderPricing,
	cfg *config.Config,
) model.LLMProviderPricing {
	if stored != nil {
		if p, ok := stored[provider]; ok && (p.InputPer1K > 0 || p.OutputPer1K > 0) {
			if p.Currency == "" {
				p.Currency = defaultCurrencyForProvider(provider)
			}
			return p
		}
	}
	return defaultPricingFromEnv(provider, cfg)
}

func defaultCurrencyForProvider(provider model.LLMProvider) string {
	if provider == model.LLMProviderYandex {
		return "RUB"
	}
	return "USD"
}

func defaultPricingFromEnv(provider model.LLMProvider, cfg *config.Config) model.LLMProviderPricing {
	switch provider {
	case model.LLMProviderYandex:
		return model.LLMProviderPricing{
			InputPer1K:  cfg.YandexPriceInputRUBPer1K,
			OutputPer1K: cfg.YandexPriceOutputRUBPer1K,
			Currency:    "RUB",
		}
	case model.LLMProviderDeepSeek:
		return model.LLMProviderPricing{
			InputPer1K:  cfg.DeepSeekPriceInputUSDPer1K,
			OutputPer1K: cfg.DeepSeekPriceOutputUSDPer1K,
			Currency:    "USD",
		}
	default:
		return model.LLMProviderPricing{
			InputPer1K:  cfg.OpenRouterPriceInputUSDPer1K,
			OutputPer1K: cfg.OpenRouterPriceOutputUSDPer1K,
			Currency:    "USD",
		}
	}
}

func UsageLogCosts(
	provider model.LLMProvider,
	stored map[model.LLMProvider]model.LLMProviderPricing,
	cfg *config.Config,
	promptTokens, completionTokens int,
) (costUSD, costRUB float64) {
	p := ResolveProviderPricing(provider, stored, cfg)
	return CalculateTokenCost(p, promptTokens, completionTokens)
}

func YandexModelURI(folderID, modelID string) string {
	modelID = strings.TrimSpace(modelID)
	if modelID == "" {
		modelID = "yandexgpt/latest"
	}
	if strings.HasPrefix(modelID, "gpt://") {
		return modelID
	}
	folderID = strings.TrimSpace(folderID)
	if folderID == "" {
		return modelID
	}
	return "gpt://" + folderID + "/" + strings.TrimPrefix(modelID, "/")
}

func DefaultYandexModels(folderID string) []string {
	folderID = strings.TrimSpace(folderID)
	if folderID == "" {
		return []string{"yandexgpt/latest", "yandexgpt-lite/latest"}
	}
	return []string{
		"gpt://" + folderID + "/yandexgpt/latest",
		"gpt://" + folderID + "/yandexgpt-lite/latest",
	}
}
