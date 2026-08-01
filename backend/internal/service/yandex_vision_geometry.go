package service

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

type yandexVisionVertex struct {
	X json.Number `json:"x"`
	Y json.Number `json:"y"`
}

type yandexVisionBoundingBox struct {
	Vertices []yandexVisionVertex `json:"vertices"`
}

type yandexVisionWord struct {
	Text        string                  `json:"text"`
	BoundingBox yandexVisionBoundingBox `json:"boundingBox"`
}

type yandexVisionLine struct {
	Text        string                  `json:"text"`
	BoundingBox yandexVisionBoundingBox `json:"boundingBox"`
	Words       []yandexVisionWord      `json:"words"`
}

type yandexVisionBlock struct {
	Lines []yandexVisionLine `json:"lines"`
}

type yandexVisionTextAnnotation struct {
	FullText string              `json:"fullText"`
	Width    json.Number         `json:"width"`
	Height   json.Number         `json:"height"`
	Blocks   []yandexVisionBlock `json:"blocks"`
}

type yandexVisionAnnotationResponse struct {
	Result *struct {
		TextAnnotation *yandexVisionTextAnnotation `json:"textAnnotation"`
	} `json:"result"`
	Error *llmAPIError `json:"error"`
}

// VisionWord is a recognized word with pixel bounding box.
type VisionWord struct {
	Text string
	XMin float64
	YMin float64
	XMax float64
	YMax float64
}

// VisionAnnotation holds OCR text plus word geometry.
type VisionAnnotation struct {
	FullText string
	Words    []VisionWord
	Width    int
	Height   int
}

func parseVisionNumber(raw json.Number) float64 {
	if raw == "" {
		return 0
	}
	if v, err := raw.Float64(); err == nil {
		return v
	}
	if i, err := raw.Int64(); err == nil {
		return float64(i)
	}
	return 0
}

func visionBoundsFromVertices(vertices []yandexVisionVertex) (xMin, yMin, xMax, yMax float64, ok bool) {
	if len(vertices) == 0 {
		return 0, 0, 0, 0, false
	}
	xMin, yMin = math.MaxFloat64, math.MaxFloat64
	xMax, yMax = -math.MaxFloat64, -math.MaxFloat64
	for _, vertex := range vertices {
		x := parseVisionNumber(vertex.X)
		y := parseVisionNumber(vertex.Y)
		if x < xMin {
			xMin = x
		}
		if x > xMax {
			xMax = x
		}
		if y < yMin {
			yMin = y
		}
		if y > yMax {
			yMax = y
		}
	}
	return xMin, yMin, xMax, yMax, xMax >= xMin && yMax >= yMin
}

func parseYandexVisionAnnotation(raw []byte) (*VisionAnnotation, error) {
	var parsed yandexVisionAnnotationResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("yandex vision ocr response parse failed: %w", err)
	}
	if parsed.Error != nil && strings.TrimSpace(parsed.Error.Message) != "" {
		return nil, fmt.Errorf("yandex vision ocr: %s", parsed.Error.Message)
	}
	if parsed.Result == nil || parsed.Result.TextAnnotation == nil {
		return nil, ErrYandexVisionEmptyText
	}
	annotation := parsed.Result.TextAnnotation
	fullText := strings.TrimSpace(annotation.FullText)
	if fullText == "" {
		return nil, ErrYandexVisionEmptyText
	}

	out := &VisionAnnotation{
		FullText: fullText,
		Words:    make([]VisionWord, 0, 256),
		Width:    int(parseVisionNumber(annotation.Width)),
		Height:   int(parseVisionNumber(annotation.Height)),
	}
	for _, block := range annotation.Blocks {
		for _, line := range block.Lines {
			if len(line.Words) > 0 {
				for _, word := range line.Words {
					text := strings.TrimSpace(word.Text)
					if text == "" {
						continue
					}
					xMin, yMin, xMax, yMax, ok := visionBoundsFromVertices(word.BoundingBox.Vertices)
					if !ok {
						xMin, yMin, xMax, yMax, ok = visionBoundsFromVertices(line.BoundingBox.Vertices)
						if !ok {
							continue
						}
					}
					out.Words = append(out.Words, VisionWord{
						Text: text,
						XMin: xMin,
						YMin: yMin,
						XMax: xMax,
						YMax: yMax,
					})
				}
				continue
			}
			text := strings.TrimSpace(line.Text)
			if text == "" {
				continue
			}
			xMin, yMin, xMax, yMax, ok := visionBoundsFromVertices(line.BoundingBox.Vertices)
			if !ok {
				continue
			}
			out.Words = append(out.Words, VisionWord{
				Text: text,
				XMin: xMin,
				YMin: yMin,
				XMax: xMax,
				YMax: yMax,
			})
		}
	}
	if out.Width <= 0 || out.Height <= 0 {
		out.Width, out.Height = inferVisionAnnotationSize(out.Words)
	}
	return out, nil
}

