package service

import (
	"strings"
	"testing"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/stretchr/testify/require"
)

func TestEstimateGeologicalJournalRowCountForHint(t *testing.T) {
	ocr := `Дата Диам. Глуб. от Глуб. до Проходка
24.07 76 155.4 156.2 0.8
24.07 76 156.2 157.0 0.8
155.4
156.2
24.07 76 158.0 158.8 0.8`
	require.Equal(t, 3, estimateGeologicalJournalRowCountForHint(ocr))
}

func TestFillGeologicalJournalDepthGapsInsertsAtMostOne(t *testing.T) {
	from := 155.4
	to := 156.2
	nextFrom := 160.0
	nextTo := 160.8
	run := 0.8
	rows := []model.GeologicalJournalRow{
		{DepthFromM: &from, DepthToM: &to, DrillingRunM: &run, Uncertainties: []string{}},
		{DepthFromM: &nextFrom, DepthToM: &nextTo, DrillingRunM: &run, Uncertainties: []string{}},
	}
	out := fillGeologicalJournalDepthGaps(rows, 1)
	require.Len(t, out, 3)
}

func TestNormalizeGeologicalJournalRowsDoesNotPadToOCRLines(t *testing.T) {
	output := &model.GeologicalJournalOutput{
		Rows: make([]model.GeologicalJournalRow, 50),
	}
	for i := range output.Rows {
		output.Rows[i] = model.GeologicalJournalRow{Uncertainties: []string{}}
	}
	ocr := strings.Repeat("155.4\n", 200)
	NormalizeGeologicalJournalRows(output, ocr)
	require.LessOrEqual(t, len(output.Rows), 60)
}

func TestNormalizeGeologicalJournalRowsSmallGapNoInsert(t *testing.T) {
	from := 155.4
	to := 156.2
	nextFrom := 156.3
	nextTo := 157.0
	run := 0.8
	output := &model.GeologicalJournalOutput{
		Rows: []model.GeologicalJournalRow{
			{DepthFromM: &from, DepthToM: &to, DrillingRunM: &run, Uncertainties: []string{}},
			{DepthFromM: &nextFrom, DepthToM: &nextTo, DrillingRunM: &run, Uncertainties: []string{}},
		},
	}
	NormalizeGeologicalJournalRows(output, "")
	require.Len(t, output.Rows, 2)
}
