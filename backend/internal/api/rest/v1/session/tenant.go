package sessionapi

import (
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	"github.com/l8ai-cn/agentcloud/backend/pkg/apierr"
	"github.com/gin-gonic/gin"
)

func headerTenant(orgSvc middleware.OrganizationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgSlug := c.GetHeader("X-Organization-Slug")
		if orgSlug == "" {
			orgSlug = c.Query("org_slug")
		}
		if orgSlug == "" {
			apierr.AbortBadRequest(c, apierr.VALIDATION_FAILED, "X-Organization-Slug header is required")
			return
		}
		c.Params = append(c.Params, gin.Param{Key: "slug", Value: orgSlug})
		middleware.TenantMiddleware(orgSvc)(c)
	}
}