func inferVisionAnnotationSize(words []VisionWord) (width, height int) {
	maxX, maxY := 0.0, 0.0
	for _, word := range words {
		if word.XMax > maxX {
			maxX = word.XMax
		}
		if word.YMax > maxY {
			maxY = word.YMax
		}
	}
	return int(math.Ceil(maxX)), int(math.Ceil(maxY))
}

func visionWordCenterY(word VisionWord) float64 {
	return (word.YMin + word.YMax) / 2
}

func visionWordCenterX(word VisionWord) float64 {
	return (word.XMin + word.XMax) / 2
}

func sortVisionWordsByX(words []VisionWord) []VisionWord {
	out := append([]VisionWord(nil), words...)
	sort.Slice(out, func(i, j int) bool {
		return visionWordCenterX(out[i]) < visionWordCenterX(out[j])
	})
	return out
}

func offsetVisionWords(words []VisionWord, dx float64) []VisionWord {
	if dx == 0 {
		return words
	}
	out := make([]VisionWord, len(words))
	for i, word := range words {
		out[i] = VisionWord{
			Text: word.Text,
			XMin: word.XMin + dx,
			YMin: word.YMin,
			XMax: word.XMax + dx,
			YMax: word.YMax + dx,
		}
	}
	return out
}

func mergeVisionAnnotations(parts ...*VisionAnnotation) *VisionAnnotation {
	if len(parts) == 0 {
		return &VisionAnnotation{Words: []VisionWord{}}
	}
	merged := &VisionAnnotation{
		FullText: strings.TrimSpace(strings.Join(collectVisionFullTexts(parts), "\n")),
		Words:    make([]VisionWord, 0, 512),
	}
	for _, part := range parts {
		if part == nil {
			continue
		}
		if part.Width > merged.Width {
			merged.Width = part.Width
		}
		if part.Height > merged.Height {
			merged.Height = part.Height
		}
		merged.Words = append(merged.Words, part.Words...)
	}
	if merged.Width <= 0 || merged.Height <= 0 {
		merged.Width, merged.Height = inferVisionAnnotationSize(merged.Words)
	}
	return merged
}

func collectVisionFullTexts(parts []*VisionAnnotation) []string {
	texts := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == nil {
			continue
		}
		text := strings.TrimSpace(part.FullText)
		if text != "" {
			texts = append(texts, text)
		}
	}
	return texts
}

func countVisionDepthPairs(words []VisionWord) int {
	joined := strings.Join(collectVisionWordTexts(words), " ")
	matches := geologicalJournalPairedNumbersLine.FindAllString(joined, -1)
	return len(matches)
}

func collectVisionWordTexts(words []VisionWord) []string {
	out := make([]string, 0, len(words))
	for _, word := range words {
		out = append(out, word.Text)
	}
	return out
}

func truncateVisionPreview(text string, limit int) string {
	text = strings.TrimSpace(text)
	if limit <= 0 || len(text) <= limit {
		return text
	}
	return text[:limit] + "…"
}

