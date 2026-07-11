package auth

import (
	"fmt"
	"os"
	"time"
)

type AccessTokenFileConfig struct {
	PrivateKeyFile string
	PublicKeyFile  string
	KeyID          string
	Issuer         string
	Audiences      []string
	Duration       time.Duration
}

func LoadAccessTokenManager(config AccessTokenFileConfig) (*AccessTokenManager, error) {
	privatePEM, err := os.ReadFile(config.PrivateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("read access token private key: %w", err)
	}
	publicPEM, err := os.ReadFile(config.PublicKeyFile)
	if err != nil {
		return nil, fmt.Errorf("read access token public key: %w", err)
	}
	privateKey, err := ParseRSAPrivateKeyPEM(privatePEM)
	if err != nil {
		return nil, fmt.Errorf("parse access token private key: %w", err)
	}
	publicKey, err := ParseRSAPublicKeyPEM(publicPEM)
	if err != nil {
		return nil, fmt.Errorf("parse access token public key: %w", err)
	}
	manager, err := NewAccessTokenManager(AccessTokenConfig{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		KeyID:      config.KeyID,
		Issuer:     config.Issuer,
		Audiences:  config.Audiences,
		Duration:   config.Duration,
	})
	if err != nil {
		return nil, fmt.Errorf("configure access token manager: %w", err)
	}
	return manager, nil
}
