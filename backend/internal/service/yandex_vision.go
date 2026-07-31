package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

const yandexVisionRecognizeURL = "https://ocr.api.cloud.yandex.net/ocr/v1/recognizeText"

var (
	ErrYandexVisionNotConfigured = errors.New("yandex vision ocr not configured")
	ErrYandexVisionEmptyText     = errors.New("yandex vision ocr returned empty text")
)

type yandexVisionRecognizeRequest struct {
	MimeType      string   `json:"mimeType"`
	LanguageCodes []string `json:"languageCodes"`
	Model         string   `json:"model"`
	Content       string   `json:"content"`
}

type yandexVisionRecognizeResponse struct {
	Result *struct {
		TextAnnotation *struct {
			FullText string `json:"fullText"`
		} `json:"textAnnotation"`
	} `json:"result"`
	Error *llmAPIError `json:"error"`
}

type YandexVisionRecognizeParams struct {
	APIKey   string
	FolderID string
	Image    []byte
	MIME     string
	Model    string
	Proxy    *model.LLMHTTPProxySettings
}

func NormalizeYandexOCRModel(model string) string {
	switch strings.TrimSpace(strings.ToLower(model)) {
	case "table", "page":
		return strings.TrimSpace(strings.ToLower(model))
	default:
		return "handwritten"
	}
}

func yandexVisionMimeType(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/png":
		return "PNG"
	case "image/gif":
		return "GIF"
	case "image/webp":
		return "WEBP"
	default:
		return "JPEG"
	}
}

func (s *LLMService) RecognizeYandexVisionText(ctx context.Context, params YandexVisionRecognizeParams) (string, error) {
	apiKey := strings.TrimSpace(params.APIKey)
	folderID := strings.TrimSpace(params.FolderID)
	if apiKey == "" || folderID == "" {
		return "", ErrYandexVisionNotConfigured
	}
	if len(params.Image) == 0 {
		return "", errors.New("image required for yandex vision ocr")
	}

	body, err := json.Marshal(yandexVisionRecognizeRequest{
		MimeType:      yandexVisionMimeType(params.MIME),
		LanguageCodes: []string{"ru", "en"},
		Model:         NormalizeYandexOCRModel(params.Model),
		Content:       base64.StdEncoding.EncodeToString(params.Image),
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, yandexVisionRecognizeURL, strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Api-Key "+apiKey)
	req.Header.Set("x-folder-id", folderID)
	req.Header.Set("User-Agent", llmUserAgent)

	resp, err := doHTTPWithProxy(ctx, s.client, params.Proxy, func(client *http.Client) (*http.Response, error) {
		return client.Do(req)
	})
	if err != nil {
		return "", fmt.Errorf("yandex vision ocr request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(raw))
		if msg == "" {
			msg = resp.Status
		}
		return "", fmt.Errorf("yandex vision ocr: %s", msg)
	}

	var parsed yandexVisionRecognizeResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("yandex vision ocr response parse failed: %w", err)
	}
	if parsed.Error != nil && strings.TrimSpace(parsed.Error.Message) != "" {
		return "", fmt.Errorf("yandex vision ocr: %s", parsed.Error.Message)
	}
	if parsed.Result == nil || parsed.Result.TextAnnotation == nil {
		return "", ErrYandexVisionEmptyText
	}
	fullText := strings.TrimSpace(parsed.Result.TextAnnotation.FullText)
	if fullText == "" {
		return "", ErrYandexVisionEmptyText
	}
	return fullText, nil
}

func geologicalJournalOCRUserPrompt(ocrText string) string {
	return strings.TrimSpace(`The following text was extracted from a geological field journal page via Yandex Vision OCR.
Structure it into the required JSON table. Correct obvious OCR mistakes using geological context when the meaning is clear.
If a value is unreadable, leave the field empty/null and add a short note to uncertainties.

--- OCR TEXT ---
` + ocrText + `
---`)
}
