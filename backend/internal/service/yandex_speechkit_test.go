package service

import "testing"

func TestExtractYandexSTTFinalRefinementText(t *testing.T) {
	line := `{"finalRefinement":{"finalIndex":"0","normalizedText":{"alternatives":[{"text":"Я яндекс спичкит."}]}}}`
	got := extractYandexSTTFinalRefinementText(line)
	if got != "Я яндекс спичкит." {
		t.Fatalf("unexpected normalized text %q", got)
	}
}

func TestExtractYandexSTTFinalText(t *testing.T) {
	line := `{"final":{"alternatives":[{"text":"я яндекс спичкит"}]}}`
	got := extractYandexSTTFinalText(line)
	if got != "я яндекс спичкит" {
		t.Fatalf("unexpected final text %q", got)
	}
}

func TestExtractYandexSTTLinePrefersNormalizedWhenPresent(t *testing.T) {
	line := `{"final":{"alternatives":[{"text":"raw text"}]},"finalRefinement":{"normalizedText":{"alternatives":[{"text":"Normalized text."}]}}}`
	if got := extractYandexSTTFinalRefinementText(line); got != "Normalized text." {
		t.Fatalf("expected normalized text, got %q", got)
	}
	if got := extractYandexSTTFinalText(line); got != "raw text" {
		t.Fatalf("expected raw final text, got %q", got)
	}
}

func TestExtractYandexSTTLineObjectsWithResultEnvelope(t *testing.T) {
	line := `{"result":{"finalRefinement":{"normalizedText":{"alternatives":[{"text":"Hello."}]}}}}`
	got := extractYandexSTTFinalRefinementText(line)
	if got != "Hello." {
		t.Fatalf("unexpected text from result envelope %q", got)
	}
}
