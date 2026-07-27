package service

import (
	"testing"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/model"
)

func TestFilterYandexChatModels(t *testing.T) {
	models := []string{
		"emb://b1g/folder/text-embeddings-v2-doc/latest",
		"gpt://b1g/folder/yandexgpt-lite/latest",
		"gpt://b1g/folder/yandexgpt/latest",
		"ds://b1g/folder/something",
	}
	got := FilterYandexChatModels(models)
	if len(got) != 2 {
		t.Fatalf("expected 2 chat models, got %d: %v", len(got), got)
	}
}

func TestPickYandexChatModelSkipsEmbedding(t *testing.T) {
	models := []string{
		"emb://b1g/folder/text-embeddings-v2-doc/latest",
		"gpt://b1g/folder/yandexgpt-lite/latest",
		"gpt://b1g/folder/yandexgpt/latest",
	}
	got := PickYandexChatModel(models, "b1g/folder", "")
	chatModels := MergeYandexChatModels(models, "b1g/folder")
	if got != chatModels[0] {
		t.Fatalf("pick = %q, want first sorted chat model %q", got, chatModels[0])
	}
	if !IsYandexChatModel(got) {
		t.Fatalf("pick must be chat model, got %q", got)
	}
}

func TestResolveModelPricingUsesModelSpecificRates(t *testing.T) {
	modelPricing := map[string]model.LLMProviderPricing{
		"gpt://b1g/folder/yandexgpt-5-lite/latest": {
			InputPer1K:  0.2,
			OutputPer1K: 0.4,
			Currency:    "RUB",
		},
	}
	got := ResolveModelPricing(
		model.LLMProviderYandex,
		"gpt://b1g/folder/yandexgpt-5-lite/latest",
		modelPricing,
		nil,
		&config.Config{YandexPriceInputRUBPer1K: 0.6, YandexPriceOutputRUBPer1K: 1.8},
	)
	if got.InputPer1K != 0.2 || got.OutputPer1K != 0.4 {
		t.Fatalf("unexpected pricing: %+v", got)
	}
}

func TestKnownYandexChatModelsIncludesCatalog(t *testing.T) {
	got := KnownYandexChatModels("b1gtest")
	if len(got) < 8 {
		t.Fatalf("expected full yandex catalog, got %d: %v", len(got), got)
	}
}
