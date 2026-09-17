package config

import (
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaultsDevelopment(t *testing.T) {
	t.Setenv("APP_ENV", "")
	_ = os.Unsetenv("APP_ENV")
	_ = os.Unsetenv("JWT_SECRET")
	_ = os.Unsetenv("ADMIN_PASSWORD")
	_ = os.Unsetenv("SWAGGER_ENABLED")
	_ = os.Unsetenv("TRUSTED_PROXIES")

	// Clear may not work with t.Setenv siblings; set explicitly.
	t.Setenv("APP_ENV", "development")
	t.Setenv("JWT_SECRET", DevJWTSecret)
	t.Setenv("ADMIN_PASSWORD", DevAdminPassword)

	cfg := Load()
	assert.Equal(t, EnvDevelopment, cfg.AppEnv)
	assert.True(t, cfg.SwaggerEnabled)
	assert.NoError(t, cfg.Validate())
}

func TestProductionValidateJWTEmpty(t *testing.T) {
	cfg := Config{AppEnv: EnvProduction, JWTSecret: "", AdminPassword: "long-enough-password"}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET is empty")
	assert.NotContains(t, err.Error(), cfg.AdminPassword)
}

func TestProductionValidateJWTDevDefault(t *testing.T) {
	cfg := Config{AppEnv: EnvProduction, JWTSecret: DevJWTSecret, AdminPassword: "long-enough-password"}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET equals the documented development default")
}

func TestProductionValidateJWTTooShort(t *testing.T) {
	cfg := Config{AppEnv: EnvProduction, JWTSecret: "short-but-not-dev-default", AdminPassword: "long-enough-password"}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET is too short")
}

func TestProductionValidateAdminEmpty(t *testing.T) {
	secret := "this-is-a-production-jwt-secret-32+"
	cfg := Config{AppEnv: EnvProduction, JWTSecret: secret, AdminPassword: ""}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ADMIN_PASSWORD is empty")
	assert.NotContains(t, err.Error(), secret)
}

func TestProductionValidateAdminDevDefault(t *testing.T) {
	secret := "this-is-a-production-jwt-secret-32+"
	cfg := Config{AppEnv: EnvProduction, JWTSecret: secret, AdminPassword: DevAdminPassword}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ADMIN_PASSWORD equals the documented development default")
	assert.NotContains(t, err.Error(), DevAdminPassword)
}

func TestProductionValidateAdminTooShort(t *testing.T) {
	secret := "this-is-a-production-jwt-secret-32+"
	cfg := Config{AppEnv: EnvProduction, JWTSecret: secret, AdminPassword: "shortpass1"}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ADMIN_PASSWORD is too short")
}

func TestProductionValidateOK(t *testing.T) {
	cfg := Config{
		AppEnv:        EnvProduction,
		JWTSecret:     "this-is-a-production-jwt-secret-32+",
		AdminPassword: "SecureAdmin!234",
	}
	assert.NoError(t, cfg.Validate())
}

func TestSwaggerDefaults(t *testing.T) {
	t.Setenv("JWT_SECRET", DevJWTSecret)
	t.Setenv("ADMIN_PASSWORD", DevAdminPassword)

	t.Setenv("APP_ENV", "development")
	_ = os.Unsetenv("SWAGGER_ENABLED")
	assert.True(t, Load().SwaggerEnabled)

	t.Setenv("APP_ENV", "production")
	_ = os.Unsetenv("SWAGGER_ENABLED")
	assert.False(t, Load().SwaggerEnabled)

	t.Setenv("APP_ENV", "production")
	t.Setenv("SWAGGER_ENABLED", "true")
	assert.True(t, Load().SwaggerEnabled)
}

func TestParseTrustedProxies(t *testing.T) {
	nets := parseCIDRs("127.0.0.1/32, 10.0.0.0/8, bad, 192.168.1.1")
	require.Len(t, nets, 3)
	assert.True(t, nets[0].Contains(net.ParseIP("127.0.0.1")))
	assert.True(t, nets[1].Contains(net.ParseIP("10.1.2.3")))
	assert.True(t, nets[2].Contains(net.ParseIP("192.168.1.1")))
}

func TestProductionLoadEmptyJWTSecret(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("ADMIN_PASSWORD", "SecureAdmin!234")
	cfg := Load()
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET is empty")
}

func TestProductionLoadEmptyAdminPassword(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "this-is-a-production-jwt-secret-32+")
	t.Setenv("ADMIN_PASSWORD", "")
	cfg := Load()
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ADMIN_PASSWORD is empty")
}

func TestLoadUnsetAPP_ENVDefaultsDevelopment(t *testing.T) {
	t.Setenv("JWT_SECRET", DevJWTSecret)
	t.Setenv("ADMIN_PASSWORD", DevAdminPassword)

	prev, had := os.LookupEnv("APP_ENV")
	require.NoError(t, os.Unsetenv("APP_ENV"))
	t.Cleanup(func() {
		if had {
			_ = os.Setenv("APP_ENV", prev)
		} else {
			_ = os.Unsetenv("APP_ENV")
		}
	})

	cfg := Load()
	assert.Equal(t, EnvDevelopment, cfg.AppEnv)
	assert.NoError(t, cfg.Validate())
}

func TestLoadAPP_ENVCaseInsensitive(t *testing.T) {
	t.Setenv("JWT_SECRET", "this-is-a-production-jwt-secret-32+")
	t.Setenv("ADMIN_PASSWORD", "SecureAdmin!234")

	t.Setenv("APP_ENV", "Development")
	cfg := Load()
	assert.Equal(t, EnvDevelopment, cfg.AppEnv)
	assert.NoError(t, cfg.Validate())

	t.Setenv("APP_ENV", "PRODUCTION")
	cfg = Load()
	assert.Equal(t, EnvProduction, cfg.AppEnv)
	assert.NoError(t, cfg.Validate())
}

func TestLoadInvalidAPP_ENVFails(t *testing.T) {
	t.Setenv("JWT_SECRET", DevJWTSecret)
	t.Setenv("ADMIN_PASSWORD", DevAdminPassword)

	for _, bad := range []string{"prodution", "prod", "staging", "dev", " "} {
		t.Run(bad, func(t *testing.T) {
			t.Setenv("APP_ENV", bad)
			cfg := Load()
			err := cfg.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "APP_ENV")
			assert.NotContains(t, err.Error(), DevJWTSecret)
			assert.NotContains(t, err.Error(), DevAdminPassword)
		})
	}
}

func TestLoadEmptyAPP_ENVWhenSetFails(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("JWT_SECRET", DevJWTSecret)
	t.Setenv("ADMIN_PASSWORD", DevAdminPassword)
	cfg := Load()
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "APP_ENV")
}

func TestLoadProductionWithValidSecretsOK(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "this-is-a-production-jwt-secret-32+")
	t.Setenv("ADMIN_PASSWORD", "SecureAdmin!234")
	cfg := Load()
	assert.Equal(t, EnvProduction, cfg.AppEnv)
	assert.NoError(t, cfg.Validate())
}
