package internal

import (
	"crypto/subtle"
	"log/slog"

	"github.com/l8ai-cn/agentcloud/backend/internal/infra/acme"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/geo"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/relay"
	"github.com/l8ai-cn/agentcloud/backend/pkg/apierr"
	"github.com/gin-gonic/gin"
)

type RelayHandler struct {
	relayManager *relay.Manager
	dnsService   *relay.DNSService
	acmeManager  *acme.Manager
	geoResolver  geo.Resolver
	logger       *slog.Logger
}

func NewRelayHandler(relayManager *relay.Manager, dnsService *relay.DNSService, acmeManager *acme.Manager, geoResolver geo.Resolver) *RelayHandler {
	return &RelayHandler{
		relayManager: relayManager,
		dnsService:   dnsService,
		acmeManager:  acmeManager,
		geoResolver:  geoResolver,
		logger:       slog.With("component", "relay_handler"),
	}
}

type RelayRouterDeps struct {
	RelayManager    *relay.Manager
	DNSService      *relay.DNSService
	ACMEManager     *acme.Manager
	GeoResolver     geo.Resolver
	InternalSecret  string
	PreviewSessions previewSessionService
}

func RegisterRelayRoutes(router *gin.RouterGroup, deps *RelayRouterDeps) {
	handler := NewRelayHandler(deps.RelayManager, deps.DNSService, deps.ACMEManager, deps.GeoResolver)

	router.Use(InternalAPIAuth(deps.InternalSecret))
	RegisterPreviewSessionRoutes(router, deps.PreviewSessions)

	router.POST("/register", handler.Register)
	router.POST("/heartbeat", handler.Heartbeat)
	router.POST("/unregister", handler.Unregister)
	router.GET("/stats", handler.Stats)
	router.GET("", handler.List)
	router.GET("/:relay_id", handler.Get)
	router.DELETE("/:relay_id", handler.ForceUnregister)
}

func InternalAPIAuth(secret string) gin.HandlerFunc {
	if secret == "" {
		panic("internal API secret must not be empty")
	}
	secretBytes := []byte(secret)
	return func(c *gin.Context) {
		auth := []byte(c.GetHeader("X-Internal-Secret"))
		if subtle.ConstantTimeCompare(auth, secretBytes) != 1 {
			apierr.AbortUnauthorized(c, apierr.AUTH_REQUIRED, "unauthorized")
			return
		}
		c.Next()
	}
}
