package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadPreviewPublicOrigin(t *testing.T) {
	t.Setenv("PREVIEW_PUBLIC_ORIGIN", "HTTPS://Preview.Example.com.:443/")
	t.Setenv("PRIMARY_DOMAIN", "api.example.com")
	t.Setenv("PUBLIC_WEB_URL", "https://app.example.com")
	t.Setenv("MOBILE_PUBLIC_BASE_URL", "https://mobile.example.com")

	cfg, err := Load()

	require.NoError(t, err)
	require.Equal(t, "https://preview.example.com", cfg.PreviewPublicOrigin)
}

func TestLoadPreviewPublicOriginRequired(t *testing.T) {
	t.Setenv("PREVIEW_PUBLIC_ORIGIN", "")

	_, err := Load()

	require.ErrorContains(t, err, "PREVIEW_PUBLIC_ORIGIN")
	require.ErrorContains(t, err, "is required")
}

func TestLoadPreviewPublicOriginRejectsNonOriginURL(t *testing.T) {
	for _, raw := range []string{
		"ftp://preview.example.com",
		"https://preview.example.com/path",
		"https://user@preview.example.com",
		"https://preview.example.com?mode=app",
		"http://127.0.0.1:10000",
		"https://.",
	} {
		t.Run(raw, func(t *testing.T) {
			t.Setenv("PREVIEW_PUBLIC_ORIGIN", raw)

			_, err := Load()

			require.ErrorContains(t, err, "PREVIEW_PUBLIC_ORIGIN")
		})
	}
}

func TestLoadPreviewPublicOriginRejectsAuthenticatedApplicationOrigins(t *testing.T) {
	for name, env := range map[string]map[string]string{
		"primary": {
			"PREVIEW_PUBLIC_ORIGIN": "https://api.example.com.:443",
			"PRIMARY_DOMAIN":        "api.example.com",
			"USE_HTTPS":             "true",
		},
		"web": {
			"PREVIEW_PUBLIC_ORIGIN": "https://app.example.com",
			"PUBLIC_WEB_URL":        "https://app.example.com/",
		},
		"mobile": {
			"PREVIEW_PUBLIC_ORIGIN":  "https://mobile.example.com",
			"MOBILE_PUBLIC_BASE_URL": "https://mobile.example.com/",
		},
	} {
		t.Run(name, func(t *testing.T) {
			for key, value := range env {
				t.Setenv(key, value)
			}

			_, err := Load()

			require.ErrorContains(t, err, "must use a dedicated origin")
		})
	}
}
