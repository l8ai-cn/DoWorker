package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	Ready func(context.Context) error
}

func NewRouter(deps Dependencies) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "live"})
	})
	router.GET("/health/ready", func(c *gin.Context) {
		if err := deps.Ready(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, errorEnvelope{
				Error: errorBody{
					Code:    "SERVICE_NOT_READY",
					Message: "市场服务尚未就绪",
				},
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	return router
}
