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
	t.Setenv("PREVIEW_COOKIE_MODE", "same-site")

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
	t.Setenv("PREVIEW_COOKIE_MODE", "same-site")

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
	t.Setenv("PREVIEW_COOKIE_MODE", "same-site")

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
	t.Setenv("PREVIEW_COOKIE_MODE", "same-site")

	if _, err := Load(); err == nil {
		t.Fatal("expected normalized empty host to fail")
	}
}

func TestLoadPreviewPublicOriginRejectsIPBase(t *testing.T) {
	clearEnv()
	defer clearEnv()
	t.Setenv("JWT_SECRET", "test-jwt")
	t.Setenv("INTERNAL_API_SECRET", "test-internal")
	t.Setenv("PREVIEW_PUBLIC_ORIGIN", "http://127.0.0.1:10000")
	t.Setenv("PREVIEW_COOKIE_MODE", "same-site")

	if _, err := Load(); err == nil {
		t.Fatal("expected IP preview base to fail because per-pod origins require DNS")
	}
}

func TestLoadPreviewCookieModeRequired(t *testing.T) {
	clearEnv()
	defer clearEnv()
	t.Setenv("JWT_SECRET", "test-jwt")
	t.Setenv("INTERNAL_API_SECRET", "test-internal")
	t.Setenv("PREVIEW_PUBLIC_ORIGIN", "https://preview.example.com")

	if _, err := Load(); err == nil {
		t.Fatal("expected PREVIEW_COOKIE_MODE to be required")
	}
}

func TestLoadPartitionedPreviewCookieRequiresHTTPS(t *testing.T) {
	clearEnv()
	defer clearEnv()
	t.Setenv("JWT_SECRET", "test-jwt")
	t.Setenv("INTERNAL_API_SECRET", "test-internal")
	t.Setenv("PREVIEW_PUBLIC_ORIGIN", "http://preview.example.com")
	t.Setenv("PREVIEW_COOKIE_MODE", "partitioned")

	if _, err := Load(); err == nil {
		t.Fatal("expected partitioned preview cookies over HTTP to fail")
	}
}

func TestLoadPartitionedPreviewCookie(t *testing.T) {
	clearEnv()
	defer clearEnv()
	t.Setenv("JWT_SECRET", "test-jwt")
	t.Setenv("INTERNAL_API_SECRET", "test-internal")
	t.Setenv("PREVIEW_PUBLIC_ORIGIN", "https://preview.example.com")
	t.Setenv("PREVIEW_COOKIE_MODE", "partitioned")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PreviewCookieMode != PreviewCookiePartitioned {
		t.Fatalf("PreviewCookieMode = %q", cfg.PreviewCookieMode)
	}
}
