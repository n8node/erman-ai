package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeYandexOCRModel(t *testing.T) {
	require.Equal(t, "handwritten", NormalizeYandexOCRModel(""))
	require.Equal(t, "handwritten", NormalizeYandexOCRModel("unknown"))
	require.Equal(t, "table", NormalizeYandexOCRModel("table"))
	require.Equal(t, "page", NormalizeYandexOCRModel("PAGE"))
}

func TestYandexVisionMimeType(t *testing.T) {
	require.Equal(t, "JPEG", yandexVisionMimeType("image/jpeg"))
	require.Equal(t, "PNG", yandexVisionMimeType("image/png"))
	require.Equal(t, "JPEG", yandexVisionMimeType("application/octet-stream"))
}

func TestParseYandexVisionRecognizeResponse(t *testing.T) {
	raw := `{"result":{"textAnnotation":{"fullText":"depth 1-2 m"}}}`
	var parsed yandexVisionRecognizeResponse
	require.NoError(t, json.Unmarshal([]byte(raw), &parsed))
	require.NotNil(t, parsed.Result)
	require.Equal(t, "depth 1-2 m", parsed.Result.TextAnnotation.FullText)
}

func TestGeologicalJournalOCRUserPrompt(t *testing.T) {
	prompt := geologicalJournalOCRUserPrompt("sample text", 5, true)
	require.Contains(t, prompt, "sample text")
	require.Contains(t, prompt, "OCR TEXT")
	require.Contains(t, prompt, "Expected logical rows")
	require.Contains(t, prompt, "5")
	require.Contains(t, prompt, "two-page spread")
}
