package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gov-waf/waf-control/internal/httpx"
)

type memLoader struct {
	users map[uuid.UUID]*User
	roles map[uuid.UUID][]string
	perms map[uuid.UUID][]string
	err   error
}

func (m *memLoader) LoadUserAuthz(ctx context.Context, userID uuid.UUID) (*User, []string, []string, error) {
	if m.err != nil {
		return nil, nil, nil, m.err
	}
	u := m.users[userID]
	if u == nil {
		return nil, nil, nil, nil
	}
	return u, m.roles[userID], m.perms[userID], nil
}

func TestJWTAuthValidPass(t *testing.T) {
	secret := "test-secret-at-least-32-chars-long!!"
	tokens := NewJWTService(secret)
	uid := uuid.New()
	loader := &memLoader{
		users: map[uuid.UUID]*User{uid: {ID: uid, Username: "alice", Status: "enabled"}},
		roles: map[uuid.UUID][]string{uid: {"admin"}},
		perms: map[uuid.UUID][]string{uid: {"site:read"}},
	}
	tok, _, err := tokens.Issue(&Claims{UserID: uid.String(), Username: "alice"}, 1)
	require.NoError(t, err)

	var gotPerms []string
	h := JWTAuth(tokens, loader)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := ClaimsFromRequest(r)
		require.NotNil(t, c)
		gotPerms = c.Permissions
		httpx.OK(w, map[string]string{"ok": "1"})
	}))
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	assert.Equal(t, 200, rr.Code)
	assert.Equal(t, []string{"site:read"}, gotPerms)
}

func TestJWTAuthInvalidFail(t *testing.T) {
	tokens := NewJWTService("test-secret-at-least-32-chars-long!!")
	loader := &memLoader{users: map[uuid.UUID]*User{}}
	h := JWTAuth(tokens, loader)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach")
	}))
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer bad")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	assert.Equal(t, 401, rr.Code)
}

func TestJWTAuthExpiredFail(t *testing.T) {
	secret := "test-secret-at-least-32-chars-long!!"
	uid := uuid.New()
	loader := &memLoader{
		users: map[uuid.UUID]*User{uid: {ID: uid, Username: "alice", Status: "enabled"}},
		perms: map[uuid.UUID][]string{uid: {"site:read"}},
	}
	expired := mustExpiredToken(t, secret, uid.String(), "alice")
	h := JWTAuth(NewJWTService(secret), loader)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach")
	}))
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+expired)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	assert.Equal(t, 401, rr.Code)
}

func TestJWTAuthDisabledUserFail(t *testing.T) {
	secret := "test-secret-at-least-32-chars-long!!"
	tokens := NewJWTService(secret)
	uid := uuid.New()
	loader := &memLoader{
		users: map[uuid.UUID]*User{uid: {ID: uid, Username: "alice", Status: "disabled"}},
		perms: map[uuid.UUID][]string{uid: {"site:read"}},
	}
	tok, _, err := tokens.Issue(&Claims{UserID: uid.String(), Username: "alice"}, 1)
	require.NoError(t, err)
	h := JWTAuth(tokens, loader)(RequirePermission("site:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach")
	})))
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	assert.Equal(t, 401, rr.Code)
}

func TestJWTAuthRoleRemovedLosesPermission(t *testing.T) {
	secret := "test-secret-at-least-32-chars-long!!"
	tokens := NewJWTService(secret)
	uid := uuid.New()
	loader := &memLoader{
		users: map[uuid.UUID]*User{uid: {ID: uid, Username: "alice", Status: "enabled"}},
		roles: map[uuid.UUID][]string{uid: {}},
		perms: map[uuid.UUID][]string{uid: {}},
	}
	tok, _, err := tokens.Issue(&Claims{UserID: uid.String(), Username: "alice"}, 1)
	require.NoError(t, err)
	h := JWTAuth(tokens, loader)(RequirePermission("site:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach")
	})))
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	assert.Equal(t, 403, rr.Code)
}

func TestJWTAuthMissingUserFail(t *testing.T) {
	secret := "test-secret-at-least-32-chars-long!!"
	tokens := NewJWTService(secret)
	uid := uuid.New()
	loader := &memLoader{users: map[uuid.UUID]*User{}}
	tok, _, err := tokens.Issue(&Claims{UserID: uid.String(), Username: "gone"}, 1)
	require.NoError(t, err)
	h := JWTAuth(tokens, loader)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach")
	}))
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	assert.Equal(t, 401, rr.Code)
	var env httpx.Envelope
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&env))
	assert.Contains(t, env.Message, "user not found")
}

func mustExpiredToken(t *testing.T, secret, userID, username string) string {
	t.Helper()
	now := time.Now().Add(-2 * time.Hour)
	claims := jwtClaims{
		UserID: userID, Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   userID,
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(secret))
	require.NoError(t, err)
	return signed
}
