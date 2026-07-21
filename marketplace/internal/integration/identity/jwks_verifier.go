package identity

import (
	"context"
	"crypto/rsa"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	authpkg "github.com/l8ai-cn/agentcloud/backend/pkg/auth"
	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidAccessToken = errors.New("invalid access token")

type JWKSConfig struct {
	URL             string
	Issuer          string
	Audience        string
	RefreshInterval time.Duration
}

type JWKSVerifier struct {
	config     JWKSConfig
	client     *http.Client
	mu         sync.Mutex
	keys       map[string]*rsa.PublicKey
	expiresAt  time.Time
	refreshing chan struct{}
	refreshErr error
	retryAt    time.Time
}

func NewJWKSVerifier(config JWKSConfig, client *http.Client) (*JWKSVerifier, error) {
	if strings.TrimSpace(config.URL) == "" ||
		strings.TrimSpace(config.Issuer) == "" ||
		strings.TrimSpace(config.Audience) == "" ||
		config.RefreshInterval <= 0 ||
		client == nil {
		return nil, errors.New("invalid JWKS verifier configuration")
	}
	return &JWKSVerifier{config: config, client: client}, nil
}

func (v *JWKSVerifier) Verify(ctx context.Context, tokenString string) (*authpkg.Claims, error) {
	claims := &authpkg.Claims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodRS256 {
				return nil, ErrInvalidAccessToken
			}
			keyID, ok := token.Header["kid"].(string)
			if !ok || keyID == "" {
				return nil, ErrInvalidAccessToken
			}
			return v.key(ctx, keyID)
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
		jwt.WithIssuer(v.config.Issuer),
		jwt.WithAudience(v.config.Audience),
		jwt.WithExpirationRequired(),
	)
	if err != nil || !token.Valid {
		return nil, ErrInvalidAccessToken
	}
	return claims, nil
}

func (v *JWKSVerifier) key(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	keys, fresh := v.cachedKeys()
	if key := keys[keyID]; key != nil && fresh {
		return key, nil
	}
	if fresh {
		return nil, ErrInvalidAccessToken
	}
	keys, err := v.refreshKeys(ctx)
	if err != nil {
		return nil, err
	}
	if key := keys[keyID]; key != nil {
		return key, nil
	}
	return nil, ErrInvalidAccessToken
}

func (v *JWKSVerifier) cachedKeys() (map[string]*rsa.PublicKey, bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.keys, len(v.keys) > 0 && time.Now().Before(v.expiresAt)
}

func (v *JWKSVerifier) refreshKeys(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	v.mu.Lock()
	if len(v.keys) > 0 && time.Now().Before(v.expiresAt) {
		keys := v.keys
		v.mu.Unlock()
		return keys, nil
	}
	if time.Now().Before(v.retryAt) {
		err := v.refreshErr
		v.mu.Unlock()
		return nil, err
	}
	if v.refreshing != nil {
		done := v.refreshing
		v.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-done:
		}
		v.mu.Lock()
		defer v.mu.Unlock()
		if v.refreshErr != nil {
			return nil, v.refreshErr
		}
		return v.keys, nil
	}
	v.refreshing = make(chan struct{})
	done := v.refreshing
	v.mu.Unlock()

	keys, err := fetchKeys(ctx, v.client, v.config.URL)
	v.mu.Lock()
	if err == nil {
		v.keys = keys
		v.expiresAt = time.Now().Add(v.config.RefreshInterval)
		v.retryAt = time.Time{}
	} else {
		v.retryAt = time.Now().Add(5 * time.Second)
	}
	v.refreshErr = err
	close(done)
	v.refreshing = nil
	v.mu.Unlock()
	if err != nil {
		return nil, err
	}
	return keys, nil
}
