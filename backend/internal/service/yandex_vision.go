package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
		return "", fmt.Errorf("yandex vision ocr: %s", formatYandexVisionAPIError(raw, resp.Status))
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

type yandexVisionAPIErrorBody struct {
	Error *struct {
		GRPCCode int    `json:"grpcCode"`
		HTTPCode int    `json:"httpCode"`
		Message  string `json:"message"`
	} `json:"error"`
}

func formatYandexVisionAPIError(raw []byte, status string) string {
	msg := strings.TrimSpace(string(raw))
	if msg == "" {
		return status
	}
	var parsed yandexVisionAPIErrorBody
	if err := json.Unmarshal(raw, &parsed); err == nil && parsed.Error != nil {
		if m := strings.TrimSpace(parsed.Error.Message); m != "" {
			if strings.Contains(m, "does not match with service account folder ID") {
				return m + " — проверьте Folder ID в AI Strategy LLM (должен быть b1g..., не email)"
			}
			if strings.Contains(strings.ToLower(m), "permission") || strings.Contains(strings.ToLower(m), "access") {
				return m + " — сервисному аккаунту нужна роль ai.vision.user"
			}
			return m
		}
	}
	return msg
}

func geologicalJournalOCRUserPrompt(ocrText string, estimatedRows int) string {
	countHint := ""
	if estimatedRows >= 3 && estimatedRows <= 120 {
		countHint = "OCR rough hint: about " + strconv.Itoa(estimatedRows) +
			" lines contain paired numeric values. " +
			"Many OCR lines are fragments of the same table row — do not emit one JSON row per OCR line.\n"
	}
	return strings.TrimSpace(`Structure the geological journal table from this OCR text.
Return one JSON object {"rows":[...]} with exactly one object per physical table row on the page.
Include empty rows: blank cells become null (numbers) or "" (text).
Preserve top-to-bottom order. Do not skip rows without depth values.
Do not inflate row count: merge OCR fragments that belong to the same handwritten line.
` + countHint + `
--- OCR TEXT START ---
` + ocrText + `
--- OCR TEXT END ---`)
}
