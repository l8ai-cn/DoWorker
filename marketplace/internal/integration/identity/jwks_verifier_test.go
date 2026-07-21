package identity

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	authpkg "github.com/l8ai-cn/agentcloud/backend/pkg/auth"
	"github.com/stretchr/testify/require"
)

func TestJWKSVerifierValidatesCoreAccessToken(t *testing.T) {
	manager := newCoreTokenManager(t, []string{"marketplace-api"})
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		require.NoError(t, json.NewEncoder(w).Encode(manager.JWKS()))
	}))
	t.Cleanup(server.Close)
	verifier, err := NewJWKSVerifier(JWKSConfig{
		URL:             server.URL,
		Issuer:          "core-auth",
		Audience:        "marketplace-api",
		RefreshInterval: time.Minute,
	}, server.Client())
	require.NoError(t, err)
	token, err := manager.GenerateToken(42, "user@example.com", "user", 9, "admin")
	require.NoError(t, err)

	first, err := verifier.Verify(context.Background(), token)
	require.NoError(t, err)
	second, err := verifier.Verify(context.Background(), token)
	require.NoError(t, err)

	require.Equal(t, int64(42), first.UserID)
	require.Equal(t, int64(9), first.OrganizationID)
	require.Equal(t, first.UserID, second.UserID)
	require.Equal(t, int32(1), requests.Load())
}

func TestJWKSVerifierRejectsWrongAudience(t *testing.T) {
	manager := newCoreTokenManager(t, []string{"agentcloud-api"})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode(manager.JWKS()))
	}))
	t.Cleanup(server.Close)
	verifier, err := NewJWKSVerifier(JWKSConfig{
		URL:             server.URL,
		Issuer:          "core-auth",
		Audience:        "marketplace-api",
		RefreshInterval: time.Minute,
	}, server.Client())
	require.NoError(t, err)
	token, err := manager.GenerateToken(42, "user@example.com", "user", 0, "")
	require.NoError(t, err)

	_, err = verifier.Verify(context.Background(), token)

	require.Error(t, err)
}

func TestJWKSVerifierDoesNotRefreshForUnknownKeyWhileCacheIsFresh(t *testing.T) {
	manager := newCoreTokenManager(t, []string{"marketplace-api"})
	unknownManager := newCoreTokenManagerWithKey(t, []string{"marketplace-api"}, "unknown-key")
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		require.NoError(t, json.NewEncoder(w).Encode(manager.JWKS()))
	}))
	t.Cleanup(server.Close)
	verifier, err := NewJWKSVerifier(JWKSConfig{
		URL:             server.URL,
		Issuer:          "core-auth",
		Audience:        "marketplace-api",
		RefreshInterval: time.Minute,
	}, server.Client())
	require.NoError(t, err)
	validToken, err := manager.GenerateToken(42, "user@example.com", "user", 0, "")
	require.NoError(t, err)
	_, err = verifier.Verify(context.Background(), validToken)
	require.NoError(t, err)
	unknownToken, err := unknownManager.GenerateToken(42, "user@example.com", "user", 0, "")
	require.NoError(t, err)

	for range 3 {
		_, err = verifier.Verify(context.Background(), unknownToken)
		require.Error(t, err)
	}

	require.Equal(t, int32(1), requests.Load())
}

func TestJWKSVerifierRejectsInvalidConfiguration(t *testing.T) {
	_, err := NewJWKSVerifier(JWKSConfig{}, http.DefaultClient)
	require.Error(t, err)
}

func newCoreTokenManager(t *testing.T, audiences []string) *authpkg.AccessTokenManager {
	return newCoreTokenManagerWithKey(t, audiences, "core-test-key")
}

func newCoreTokenManagerWithKey(
	t *testing.T,
	audiences []string,
	keyID string,
) *authpkg.AccessTokenManager {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	manager, err := authpkg.NewAccessTokenManager(authpkg.AccessTokenConfig{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
		KeyID:      keyID,
		Issuer:     "core-auth",
		Audiences:  audiences,
		Duration:   time.Hour,
	})
	require.NoError(t, err)
	return manager
}
