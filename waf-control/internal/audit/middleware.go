package audit

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/gov-waf/waf-control/internal/auth"
	"github.com/gov-waf/waf-control/internal/httpx"
)

type statusCapture struct {
	http.ResponseWriter
	status int
}

func (w *statusCapture) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Middleware writes audit logs for mutating methods under /api/v1.
func Middleware(svc *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			method := r.Method
			mutating := method == http.MethodPost || method == http.MethodPut ||
				method == http.MethodPatch || method == http.MethodDelete
			if !mutating || !strings.HasPrefix(r.URL.Path, "/api/v1") {
				next.ServeHTTP(w, r)
				return
			}
			sw := &statusCapture{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)

			log := &Log{
				ID:         uuid.New(),
				Method:     method,
				Path:       r.URL.Path,
				Resource:   extractResource(r.URL.Path),
				ResourceID: extractResourceID(r.URL.Path),
				StatusCode: sw.status,
				IP:         httpx.ClientIP(r),
				UserAgent:  r.UserAgent(),
				CreatedAt:  time.Now().UTC(),
			}
			if c := auth.ClaimsFromRequest(r); c != nil {
				log.ActorName = c.Username
				if id, err := uuid.Parse(c.UserID); err == nil {
					log.ActorID = &id
				}
			}
			// fire-and-forget; don't block response on audit failure
			_ = svc.Write(r.Context(), log)
		})
	}
}

func extractResource(path string) string {
	// /api/v1/sites/xxx -> sites
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

func extractResourceID(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 4 {
		return parts[3]
	}
	return ""
}
