package rest

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authpkg "github.com/l8ai-cn/agentcloud/backend/pkg/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterJWKSRoute(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	manager, err := authpkg.NewAccessTokenManager(authpkg.AccessTokenConfig{
		PublicKey: &privateKey.PublicKey,
		KeyID:     "core-2026-07",
		Issuer:    "agentcloud",
		Audiences: []string{"agentcloud-api"},
		Duration:  time.Hour,
	})
	require.NoError(t, err)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerJWKSRoute(router, manager)

	request := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "public, max-age=300", response.Header().Get("Cache-Control"))
	var body authpkg.JSONWebKeySet
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Len(t, body.Keys, 1)
	assert.Equal(t, "core-2026-07", body.Keys[0].KeyID)
	assert.Equal(t, "RS256", body.Keys[0].Algorithm)
	assert.NotEmpty(t, body.Keys[0].Modulus)
	assert.NotEmpty(t, body.Keys[0].Exponent)
}
