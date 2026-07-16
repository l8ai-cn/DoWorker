package sessionapi

import (
	"net/http"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/pkg/embedtoken"
	"github.com/gin-gonic/gin"
)

const embedClaimsContextKey = "agent_embed_claims"

func (d *Deps) embedSessionAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if d.EmbedTokens == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "embed sessions unavailable"})
			c.Abort()
			return
		}
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "embed session is required"})
			c.Abort()
			return
		}
		claims, err := d.EmbedTokens.ValidateSession(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid embed session"})
			c.Abort()
			return
		}
		tenant := &middleware.TenantContext{
			OrganizationID:   claims.OrganizationID,
			OrganizationSlug: claims.OrganizationSlug,
			UserID:           claims.UserID,
			UserRole:         "embed",
		}
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("tenant", tenant)
		c.Set(embedClaimsContextKey, claims)
		c.Request = c.Request.WithContext(middleware.SetTenant(c.Request.Context(), tenant))
		c.Next()
	}
}

func requireEmbedCapability(capability string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := embedClaims(c)
		if claims == nil || !hasEmbedCapability(claims, capability) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func embedClaims(c *gin.Context) *embedtoken.Claims {
	value, ok := c.Get(embedClaimsContextKey)
	if !ok {
		return nil
	}
	claims, _ := value.(*embedtoken.Claims)
	return claims
}

func hasEmbedCapability(claims *embedtoken.Claims, capability string) bool {
	for _, value := range claims.Capabilities {
		if value == capability {
			return true
		}
	}
	return false
}
