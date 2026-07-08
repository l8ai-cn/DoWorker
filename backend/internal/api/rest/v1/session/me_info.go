package sessionapi

import (
	"net/http"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	orgservice "github.com/anthropics/agentsmesh/backend/internal/service/organization"
	"github.com/gin-gonic/gin"
)

func (d *Deps) handleMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"user_id":   nil,
			"login_url": "/login",
		})
		return
	}
	email, _ := c.Get("email")
	emailStr, _ := email.(string)
	if emailStr == "" {
		emailStr = "user"
	}
	isAdmin := false
	if d.User != nil {
		if u, err := d.User.GetByID(c.Request.Context(), userID); err == nil && u != nil {
			isAdmin = u.IsSystemAdmin
			if u.Email != "" {
				emailStr = u.Email
			}
		}
	}
	resp := gin.H{
		"user_id":  emailStr,
		"is_admin": isAdmin,
	}
	if d.Org != nil {
		if orgs, err := d.Org.ListByUser(c.Request.Context(), userID); err == nil && len(orgs) > 0 {
			resp["org_slugs"] = orgservice.CollectOrgSlugs(orgs)
			resp["org_slug"] = orgservice.PickDefaultOrgSlug(orgs)
		}
	}
	c.JSON(http.StatusOK, resp)
}

func (d *Deps) handleInfo(c *gin.Context) {
	version := d.Version
	if version == "" {
		version = "do-worker-dev"
	}
	c.JSON(http.StatusOK, gin.H{
		"accounts_enabled":          true,
		"login_url":                 "/login",
		"needs_setup":               false,
		"databricks_features":       false,
		"managed_sandboxes_enabled": false,
		"product_name":              "Do Worker",
		"sandbox_provider":          nil,
		"server_version":            version,
		"smart_routing_enabled":     false,
	})
}
