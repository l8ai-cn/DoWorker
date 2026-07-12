package config

import (
	"errors"
	"os"
)

type Config struct {
	HTTPAddress       string
	DatabaseURL       string
	IdentityIssuer    string
	IdentityAudience  string
	IdentityJWKSURL   string
	RuntimeBridgeURL  string
	InternalAPISecret string
}

func Load() (Config, error) {
	return LoadFrom(os.Getenv)
}

func LoadFrom(getenv func(string) string) (Config, error) {
	databaseURL := getenv("MARKETPLACE_DATABASE_URL")
	if databaseURL == "" {
		return Config{}, errors.New("MARKETPLACE_DATABASE_URL is required")
	}
	address := getenv("MARKETPLACE_HTTP_ADDRESS")
	if address == "" {
		address = ":8080"
	}
	issuer := getenv("MARKETPLACE_IDENTITY_ISSUER")
	audience := getenv("MARKETPLACE_IDENTITY_AUDIENCE")
	jwksURL := getenv("MARKETPLACE_IDENTITY_JWKS_URL")
	if issuer == "" || audience == "" || jwksURL == "" {
		return Config{}, errors.New("marketplace identity issuer, audience, and JWKS URL are required")
	}
	runtimeBridgeURL := getenv("MARKETPLACE_RUNTIME_BRIDGE_URL")
	internalSecret := getenv("INTERNAL_API_SECRET")
	if runtimeBridgeURL == "" || internalSecret == "" {
		return Config{}, errors.New("marketplace runtime bridge URL and internal API secret are required")
	}
	return Config{
		HTTPAddress:       address,
		DatabaseURL:       databaseURL,
		IdentityIssuer:    issuer,
		IdentityAudience:  audience,
		IdentityJWKSURL:   jwksURL,
		RuntimeBridgeURL:  runtimeBridgeURL,
		InternalAPISecret: internalSecret,
	}, nil
}
