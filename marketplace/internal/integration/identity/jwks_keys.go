package identity

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"

	authpkg "github.com/l8ai-cn/agentcloud/backend/pkg/auth"
)

func fetchKeys(
	ctx context.Context,
	client *http.Client,
	url string,
) (map[string]*rsa.PublicKey, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create JWKS request: %w", err)
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch JWKS: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch JWKS: status %d", response.StatusCode)
	}
	var set authpkg.JSONWebKeySet
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&set); err != nil {
		return nil, fmt.Errorf("decode JWKS: %w", err)
	}
	return parseKeys(set)
}

func parseKeys(set authpkg.JSONWebKeySet) (map[string]*rsa.PublicKey, error) {
	keys := make(map[string]*rsa.PublicKey, len(set.Keys))
	for _, key := range set.Keys {
		if key.KeyType != "RSA" || key.Use != "sig" || key.Algorithm != "RS256" || key.KeyID == "" {
			continue
		}
		modulus, err := base64.RawURLEncoding.DecodeString(key.Modulus)
		if err != nil {
			return nil, fmt.Errorf("decode JWKS modulus: %w", err)
		}
		exponent, err := base64.RawURLEncoding.DecodeString(key.Exponent)
		if err != nil {
			return nil, fmt.Errorf("decode JWKS exponent: %w", err)
		}
		exponentValue := new(big.Int).SetBytes(exponent)
		if !exponentValue.IsInt64() || exponentValue.Sign() <= 0 {
			return nil, errors.New("invalid JWKS exponent")
		}
		publicKey := &rsa.PublicKey{N: new(big.Int).SetBytes(modulus), E: int(exponentValue.Int64())}
		if publicKey.N.BitLen() < 2048 || publicKey.E < 3 {
			return nil, errors.New("invalid JWKS RSA key")
		}
		keys[key.KeyID] = publicKey
	}
	if len(keys) == 0 {
		return nil, errors.New("JWKS contains no usable signing key")
	}
	return keys, nil
}
