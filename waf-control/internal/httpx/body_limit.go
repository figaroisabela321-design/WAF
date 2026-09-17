package httpx

import (
	"net/http"
)

// MaxBodyBytes limits request body size (default ~1 MiB).
// Oversized bodies receive HTTP 413 with the project error envelope when handlers
// decode JSON via DecodeJSON / check IsBodyTooLarge.
func MaxBodyBytes(max int64) func(http.Handler) http.Handler {
	if max <= 0 {
		max = 1 << 20
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost, http.MethodPut, http.MethodPatch:
				r.Body = http.MaxBytesReader(w, r.Body, max)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// WriteBodyTooLarge writes a 413 envelope.
func WriteBodyTooLarge(w http.ResponseWriter) {
	Fail(w, 413, CodePayloadTooLarge, "request body too large")
}

// IsBodyTooLarge reports whether err is from MaxBytesReader.
func IsBodyTooLarge(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(*http.MaxBytesError); ok {
		return true
	}
	return false
}
