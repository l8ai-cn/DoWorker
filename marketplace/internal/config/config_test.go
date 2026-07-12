package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadRequiresMarketplaceDatabaseURL(t *testing.T) {
	_, err := LoadFrom(func(string) string { return "" })
	require.Error(t, err)
}

func TestLoadUsesIndependentAddress(t *testing.T) {
	values := map[string]string{
		"MARKETPLACE_DATABASE_URL":       "postgres://marketplace:test@localhost/marketplace",
		"MARKETPLACE_HTTP_ADDRESS":       ":18080",
		"MARKETPLACE_IDENTITY_ISSUER":    "https://dowork.l8ai.cn",
		"MARKETPLACE_IDENTITY_AUDIENCE":  "marketplace-api",
		"MARKETPLACE_IDENTITY_JWKS_URL":  "https://dowork.l8ai.cn/.well-known/jwks.json",
		"MARKETPLACE_RUNTIME_BRIDGE_URL": "http://backend:8080/api/internal/marketplace/installations",
		"INTERNAL_API_SECRET":            "secret",
	}
	cfg, err := LoadFrom(func(key string) string { return values[key] })
	require.NoError(t, err)
	require.Equal(t, ":18080", cfg.HTTPAddress)
	require.Equal(t, "https://dowork.l8ai.cn", cfg.IdentityIssuer)
	require.Equal(t, "marketplace-api", cfg.IdentityAudience)
	require.Equal(t, "https://dowork.l8ai.cn/.well-known/jwks.json", cfg.IdentityJWKSURL)
	require.Equal(t, "http://backend:8080/api/internal/marketplace/installations", cfg.RuntimeBridgeURL)
}

func TestLoadRequiresIdentityConfiguration(t *testing.T) {
	values := map[string]string{
		"MARKETPLACE_DATABASE_URL": "postgres://marketplace:test@localhost/marketplace",
	}
	_, err := LoadFrom(func(key string) string { return values[key] })
	require.Error(t, err)
}
