package audit

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/gov-waf/waf-control/internal/auth"
	"github.com/gov-waf/waf-control/internal/httpx"
	applog "github.com/gov-waf/waf-control/internal/log"
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
// On Write failure: do NOT fail the business HTTP response; ERROR-log
// request_id, actor_id, actor_name, method, path, resource, resource_id, error.
// Never log password/JWT/Authorization/secret/full sensitive body.
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

			logEntry := &Log{
				ID:         uuid.New(),
				Method:     method,
				Path:       r.URL.Path,
				Resource:   extractResource(r.URL.Path),
				ResourceID: extractResourceID(r.URL.Path),
				StatusCode: sw.status,
				IP:         httpx.ClientIPFromRequest(r, httpx.TrustedProxiesFromContext(r.Context())),
				UserAgent:  r.UserAgent(),
				CreatedAt:  time.Now().UTC(),
			}

			var actorIDStr, actorName string
			if c := auth.ClaimsFromRequest(r); c != nil {
				actorName = c.Username
				logEntry.ActorName = c.Username
				if id, err := uuid.Parse(c.UserID); err == nil {
					logEntry.ActorID = &id
					actorIDStr = id.String()
				}
			}
			if err := svc.Write(r.Context(), logEntry); err != nil {
				logger := applog.FromContext(r.Context())
				if logger == nil {
					logger = slog.Default()
				}
				logger.Error("audit write failed",
					"request_id", httpx.RequestIDFromContext(r.Context()),
					"actor_id", actorIDStr,
					"actor_name", actorName,
					"method", method,
					"path", r.URL.Path,
					"resource", logEntry.Resource,
					"resource_id", logEntry.ResourceID,
					"error", err.Error(),
				)
			}
		})
	}
}

func extractResource(path string) string {
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