func normalizeGeologicalJournalLayoutMode(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "spread", "two-page", "twopage":
		return "spread"
	case "single", "one-page", "onepage":
		return "single"
	default:
		return "auto"
	}
}

func shouldSplitGeologicalJournalSpread(layoutMode string, width, height int) bool {
	switch normalizeGeologicalJournalLayoutMode(layoutMode) {
	case "spread":
		return true
	case "single":
		return false
	default:
		if width <= 0 || height <= 0 {
			return false
		}
		return float64(width)/float64(height) >= 1.35
	}
}

func geologicalJournalStructuringMaxTokens(configured int) int {
	if configured < 8192 {
		return 8192
	}
	return configured
}

func estimateVisionRowBands(annotation *VisionAnnotation) int {
	if annotation == nil {
		return 0
	}
	imageHeight := annotation.Height
	if imageHeight <= 0 {
		imageHeight = 1
	}
	leftWords := make([]VisionWord, 0, len(annotation.Words)/2)
	rightWords := make([]VisionWord, 0, len(annotation.Words)/2)
	splitX := float64(annotation.Width) / 2
	if splitX <= 0 {
		splitX = 500
	}
	for _, word := range annotation.Words {
		if visionWordCenterX(word) < splitX {
			leftWords = append(leftWords, word)
		} else {
			rightWords = append(rightWords, word)
		}
	}
	pairs := pairSpreadVisionRows(
		groupVisionWordsIntoRows(leftWords, imageHeight),
		groupVisionWordsIntoRows(rightWords, imageHeight),
		imageHeight,
	)
	return len(buildLogicalRecordsFromSpreadPairs(pairs))
}

func formatVisionRowWords(words []VisionWord) string {
	parts := make([]string, 0, len(words))
	for _, word := range words {
		parts = append(parts, word.Text)
	}
	return strings.Join(parts, " | ")
}

func buildSpreadStructuredOCRText(annotation *VisionAnnotation, splitX float64) string {
	if annotation == nil {
		return ""
	}
	if len(annotation.Words) == 0 {
		return strings.TrimSpace(annotation.FullText)
	}
	imageHeight := annotation.Height
	if imageHeight <= 0 {
		imageHeight = 1
	}
	leftWords := make([]VisionWord, 0, len(annotation.Words)/2)
	rightWords := make([]VisionWord, 0, len(annotation.Words)/2)
	for _, word := range annotation.Words {
		cx := visionWordCenterX(word)
		if cx < splitX {
			leftWords = append(leftWords, word)
		} else {
			rightWords = append(rightWords, word)
		}
	}
	pairs := pairSpreadVisionRows(
		groupVisionWordsIntoRows(leftWords, imageHeight),
		groupVisionWordsIntoRows(rightWords, imageHeight),
		imageHeight,
	)
	records := buildLogicalRecordsFromSpreadPairs(pairs)
	if len(records) == 0 {
		return strings.TrimSpace(annotation.FullText)
	}
	return formatLogicalRecordsText(records, true, len(annotation.Words))
}

func buildSingleStructuredOCRText(annotation *VisionAnnotation) string {
	if annotation == nil {
		return ""
	}
	if len(annotation.Words) == 0 {
		return strings.TrimSpace(annotation.FullText)
	}
	imageHeight := annotation.Height
	if imageHeight <= 0 {
		imageHeight = 1
	}
	rows := groupVisionWordsIntoRows(annotation.Words, imageHeight)
	records := buildLogicalRecordsFromSingleRows(rows, imageHeight)
	if len(records) == 0 {
		return strings.TrimSpace(annotation.FullText)
	}
	return formatLogicalRecordsText(records, false, len(annotation.Words))
}

type spreadVisionRowPair struct {
	YNorm float64
	Left  []VisionWord
	Right []VisionWord
}

func visionRowBandYNorm(words []VisionWord, imageHeight int) float64 {
	if len(words) == 0 || imageHeight <= 0 {
		return 0
	}
	total := 0.0
	for _, word := range words {
		total += visionWordCenterY(word)
	}
	return (total / float64(len(words))) / float64(imageHeight)
}

