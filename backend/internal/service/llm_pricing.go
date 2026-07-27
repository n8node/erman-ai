package service

import (
	"sort"
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
	modelName string,
	modelPricing map[string]model.LLMProviderPricing,
	providerPricing map[model.LLMProvider]model.LLMProviderPricing,
	cfg *config.Config,
	promptTokens, completionTokens int,
) (costUSD, costRUB float64) {
	p := ResolveModelPricing(provider, modelName, modelPricing, providerPricing, cfg)
	return CalculateTokenCost(p, promptTokens, completionTokens)
}

func ResolveModelPricing(
	provider model.LLMProvider,
	modelName string,
	modelPricing map[string]model.LLMProviderPricing,
	providerPricing map[model.LLMProvider]model.LLMProviderPricing,
	cfg *config.Config,
) model.LLMProviderPricing {
	modelName = strings.TrimSpace(modelName)
	if modelPricing != nil && modelName != "" {
		if p, ok := modelPricing[modelName]; ok && (p.InputPer1K > 0 || p.OutputPer1K > 0) {
			if p.Currency == "" {
				p.Currency = defaultCurrencyForProvider(provider)
			}
			return p
		}
	}
	return ResolveProviderPricing(provider, providerPricing, cfg)
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
	return KnownYandexChatModels(folderID)
}

func KnownYandexChatModelSuffixes() []string {
	return []string{
		"yandexgpt/latest",
		"yandexgpt-lite/latest",
		"yandexgpt-5/latest",
		"yandexgpt-5-lite/latest",
		"yandexgpt-pro/latest",
		"yandexgpt-pro-5/latest",
		"yandexgpt-pro-5.1/latest",
		"aliceai-llm/latest",
		"qwen3-235b-a22b-fp8/latest",
		"gpt-oss-120b/latest",
		"gpt-oss-20b/latest",
	}
}

func KnownYandexChatModels(folderID string) []string {
	folderID = strings.TrimSpace(folderID)
	suffixes := KnownYandexChatModelSuffixes()
	if folderID == "" {
		return suffixes
	}
	out := make([]string, 0, len(suffixes))
	for _, suffix := range suffixes {
		out = append(out, "gpt://"+folderID+"/"+suffix)
	}
	return out
}

// IsYandexChatModel returns true for gpt:// completion models (not emb:// embeddings).
func IsYandexChatModel(modelID string) bool {
	return strings.HasPrefix(strings.TrimSpace(modelID), "gpt://")
}

func FilterYandexChatModels(models []string) []string {
	out := make([]string, 0, len(models))
	seen := make(map[string]struct{})
	for _, m := range models {
		m = strings.TrimSpace(m)
		if !IsYandexChatModel(m) {
			continue
		}
		if _, ok := seen[m]; ok {
			continue
		}
		seen[m] = struct{}{}
		out = append(out, m)
	}
	return out
}

func YandexFolderFromModelURI(modelURI string) string {
	modelURI = strings.TrimSpace(modelURI)
	if !strings.HasPrefix(modelURI, "gpt://") {
		return ""
	}
	rest := strings.TrimPrefix(modelURI, "gpt://")
	idx := strings.Index(rest, "/")
	if idx <= 0 {
		return ""
	}
	return rest[:idx]
}

func MergeYandexChatModels(apiModels []string, folderID string) []string {
	merged := FilterYandexChatModels(apiModels)
	seen := make(map[string]struct{}, len(merged))
	for _, m := range merged {
		seen[m] = struct{}{}
	}
	for _, d := range KnownYandexChatModels(folderID) {
		if _, ok := seen[d]; ok {
			continue
		}
		seen[d] = struct{}{}
		merged = append(merged, d)
	}
	sortYandexChatModels(merged)
	return merged
}

func sortYandexChatModels(models []string) {
	sort.Slice(models, func(i, j int) bool {
		pi := yandexChatModelPriority(models[i])
		pj := yandexChatModelPriority(models[j])
		if pi != pj {
			return pi < pj
		}
		return models[i] < models[j]
	})
}

func yandexChatModelPriority(id string) int {
	lower := strings.ToLower(id)
	switch {
	case strings.Contains(lower, "yandexgpt-lite"):
		return 20
	case strings.Contains(lower, "yandexgpt"):
		return 10
	default:
		return 50
	}
}

// PickYandexChatModel chooses a chat completion model, never an embedding model.
func PickYandexChatModel(models []string, folderID, preferred string) string {
	chatModels := MergeYandexChatModels(models, folderID)
	if preferred != "" {
		uri := YandexModelURI(folderID, preferred)
		if containsString(chatModels, uri) {
			return uri
		}
		if containsString(chatModels, preferred) {
			return preferred
		}
	}
	if len(chatModels) > 0 {
		return chatModels[0]
	}
	return YandexModelURI(folderID, preferred)
}

func containsString(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
