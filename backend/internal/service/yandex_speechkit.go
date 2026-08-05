package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	yandexSTTAsyncRecognizeURL = "https://stt.api.cloud.yandex.net/stt/v3/recognizeFileAsync"
	yandexSTTGetRecognitionURL = "https://stt.api.cloud.yandex.net/stt/v3/getRecognition"
	yandexSTTSyncRecognizeURL  = "https://stt.api.cloud.yandex.net/speech/v1/stt:recognize"
	yandexOperationsURL        = "https://operation.api.cloud.yandex.net/operations/"
)

var (
	ErrYandexSpeechKitNotConfigured = errors.New("yandex speechkit not configured")
	ErrYandexSpeechKitEmptyResult   = errors.New("yandex speechkit returned empty transcript")
)

type YandexSpeechKitParams struct {
	APIKey                   string
	FolderID                 string
	LanguageCode             string
	Model                    string
	TextNormalizationEnabled bool
	LiteratureText           bool
	ProfanityFilter          bool
}

type yandexSTTAsyncRequest struct {
	Content          string                    `json:"content"`
	RecognitionModel yandexSTTRecognitionModel `json:"recognitionModel"`
}

type yandexSTTRecognitionModel struct {
	Model               string                       `json:"model"`
	AudioFormat         yandexSTTAudioFormatOptions  `json:"audioFormat"`
	TextNormalization   *yandexSTTTextNormalization    `json:"textNormalization,omitempty"`
	LanguageRestriction yandexSTTLanguageRestriction `json:"languageRestriction"`
}

type yandexSTTTextNormalization struct {
	TextNormalization string `json:"textNormalization"`
	ProfanityFilter   bool   `json:"profanityFilter"`
	LiteratureText    bool   `json:"literatureText"`
}

type yandexSTTAudioFormatOptions struct {
	ContainerAudio *yandexSTTContainerAudio `json:"containerAudio,omitempty"`
}

type yandexSTTContainerAudio struct {
	ContainerAudioType string `json:"containerAudioType"`
}

type yandexSTTLanguageRestriction struct {
	RestrictionType string   `json:"restrictionType"`
	LanguageCode    []string `json:"languageCode"`
}

type yandexSTTAsyncResponse struct {
	ID   string `json:"id"`
	Done bool   `json:"done"`
}

type yandexSTTOperation struct {
	ID     string          `json:"id"`
	Done   bool            `json:"done"`
	Error  *llmAPIError    `json:"error"`
	Result json.RawMessage `json:"result"`
}

type yandexSTTSyncResponse struct {
	Result string       `json:"result"`
	Error  *llmAPIError `json:"error"`
}

func normalizeSpeechKitModel(model string) string {
	switch strings.TrimSpace(model) {
	case "", "general", "general:rc", "deferred-general", "deferred-general:rc":
		return strings.TrimSpace(model)
	default:
		return "general"
	}
}

func normalizeSpeechKitLanguage(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return "ru-RU"
	}
	return code
}

func speechKitContainerType(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "audio/wav", "audio/x-wav", "audio/wave":
		return "WAV"
	case "audio/mpeg", "audio/mp3":
		return "MP3"
	default:
		return "OGG_OPUS"
	}
}

func buildSpeechKitTextNormalization(params YandexSpeechKitParams) *yandexSTTTextNormalization {
	if !params.TextNormalizationEnabled && !params.LiteratureText && !params.ProfanityFilter {
		return nil
	}
	mode := "TEXT_NORMALIZATION_DISABLED"
	if params.TextNormalizationEnabled || params.LiteratureText {
		mode = "TEXT_NORMALIZATION_ENABLED"
	}
	return &yandexSTTTextNormalization{
		TextNormalization: mode,
		ProfanityFilter:   params.ProfanityFilter,
		LiteratureText:    params.LiteratureText,
	}
}

func (params YandexSpeechKitParams) PreferAsyncRecognition() bool {
	return params.TextNormalizationEnabled || params.LiteratureText
}

