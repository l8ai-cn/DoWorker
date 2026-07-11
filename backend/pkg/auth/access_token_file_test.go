package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadAccessTokenManager(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)

	directory := t.TempDir()
	privatePath := filepath.Join(directory, "private.pem")
	publicPath := filepath.Join(directory, "public.pem")
	require.NoError(t, os.WriteFile(
		privatePath,
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER}),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		publicPath,
		pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER}),
		0o600,
	))

	manager, err := LoadAccessTokenManager(AccessTokenFileConfig{
		PrivateKeyFile: privatePath,
		PublicKeyFile:  publicPath,
		KeyID:          "core-2026-07",
		Issuer:         "agentsmesh",
		Audiences:      []string{"agentsmesh-api"},
		Duration:       time.Hour,
	})
	require.NoError(t, err)

	token, err := manager.GenerateToken(42, "user@example.com", "user", 0, "")
	require.NoError(t, err)
	claims, err := manager.ValidateToken(token, "agentsmesh-api")
	require.NoError(t, err)
	require.Equal(t, int64(42), claims.UserID)
}

func TestLoadAccessTokenManagerRejectsMissingKeyFile(t *testing.T) {
	_, err := LoadAccessTokenManager(AccessTokenFileConfig{
		PrivateKeyFile: filepath.Join(t.TempDir(), "missing-private.pem"),
		PublicKeyFile:  filepath.Join(t.TempDir(), "missing-public.pem"),
		KeyID:          "core-2026-07",
		Issuer:         "agentsmesh",
		Audiences:      []string{"agentsmesh-api"},
		Duration:       time.Hour,
	})
	require.Error(t, err)
}
