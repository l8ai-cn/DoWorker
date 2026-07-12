package api

import (
	"context"
	"net/http"

	actorapi "github.com/anthropics/agentsmesh/marketplace/internal/api/actor"
	consoleapi "github.com/anthropics/agentsmesh/marketplace/internal/api/console"
	consumerapi "github.com/anthropics/agentsmesh/marketplace/internal/api/consumer"
	publicapi "github.com/anthropics/agentsmesh/marketplace/internal/api/public"
	"github.com/anthropics/agentsmesh/marketplace/internal/service"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	Ready         func(context.Context) error
	Storefront    *service.StorefrontService
	Identity      actorapi.TokenVerifier
	Installations consumerapi.InstallationOrchestrator
}

func NewRouter(deps Dependencies) *gin.Engine {
	if deps.Ready == nil || deps.Storefront == nil || deps.Identity == nil ||
		deps.Installations == nil {
		panic("marketplace router dependencies are required")
	}
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
	publicapi.NewHandler(deps.Storefront).RegisterRoutes(
		router.Group("/api/marketplace/v1"),
	)
	console := router.Group("/api/marketplace/v1/console")
	console.Use(actorapi.Middleware(deps.Identity))
	consoleapi.NewSessionHandler().RegisterRoutes(console)
	consumer := router.Group("/api/marketplace/v1")
	consumer.Use(actorapi.Middleware(deps.Identity))
	consumerapi.NewInstallationHandler(deps.Installations).RegisterRoutes(consumer)
	return router
}
