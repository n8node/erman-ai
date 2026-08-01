package service

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

var geologicalJournalNoisePattern = regexp.MustCompile(`(?i)^(стр\.?|столб|col\.?|column|\d{1,2})$`)

type logicalJournalRecord struct {
	YNorm      float64
	Left       []VisionWord
	Right      []VisionWord
	DepthFrom  *float64
	DepthTo    *float64
	Empty      bool
}

func parseVisionDecimal(text string) (float64, bool) {
	text = normalizeGeologicalJournalNumberString(text)
	if text == "" {
		return 0, false
	}
	value, err := strconv.ParseFloat(text, 64)
	if err != nil || value < 0 || value > 15000 {
		return 0, false
	}
	return value, true
}

func extractDepthPairFromWords(words []VisionWord) (*float64, *float64, bool) {
	var nums []float64
	for _, word := range words {
		value, ok := parseVisionDecimal(word.Text)
		if !ok {
			continue
		}
		nums = append(nums, value)
	}
	if len(nums) < 2 {
		return nil, nil, false
	}
	bestFrom, bestTo := -1.0, -1.0
	found := false
	for i := 0; i < len(nums); i++ {
		for j := i + 1; j < len(nums); j++ {
			from, to := nums[i], nums[j]
			if from > to {
				from, to = to, from
			}
			run := to - from
			if run <= 0.05 || run > 50 {
				continue
			}
			if !found || run < bestTo-bestFrom {
				bestFrom, bestTo = from, to
				found = true
			}
		}
	}
	if !found {
		return nil, nil, false
	}
	fromCopy, toCopy := bestFrom, bestTo
	return &fromCopy, &toCopy, true
}

func bandText(words []VisionWord) string {
	parts := make([]string, 0, len(words))
	for _, word := range words {
		text := strings.TrimSpace(word.Text)
		if text == "" || geologicalJournalNoisePattern.MatchString(text) {
			continue
		}
		parts = append(parts, text)
	}
	return strings.Join(parts, " ")
}

func isJournalNoiseBand(left, right string) bool {
	combined := strings.TrimSpace(left + " " + right)
	if combined == "" {
		return false
	}
	lower := strings.ToLower(combined)
	if geologicalJournalHeaderLinePattern.MatchString(lower) &&
		!geologicalJournalDepthNumberPattern.MatchString(combined) {
		return true
	}
	return isGeologicalJournalHeaderOnlyLine(lower)
}

func appendVisionWords(base []VisionWord, extra []VisionWord) []VisionWord {
	if len(extra) == 0 {
		return base
	}
	return append(append([]VisionWord(nil), base...), extra...)
}

func hasJournalContentBelow(pairs []spreadVisionRowPair, fromIndex int) bool {
	for i := fromIndex + 1; i < len(pairs); i++ {
		left := bandText(pairs[i].Left)
		right := bandText(pairs[i].Right)
		if left == "" && right == "" {
			continue
		}
		if isJournalNoiseBand(left, right) {
			continue
		}
		return true
	}
	return false
}

func buildLogicalRecordsFromSpreadPairs(pairs []spreadVisionRowPair) []logicalJournalRecord {
	records := make([]logicalJournalRecord, 0, len(pairs))
	for i, pair := range pairs {
		leftText := bandText(pair.Left)
		rightText := bandText(pair.Right)
		if isJournalNoiseBand(leftText, rightText) {
			continue
		}
		from, to, hasDepth := extractDepthPairFromWords(pair.Left)
		switch {
		case hasDepth:
			records = append(records, logicalJournalRecord{
				YNorm: pair.YNorm, Left: pair.Left, Right: pair.Right,
				DepthFrom: from, DepthTo: to,
			})
		case leftText == "" && rightText == "":
			if len(records) > 0 && hasJournalContentBelow(pairs, i) {
				records = append(records, logicalJournalRecord{
					YNorm: pair.YNorm, Empty: true,
				})
			}
		case rightText != "" && !hasDepth && leftText == "":
			if len(records) == 0 {
				continue
			}
			last := &records[len(records)-1]
			last.Right = appendVisionWords(last.Right, pair.Right)
		case leftText != "" && !hasDepth:
			if len(records) == 0 {
				continue
			}
			last := &records[len(records)-1]
			last.Left = appendVisionWords(last.Left, pair.Left)
			if rightText != "" {
				last.Right = appendVisionWords(last.Right, pair.Right)
			}
		}
	}
	return records
}

