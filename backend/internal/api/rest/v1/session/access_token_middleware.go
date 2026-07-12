package sessionapi

import (
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func accessTokenMiddleware(d Deps) gin.HandlerFunc {
	if d.Auth == nil {
		return middleware.AuthMiddleware(nil, "")
	}
	return middleware.AuthMiddleware(
		d.Auth.AccessTokenManager(),
		d.Auth.AccessTokenAudience(),
	)
}
