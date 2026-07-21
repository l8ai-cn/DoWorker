package rest

import (
	"net/http"

	authpkg "github.com/l8ai-cn/agentcloud/backend/pkg/auth"
	"github.com/gin-gonic/gin"
)

func registerJWKSRoute(router gin.IRouter, manager *authpkg.AccessTokenManager) {
	router.GET("/.well-known/jwks.json", func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=300")
		c.JSON(http.StatusOK, manager.JWKS())
	})
}
