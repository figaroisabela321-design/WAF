package httpx

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeEnvelope_Success(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		message string
		data    any
	}{
		{"ok with object", 0, "ok", map[string]string{"id": "1"}},
		{"ok with nil data", 0, "ok", nil},
		{"ok with list", 0, "ok", map[string]any{"items": []int{1, 2}, "total": 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := EncodeEnvelope(tt.code, tt.message, tt.data)
			require.NoError(t, err)
			var env Envelope
			require.NoError(t, json.Unmarshal(b, &env))
			assert.Equal(t, tt.code, env.Code)
			assert.Equal(t, tt.message, env.Message)
		})
	}
}

func TestEncodeEnvelope_Error(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		message string
	}{
		{"bad request", 40000, "bad request"},
		{"unauthorized", 40100, "unauthorized"},
		{"not found", 40400, "not found"},
		{"internal", 50000, "internal error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := EncodeEnvelope(tt.code, tt.message, nil)
			require.NoError(t, err)
			var env Envelope
			require.NoError(t, json.Unmarshal(b, &env))
			assert.Equal(t, tt.code, env.Code)
			assert.Equal(t, tt.message, env.Message)
			assert.Nil(t, env.Data)
		})
	}
}