func buildLogicalRecordsFromSingleRows(rows [][]VisionWord, imageHeight int) []logicalJournalRecord {
	records := make([]logicalJournalRecord, 0, len(rows))
	if imageHeight <= 0 {
		imageHeight = 1
	}
	for i, row := range rows {
		text := bandText(row)
		if text == "" {
			if len(records) > 0 && i < len(rows)-1 {
				records = append(records, logicalJournalRecord{
					YNorm: visionRowBandYNorm(row, imageHeight), Empty: true,
				})
			}
			continue
		}
		if isJournalNoiseBand(text, "") {
			continue
		}
		from, to, hasDepth := extractDepthPairFromWords(row)
		if hasDepth {
			records = append(records, logicalJournalRecord{
				YNorm: visionRowBandYNorm(row, imageHeight), Left: row,
				DepthFrom: from, DepthTo: to,
			})
			continue
		}
		if len(records) == 0 {
			continue
		}
		last := &records[len(records)-1]
		last.Left = appendVisionWords(last.Left, row)
	}
	return records
}

func formatLogicalRecord(index int, record logicalJournalRecord, spread bool) string {
	if record.Empty {
		return fmt.Sprintf("--- RECORD %03d type=empty y=%.3f ---\n(empty printed table row)\n", index+1, record.YNorm)
	}
	left := formatVisionRowWords(record.Left)
	right := formatVisionRowWords(record.Right)
	depthHint := ""
	if record.DepthFrom != nil && record.DepthTo != nil {
		depthHint = fmt.Sprintf(" depth=%.1f-%.1f", *record.DepthFrom, *record.DepthTo)
	}
	if spread {
		return fmt.Sprintf("--- RECORD %03d type=data%s y=%.3f ---\nL: %s\nR: %s\n",
			index+1, depthHint, record.YNorm, left, right)
	}
	return fmt.Sprintf("--- RECORD %03d type=data%s y=%.3f ---\n%s\n",
		index+1, depthHint, record.YNorm, left)
}

func formatLogicalRecordsText(records []logicalJournalRecord, spread bool, wordCount int) string {
	var builder strings.Builder
	builder.WriteString("# Geological journal OCR — depth-anchored records\n")
	builder.WriteString(fmt.Sprintf("# OCR words: %d, logical records: %d\n", wordCount, len(records)))
	builder.WriteString("# One RECORD = one table row: depth interval, intentionally blank line, or merged description block.\n\n")
	for i, record := range records {
		builder.WriteString(formatLogicalRecord(i, record, spread))
		builder.WriteByte('\n')
	}
	return strings.TrimSpace(builder.String())
}

func visionRowGroupingTolerance(imageHeight int) float64 {
	tolerance := float64(imageHeight) * 0.028
	if tolerance < 14 {
		tolerance = 14
	}
	if tolerance > 36 {
		tolerance = 36
	}
	return tolerance
}

func groupVisionWordsIntoRows(words []VisionWord, imageHeight int) [][]VisionWord {
	if len(words) == 0 {
		return nil
	}
	tolerance := visionRowGroupingTolerance(imageHeight)

	sorted := append([]VisionWord(nil), words...)
	sort.Slice(sorted, func(i, j int) bool {
		cyI := visionWordCenterY(sorted[i])
		cyJ := visionWordCenterY(sorted[j])
		if math.Abs(cyI-cyJ) > tolerance/2 {
			return cyI < cyJ
		}
		return visionWordCenterX(sorted[i]) < visionWordCenterX(sorted[j])
	})

	var rows [][]VisionWord
	var current []VisionWord
	var currentY float64
	for idx, word := range sorted {
		cy := visionWordCenterY(word)
		if idx == 0 || math.Abs(cy-currentY) > tolerance {
			if len(current) > 0 {
				rows = append(rows, sortVisionWordsByX(current))
			}
			current = []VisionWord{word}
			currentY = cy
			continue
		}
		current = append(current, word)
	}
	if len(current) > 0 {
		rows = append(rows, sortVisionWordsByX(current))
	}
	return rows
}

func countLogicalRecords(text string) int {
	count := 0
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "--- RECORD ") {
			count++
		}
	}
	return count
}

