package middleware_test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	authpkg "github.com/anthropics/agentsmesh/backend/pkg/auth"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const middlewareAudience = "agentsmesh-api"

func newAccessTokenManager(t *testing.T) (*authpkg.AccessTokenManager, *rsa.PrivateKey) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	manager, err := authpkg.NewAccessTokenManager(authpkg.AccessTokenConfig{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
		KeyID:      "middleware-test-key",
		Issuer:     "middleware-test",
		Audiences:  []string{middlewareAudience},
		Duration:   time.Hour,
	})
	require.NoError(t, err)
	return manager, privateKey
}

func performRequest(handler gin.HandlerFunc, authorization, queryToken string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/", handler, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id":  c.GetInt64("user_id"),
			"username": c.GetString("username"),
		})
	})
	request := httptest.NewRequest(http.MethodGet, "/?token="+queryToken, nil)
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestAuthMiddlewareAcceptsRS256AccessToken(t *testing.T) {
	manager, _ := newAccessTokenManager(t)
	token, err := manager.GenerateToken(42, "user@example.com", "user", 9, "admin")
	require.NoError(t, err)

	response := performRequest(
		middleware.AuthMiddleware(manager, middlewareAudience),
		"Bearer "+token,
		"",
	)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"user_id":42,"username":"user"}`, response.Body.String())
}

func TestAuthMiddlewareAcceptsQueryToken(t *testing.T) {
	manager, _ := newAccessTokenManager(t)
	token, err := manager.GenerateToken(7, "query@example.com", "query-user", 0, "")
	require.NoError(t, err)

	response := performRequest(middleware.AuthMiddleware(manager, middlewareAudience), "", token)

	assert.Equal(t, http.StatusOK, response.Code)
}

func TestAuthMiddlewareRejectsWrongAudience(t *testing.T) {
	manager, _ := newAccessTokenManager(t)
	token, err := manager.GenerateToken(42, "user@example.com", "user", 0, "")
	require.NoError(t, err)

	response := performRequest(middleware.AuthMiddleware(manager, "marketplace-api"), "Bearer "+token, "")

	assert.Equal(t, http.StatusUnauthorized, response.Code)
}

func TestAuthMiddlewareRejectsHS256Token(t *testing.T) {
	manager, _ := newAccessTokenManager(t)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 42,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte("legacy-secret"))
	require.NoError(t, err)

	response := performRequest(middleware.AuthMiddleware(manager, middlewareAudience), "Bearer "+token, "")

	assert.Equal(t, http.StatusUnauthorized, response.Code)
}

func TestOptionalAuthMiddlewareIgnoresInvalidToken(t *testing.T) {
	manager, _ := newAccessTokenManager(t)

	response := performRequest(
		middleware.OptionalAuthMiddleware(manager, middlewareAudience),
		"Bearer invalid",
		"",
	)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"user_id":0,"username":""}`, response.Body.String())
}
