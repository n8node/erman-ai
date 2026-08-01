package service

import (
	"math"
	"regexp"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

var (
	geologicalJournalDepthNumberPattern = regexp.MustCompile(`\d+[.,]\d+`)
	geologicalJournalPairedNumbersLine  = regexp.MustCompile(`\d+[.,]\d+.*\d+[.,]\d+`)
	geologicalJournalHeaderLinePattern  = regexp.MustCompile(`(?i)(дата|диам|глубин|проход|керн|описан|проб|интервал|date|depth|diameter|core|sample|recovery|run)`)
)

const (
	geologicalJournalDefaultRunM   = 0.8
	geologicalJournalMaxGapInserts = 1
	geologicalJournalMaxExtraRows  = 10
	geologicalJournalGapMinFactor  = 2.0
)

func emptyGeologicalJournalRow() model.GeologicalJournalRow {
	return model.GeologicalJournalRow{Uncertainties: []string{}}
}

// estimateGeologicalJournalRowCountForHint counts OCR lines that likely represent
// one table row (paired numeric values on the same line). It intentionally ignores
// blank lines and single-number fragments produced by column-wise OCR.
func estimateGeologicalJournalRowCountForHint(ocrText string) int {
	headerSeen := false
	count := 0
	for _, line := range strings.Split(ocrText, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		lower := strings.ToLower(trimmed)
		if !headerSeen && geologicalJournalHeaderLinePattern.MatchString(lower) &&
			!geologicalJournalDepthNumberPattern.MatchString(trimmed) {
			headerSeen = true
			continue
		}
		if !headerSeen {
			headerSeen = true
		}
		if isGeologicalJournalHeaderOnlyLine(lower) {
			continue
		}
		if geologicalJournalPairedNumbersLine.MatchString(trimmed) {
			count++
		}
	}
	return count
}

func isGeologicalJournalHeaderOnlyLine(lower string) bool {
	if geologicalJournalDepthNumberPattern.MatchString(lower) {
		return false
	}
	matches := 0
	for _, keyword := range []string{"дата", "диам", "глуб", "проход", "керн", "опис", "проб", "date", "depth", "diam", "core", "sample"} {
		if strings.Contains(lower, keyword) {
			matches++
		}
	}
	return matches >= 3
}

func medianGeologicalJournalRun(rows []model.GeologicalJournalRow) float64 {
	var runs []float64
	for _, row := range rows {
		if row.DepthFromM != nil && row.DepthToM != nil {
			run := *row.DepthToM - *row.DepthFromM
			if run > 0.05 && run < 50 {
				runs = append(runs, run)
			}
		}
		if row.DrillingRunM != nil && *row.DrillingRunM > 0.05 && *row.DrillingRunM < 50 {
			runs = append(runs, *row.DrillingRunM)
		}
	}
	if len(runs) == 0 {
		return geologicalJournalDefaultRunM
	}
	total := 0.0
	for _, value := range runs {
		total += value
	}
	return total / float64(len(runs))
}

func fillGeologicalJournalDepthGaps(
	rows []model.GeologicalJournalRow,
	insertBudget int,
) []model.GeologicalJournalRow {
	if len(rows) < 2 || insertBudget <= 0 {
		return rows
	}
	avgRun := medianGeologicalJournalRun(rows)
	minGap := math.Max(geologicalJournalDepthToleranceM*2, avgRun*geologicalJournalGapMinFactor)
	out := make([]model.GeologicalJournalRow, 0, len(rows)+insertBudget)
	inserted := 0

	for i, row := range rows {
		out = append(out, row)
		if i >= len(rows)-1 || inserted >= insertBudget {
			continue
		}
		next := rows[i+1]
		if row.DepthToM == nil || next.DepthFromM == nil {
			continue
		}
		gap := *next.DepthFromM - *row.DepthToM
		if gap < minGap {
			continue
		}
		for range geologicalJournalMaxGapInserts {
			if inserted >= insertBudget {
				break
			}
			placeholder := emptyGeologicalJournalRow()
			placeholder.Uncertainties = []string{"row_gap:depth discontinuity"}
			out = append(out, placeholder)
			inserted++
		}
	}
	return out
}

// NormalizeGeologicalJournalRows inserts at most a few placeholder rows for large
// depth discontinuities. It never pads to OCR line count.
func NormalizeGeologicalJournalRows(output *model.GeologicalJournalOutput, _ string) {
	if output == nil || len(output.Rows) == 0 {
		return
	}
	original := len(output.Rows)
	extraCap := original / 10
	if extraCap < 1 {
		extraCap = 1
	}
	if extraCap > geologicalJournalMaxExtraRows {
		extraCap = geologicalJournalMaxExtraRows
	}
	output.Rows = fillGeologicalJournalDepthGaps(output.Rows, extraCap)
}
