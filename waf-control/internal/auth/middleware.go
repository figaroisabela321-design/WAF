package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/gov-waf/waf-control/internal/httpx"
)

// PermissionLoader loads live roles/permissions for a user.
// Implementations may add caching later (e.g. Redis) without changing callers.
type PermissionLoader interface {
	LoadUserAuthz(ctx context.Context, userID uuid.UUID) (user *User, roles []string, permissions []string, err error)
}

// JWTAuth middleware: verify JWT → load user from DB → require enabled →
// load current roles/permissions → inject Claims. Embedded token permissions
// are never used as the source of truth.
func JWTAuth(tokens TokenService, loader PermissionLoader) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authz := r.Header.Get("Authorization")
			if authz == "" || !strings.HasPrefix(authz, "Bearer ") {
				httpx.Fail(w, 401, httpx.CodeUnauthorized, "missing or invalid authorization header")
				return
			}
			token := strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
			idClaims, err := tokens.Parse(token)
			if err != nil {
				httpx.Fail(w, 401, httpx.CodeUnauthorized, "invalid or expired token")
				return
			}
			uid, err := uuid.Parse(idClaims.UserID)
			if err != nil {
				httpx.Fail(w, 401, httpx.CodeUnauthorized, "invalid or expired token")
				return
			}
			user, roles, perms, err := loader.LoadUserAuthz(r.Context(), uid)
			if err != nil {
				httpx.Fail(w, 500, httpx.CodeInternal, "authorization lookup failed")
				return
			}
			if user == nil {
				httpx.Fail(w, 401, httpx.CodeUnauthorized, "user not found")
				return
			}
			if user.Status != "enabled" {
				httpx.Fail(w, 401, httpx.CodeUnauthorized, "user disabled")
				return
			}
			claims := &Claims{
				UserID:      user.ID.String(),
				Username:    user.Username,
				Roles:       roles,
				Permissions: perms,
			}
			ctx := httpx.WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// JWT is retained as a name alias for identity-only parse without DB load.
// Prefer JWTAuth for protected routes.
func JWT(tokens TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authz := r.Header.Get("Authorization")
			if authz == "" || !strings.HasPrefix(authz, "Bearer ") {
				httpx.Fail(w, 401, httpx.CodeUnauthorized, "missing or invalid authorization header")
				return
			}
			token := strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
			claims, err := tokens.Parse(token)
			if err != nil {
				httpx.Fail(w, 401, httpx.CodeUnauthorized, "invalid or expired token")
				return
			}
			ctx := httpx.WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission checks that the current user has the given permission code
// from live Claims.Permissions (loaded from DB, not from JWT).
func RequirePermission(code string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := httpx.ClaimsFromContext(r.Context())
			claims, ok := raw.(*Claims)
			if !ok || claims == nil {
				httpx.Fail(w, 401, httpx.CodeUnauthorized, "unauthorized")
				return
			}
			if !hasPermission(claims.Permissions, code) {
				httpx.Fail(w, 403, httpx.CodeForbidden, "permission denied: "+code)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func hasPermission(perms []string, code string) bool {
	for _, p := range perms {
		if p == code {
			return true
		}
	}
	return false
}

// ClaimsFromRequest is a helper to get *Claims.
func ClaimsFromRequest(r *http.Request) *Claims {
	raw := httpx.ClaimsFromContext(r.Context())
	if c, ok := raw.(*Claims); ok {
		return c
	}
	return nil
}
