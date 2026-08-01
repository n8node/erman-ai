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
		_, _ = w.Write(expected)
	}))
	defer server.Close()

	preprocessor := NewJournalImagePreprocessor(server.URL)
	actual, contentType, err := preprocessor.Preprocess(
		context.Background(),
		[]byte("source-image"),
		"image/jpeg",
	)

	require.NoError(t, err)
	assert.Equal(t, expected, actual)
	assert.Equal(t, "image/png", contentType)
}

func TestJournalImagePreprocessorDisabledForInvalidURL(t *testing.T) {
	t.Parallel()
	preprocessor := NewJournalImagePreprocessor("not-a-url")

	assert.False(t, preprocessor.Enabled())
	_, _, err := preprocessor.Preprocess(
		context.Background(),
		[]byte("source-image"),
		"image/jpeg",
	)
	assert.Error(t, err)
}