func rowHasJournalData(row model.GeologicalJournalRow) bool {
	if row.Date != "" || row.RockDescription != "" || row.SamplingInterval != "" ||
		row.SampleNumber != "" || row.Notes != "" {
		return true
	}
	if row.DrillingDiameterMM != nil || row.DepthFromM != nil || row.DepthToM != nil ||
		row.DrillingRunM != nil || row.CoreRecoveryM != nil || row.CoreRecoveryPct != nil {
		return true
	}
	return false
}

func rowsShareDepth(a, b model.GeologicalJournalRow) bool {
	if a.DepthFromM == nil || a.DepthToM == nil || b.DepthFromM == nil || b.DepthToM == nil {
		return false
	}
	const eps = 0.05
	return math.Abs(*a.DepthFromM-*b.DepthFromM) < eps && math.Abs(*a.DepthToM-*b.DepthToM) < eps
}

func mergeRockDescription(base, extra string) string {
	base = strings.TrimSpace(base)
	extra = strings.TrimSpace(extra)
	if base == "" {
		return extra
	}
	if extra == "" {
		return base
	}
	if strings.Contains(base, extra) {
		return base
	}
	return base + " " + extra
}

// ConsolidateGeologicalJournalRows merges LLM over-segmentation: continuation rows,
// duplicate depth intervals, and trailing empty noise rows.
func ConsolidateGeologicalJournalRows(output *model.GeologicalJournalOutput) {
	if output == nil || len(output.Rows) == 0 {
		return
	}
	merged := make([]model.GeologicalJournalRow, 0, len(output.Rows))
	for _, row := range output.Rows {
		if !rowHasJournalData(row) {
			if len(merged) == 0 {
				continue
			}
			merged = append(merged, row)
			continue
		}
		if len(merged) == 0 {
			merged = append(merged, row)
			continue
		}
		lastIdx := len(merged) - 1
		last := merged[lastIdx]
		if row.DepthFromM == nil && row.DepthToM == nil &&
			(row.RockDescription != "" || row.Notes != "" || row.SamplingInterval != "" || row.SampleNumber != "") {
			last.RockDescription = mergeRockDescription(last.RockDescription, row.RockDescription)
			last.Notes = mergeRockDescription(last.Notes, row.Notes)
			if last.SamplingInterval == "" {
				last.SamplingInterval = row.SamplingInterval
			}
			if last.SampleNumber == "" {
				last.SampleNumber = row.SampleNumber
			}
			merged[lastIdx] = last
			continue
		}
		if rowsShareDepth(last, row) {
			last.RockDescription = mergeRockDescription(last.RockDescription, row.RockDescription)
			if last.Date == "" {
				last.Date = row.Date
			}
			if last.DrillingDiameterMM == nil {
				last.DrillingDiameterMM = row.DrillingDiameterMM
			}
			if last.DrillingRunM == nil {
				last.DrillingRunM = row.DrillingRunM
			}
			if last.CoreRecoveryM == nil {
				last.CoreRecoveryM = row.CoreRecoveryM
			}
			if last.CoreRecoveryPct == nil {
				last.CoreRecoveryPct = row.CoreRecoveryPct
			}
			merged[lastIdx] = last
			continue
		}
		merged = append(merged, row)
	}
	output.Rows = merged
}

func prunePhantomJournalRows(output *model.GeologicalJournalOutput, maxRows int) {
	if output == nil || maxRows <= 0 || len(output.Rows) <= maxRows {
		return
	}
	depthRows := 0
	for _, row := range output.Rows {
		if row.DepthFromM != nil && row.DepthToM != nil {
			depthRows++
		}
	}
	if depthRows == 0 {
		return
	}
	allowedEmpty := maxRows - depthRows
	if allowedEmpty < 0 {
		allowedEmpty = 0
	}
	kept := make([]model.GeologicalJournalRow, 0, maxRows)
	emptyKept := 0
	for _, row := range output.Rows {
		if row.DepthFromM != nil && row.DepthToM != nil {
			kept = append(kept, row)
			continue
		}
		if !rowHasJournalData(row) {
			if emptyKept < allowedEmpty {
				kept = append(kept, row)
				emptyKept++
			}
			continue
		}
		kept = append(kept, row)
	}
	if len(kept) < len(output.Rows) {
		output.Rows = kept
	}
}
