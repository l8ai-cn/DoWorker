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
		"MARKETPLACE_DATABASE_URL": "postgres://marketplace:test@localhost/marketplace",
		"MARKETPLACE_HTTP_ADDRESS": ":18080",
	}
	cfg, err := LoadFrom(func(key string) string { return values[key] })
	require.NoError(t, err)
	require.Equal(t, ":18080", cfg.HTTPAddress)
}
