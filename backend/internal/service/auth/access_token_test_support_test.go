package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	authpkg "github.com/l8ai-cn/agentcloud/backend/pkg/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

const testAccessTokenAudience = "agentcloud-api"

type testAccessTokenFixture struct {
	manager    *authpkg.AccessTokenManager
	privateKey *rsa.PrivateKey
}

func newTestAccessTokenFixture(
	t *testing.T,
	issuer string,
	duration time.Duration,
) testAccessTokenFixture {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	manager, err := authpkg.NewAccessTokenManager(authpkg.AccessTokenConfig{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
		KeyID:      "test-access-token-key",
		Issuer:     issuer,
		Audiences:  []string{testAccessTokenAudience, "marketplace-api"},
		Duration:   duration,
	})
	require.NoError(t, err)
	return testAccessTokenFixture{manager: manager, privateKey: privateKey}
}

func configureTestAccessTokens(t *testing.T, config *Config) testAccessTokenFixture {
	t.Helper()
	fixture := newTestAccessTokenFixture(t, config.Issuer, config.JWTExpiration)
	config.AccessTokens = fixture.manager
	config.AccessTokenAudience = testAccessTokenAudience
	return fixture
}

func signExpiredTestAccessToken(
	t *testing.T,
	fixture testAccessTokenFixture,
	userID int64,
) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "test-issuer",
			Audience:  jwt.ClaimStrings{testAccessTokenAudience},
		},
	})
	token.Header["kid"] = "test-access-token-key"
	value, err := token.SignedString(fixture.privateKey)
	require.NoError(t, err)
	return value
}
