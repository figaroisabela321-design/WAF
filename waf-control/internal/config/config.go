package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"

	// Documented development defaults — refused in production.
	DevJWTSecret     = "dev-jwt-secret-change-me"
	DevAdminPassword = "Admin@123"

	MinJWTSecretLen     = 32
	MinAdminPasswordLen = 12
	DefaultMaxBodyBytes = 1 << 20 // 1 MiB
)

// Config holds all runtime configuration from environment variables.
type Config struct {
	AppEnv         string
	HTTPAddr       string
	DatabaseURL    string
	JWTSecret      string
	JWTExpireHours int
	LogLevel       string
	AdminUsername  string
	AdminPassword  string
	TrustedProxies []*net.IPNet
	SwaggerEnabled bool
	MaxBodyBytes   int64
}

// Load reads configuration from environment with sane local defaults.
// APP_ENV: unset → development; otherwise case-insensitive normalize to lowercase.
// Only "development" and "production" are allowed — unknown values (typos like
// "prodution", "prod", empty when set) are preserved so Validate fails closed.
// Call Validate after Load; production will fail fast on unsafe secrets.
func Load() Config {
	appEnv := EnvDevelopment
	if v, ok := os.LookupEnv("APP_ENV"); ok {
		// Set (including empty): normalize; do not silently remap unknowns to development.
		appEnv = strings.ToLower(strings.TrimSpace(v))
	}

	cfg := Config{
		AppEnv:         appEnv,
		HTTPAddr:       getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://waf:waf@localhost:5432/waf?sslmode=disable"),
		JWTSecret:      getEnvAllowEmpty("JWT_SECRET", DevJWTSecret),
		JWTExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 24),
		LogLevel:       strings.ToLower(getEnv("LOG_LEVEL", "info")),
		AdminUsername:  getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:  getEnvAllowEmpty("ADMIN_PASSWORD", DevAdminPassword),
		TrustedProxies: parseCIDRs(getEnv("TRUSTED_PROXIES", "")),
		MaxBodyBytes:   int64(getEnvInt("MAX_BODY_BYTES", DefaultMaxBodyBytes)),
	}

	// Swagger: development default on; production default off unless SWAGGER_ENABLED=true.
	if v, ok := os.LookupEnv("SWAGGER_ENABLED"); ok {
		cfg.SwaggerEnabled = parseBool(v, appEnv != EnvProduction)
	} else {
		cfg.SwaggerEnabled = appEnv != EnvProduction
	}

	if cfg.MaxBodyBytes <= 0 {
		cfg.MaxBodyBytes = DefaultMaxBodyBytes
	}
	return cfg
}

// Validate fails fast for invalid APP_ENV and production misconfiguration.
// Never logs secret values.
func (c Config) Validate() error {
	switch c.AppEnv {
	case EnvDevelopment:
		return nil
	case EnvProduction:
		// continue to secret checks below
	default:
		return fmt.Errorf("invalid APP_ENV %q: must be %q or %q (case-insensitive normalization); unset APP_ENV defaults to development", c.AppEnv, EnvDevelopment, EnvProduction)
	}

	if strings.TrimSpace(c.JWTSecret) == "" {
		return fmt.Errorf("production config refused: JWT_SECRET is empty")
	}
	if c.JWTSecret == DevJWTSecret {
		return fmt.Errorf("production config refused: JWT_SECRET equals the documented development default")
	}
	if len(c.JWTSecret) < MinJWTSecretLen {
		return fmt.Errorf("production config refused: JWT_SECRET is too short (minimum %d characters)", MinJWTSecretLen)
	}
	if c.AdminPassword == "" {
		return fmt.Errorf("production config refused: ADMIN_PASSWORD is empty")
	}
	if c.AdminPassword == DevAdminPassword {
		return fmt.Errorf("production config refused: ADMIN_PASSWORD equals the documented development default")
	}
	if len(c.AdminPassword) < MinAdminPasswordLen {
		return fmt.Errorf("production config refused: ADMIN_PASSWORD is too short (minimum %d characters)", MinAdminPasswordLen)
	}
	return nil
}

// IsProduction reports whether APP_ENV is production.
func (c Config) IsProduction() bool {
	return c.AppEnv == EnvProduction
}

func parseCIDRs(raw string) []*net.IPNet {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]*net.IPNet, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.Contains(p, "/") {
			if ip := net.ParseIP(p); ip != nil {
				if ip.To4() != nil {
					p = p + "/32"
				} else {
					p = p + "/128"
				}
			}
		}
		_, n, err := net.ParseCIDR(p)
		if err != nil {
			continue
		}
		out = append(out, n)
	}
	return out
}

func parseBool(v string, def bool) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}

// getEnvAllowEmpty returns the env value even if empty when the variable is set.
// Unset variables fall back to def. Used so production can detect explicit empty secrets.
func getEnvAllowEmpty(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
