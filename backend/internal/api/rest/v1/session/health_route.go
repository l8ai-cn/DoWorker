package sessionapi

import (
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterHealthRoute mounts the shared /health endpoint. Without session_ids
// it answers the infra probe; with session_ids it returns Omnigent liveness.
func RegisterHealthRoute(r *gin.Engine, d Deps) {
	r.GET("/health", func(c *gin.Context) {
		if c.Query("session_ids") != "" {
			middleware.AuthMiddleware(d.JWTSecret)(c)
			if c.IsAborted() {
				return
			}
			d.handleSessionHealth(c)
			return
		}
		c.JSON(200, gin.H{"status": "ok", "service": "do-worker-api"})
	})
}
