package service

import "testing"

func TestGeologicalJournalMaxPDFBytes(t *testing.T) {
	if GeologicalJournalMaxPDFBytes != 250<<20 {
		t.Fatalf("unexpected PDF limit: %d", GeologicalJournalMaxPDFBytes)
	}
}
