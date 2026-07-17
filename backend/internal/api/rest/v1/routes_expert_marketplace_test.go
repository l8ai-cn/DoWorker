package v1

import (
	"testing"

	expertsvc "github.com/anthropics/agentsmesh/backend/internal/service/expert"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRegisterExpertRoutesIncludesUserMarketplaceOperations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	registerExpertRoutes(
		engine.Group("/api/v1/organizations/:orgSlug"),
		&Services{Expert: expertsvc.NewService(expertsvc.Deps{})},
	)

	routes := make(map[string]struct{})
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = struct{}{}
	}
	for _, expected := range []string{
		"POST /api/v1/organizations/:orgSlug/experts/:expertSlug/market-submissions",
		"GET /api/v1/organizations/:orgSlug/marketplace/submissions",
		"POST /api/v1/organizations/:orgSlug/marketplace/releases/:releaseID/withdraw",
		"POST /api/v1/organizations/:orgSlug/experts/:expertSlug/market-upgrade",
		"GET /api/v1/organizations/:orgSlug/experts/:expertSlug/market-upgrade",
	} {
		_, exists := routes[expected]
		require.True(t, exists, expected)
	}
}
