package interceptors_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/l8ai-cn/agentcloud/backend/internal/api/connect/interceptors"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	authpkg "github.com/l8ai-cn/agentcloud/backend/pkg/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const connectAudience = "agentcloud-api"

type echoReq struct{ Msg string }
type echoRes struct{ Echo string }

type tokenFixture struct {
	manager    *authpkg.AccessTokenManager
	privateKey *rsa.PrivateKey
}

func newTokenFixture(t *testing.T) tokenFixture {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	manager, err := authpkg.NewAccessTokenManager(authpkg.AccessTokenConfig{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
		KeyID:      "connect-test-key",
		Issuer:     "connect-test",
		Audiences:  []string{connectAudience},
		Duration:   time.Hour,
	})
	require.NoError(t, err)
	return tokenFixture{manager: manager, privateKey: privateKey}
}

func (f tokenFixture) issue(t *testing.T, userID int64) string {
	t.Helper()
	token, err := f.manager.GenerateToken(userID, "u@example.com", "user", 0, "")
	require.NoError(t, err)
	return token
}

func (f tokenFixture) issueExpired(t *testing.T, userID int64) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, middleware.JWTClaims{
		UserID:   userID,
		Email:    "u@example.com",
		Username: "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "connect-test",
			Audience:  jwt.ClaimStrings{connectAudience},
		},
	})
	token.Header["kid"] = "connect-test-key"
	value, err := token.SignedString(f.privateKey)
	require.NoError(t, err)
	return value
}

func runInterceptor(
	t *testing.T,
	manager *authpkg.AccessTokenManager,
	audience string,
	header string,
	next connect.UnaryFunc,
) (connect.AnyResponse, error) {
	t.Helper()
	interceptor := interceptors.NewAuthInterceptor(manager, audience)
	req := connect.NewRequest(&echoReq{Msg: "hi"})
	if header != "" {
		req.Header().Set("Authorization", header)
	}
	return interceptor.WrapUnary(next)(context.Background(), req)
}

func okHandler(capture *context.Context) connect.UnaryFunc {
	return func(ctx context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		*capture = ctx
		return connect.NewResponse(&echoRes{Echo: "ok"}), nil
	}
}

func TestAuthInterceptorValidTokenPopulatesContext(t *testing.T) {
	fixture := newTokenFixture(t)
	var captured context.Context

	response, err := runInterceptor(
		t,
		fixture.manager,
		connectAudience,
		"Bearer "+fixture.issue(t, 42),
		okHandler(&captured),
	)

	require.NoError(t, err)
	require.NotNil(t, response)
	tenant := middleware.GetTenant(captured)
	require.NotNil(t, tenant)
	assert.Equal(t, int64(42), tenant.UserID)
	claims := interceptors.ClaimsFromContext(captured)
	require.NotNil(t, claims)
	assert.Equal(t, "u@example.com", claims.Email)
}

func TestAuthInterceptorRejectsMissingOrInvalidToken(t *testing.T) {
	fixture := newTokenFixture(t)
	otherFixture := newTokenFixture(t)
	cases := map[string]string{
		"missing":        "",
		"malformed":      "Bearer not-a-jwt",
		"wrong key":      "Bearer " + otherFixture.issue(t, 42),
		"missing scheme": fixture.issue(t, 42),
		"empty bearer":   "Bearer ",
		"non bearer":     "Basic dXNlcjpwYXNz",
		"expired":        "Bearer " + fixture.issueExpired(t, 42),
	}

	for name, header := range cases {
		t.Run(name, func(t *testing.T) {
			called := false
			response, err := runInterceptor(
				t,
				fixture.manager,
				connectAudience,
				header,
				func(context.Context, connect.AnyRequest) (connect.AnyResponse, error) {
					called = true
					return nil, nil
				},
			)
			require.Error(t, err)
			assert.Nil(t, response)
			assert.False(t, called)
			var connectErr *connect.Error
			require.True(t, errors.As(err, &connectErr))
			assert.Equal(t, connect.CodeUnauthenticated, connectErr.Code())
		})
	}
}

func TestAuthInterceptorRejectsWrongAudience(t *testing.T) {
	fixture := newTokenFixture(t)

	_, err := runInterceptor(
		t,
		fixture.manager,
		"marketplace-api",
		"Bearer "+fixture.issue(t, 42),
		okHandler(new(context.Context)),
	)

	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeUnauthenticated, connectErr.Code())
}

func TestAuthInterceptorPassesThroughHandlerErrors(t *testing.T) {
	fixture := newTokenFixture(t)
	downstreamErr := connect.NewError(connect.CodeNotFound, errors.New("resource missing"))

	response, err := runInterceptor(
		t,
		fixture.manager,
		connectAudience,
		"Bearer "+fixture.issue(t, 42),
		func(context.Context, connect.AnyRequest) (connect.AnyResponse, error) {
			return nil, downstreamErr
		},
	)

	assert.Nil(t, response)
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeNotFound, connectErr.Code())
}
