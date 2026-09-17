package audit

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gov-waf/waf-control/internal/auth"
	"github.com/gov-waf/waf-control/internal/httpx"
	applog "github.com/gov-waf/waf-control/internal/log"
)

type failRepo struct{}

func (failRepo) Create(ctx context.Context, log *Log) error {
	return errors.New("db down")
}
func (failRepo) List(ctx context.Context, offset, limit int) ([]Log, int, error) {
	return nil, 0, nil
}
func (failRepo) GetByID(ctx context.Context, id uuid.UUID) (*Log, error) {
	return nil, nil
}

func TestAuditWriteFailureLogsErrorDoesNotFailHTTP(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := NewService(failRepo{})
	uid := uuid.New()

	h := httpx.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(applog.IntoContext(r.Context(), logger))
		r = r.WithContext(httpx.WithClaims(r.Context(), &auth.Claims{
			UserID: uid.String(), Username: "auditor",
		}))
		Middleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpx.OK(w, map[string]string{"ok": "1"})
		})).ServeHTTP(w, r)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sites", nil)
	req.Header.Set("X-Request-ID", "req-audit-test-1")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	assert.Equal(t, 200, rr.Code)

	out := buf.String()
	require.Contains(t, out, "audit write failed")
	assert.Contains(t, out, "auditor")
	assert.Contains(t, out, "req-audit-test-1")
	assert.Contains(t, out, `"method":"POST"`)
	assert.Contains(t, out, "/api/v1/sites")
	assert.Contains(t, out, `"resource":"sites"`)
	assert.NotContains(t, out, "Authorization")
	assert.NotContains(t, strings.ToLower(out), "password")
	assert.NotContains(t, out, "Bearer ")
}
