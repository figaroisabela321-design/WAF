package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime configuration from environment variables.
type Config struct {
	HTTPAddr       string
	DatabaseURL    string
	JWTSecret      string
	JWTExpireHours int
	LogLevel       string
	AdminUsername  string
	AdminPassword  string
}

// Load reads configuration from environment with sane local defaults.
func Load() Config {
	return Config{
		HTTPAddr:       getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://waf:waf@localhost:5432/waf?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "dev-jwt-secret-change-me"),
		JWTExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 24),
		LogLevel:       strings.ToLower(getEnv("LOG_LEVEL", "info")),
		AdminUsername:  getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:  getEnv("ADMIN_PASSWORD", "Admin@123"),
	}
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
