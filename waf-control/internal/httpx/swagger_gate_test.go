package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/gov-waf/waf-control/internal/config"
)

func TestSwaggerProductionDefaultOff(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "this-is-a-production-jwt-secret-32+")
	t.Setenv("ADMIN_PASSWORD", "SecureAdmin!234")
	_ = os.Unsetenv("SWAGGER_ENABLED")
	cfg := config.Load()
	assert.False(t, cfg.SwaggerEnabled)

	r := chi.NewRouter()
	if cfg.SwaggerEnabled {
		r.Get("/swagger/*", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		})
	} else {
		r.Get("/swagger/*", func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		})
	}
	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, 404, rr.Code)
}

func TestSwaggerDevelopmentDefaultOn(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	_ = os.Unsetenv("SWAGGER_ENABLED")
	cfg := config.Load()
	assert.True(t, cfg.SwaggerEnabled)
}
