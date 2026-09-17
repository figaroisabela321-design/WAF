package auth

import (
	"net/http"
	"strings"

	"github.com/gov-waf/waf-control/internal/httpx"
)

// JWT middleware validates Bearer token and injects Claims into context.
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

// RequirePermission checks that the current user has the given permission code.
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
