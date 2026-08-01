package service

import (
	"testing"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/stretchr/testify/require"
)

func TestBuildLogicalRecordsFromSpreadPairs(t *testing.T) {
	pairs := []spreadVisionRowPair{
		{YNorm: 0.1, Left: []VisionWord{{Text: "49.0"}, {Text: "52.2"}}, Right: []VisionWord{{Text: "кварциты"}}},
		{YNorm: 0.11, Left: nil, Right: []VisionWord{{Text: "железистые"}}},
		{YNorm: 0.12, Left: nil, Right: nil},
		{YNorm: 0.2, Left: []VisionWord{{Text: "52.2"}, {Text: "55.0"}}, Right: []VisionWord{{Text: "сланцы"}}},
	}
	records := buildLogicalRecordsFromSpreadPairs(pairs)
	require.Len(t, records, 3)
	require.NotNil(t, records[0].DepthFrom)
	require.InDelta(t, 49.0, *records[0].DepthFrom, 0.001)
	require.Contains(t, bandText(records[0].Right), "кварциты")
	require.Contains(t, bandText(records[0].Right), "железистые")
	require.True(t, records[1].Empty)
	require.NotNil(t, records[2].DepthFrom)
}

func TestConsolidateGeologicalJournalRows(t *testing.T) {
	output := &model.GeologicalJournalOutput{
		Rows: []model.GeologicalJournalRow{
			{DepthFromM: floatPtr(49), DepthToM: floatPtr(52.2), RockDescription: "кварциты", Uncertainties: []string{}},
			{RockDescription: "железистые", Uncertainties: []string{}},
			{DepthFromM: floatPtr(52.2), DepthToM: floatPtr(55), RockDescription: "сланцы", Uncertainties: []string{}},
			{DepthFromM: floatPtr(52.2), DepthToM: floatPtr(55), RockDescription: "дубликат", Uncertainties: []string{}},
		},
	}
	ConsolidateGeologicalJournalRows(output)
	require.Len(t, output.Rows, 2)
	require.Contains(t, output.Rows[0].RockDescription, "железистые")
	require.Contains(t, output.Rows[1].RockDescription, "дубликат")
}

func TestBuildSpreadStructuredOCRTextUsesRecords(t *testing.T) {
	annotation := &VisionAnnotation{
		Width:  2000,
		Height: 1000,
		Words: []VisionWord{
			{Text: "24.08.83", XMin: 50, YMin: 100, XMax: 120, YMax: 115},
			{Text: "49.0", XMin: 300, YMin: 100, XMax: 340, YMax: 115},
			{Text: "52.2", XMin: 360, YMin: 100, XMax: 400, YMax: 115},
			{Text: "кварциты", XMin: 1200, YMin: 101, XMax: 1300, YMax: 118},
		},
	}
	text := buildSpreadStructuredOCRText(annotation, 1000)
	require.Contains(t, text, "depth-anchored records")
	require.Contains(t, text, "RECORD 001")
	require.Contains(t, text, "49.0")
	require.Contains(t, text, "кварциты")
}
