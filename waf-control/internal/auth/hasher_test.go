package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBcryptHasher(t *testing.T) {
	h := NewBcryptHasher()
	tests := []struct {
		name     string
		password string
	}{
		{"simple", "password"},
		{"admin default", "Admin@123"},
		{"unicode", "密码测试123!"},
		{"long", "a-very-long-password-with-special-chars-!@#$%^&*()"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := h.Hash(tt.password)
			require.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.NotEqual(t, tt.password, hash)
			assert.NoError(t, h.Compare(hash, tt.password))
			assert.Error(t, h.Compare(hash, tt.password+"wrong"))
		})
	}
}

func TestBcryptHasher_DifferentHashes(t *testing.T) {
	h := NewBcryptHasher()
	h1, err := h.Hash("Admin@123")
	require.NoError(t, err)
	h2, err := h.Hash("Admin@123")
	require.NoError(t, err)
	assert.NotEqual(t, h1, h2) // salt differs
	assert.NoError(t, h.Compare(h1, "Admin@123"))
	assert.NoError(t, h.Compare(h2, "Admin@123"))
}
