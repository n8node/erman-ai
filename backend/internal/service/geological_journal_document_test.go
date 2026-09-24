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
