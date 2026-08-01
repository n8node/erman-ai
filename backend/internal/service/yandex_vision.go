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
		TextAnnotation *yandexVisionTextAnnotation `json:"textAnnotation"`
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

func (s *LLMService) recognizeYandexVisionRaw(ctx context.Context, params YandexVisionRecognizeParams) ([]byte, error) {
	apiKey := strings.TrimSpace(params.APIKey)
	folderID := strings.TrimSpace(params.FolderID)
	if apiKey == "" || folderID == "" {
		return nil, ErrYandexVisionNotConfigured
	}
	if len(params.Image) == 0 {
		return nil, errors.New("image required for yandex vision ocr")
	}

	body, err := json.Marshal(yandexVisionRecognizeRequest{
		MimeType:      yandexVisionMimeType(params.MIME),
		LanguageCodes: []string{"ru", "en"},
		Model:         NormalizeYandexOCRModel(params.Model),
		Content:       base64.StdEncoding.EncodeToString(params.Image),
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, yandexVisionRecognizeURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Api-Key "+apiKey)
	req.Header.Set("x-folder-id", folderID)
	req.Header.Set("User-Agent", llmUserAgent)

	resp, err := doHTTPWithProxy(ctx, s.client, params.Proxy, func(client *http.Client) (*http.Response, error) {
		return client.Do(req)
	})
	if err != nil {
		return nil, fmt.Errorf("yandex vision ocr request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("yandex vision ocr: %s", formatYandexVisionAPIError(raw, resp.Status))
	}
	return raw, nil
}

func (s *LLMService) RecognizeYandexVisionText(ctx context.Context, params YandexVisionRecognizeParams) (string, error) {
	raw, err := s.recognizeYandexVisionRaw(ctx, params)
	if err != nil {
		return "", err
	}
	annotation, err := parseYandexVisionAnnotation(raw)
	if err != nil {
		return "", err
	}
	return annotation.FullText, nil
}

func (s *LLMService) RecognizeYandexVisionAnnotation(ctx context.Context, params YandexVisionRecognizeParams) (*VisionAnnotation, error) {
	raw, err := s.recognizeYandexVisionRaw(ctx, params)
	if err != nil {
		return nil, err
	}
	return parseYandexVisionAnnotation(raw)
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

func geologicalJournalOCRUserPrompt(ocrText string, estimatedRows int, spreadLayout bool) string {
	countHint := ""
	if estimatedRows >= 3 && estimatedRows <= 120 {
		countHint = "Expected logical rows (depth intervals): about " + strconv.Itoa(estimatedRows) + ".\n"
	}
	layoutHint := ""
	if spreadLayout || strings.Contains(ocrText, "spread geometry") {
		layoutHint = `This is a two-page spread journal (columns 1-10 on LEFT, 11-15 on RIGHT).
Each --- ROW band is one logical record. Anchor rows on depth_from_m / depth_to_m when present.
Rock description in RIGHT may span multiple printed grid lines — keep it in one JSON row per ROW band.
Date may appear only once per day block — copy forward only when clearly the same drilling day.
Skip completely empty ROW bands (no numbers and no text on both sides).
`
	}
	return strings.TrimSpace(`Structure the geological journal table from this OCR text.
Return one JSON object {"rows":[...]} with one object per logical ROW band in order.
A logical row is defined by a depth interval (depth_from_m, depth_to_m) when visible; otherwise one ROW band.
Include sparse rows: blank cells become null (numbers) or "" (text).
Preserve top-to-bottom order. Do not skip ROW bands that contain depth values or rock descriptions.
Do not merge two ROW bands into one JSON row. Do not emit one JSON row per isolated OCR word.
` + layoutHint + countHint + `
--- OCR TEXT START ---
` + ocrText + `
--- OCR TEXT END ---`)
}
