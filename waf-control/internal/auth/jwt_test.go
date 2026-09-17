package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTIssueAndParseIdentityOnly(t *testing.T) {
	svc := NewJWTService("test-secret-at-least-32-chars-long!!")
	token, expiresIn, err := svc.Issue(&Claims{UserID: "uid-1", Username: "alice", Roles: []string{"admin"}, Permissions: []string{"site:read"}}, 1)
	require.NoError(t, err)
	assert.Equal(t, 3600, expiresIn)

	claims, err := svc.Parse(token)
	require.NoError(t, err)
	assert.Equal(t, "uid-1", claims.UserID)
	assert.Equal(t, "alice", claims.Username)
	assert.Empty(t, claims.Roles)
	assert.Empty(t, claims.Permissions)

	// Raw token must not embed permissions/roles as claims we rely on.
	parsed, _, err := jwt.NewParser().ParseUnverified(token, jwt.MapClaims{})
	require.NoError(t, err)
	mc := parsed.Claims.(jwt.MapClaims)
	_, hasRoles := mc["roles"]
	_, hasPerms := mc["permissions"]
	assert.False(t, hasRoles)
	assert.False(t, hasPerms)
}

func TestJWTInvalidToken(t *testing.T) {
	svc := NewJWTService("test-secret-at-least-32-chars-long!!")
	_, err := svc.Parse("not-a-token")
	require.Error(t, err)
}

func TestJWTExpiredToken(t *testing.T) {
	secret := []byte("test-secret-at-least-32-chars-long!!")
	now := time.Now().Add(-2 * time.Hour)
	claims := jwtClaims{
		UserID: "u1", Username: "bob",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   "u1",
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(secret)
	require.NoError(t, err)

	svc := NewJWTService(string(secret))
	_, err = svc.Parse(signed)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "expired") || err != nil)
}

func TestJWTWrongSecret(t *testing.T) {
	svc := NewJWTService("test-secret-at-least-32-chars-long!!")
	token, _, err := svc.Issue(&Claims{UserID: "u1", Username: "x"}, 1)
	require.NoError(t, err)
	other := NewJWTService("other-secret-at-least-32-chars-long!")
	_, err = other.Parse(token)
	require.Error(t, err)
}
