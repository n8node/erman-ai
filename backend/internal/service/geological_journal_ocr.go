package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

const geologicalJournalStructuringChunkRows = 14

type geologicalJournalOCRResult struct {
	StructuredText string
	FullText       string
	Annotation     *VisionAnnotation
	Diagnostics    model.GeologicalJournalOCRDiagnostics
}

func (s *GeologicalJournalService) recognizeGeologicalJournalOCR(
	ctx context.Context,
	imageData []byte,
	mime string,
	imageWidth int,
	imageHeight int,
	layoutMode string,
	yandexKey string,
	yandexFolder string,
	ocrModel string,
	proxy *model.LLMHTTPProxySettings,
) (*geologicalJournalOCRResult, error) {
	normalizedLayout := normalizeGeologicalJournalLayoutMode(layoutMode)
	spreadSplit := shouldSplitGeologicalJournalSpread(normalizedLayout, imageWidth, imageHeight)
	visionParams := YandexVisionRecognizeParams{
		APIKey:   yandexKey,
		FolderID: yandexFolder,
		MIME:     mime,
		Model:    ocrModel,
		Proxy:    proxy,
	}

	var merged *VisionAnnotation
	if spreadSplit {
		src, err := decodeImage(imageData)
		if err != nil {
			return nil, fmt.Errorf("ocr image decode: %w", err)
		}
		leftImg, rightImg, splitX := splitImageVerticalHalves(src)
		leftPNG, err := encodeImagePNG(leftImg)
		if err != nil {
			return nil, fmt.Errorf("ocr left encode: %w", err)
		}
		rightPNG, err := encodeImagePNG(rightImg)
		if err != nil {
			return nil, fmt.Errorf("ocr right encode: %w", err)
		}
		leftAnn, err := s.llm.RecognizeYandexVisionAnnotation(ctx, YandexVisionRecognizeParams{
			APIKey: visionParams.APIKey, FolderID: visionParams.FolderID,
			Image: leftPNG, MIME: "image/png", Model: visionParams.Model, Proxy: visionParams.Proxy,
		})
		if err != nil {
			return nil, fmt.Errorf("ocr left: %w", err)
		}
		rightAnn, err := s.llm.RecognizeYandexVisionAnnotation(ctx, YandexVisionRecognizeParams{
			APIKey: visionParams.APIKey, FolderID: visionParams.FolderID,
			Image: rightPNG, MIME: "image/png", Model: visionParams.Model, Proxy: visionParams.Proxy,
		})
		if err != nil {
			return nil, fmt.Errorf("ocr right: %w", err)
		}
		rightWords := offsetVisionWords(rightAnn.Words, float64(splitX))
		rightAnn.Words = rightWords
		if rightAnn.Width > 0 {
			rightAnn.Width += splitX
		}
		merged = mergeVisionAnnotations(leftAnn, rightAnn)
		if merged.Width <= 0 {
			merged.Width = imageWidth
		}
		if merged.Height <= 0 {
			merged.Height = imageHeight
		}
	} else {
		visionParams.Image = imageData
		annotation, err := s.llm.RecognizeYandexVisionAnnotation(ctx, visionParams)
		if err != nil {
			return nil, err
		}
		merged = annotation
		if merged.Width <= 0 {
			merged.Width = imageWidth
		}
		if merged.Height <= 0 {
			merged.Height = imageHeight
		}
	}

	structuredText := buildSingleStructuredOCRText(merged)
	if spreadSplit {
		splitX := float64(merged.Width) / 2
		if splitX <= 0 {
			splitX = float64(imageWidth) / 2
		}
		structuredText = buildSpreadStructuredOCRText(merged, splitX)
	}
	if len(strings.TrimSpace(structuredText)) < 20 {
		structuredText = merged.FullText
	}

	diagnostics := model.GeologicalJournalOCRDiagnostics{
		CharCount:          len(structuredText),
		WordCount:          len(merged.Words),
		EstimatedDepthRows: countVisionDepthPairs(merged.Words),
		StructuredRowBands: countLogicalRecords(structuredText),
		GeometryUsed:       len(merged.Words) > 0,
		SpreadSplit:        spreadSplit,
		LayoutMode:         normalizedLayout,
		OCRPreview:         truncateVisionPreview(structuredText, 800),
	}
	if diagnostics.StructuredRowBands == 0 {
		diagnostics.StructuredRowBands = estimateGeologicalJournalRowCountForHint(structuredText)
	}
	if diagnostics.EstimatedDepthRows == 0 {
		diagnostics.EstimatedDepthRows = estimateGeologicalJournalRowCountForHint(structuredText)
	}

	return &geologicalJournalOCRResult{
		StructuredText: structuredText,
		FullText:       merged.FullText,
		Annotation:     merged,
		Diagnostics:    diagnostics,
	}, nil
}

type geologicalJournalStructureResult struct {
	Output              *model.GeologicalJournalOutput
	PromptTokens        int
	CompletionTokens    int
	TotalTokens         int
	Model               string
	StructuringChunks   int
}

func (s *GeologicalJournalService) structureGeologicalJournalOCR(
	ctx context.Context,
	structuredText string,
	spreadLayout bool,
	settings model.GeologicalJournalSettings,
	systemPrompt string,
	apiKey string,
	folderID string,
	proxy *model.LLMHTTPProxySettings,
) (*geologicalJournalStructureResult, error) {
	estimatedRows := countStructuredOCRRows(structuredText)
	chunks := splitStructuredOCRTextIntoChunks(structuredText, geologicalJournalStructuringChunkRows)
	if len(chunks) == 0 {
		chunks = []string{structuredText}
	}

	merged := &model.GeologicalJournalOutput{Rows: make([]model.GeologicalJournalRow, 0, estimatedRows)}
	result := &geologicalJournalStructureResult{StructuringChunks: len(chunks)}
	maxTokens := geologicalJournalStructuringMaxTokens(settings.MaxTokens)

	for i, chunk := range chunks {
		chunkRows := countStructuredOCRRows(chunk)
		userPrompt := geologicalJournalOCRUserPrompt(
			formatChunkStructuringPrompt(chunk, i, len(chunks), chunkRows),
			chunkRows,
			spreadLayout,
		)
		completion, err := s.llm.Complete(ctx, LLMCompletionRequest{
			Provider:     settings.Provider,
			Model:        settings.ActiveModel(),
			SystemPrompt: systemPrompt,
			UserPrompt:   userPrompt,
			Temperature:  settings.Temperature,
			MaxTokens:    maxTokens,
			APIKey:       apiKey,
			FolderID:     folderID,
			Proxy:        proxy,
		})
		if err != nil {
			return nil, err
		}
		output, err := ParseGeologicalJournalOutput(completion.Content)
		if err != nil {
			return nil, err
		}
		merged.Rows = append(merged.Rows, output.Rows...)
		result.PromptTokens += completion.PromptTokens
		result.CompletionTokens += completion.CompletionTokens
		result.TotalTokens += completion.TotalTokens
		if result.Model == "" {
			result.Model = completion.Model
		}
	}
	result.Output = merged
	return result, nil
}
