package httpx

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"

	applog "github.com/gov-waf/waf-control/internal/log"
)

type ctxKeyRequestID struct{}
type ctxKeyClaims struct{}

// RequestIDFromContext returns the request ID.
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyRequestID{}).(string); ok {
		return v
	}
	return ""
}

// WithClaims stores JWT claims in context.
func WithClaims(ctx context.Context, claims any) context.Context {
	return context.WithValue(ctx, ctxKeyClaims{}, claims)
}

// ClaimsFromContext retrieves JWT claims.
func ClaimsFromContext(ctx context.Context) any {
	return ctx.Value(ctxKeyClaims{})
}

// Recover middleware recovers panics and returns 500.
func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recovered",
						"panic", rec,
						"stack", string(debug.Stack()),
						"path", r.URL.Path,
					)
					Fail(w, http.StatusInternalServerError, CodeInternal, "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// RequestID middleware injects X-Request-ID.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-ID")
		if rid == "" {
			rid = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", rid)
		ctx := context.WithValue(r.Context(), ctxKeyRequestID{}, rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AccessLog logs method, path, status, duration with request_id.
func AccessLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rid := RequestIDFromContext(r.Context())
			reqLogger := applog.WithRequestID(logger, rid)
			ctx := applog.IntoContext(r.Context(), reqLogger)
			ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(ww, r.WithContext(ctx))
			reqLogger.Info("access",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"ip", ClientIP(r),
			)
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// ClientIP extracts client IP from headers or RemoteAddr.
func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}

// StatusFromWriter returns status from statusWriter if wrapped.
func StatusFromWriter(w http.ResponseWriter) int {
	if sw, ok := w.(*statusWriter); ok {
		return sw.status
	}
	return 0
}
