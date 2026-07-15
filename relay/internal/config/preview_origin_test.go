package config

import (
	"testing"
)

func TestLoadPreviewPublicOrigin(t *testing.T) {
	clearEnv()
	defer clearEnv()
	t.Setenv("JWT_SECRET", "test-jwt")
	t.Setenv("INTERNAL_API_SECRET", "test-internal")
	t.Setenv("PRIMARY_DOMAIN", "app.example.com")
	t.Setenv("USE_HTTPS", "true")
	t.Setenv("PREVIEW_PUBLIC_ORIGIN", "HTTPS://Preview.Example.com.:443/")

	cfg, err := Load()

	if err != nil {
		t.Fatal(err)
	}
	if cfg.PreviewPublicOrigin != "https://preview.example.com" {
		t.Fatalf("PreviewPublicOrigin = %q", cfg.PreviewPublicOrigin)
	}
}

func TestLoadPreviewPublicOriginRequired(t *testing.T) {
	clearEnv()
	defer clearEnv()
	t.Setenv("JWT_SECRET", "test-jwt")
	t.Setenv("INTERNAL_API_SECRET", "test-internal")

	if _, err := Load(); err == nil {
		t.Fatal("expected PREVIEW_PUBLIC_ORIGIN to be required")
	}
}

func TestLoadPreviewPublicOriginRejectsApplicationOrigin(t *testing.T) {
	clearEnv()
	defer clearEnv()
	t.Setenv("JWT_SECRET", "test-jwt")
	t.Setenv("INTERNAL_API_SECRET", "test-internal")
	t.Setenv("PRIMARY_DOMAIN", "app.example.com")
	t.Setenv("USE_HTTPS", "true")
	t.Setenv("PREVIEW_PUBLIC_ORIGIN", "https://app.example.com")

	if _, err := Load(); err == nil {
		t.Fatal("expected application origin collision to fail")
	}
}

func TestLoadPreviewPublicOriginRejectsEmptyNormalizedHost(t *testing.T) {
	clearEnv()
	defer clearEnv()
	t.Setenv("JWT_SECRET", "test-jwt")
	t.Setenv("INTERNAL_API_SECRET", "test-internal")
	t.Setenv("PREVIEW_PUBLIC_ORIGIN", "https://.")

	if _, err := Load(); err == nil {
		t.Fatal("expected normalized empty host to fail")
	}
}
