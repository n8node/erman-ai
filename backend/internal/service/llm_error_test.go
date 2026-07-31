package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLLMAPIErrorUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    string
	}{
		{name: "string", payload: `"model does not support images"`, want: "model does not support images"},
		{name: "message object", payload: `{"message":"invalid model"}`, want: "invalid model"},
		{name: "detail object", payload: `{"detail":"permission denied"}`, want: "permission denied"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got llmAPIError
			require.NoError(t, json.Unmarshal([]byte(tt.payload), &got))
			assert.Equal(t, tt.want, got.Message)
		})
	}
}
