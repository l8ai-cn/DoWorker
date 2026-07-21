package actor

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	authpkg "github.com/l8ai-cn/agentcloud/backend/pkg/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type verifierStub struct {
	claims *authpkg.Claims
	err    error
}

func (s verifierStub) Verify(context.Context, string) (*authpkg.Claims, error) {
	return s.claims, s.err
}

func TestMiddlewareInjectsActor(t *testing.T) {
	router := actorRouter(verifierStub{claims: &authpkg.Claims{
		UserID:         42,
		Email:          "user@example.com",
		Username:       "user",
		OrganizationID: 9,
		Role:           "admin",
	}})
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"user_id":42,"organization_id":9,"role":"admin"}`, response.Body.String())
}

func TestMiddlewareRejectsMissingOrInvalidToken(t *testing.T) {
	cases := map[string]struct {
		verifier TokenVerifier
		header   string
	}{
		"missing": {verifier: verifierStub{}},
		"invalid": {
			verifier: verifierStub{err: errors.New("invalid token")},
			header:   "Bearer invalid",
		},
		"query token": {
			verifier: verifierStub{claims: &authpkg.Claims{UserID: 42}},
		},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			router := actorRouter(testCase.verifier)
			path := "/"
			if name == "query token" {
				path = "/?token=access-token"
			}
			request := httptest.NewRequest(http.MethodGet, path, nil)
			if testCase.header != "" {
				request.Header.Set("Authorization", testCase.header)
			}
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			assert.Equal(t, http.StatusUnauthorized, response.Code)
		})
	}
}

func actorRouter(verifier TokenVerifier) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Middleware(verifier))
	router.GET("/", func(c *gin.Context) {
		current, ok := FromContext(c)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"user_id":         current.UserID,
			"organization_id": current.OrganizationID,
			"role":            current.Role,
		})
	})
	return router
}
