package service

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateGeologicalJournalImage(t *testing.T) {
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 2, 3))
	img.Set(0, 0, color.White)
	require.NoError(t, png.Encode(&buf, img))

	got, err := ValidateGeologicalJournalImage(buf.Bytes())
	require.NoError(t, err)
	require.Equal(t, "image/png", got.ContentType)
	require.Equal(t, 2, got.Width)
	require.Equal(t, 3, got.Height)

	_, err = ValidateGeologicalJournalImage([]byte("not an image"))
	require.ErrorIs(t, err, ErrGeologicalJournalInvalidImage)
}

func TestParseGeologicalJournalOutputStrict(t *testing.T) {
	raw := `{"rows":[{"date":"2026-01-01","drilling_diameter_mm":76,"depth_from_m":1,"depth_to_m":2,"drilling_run_m":1,"core_recovery_m":0.8,"core_recovery_pct":80,"rock_description":"granite","sampling_interval":"1-2","sample_number":"A1","notes":"","uncertainties":[]}]}`
	out, err := ParseGeologicalJournalOutput(raw)
	require.NoError(t, err)
	require.Len(t, out.Rows, 1)
	require.Equal(t, "granite", out.Rows[0].RockDescription)

	_, err = ParseGeologicalJournalOutput(`{"rows":[],"unexpected":true}`)
	require.Error(t, err)
}

func TestGeologicalJournalHasAccess(t *testing.T) {
	require.True(t, GeologicalJournalHasAccess("superadmin", false))
	require.True(t, GeologicalJournalHasAccess("user", true))
	require.False(t, GeologicalJournalHasAccess("user", false))
}
