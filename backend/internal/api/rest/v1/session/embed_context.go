package sessionapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/pkg/embedtoken"
	"github.com/gin-gonic/gin"
)

type createEmbedContextBody struct {
	ParentOrigins []string `json:"parent_origins"`
	Capabilities  []string `json:"capabilities"`
}

type redeemEmbedContextBody struct {
	RedemptionProof string `json:"redemption_proof"`
}

func (d *Deps) handleCreateEmbedContext(c *gin.Context) {
	row, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok || !d.requireSessionLevel(c, row, levelManage) {
		return
	}
	if d.EmbedTokens == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "embed contexts unavailable"})
		return
	}
	var body createEmbedContextBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid embed context"})
		return
	}
	origins, err := validateEmbedOrigins(body.ParentOrigins)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	capabilities, err := validateEmbedCapabilities(body.Capabilities)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenant := middleware.GetTenant(c)
	if tenant == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	grant, err := d.EmbedTokens.IssueContext(c.Request.Context(), embedtoken.ContextInput{
		SessionID:            row.ID,
		OrganizationID:       tenant.OrganizationID,
		OrganizationSlug:     tenant.OrganizationSlug,
		UserID:               tenant.UserID,
		Email:                d.viewerEmail(c),
		Capabilities:         capabilities,
		AllowedParentOrigins: origins,
	})
	if err != nil {
		if errors.Is(err, embedtoken.ErrContextStore) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "embed contexts unavailable"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "embed context failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"embed_context":    grant.Token,
		"redemption_proof": grant.RedemptionProof,
		"expires_at":       grant.ExpiresAt.Unix(),
	})
}

func (d *Deps) handleInspectEmbedContext(c *gin.Context) {
	if d.EmbedTokens == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "embed contexts unavailable"})
		return
	}
	token, ok := bearerToken(c.GetHeader("Authorization"))
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "embed context is required"})
		return
	}
	claims, err := d.EmbedTokens.InspectContext(c.Request.Context(), token)
	if err != nil {
		if errors.Is(err, embedtoken.ErrContextStore) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "embed contexts unavailable"})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid embed context"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"expires_at":     claims.ExpiresAt.Unix(),
		"parent_origins": claims.AllowedParentOrigins,
	})
}

func (d *Deps) handleRedeemEmbedContext(c *gin.Context) {
	if d.EmbedTokens == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "embed contexts unavailable"})
		return
	}
	token, ok := bearerToken(c.GetHeader("Authorization"))
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "embed context is required"})
		return
	}
	var body redeemEmbedContextBody
	if err := c.ShouldBindJSON(&body); err != nil || body.RedemptionProof == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid embed context"})
		return
	}
	claims, err := d.EmbedTokens.ValidateContext(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid embed context"})
		return
	}
	accessToken, expiresAt, err := d.EmbedTokens.RedeemContext(
		c.Request.Context(),
		token,
		body.RedemptionProof,
	)
	if err != nil {
		if errors.Is(err, embedtoken.ErrContextStore) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "embed contexts unavailable"})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid embed context"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":   accessToken,
		"expires_at":     expiresAt.Unix(),
		"session_id":     claims.SessionID,
		"org_slug":       claims.OrganizationSlug,
		"capabilities":   claims.Capabilities,
		"parent_origins": claims.AllowedParentOrigins,
	})
}

func bearerToken(value string) (string, bool) {
	parts := strings.SplitN(value, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}
