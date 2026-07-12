package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestAccessTokenManagerSignsRS256AndPublishesJWKS(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	manager, err := NewAccessTokenManager(AccessTokenConfig{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
		KeyID:      "auth-key-2026-07",
		Issuer:     "https://auth.example.com",
		Audiences:  []string{"agentsmesh-api", "marketplace-api"},
		Duration:   time.Hour,
	})
	require.NoError(t, err)

	tokenString, err := manager.GenerateToken(42, "user@example.com", "user", 9, "admin")
	require.NoError(t, err)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		require.Equal(t, "RS256", token.Method.Alg())
		require.Equal(t, "auth-key-2026-07", token.Header["kid"])
		return &privateKey.PublicKey, nil
	})
	require.NoError(t, err)
	require.True(t, token.Valid)

	claims, err := manager.ValidateToken(tokenString, "marketplace-api")
	require.NoError(t, err)
	require.Equal(t, int64(42), claims.UserID)
	require.Equal(t, "https://auth.example.com", claims.Issuer)
	require.Contains(t, []string(claims.Audience), "marketplace-api")

	jwks := manager.JWKS()
	require.Len(t, jwks.Keys, 1)
	require.Equal(t, "RSA", jwks.Keys[0].KeyType)
	require.Equal(t, "RS256", jwks.Keys[0].Algorithm)
	require.Equal(t, "auth-key-2026-07", jwks.Keys[0].KeyID)
	require.Equal(t, base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes()), jwks.Keys[0].Modulus)
	require.Equal(t, testEncodeRSAExponent(privateKey.PublicKey.E), jwks.Keys[0].Exponent)
}

func TestAccessTokenManagerRejectsWrongAudienceAlgorithmAndKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	manager, err := NewAccessTokenManager(AccessTokenConfig{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
		KeyID:      "auth-key",
		Issuer:     "agentsmesh",
		Audiences:  []string{"agentsmesh-api"},
		Duration:   time.Hour,
	})
	require.NoError(t, err)
	tokenString, err := manager.GenerateToken(42, "user@example.com", "user", 9, "admin")
	require.NoError(t, err)

	_, err = manager.ValidateToken(tokenString, "marketplace-api")
	require.ErrorIs(t, err, ErrInvalidToken)

	hsToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: 42,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			Issuer:    "agentsmesh",
			Audience:  jwt.ClaimStrings{"agentsmesh-api"},
		},
	})
	hsToken.Header["kid"] = "auth-key"
	hsString, err := hsToken.SignedString([]byte("shared-secret"))
	require.NoError(t, err)
	_, err = manager.ValidateToken(hsString, "agentsmesh-api")
	require.ErrorIs(t, err, ErrInvalidToken)

	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	otherManager, err := NewAccessTokenManager(AccessTokenConfig{
		PublicKey: &otherKey.PublicKey,
		KeyID:     "auth-key",
		Issuer:    "agentsmesh",
		Audiences: []string{"agentsmesh-api"},
		Duration:  time.Hour,
	})
	require.NoError(t, err)
	_, err = otherManager.ValidateToken(tokenString, "agentsmesh-api")
	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestAccessTokenManagerRequiresExplicitKeysAndAudience(t *testing.T) {
	_, err := NewAccessTokenManager(AccessTokenConfig{})
	require.Error(t, err)

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	_, err = NewAccessTokenManager(AccessTokenConfig{
		PublicKey: &privateKey.PublicKey,
		KeyID:     "auth-key",
		Issuer:    "agentsmesh",
	})
	require.Error(t, err)

	weakKey, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)
	_, err = NewAccessTokenManager(AccessTokenConfig{
		PublicKey: &weakKey.PublicKey,
		KeyID:     "auth-key",
		Issuer:    "agentsmesh",
		Audiences: []string{"agentsmesh-api"},
		Duration:  time.Hour,
	})
	require.ErrorIs(t, err, ErrAccessTokenConfig)
}

func TestParseAccessTokenPEMKeys(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER})
	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)
	publicPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})

	parsedPrivate, err := ParseRSAPrivateKeyPEM(privatePEM)
	require.NoError(t, err)
	parsedPublic, err := ParseRSAPublicKeyPEM(publicPEM)
	require.NoError(t, err)
	require.Zero(t, parsedPrivate.PublicKey.N.Cmp(parsedPublic.N))

	_, err = ParseRSAPrivateKeyPEM([]byte("not pem"))
	require.ErrorIs(t, err, ErrAccessTokenConfig)
	_, err = ParseRSAPublicKeyPEM(privatePEM)
	require.ErrorIs(t, err, ErrAccessTokenConfig)
}

func testEncodeRSAExponent(exponent int) string {
	return base64.RawURLEncoding.EncodeToString(big.NewInt(int64(exponent)).Bytes())
}
