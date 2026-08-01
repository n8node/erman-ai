package service

import (
	"math"
	"regexp"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

var (
	geologicalJournalDepthNumberPattern = regexp.MustCompile(`\d+[.,]\d+`)
	geologicalJournalHeaderLinePattern  = regexp.MustCompile(`(?i)(дата|диам|глубин|проход|керн|описан|проб|интервал|date|depth|diameter|core|sample|recovery|run)`)
)

const (
	geologicalJournalDefaultRunM = 0.8
	geologicalJournalMaxGapRows  = 12
	geologicalJournalGapRunRatio = 1.35
)

func emptyGeologicalJournalRow() model.GeologicalJournalRow {
	return model.GeologicalJournalRow{Uncertainties: []string{}}
}

// estimateGeologicalJournalRowCount heuristically counts table body lines in OCR text.
func estimateGeologicalJournalRowCount(ocrText string) int {
	lines := strings.Split(ocrText, "\n")
	headerSeen := false
	count := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if headerSeen {
				count++
			}
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
		count++
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

func fillGeologicalJournalDepthGaps(rows []model.GeologicalJournalRow) []model.GeologicalJournalRow {
	if len(rows) < 2 {
		return rows
	}
	avgRun := medianGeologicalJournalRun(rows)
	out := make([]model.GeologicalJournalRow, 0, len(rows)+4)
	for i, row := range rows {
		out = append(out, row)
		if i >= len(rows)-1 {
			break
		}
		next := rows[i+1]
		if row.DepthToM == nil || next.DepthFromM == nil {
			continue
		}
		gap := *next.DepthFromM - *row.DepthToM
		if gap <= geologicalJournalDepthToleranceM {
			continue
		}
		missing := int(math.Round(gap / (avgRun * geologicalJournalGapRunRatio)))
		if missing < 1 {
			missing = 1
		}
		if missing > geologicalJournalMaxGapRows {
			missing = geologicalJournalMaxGapRows
		}
		for range missing {
			placeholder := emptyGeologicalJournalRow()
			placeholder.Uncertainties = []string{"row_gap:depth discontinuity"}
			out = append(out, placeholder)
		}
	}
	return out
}

func padGeologicalJournalRowsToEstimate(
	rows []model.GeologicalJournalRow,
	estimated int,
) []model.GeologicalJournalRow {
	if estimated <= len(rows) {
		return rows
	}
	out := append([]model.GeologicalJournalRow{}, rows...)
	for len(out) < estimated {
		placeholder := emptyGeologicalJournalRow()
		placeholder.Uncertainties = []string{"row_padding:missing_from_ocr"}
		out = append(out, placeholder)
	}
	return out
}

// NormalizeGeologicalJournalRows aligns row count with OCR heuristics and inserts
// placeholders for likely missing blank or skipped rows.
func NormalizeGeologicalJournalRows(output *model.GeologicalJournalOutput, ocrText string) {
	if output == nil {
		return
	}
	estimated := estimateGeologicalJournalRowCount(ocrText)
	rows := fillGeologicalJournalDepthGaps(output.Rows)
	rows = padGeologicalJournalRowsToEstimate(rows, estimated)
	output.Rows = rows
}
