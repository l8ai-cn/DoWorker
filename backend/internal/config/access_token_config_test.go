package config

import (
	"reflect"
	"testing"
)

func TestLoadAccessTokenConfig(t *testing.T) {
	t.Setenv("ACCESS_TOKEN_PRIVATE_KEY_FILE", "/run/secrets/access-token-private.pem")
	t.Setenv("ACCESS_TOKEN_PUBLIC_KEY_FILE", "/run/config/access-token-public.pem")
	t.Setenv("ACCESS_TOKEN_KEY_ID", "core-2026-07")
	t.Setenv("ACCESS_TOKEN_ISSUER", "https://dowork.l8ai.cn")
	t.Setenv("ACCESS_TOKEN_AUDIENCES", "agentsmesh-api,marketplace-api")
	t.Setenv("ACCESS_TOKEN_CORE_AUDIENCE", "agentsmesh-api")
	t.Setenv("ACCESS_TOKEN_EXPIRATION_HOURS", "12")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	wantAudiences := []string{"agentsmesh-api", "marketplace-api"}
	if cfg.AccessToken.PrivateKeyFile != "/run/secrets/access-token-private.pem" {
		t.Errorf("PrivateKeyFile = %q", cfg.AccessToken.PrivateKeyFile)
	}
	if cfg.AccessToken.PublicKeyFile != "/run/config/access-token-public.pem" {
		t.Errorf("PublicKeyFile = %q", cfg.AccessToken.PublicKeyFile)
	}
	if cfg.AccessToken.KeyID != "core-2026-07" {
		t.Errorf("KeyID = %q", cfg.AccessToken.KeyID)
	}
	if cfg.AccessToken.Issuer != "https://dowork.l8ai.cn" {
		t.Errorf("Issuer = %q", cfg.AccessToken.Issuer)
	}
	if !reflect.DeepEqual(cfg.AccessToken.Audiences, wantAudiences) {
		t.Errorf("Audiences = %#v", cfg.AccessToken.Audiences)
	}
	if cfg.AccessToken.CoreAudience != "agentsmesh-api" {
		t.Errorf("CoreAudience = %q", cfg.AccessToken.CoreAudience)
	}
	if cfg.AccessToken.ExpirationHours != 12 {
		t.Errorf("ExpirationHours = %d", cfg.AccessToken.ExpirationHours)
	}
}
