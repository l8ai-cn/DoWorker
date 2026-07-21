package console

import (
	"net/http"
	"net/http/httptest"
	"testing"

	actorapi "github.com/l8ai-cn/agentcloud/marketplace/internal/api/actor"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetSessionReturnsCurrentActor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("marketplace_actor", actorapi.Actor{
			UserID:         42,
			Email:          "user@example.com",
			Username:       "user",
			OrganizationID: 9,
			Role:           "admin",
		})
	})
	NewSessionHandler().RegisterRoutes(router.Group("/api/marketplace/v1/console"))
	response := httptest.NewRecorder()

	router.ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/api/marketplace/v1/console/me", nil),
	)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{
		"user_id":"42",
		"email":"user@example.com",
		"username":"user",
		"platform_organization_id":"9",
		"platform_role":"admin"
	}`, response.Body.String())
}
