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

const (
	maxPreprocessedImageBytes = 30 * 1024 * 1024
	maxPDFBytes               = 250 * 1024 * 1024
)

type JournalPreprocessResult struct {
	Image       []byte
	ContentType string
	Metadata    model.GeologicalJournalPreprocessingInfo
}

type JournalPDFInfo struct {
	PageCount int `json:"page_count"`
}

type JournalPDFPageAnalysis struct {
	Status                string          `json:"status"`
	Phase                 string          `json:"phase"`
	OrientationDegrees    int             `json:"orientation_degrees"`
	OrientationConfidence float64         `json:"orientation_confidence"`
	ContentType           string          `json:"content_type"`
	TextCharCount         int             `json:"text_char_count"`
	TableCount            int             `json:"table_count"`
	OCRText               string          `json:"ocr_text"`
	OCRTSV                string          `json:"ocr_tsv"`
	Analysis              json.RawMessage `json:"analysis"`
	OriginalAssetPath     string          `json:"original_asset_path"`
	OrientedAssetPath     string          `json:"oriented_asset_path"`
	PreprocessedAssetPath string          `json:"preprocessed_asset_path"`
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

func (p *JournalImagePreprocessor) PDFInfo(ctx context.Context, pdfPath string) (*JournalPDFInfo, error) {
	var out JournalPDFInfo
	if err := p.postJSON(ctx, "/pdf-info", map[string]any{"pdf_path": pdfPath}, &out); err != nil {
		return nil, err
	}
	if out.PageCount <= 0 {
		return nil, errors.New("pdf has no pages")
	}
	return &out, nil
}

func (p *JournalImagePreprocessor) PDFInfoBytes(ctx context.Context, pdf []byte) (*JournalPDFInfo, error) {
	var out JournalPDFInfo
	if err := p.postPDF(ctx, "/pdf-info", pdf, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (p *JournalImagePreprocessor) AnalyzePDFPage(ctx context.Context, pdfPath string, pageNumber int, outputDir string) (*JournalPDFPageAnalysis, error) {
	var out JournalPDFPageAnalysis
	err := p.postJSON(ctx, "/analyze-page", map[string]any{
		"pdf_path": pdfPath, "page_number": pageNumber, "output_dir": outputDir,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (p *JournalImagePreprocessor) AnalyzePDFPageBytes(ctx context.Context, pdf []byte, pageNumber int, outputDir string) (*JournalPDFPageAnalysis, error) {
	var out JournalPDFPageAnalysis
	if err := p.postPDF(ctx, "/analyze-page", pdf, map[string]string{
		"X-Page-Number": fmt.Sprintf("%d", pageNumber),
		"X-Output-Dir":  outputDir,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (p *JournalImagePreprocessor) postPDF(ctx context.Context, endpoint string, pdf []byte, headers map[string]string, result any) error {
	if !p.Enabled() {
		return errors.New("journal image preprocessor is disabled")
	}
	if len(pdf) == 0 {
		return errors.New("pdf is empty")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+endpoint, bytes.NewReader(pdf))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/pdf")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("preprocessor request failed: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxPDFBytes))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("preprocessor returned %s: %s", resp.Status, strings.TrimSpace(string(raw)))
	}
	if err := json.Unmarshal(raw, result); err != nil {
		return fmt.Errorf("preprocessor response parse failed: %w", err)
	}
	return nil
}

func (p *JournalImagePreprocessor) postJSON(ctx context.Context, endpoint string, payload any, result any) error {
	if !p.Enabled() {
		return errors.New("journal image preprocessor is disabled")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("preprocessor request failed: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxPreprocessedImageBytes))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("preprocessor returned %s: %s", resp.Status, strings.TrimSpace(string(raw)))
	}
	if err := json.Unmarshal(raw, result); err != nil {
		return fmt.Errorf("preprocessor response parse failed: %w", err)
	}
	return nil
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
