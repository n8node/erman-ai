package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeologicalJournalMaxPDFBytes(t *testing.T) {
	if GeologicalJournalMaxPDFBytes != 250<<20 {
		t.Fatalf("unexpected PDF limit: %d", GeologicalJournalMaxPDFBytes)
	}
}

func TestParseDocumentLLMTableResult(t *testing.T) {
	content := `{"rows":[{"date":"24.07.25","drilling_diameter_mm":76,"depth_from_m":1,"depth_to_m":2,"drilling_run_m":1,"core_recovery_m":0.8,"core_recovery_pct":80,"rock_description":"сланец","sampling_interval":"1-2","sample_number":"A1","notes":"","uncertainties":[]}],"summary":"итог"}`

	result, ok := parseDocumentLLMTableResult(content)
	require.True(t, ok)
	require.Len(t, result.Rows, 1)
	require.Equal(t, "сланец", result.Rows[0]["rock_description"])
}

func TestParseDocumentLLMMixedResultKeepsTextAndTable(t *testing.T) {
	content := `{"transcription":"The ore in the upper levels was exceedingly rich.","columns":["Year","Production"],"rows":[{"Year":"1870","Production":"$15,000"}]}`

	transcription, raw := parseDocumentLLMResponse(content)
	table, ok := parseDocumentLLMTableResult(content)

	require.Equal(t, "The ore in the upper levels was exceedingly rich.", transcription)
	require.JSONEq(t, content, string(raw))
	require.True(t, ok)
	require.Equal(t, []string{"Year", "Production"}, table.Columns)
	require.Equal(t, "1870", table.Rows[0]["Year"])
}

func TestParseDocumentLLMTranscription(t *testing.T) {
	transcription, ok := parseDocumentLLMTranscription(`{"transcription":"paragraph text","columns":[],"rows":[]}`)

	require.True(t, ok)
	require.Equal(t, "paragraph text", transcription)
}
