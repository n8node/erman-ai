package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/erman-ai/erman-ai/internal/model"
)

const maxPreprocessedImageBytes = 30 * 1024 * 1024

type JournalPreprocessResult struct {
	Image       []byte
	ContentType string
	Metadata    model.GeologicalJournalPreprocessingInfo
}

type JournalImagePreprocessor struct {
	baseURL string
	client  *http.Client
}

func NewJournalImagePreprocessor(baseURL string) *JournalImagePreprocessor {
	normalized := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if parsed, err := url.Parse(normalized); err != nil ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") ||
		parsed.Host == "" {
		normalized = ""
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	return &JournalImagePreprocessor{
		baseURL: normalized,
		client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}

func (p *JournalImagePreprocessor) Enabled() bool {
	return p != nil && p.baseURL != ""
}

func (p *JournalImagePreprocessor) Preprocess(
	ctx context.Context,
	image []byte,
	contentType string,
) (*JournalPreprocessResult, error) {
	if !p.Enabled() {
		return nil, errors.New("journal image preprocessor is disabled")
	}
	if len(image) == 0 {
		return nil, errors.New("image required")
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		p.baseURL+"/preprocess",
		bytes.NewReader(image),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(image)))

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("preprocessor request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxPreprocessedImageBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read preprocessor response: %w", err)
	}
	if len(body) > maxPreprocessedImageBytes {
		return nil, errors.New("preprocessor response is too large")
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		message := strings.TrimSpace(string(body))
		if len(message) > 300 {
			message = message[:300]
		}
		return nil, fmt.Errorf("preprocessor returned %s: %s", resp.Status, message)
	}
	if !strings.HasPrefix(strings.ToLower(resp.Header.Get("Content-Type")), "image/") {
		return nil, errors.New("preprocessor returned non-image content")
	}
	if len(body) == 0 {
		return nil, errors.New("preprocessor returned an empty image")
	}

	metadata := parsePreprocessorMetadata(resp.Header.Get("X-Preprocessing-Metadata"))
	metadata.Applied = true
	metadata.UsedForOCR = true
	metadata.HasPreprocessedImage = true

	return &JournalPreprocessResult{
		Image:       body,
		ContentType: "image/png",
		Metadata:    metadata,
	}, nil
}

func parsePreprocessorMetadata(raw string) model.GeologicalJournalPreprocessingInfo {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return model.GeologicalJournalPreprocessingInfo{}
	}
	var metadata model.GeologicalJournalPreprocessingInfo
	_ = json.Unmarshal([]byte(raw), &metadata)
	return metadata
}
