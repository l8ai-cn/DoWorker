package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublicWebBaseURLRequiresExplicitConfig(t *testing.T) {
	cfg := &Config{
		PrimaryDomain: "api.example.com",
		PublicWebURL:  "https://app.example.com/",
		UseHTTPS:      true,
	}

	require.Equal(t, "https://api.example.com", cfg.FrontendURL())
	require.Equal(t, "https://app.example.com", cfg.PublicWebBaseURL())

	cfg.PublicWebURL = ""
	require.Empty(t, cfg.PublicWebBaseURL())
}
