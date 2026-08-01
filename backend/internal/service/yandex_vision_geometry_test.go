package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeGeologicalJournalLayoutMode(t *testing.T) {
	require.Equal(t, "spread", normalizeGeologicalJournalLayoutMode("two-page"))
	require.Equal(t, "single", normalizeGeologicalJournalLayoutMode("one-page"))
	require.Equal(t, "auto", normalizeGeologicalJournalLayoutMode(""))
}

func TestShouldSplitGeologicalJournalSpread(t *testing.T) {
	require.True(t, shouldSplitGeologicalJournalSpread("spread", 1000, 600))
	require.False(t, shouldSplitGeologicalJournalSpread("single", 2000, 1000))
	require.True(t, shouldSplitGeologicalJournalSpread("auto", 2000, 1000))
	require.False(t, shouldSplitGeologicalJournalSpread("auto", 1000, 1000))
}

func TestGroupVisionWordsIntoRows(t *testing.T) {
	words := []VisionWord{
		{Text: "49.0", XMin: 100, YMin: 100, XMax: 130, YMax: 115},
		{Text: "52.2", XMin: 150, YMin: 102, XMax: 180, YMax: 117},
		{Text: "кварциты", XMin: 900, YMin: 101, XMax: 980, YMax: 118},
		{Text: "55.0", XMin: 100, YMin: 200, XMax: 130, YMax: 215},
	}
	rows := groupVisionWordsIntoRows(words, 1000)
	require.Len(t, rows, 2)
	require.Equal(t, "49.0", rows[0][0].Text)
	require.Equal(t, "52.2", rows[0][1].Text)
	require.Equal(t, "кварциты", rows[0][2].Text)
}

func TestBuildSpreadStructuredOCRText(t *testing.T) {
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

func TestSplitStructuredOCRTextIntoChunks(t *testing.T) {
	var builder string
	builder = "# header\n\n"
	for i := 1; i <= 20; i++ {
		builder += "--- RECORD 001 type=data y=0.100 ---\nL: depth\nR: rock\n\n"
	}
	chunks := splitStructuredOCRTextIntoChunks(builder, 5)
	require.GreaterOrEqual(t, len(chunks), 2)
}

func TestParseYandexVisionAnnotation(t *testing.T) {
	raw := []byte(`{
		"result": {
			"textAnnotation": {
				"fullText": "49.0 52.2",
				"width": "1000",
				"height": "800",
				"blocks": [{
					"lines": [{
						"text": "49.0 52.2",
						"boundingBox": {"vertices":[{"x":"10","y":"20"},{"x":"100","y":"20"},{"x":"100","y":"40"},{"x":"10","y":"40"}]},
						"words": [
							{"text":"49.0","boundingBox":{"vertices":[{"x":"10","y":"20"},{"x":"40","y":"20"},{"x":"40","y":"40"},{"x":"10","y":"40"}]}},
							{"text":"52.2","boundingBox":{"vertices":[{"x":"60","y":"20"},{"x":"100","y":"20"},{"x":"100","y":"40"},{"x":"60","y":"40"}]}}
						]
					}]
				}]
			}
		}
	}`)
	annotation, err := parseYandexVisionAnnotation(raw)
	require.NoError(t, err)
	require.Equal(t, "49.0 52.2", annotation.FullText)
	require.Len(t, annotation.Words, 2)
	require.Equal(t, 1000, annotation.Width)
}

func TestGeologicalJournalStructuringMaxTokens(t *testing.T) {
	require.Equal(t, 8192, geologicalJournalStructuringMaxTokens(4096))
	require.Equal(t, 16384, geologicalJournalStructuringMaxTokens(16384))
}
