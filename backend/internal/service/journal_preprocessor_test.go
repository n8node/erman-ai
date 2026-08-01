package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJournalImagePreprocessorPreprocess(t *testing.T) {
	t.Parallel()
	expected := []byte("processed-png")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/preprocess", r.URL.Path)
		assert.Equal(t, "image/jpeg", r.Header.Get("Content-Type"))
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set(
			"X-Preprocessing-Metadata",
			`{"perspective_corrected":true,"deskew_angle":1.2,"scale":2,"width":1200,"height":800}`,
		)
		_, _ = w.Write(expected)
	}))
	defer server.Close()

	preprocessor := NewJournalImagePreprocessor(server.URL)
	result, err := preprocessor.Preprocess(
		context.Background(),
		[]byte("source-image"),
		"image/jpeg",
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, expected, result.Image)
	assert.Equal(t, "image/png", result.ContentType)
	assert.True(t, result.Metadata.Applied)
	assert.True(t, result.Metadata.HasPreprocessedImage)
	assert.True(t, result.Metadata.PerspectiveCorrected)
	assert.Equal(t, 1200, result.Metadata.Width)
}

func TestJournalImagePreprocessorDisabledForInvalidURL(t *testing.T) {
	t.Parallel()
	preprocessor := NewJournalImagePreprocessor("not-a-url")

	assert.False(t, preprocessor.Enabled())
	_, err := preprocessor.Preprocess(
		context.Background(),
		[]byte("source-image"),
		"image/jpeg",
	)
	assert.Error(t, err)
}
