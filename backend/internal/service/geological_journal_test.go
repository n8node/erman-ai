package service

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/erman-ai/erman-ai/internal/model"
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
	cases := []string{
		raw,
		"Processing the image...\n" + raw,
		"```json\n" + raw + "\n```",
		"Preliminary {not valid JSON} result:\n" + raw + "\nDone.",
	}
	for _, content := range cases {
		out, err := ParseGeologicalJournalOutput(content)
		require.NoError(t, err)
		require.Len(t, out.Rows, 1)
		require.Equal(t, "granite", out.Rows[0].RockDescription)
	}

	_, err := ParseGeologicalJournalOutput(`{"rows":[],"unexpected":true}`)
	require.Error(t, err)
	_, err = ParseGeologicalJournalOutput("Processing failed to return structured data.")
	require.Error(t, err)
}

func TestParseGeologicalJournalOutputCoercesStringNumbers(t *testing.T) {
	raw := `{"rows":[{"date":"24.07.25","drilling_diameter_mm":"76","depth_from_m":"155.4","depth_to_m":"156.2","drilling_run_m":"0,8","core_recovery_m":"0.7","core_recovery_pct":"87%","rock_description":"Сланцы","sampling_interval":"","sample_number":"","notes":"","uncertainties":[]}]}`
	out, err := ParseGeologicalJournalOutput(raw)
	require.NoError(t, err)
	require.Len(t, out.Rows, 1)
	require.NotNil(t, out.Rows[0].DrillingDiameterMM)
	require.InDelta(t, 76, *out.Rows[0].DrillingDiameterMM, 0.001)
	require.NotNil(t, out.Rows[0].DrillingRunM)
	require.InDelta(t, 0.8, *out.Rows[0].DrillingRunM, 0.001)
	require.NotNil(t, out.Rows[0].CoreRecoveryPct)
	require.InDelta(t, 87, *out.Rows[0].CoreRecoveryPct, 0.001)
}

func TestGeologicalJournalHasAccess(t *testing.T) {
	require.True(t, GeologicalJournalHasAccess("superadmin", false))
	require.True(t, GeologicalJournalHasAccess("user", true))
	require.False(t, GeologicalJournalHasAccess("user", false))
}

func TestApplyGeologicalJournalModelDefaultsIncludesDeepSeek(t *testing.T) {
	settings := model.GeologicalJournalSettings{}
	applyGeologicalJournalModelDefaults(&settings, []model.LLMProviderStatus{
		{ID: model.LLMProviderOpenRouter, DefaultModel: "openrouter-default"},
		{ID: model.LLMProviderDeepSeek, DefaultModel: "deepseek-default"},
		{ID: model.LLMProviderYandex, DefaultModel: "yandex-default"},
	})

	require.Equal(t, "openrouter-default", settings.OpenRouterModel)
	require.Equal(t, "deepseek-default", settings.DeepSeekModel)
	require.Equal(t, "yandex-default", settings.YandexModel)
	settings.Provider = model.LLMProviderDeepSeek
	require.Equal(t, "deepseek-default", settings.ActiveModel())
}