func pairSpreadVisionRows(leftRows, rightRows [][]VisionWord, imageHeight int) []spreadVisionRowPair {
	type rowMeta struct {
		words []VisionWord
		yNorm float64
		used  bool
	}
	leftMeta := make([]rowMeta, len(leftRows))
	rightMeta := make([]rowMeta, len(rightRows))
	for i, row := range leftRows {
		leftMeta[i] = rowMeta{words: row, yNorm: visionRowBandYNorm(row, imageHeight)}
	}
	for i, row := range rightRows {
		rightMeta[i] = rowMeta{words: row, yNorm: visionRowBandYNorm(row, imageHeight)}
	}

	threshold := 0.018
	if imageHeight > 0 {
		threshold = 18 / float64(imageHeight)
	}
	if threshold < 0.012 {
		threshold = 0.012
	}
	if threshold > 0.03 {
		threshold = 0.03
	}

	var pairs []spreadVisionRowPair
	for _, left := range leftMeta {
		if len(left.words) == 0 {
			continue
		}
		matchIdx := -1
		bestDelta := threshold
		for j, right := range rightMeta {
			if right.used || len(right.words) == 0 {
				continue
			}
			delta := math.Abs(left.yNorm - right.yNorm)
			if delta <= bestDelta {
				bestDelta = delta
				matchIdx = j
			}
		}
		rightWords := []VisionWord(nil)
		if matchIdx >= 0 {
			rightMeta[matchIdx].used = true
			rightWords = rightMeta[matchIdx].words
		}
		pairs = append(pairs, spreadVisionRowPair{
			YNorm: left.yNorm,
			Left:  left.words,
			Right: rightWords,
		})
	}
	for _, right := range rightMeta {
		if right.used || len(right.words) == 0 {
			continue
		}
		pairs = append(pairs, spreadVisionRowPair{
			YNorm: right.yNorm,
			Right: right.words,
		})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].YNorm < pairs[j].YNorm
	})
	return pairs
}

func splitStructuredOCRTextIntoChunks(text string, maxRowsPerChunk int) []string {
	if maxRowsPerChunk <= 0 {
		return []string{text}
	}
	lines := strings.Split(text, "\n")
	var header []string
	var sections []string
	var currentSection strings.Builder
	rowCount := 0

	flushSection := func() {
		body := strings.TrimSpace(currentSection.String())
		if body != "" {
			sections = append(sections, body)
		}
		currentSection.Reset()
		rowCount = 0
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			header = append(header, line)
			continue
		}
		if strings.HasPrefix(line, "--- RECORD ") {
			if rowCount >= maxRowsPerChunk {
				flushSection()
			}
			rowCount++
		}
		if currentSection.Len() > 0 || strings.TrimSpace(line) != "" {
			currentSection.WriteString(line)
			currentSection.WriteByte('\n')
		}
	}
	flushSection()

	if len(sections) <= 1 {
		return []string{text}
	}
	prefix := strings.Join(header, "\n")
	chunks := make([]string, 0, len(sections))
	for _, section := range sections {
		if prefix == "" {
			chunks = append(chunks, section)
			continue
		}
		chunks = append(chunks, prefix+"\n\n"+section)
	}
	return chunks
}

func countStructuredOCRRows(text string) int {
	if count := countLogicalRecords(text); count > 0 {
		return count
	}
	return estimateGeologicalJournalRowCountForHint(text)
}

func formatChunkStructuringPrompt(chunk string, chunkIndex, chunkTotal int, estimatedRows int) string {
	header := fmt.Sprintf("Structure chunk %d of %d from a geological journal OCR table.\n", chunkIndex+1, chunkTotal)
	if estimatedRows > 0 {
		header += fmt.Sprintf("This chunk contains about %d RECORD markers.\n", estimatedRows)
	}
	header += "Return exactly one JSON row per RECORD marker, in the same order. Skip nothing inside this chunk.\n\n"
	return header + chunk
}
