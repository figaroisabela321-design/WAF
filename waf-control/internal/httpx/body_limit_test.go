package httpx

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaxBodyBytesRejectsOversized(t *testing.T) {
	const max = 64
	h := MaxBodyBytes(max)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var dst map[string]any
		if !DecodeJSON(w, r, &dst) {
			return
		}
		OK(w, dst)
	}))
	body := `{"x":"` + strings.Repeat("a", 200) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	assert.Equal(t, 413, rr.Code)
	var env Envelope
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&env))
	assert.Equal(t, CodePayloadTooLarge, env.Code)
}

func TestMaxBodyBytesAllowsSmall(t *testing.T) {
	h := MaxBodyBytes(1 << 20)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var dst map[string]string
		if !DecodeJSON(w, r, &dst) {
			return
		}
		OK(w, dst)
	}))
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"a":"b"}`))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	assert.Equal(t, 200, rr.Code)
}

func TestIsBodyTooLarge(t *testing.T) {
	assert.True(t, IsBodyTooLarge(&http.MaxBytesError{Limit: 10}))
	assert.False(t, IsBodyTooLarge(io.EOF))
	assert.False(t, IsBodyTooLarge(nil))
}