func (s *LLMService) TranscribeYandexSpeechKitAsync(ctx context.Context, params YandexSpeechKitParams, audio []byte, contentType string) (string, error) {
	apiKey := strings.TrimSpace(params.APIKey)
	folderID := strings.TrimSpace(params.FolderID)
	if apiKey == "" || folderID == "" {
		return "", ErrYandexSpeechKitNotConfigured
	}
	if len(audio) == 0 {
		return "", errors.New("audio required for speechkit")
	}

	model := normalizeSpeechKitModel(params.Model)
	lang := normalizeSpeechKitLanguage(params.LanguageCode)
	containerType := speechKitContainerType(contentType)

	body, err := json.Marshal(yandexSTTAsyncRequest{
		Content: base64.StdEncoding.EncodeToString(audio),
		RecognitionModel: yandexSTTRecognitionModel{
			Model: model,
			AudioFormat: yandexSTTAudioFormatOptions{
				ContainerAudio: &yandexSTTContainerAudio{ContainerAudioType: containerType},
			},
			TextNormalization: buildSpeechKitTextNormalization(params),
			LanguageRestriction: yandexSTTLanguageRestriction{
				RestrictionType: "WHITELIST",
				LanguageCode:    []string{lang},
			},
		},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, yandexSTTAsyncRecognizeURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Api-Key "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	s.applyLLMHeaders(req, llmHTTPHeaders{yandex: true, folderID: folderID})

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("speechkit async start failed: %s", strings.TrimSpace(string(raw)))
	}

	var started yandexSTTAsyncResponse
	if err := json.Unmarshal(raw, &started); err != nil {
		return "", fmt.Errorf("speechkit async response: %w", err)
	}
	if strings.TrimSpace(started.ID) == "" {
		return "", errors.New("speechkit async response missing operation id")
	}

	if err := s.waitYandexSTTOperation(ctx, apiKey, started.ID); err != nil {
		return "", err
	}
	preferNormalized := params.TextNormalizationEnabled || params.LiteratureText
	return s.fetchYandexSTTRecognition(ctx, apiKey, folderID, started.ID, preferNormalized)
}

func (s *LLMService) TranscribeYandexSpeechKitSync(ctx context.Context, params YandexSpeechKitParams, audio []byte, format string) (string, error) {
	apiKey := strings.TrimSpace(params.APIKey)
	folderID := strings.TrimSpace(params.FolderID)
	if apiKey == "" || folderID == "" {
		return "", ErrYandexSpeechKitNotConfigured
	}
	if len(audio) == 0 {
		return "", errors.New("audio required for speechkit")
	}
	if format == "" {
		format = "oggopus"
	}
	lang := normalizeSpeechKitLanguage(params.LanguageCode)

	url := fmt.Sprintf("%s?folderId=%s&lang=%s&format=%s",
		yandexSTTSyncRecognizeURL, folderID, lang, format)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(audio))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Api-Key "+apiKey)
	s.applyLLMHeaders(req, llmHTTPHeaders{yandex: true, folderID: folderID})

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("speechkit sync failed: %s", strings.TrimSpace(string(raw)))
	}

	var parsed yandexSTTSyncResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", err
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("speechkit sync error: %s", parsed.Error.Message)
	}
	text := strings.TrimSpace(parsed.Result)
	if text == "" {
		return "", ErrYandexSpeechKitEmptyResult
	}
	return text, nil
}

func (s *LLMService) waitYandexSTTOperation(ctx context.Context, apiKey, operationID string) error {
	deadline := time.Now().Add(45 * time.Minute)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if time.Now().After(deadline) {
			return errors.New("speechkit operation timed out")
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, yandexOperationsURL+operationID, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Api-Key "+apiKey)

		resp, err := s.client.Do(req)
		if err != nil {
			return err
		}
		raw, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}
		if resp.StatusCode >= 400 {
			return fmt.Errorf("speechkit operation poll failed: %s", strings.TrimSpace(string(raw)))
		}

		var op yandexSTTOperation
		if err := json.Unmarshal(raw, &op); err != nil {
			return err
		}
		if op.Error != nil {
			return fmt.Errorf("speechkit operation error: %s", op.Error.Message)
		}
		if op.Done {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
}

func (s *LLMService) fetchYandexSTTRecognition(ctx context.Context, apiKey, folderID, operationID string, preferNormalized bool) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, yandexSTTGetRecognitionURL+"?operation_id="+operationID, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Api-Key "+apiKey)
	s.applyLLMHeaders(req, llmHTTPHeaders{yandex: true, folderID: folderID})

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var normalizedParts []string
	var rawParts []string
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if text := extractYandexSTTFinalRefinementText(line); text != "" {
			normalizedParts = append(normalizedParts, text)
		}
		if text := extractYandexSTTFinalText(line); text != "" {
			rawParts = append(rawParts, text)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("speechkit getRecognition failed: status %d", resp.StatusCode)
	}

	parts := rawParts
	if preferNormalized && len(normalizedParts) > 0 {
		parts = normalizedParts
	}
	merged := strings.TrimSpace(strings.Join(parts, " "))
	if merged == "" {
		return "", ErrYandexSpeechKitEmptyResult
	}
	return merged, nil
}

func extractYandexSTTFinalRefinementText(line string) string {
	for _, obj := range yandexSTTLineObjects(line) {
		raw, ok := obj["finalRefinement"]
		if !ok {
			continue
		}
		var refinement map[string]json.RawMessage
		if err := json.Unmarshal(raw, &refinement); err != nil {
			continue
		}
		if normRaw, ok := refinement["normalizedText"]; ok {
			if text := extractYandexSTTAlternatives(normRaw); text != "" {
				return text
			}
		}
	}
	return ""
}

func extractYandexSTTFinalText(line string) string {
	for _, obj := range yandexSTTLineObjects(line) {
		if raw, ok := obj["final"]; ok {
			if text := extractYandexSTTAlternatives(raw); text != "" {
				return text
			}
		}
	}
	return ""
}

func yandexSTTLineObjects(line string) []map[string]json.RawMessage {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(line), &payload); err != nil {
		return nil
	}
	if raw, ok := payload["result"]; ok {
		var inner map[string]json.RawMessage
		if err := json.Unmarshal(raw, &inner); err == nil {
			return []map[string]json.RawMessage{inner}
		}
	}
	return []map[string]json.RawMessage{payload}
}

func extractYandexSTTLineText(line string) string {
	if text := extractYandexSTTFinalRefinementText(line); text != "" {
		return text
	}
	return extractYandexSTTFinalText(line)
}

func extractYandexSTTAlternatives(raw json.RawMessage) string {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return ""
	}
	altRaw, ok := obj["alternatives"]
	if !ok {
		return ""
	}
	var alts []struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(altRaw, &alts); err != nil {
		return ""
	}
	if len(alts) == 0 {
		return ""
	}
	return strings.TrimSpace(alts[0].Text)
}
