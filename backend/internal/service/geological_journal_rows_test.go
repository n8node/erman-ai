package service

import (
	"testing"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/stretchr/testify/require"
)

func TestEstimateGeologicalJournalRowCount(t *testing.T) {
	ocr := `Дата Диам. Глуб. от Глуб. до Проходка
24.07 76 155.4 156.2 0.8
24.07 76 156.2 157.0 0.8

24.07 76 158.0 158.8 0.8`
	require.Equal(t, 4, estimateGeologicalJournalRowCount(ocr))
}

func TestFillGeologicalJournalDepthGaps(t *testing.T) {
	from := 155.4
	to := 156.2
	nextFrom := 158.0
	nextTo := 158.8
	run := 0.8
	rows := []model.GeologicalJournalRow{
		{
			DepthFromM:    &from,
			DepthToM:      &to,
			DrillingRunM:  &run,
			Uncertainties: []string{},
		},
		{
			DepthFromM:    &nextFrom,
			DepthToM:      &nextTo,
			DrillingRunM:  &run,
			Uncertainties: []string{},
		},
	}
	out := fillGeologicalJournalDepthGaps(rows)
	require.Greater(t, len(out), len(rows))
}

func TestPadGeologicalJournalRowsToEstimate(t *testing.T) {
	rows := []model.GeologicalJournalRow{{Uncertainties: []string{}}}
	out := padGeologicalJournalRowsToEstimate(rows, 3)
	require.Len(t, out, 3)
	require.Contains(t, out[2].Uncertainties, "row_padding:missing_from_ocr")
}

func TestNormalizeGeologicalJournalRows(t *testing.T) {
	ocr := `Дата Диам. Глуб. от Глуб. до
24.07 76 155.4 156.2
24.07 76 156.2 157.0
24.07 76 158.0 158.8`
	output := &model.GeologicalJournalOutput{
		Rows: []model.GeologicalJournalRow{
			{Uncertainties: []string{}},
		},
	}
	NormalizeGeologicalJournalRows(output, ocr)
	require.GreaterOrEqual(t, len(output.Rows), 3)
}
