package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gov-waf/waf-control/docs"
	"github.com/gov-waf/waf-control/internal/alert"
	"github.com/gov-waf/waf-control/internal/audit"
	"github.com/gov-waf/waf-control/internal/auth"
	"github.com/gov-waf/waf-control/internal/config"
	"github.com/gov-waf/waf-control/internal/db"
	"github.com/gov-waf/waf-control/internal/event"
	"github.com/gov-waf/waf-control/internal/httpx"
	applog "github.com/gov-waf/waf-control/internal/log"
	"github.com/gov-waf/waf-control/internal/node"
	"github.com/gov-waf/waf-control/internal/policy"
	"github.com/gov-waf/waf-control/internal/rule"
	"github.com/gov-waf/waf-control/internal/site"
	"github.com/gov-waf/waf-control/migrations"
	"github.com/gov-waf/waf-control/pkg/adapters/agent"
	"github.com/gov-waf/waf-control/pkg/adapters/coraza"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	cfg := config.Load()
	logger := applog.New(cfg.LogLevel)
	slog := logger

	if err := cfg.Validate(); err != nil {
		slog.Error("configuration validation failed", "error", err.Error(), "app_env", cfg.AppEnv)
		os.Exit(1)
	}

	ctx := context.Background()

	// Migrations
	if err := db.RunMigrations(cfg.DatabaseURL, migrations.FS); err != nil {
		slog.Error("migrations failed", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations applied")

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Adapters (Phase 1 noops)
	_ = coraza.NewNoop()
	_ = agent.NewNoop()

	// Auth
	hasher := auth.NewBcryptHasher()
	tokens := auth.NewJWTService(cfg.JWTSecret)
	userRepo := auth.NewPGUserRepository(pool)
	roleRepo := auth.NewPGRoleRepository(pool)
	authSvc := auth.NewService(userRepo, roleRepo, hasher, tokens, cfg.JWTExpireHours)
	authHandler := auth.NewHandler(authSvc)

	if err := authSvc.SeedAdmin(ctx, cfg.AdminUsername, cfg.AdminPassword); err != nil {
		slog.Error("seed admin failed", "error", err)
		os.Exit(1)
	}
	slog.Info("admin seed checked", "username", cfg.AdminUsername)

	// Domains
	siteSvc := site.NewService(site.NewPGRepository(pool))
	siteHandler := site.NewHandler(siteSvc)

	nodeSvc := node.NewService(node.NewPGRepository(pool))
	nodeHandler := node.NewHandler(nodeSvc)

	policySvc := policy.NewService(policy.NewPGRepository(pool))
	policyHandler := policy.NewHandler(policySvc)

	ruleSvc := rule.NewService(rule.NewPGRepository(pool))
	ruleHandler := rule.NewHandler(ruleSvc)

	auditSvc := audit.NewService(audit.NewPGRepository(pool))
	auditHandler := audit.NewHandler(auditSvc)

	eventHandler := event.NewHandler(event.NewNoopStore())
	alertHandler := alert.NewHandler(alert.NewNoopStore())

	r := chi.NewRouter()
	r.Use(httpx.Recover(slog))
	r.Use(httpx.RequestID)
	r.Use(httpx.TrustedProxies(cfg.TrustedProxies))
	r.Use(httpx.AccessLog(slog))
	r.Use(httpx.MaxBodyBytes(cfg.MaxBodyBytes))

	// Public health
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		httpx.OK(w, map[string]string{"status": "up"})
	})
	r.Get("/ready", readyHandler(pool))

	// Swagger UI + OpenAPI YAML (disabled by default in production)
	if cfg.SwaggerEnabled {
		r.Get("/swagger/openapi.yaml", serveOpenAPI)
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("/swagger/openapi.yaml"),
		))
		r.Get("/docs/openapi.yaml", serveOpenAPI)
	} else {
		r.Get("/swagger/*", swaggerDisabled)
		r.Get("/swagger/openapi.yaml", swaggerDisabled)
		r.Get("/docs/openapi.yaml", swaggerDisabled)
	}

	// Auth login (public)
	// Protected API
	r.Route("/api/v1", func(api chi.Router) {
		// Login is public; still audited (actor may be empty)
		api.With(audit.Middleware(auditSvc)).Post("/auth/login", authHandler.Login)

		api.Group(func(priv chi.Router) {
			priv.Use(auth.JWTAuth(tokens, userRepo))
			priv.Use(audit.Middleware(auditSvc))

			priv.With(auth.RequirePermission("site:read")).Get("/sites", siteHandler.List)
			priv.With(auth.RequirePermission("site:write")).Post("/sites", siteHandler.Create)
			priv.With(auth.RequirePermission("site:read")).Get("/sites/{id}", siteHandler.Get)
			priv.With(auth.RequirePermission("site:write")).Put("/sites/{id}", siteHandler.Update)
			priv.With(auth.RequirePermission("site:write")).Delete("/sites/{id}", siteHandler.Delete)

			priv.With(auth.RequirePermission("node:read")).Get("/nodes", nodeHandler.List)
			priv.With(auth.RequirePermission("node:write")).Post("/nodes", nodeHandler.Create)
			priv.With(auth.RequirePermission("node:read")).Get("/nodes/{id}", nodeHandler.Get)
			priv.With(auth.RequirePermission("node:write")).Put("/nodes/{id}", nodeHandler.Update)

			priv.With(auth.RequirePermission("policy:read")).Get("/policies", policyHandler.List)
			priv.With(auth.RequirePermission("policy:write")).Post("/policies", policyHandler.Create)
			priv.With(auth.RequirePermission("policy:read")).Get("/policies/{id}", policyHandler.Get)
			priv.With(auth.RequirePermission("policy:write")).Put("/policies/{id}", policyHandler.Update)

			priv.With(auth.RequirePermission("rule:read")).Get("/rules", ruleHandler.List)
			priv.With(auth.RequirePermission("rule:write")).Post("/rules", ruleHandler.Create)
			priv.With(auth.RequirePermission("rule:read")).Get("/rules/{id}", ruleHandler.Get)
			priv.With(auth.RequirePermission("rule:write")).Put("/rules/{id}", ruleHandler.Update)
			priv.With(auth.RequirePermission("rule:write")).Delete("/rules/{id}", ruleHandler.Delete)

			priv.With(auth.RequirePermission("user:read")).Get("/users", authHandler.ListUsers)
			priv.With(auth.RequirePermission("user:write")).Post("/users", authHandler.CreateUser)
			priv.With(auth.RequirePermission("user:read")).Get("/users/{id}", authHandler.GetUser)
			priv.With(auth.RequirePermission("user:write")).Put("/users/{id}", authHandler.UpdateUser)

			priv.With(auth.RequirePermission("user:read")).Get("/roles", authHandler.ListRoles)
			priv.With(auth.RequirePermission("user:write")).Post("/roles", authHandler.CreateRole)
			priv.With(auth.RequirePermission("user:read")).Get("/permissions", authHandler.ListPermissions)

			priv.With(auth.RequirePermission("audit:read")).Get("/audit-logs", auditHandler.List)

			priv.With(auth.RequirePermission("site:read")).Get("/events", eventHandler.List)
			priv.With(auth.RequirePermission("site:read")).Get("/alerts", alertHandler.List)
		})
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		slog.Info("waf-control listening", "addr", cfg.HTTPAddr, "app_env", cfg.AppEnv, "swagger", cfg.SwaggerEnabled)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func readyHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			httpx.Fail(w, 503, httpx.CodeInternal, "database unavailable")
			return
		}
		httpx.OK(w, map[string]string{"status": "ready"})
	}
}

func serveOpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(docs.OpenAPIYAML)
}

func swaggerDisabled(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}
