package service

import "testing"

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
	want := "gpt://b1g/folder/yandexgpt/latest"
	if got != want {
		t.Fatalf("pick = %q, want %q", got, want)
	}
}

func TestMergeYandexChatModelsAddsDefaults(t *testing.T) {
	got := MergeYandexChatModels([]string{"gpt://b1gx/custom-model/latest"}, "b1gx")
	if len(got) < 2 {
		t.Fatalf("expected merged defaults, got %v", got)
	}
}
