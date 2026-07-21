package sessionapi

import (
	"net/http"
	"strconv"

	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	tokenquotasvc "github.com/l8ai-cn/agentcloud/backend/internal/service/tokenquota"
	"github.com/gin-gonic/gin"
)

func (d *Deps) handleListTokenQuotas(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.TokenQuotas == nil {
		c.JSON(http.StatusOK, gin.H{"object": "list", "data": []any{}})
		return
	}
	quotas, err := d.TokenQuotas.List(c.Request.Context(), tenant.OrganizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list quotas"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"object": "list", "data": quotas})
}

type upsertTokenQuotaBody struct {
	UserID      *int64  `json:"user_id"`
	Model       *string `json:"model"`
	LimitTokens int64   `json:"limit_tokens"`
	Period      string  `json:"period"`
}

func (d *Deps) handleUpsertTokenQuota(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.TokenQuotas == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "quota service unavailable"})
		return
	}
	var body upsertTokenQuotaBody
	if err := c.ShouldBindJSON(&body); err != nil || body.LimitTokens < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit_tokens must be >= 0"})
		return
	}
	if err := d.TokenQuotas.Upsert(c.Request.Context(), tokenquotasvc.UpsertInput{
		OrgID:       tenant.OrganizationID,
		UserID:      body.UserID,
		Model:       body.Model,
		LimitTokens: body.LimitTokens,
		Period:      body.Period,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save quota"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (d *Deps) handleDeleteTokenQuota(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.TokenQuotas == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "quota service unavailable"})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := d.TokenQuotas.Delete(c.Request.Context(), id, tenant.OrganizationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete quota"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (d *Deps) handleQuotaReport(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.TokenQuotas == nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	report, err := d.TokenQuotas.Report(c.Request.Context(), tenant.OrganizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build report"})
		return
	}
	c.JSON(http.StatusOK, report)
}
