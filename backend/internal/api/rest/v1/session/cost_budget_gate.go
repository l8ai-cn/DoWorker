package sessionapi

import (
	"net/http"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func (d *Deps) checkCostBudget(c *gin.Context, podKey string) bool {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.Policies == nil || d.SessionUsage == nil || podKey == "" {
		return true
	}
	maxUSD, ok, err := d.Policies.OrgCostBudgetUSD(c.Request.Context(), tenant.OrganizationID)
	if err != nil || !ok {
		return true
	}
	agg, err := d.SessionUsage.Aggregate(c.Request.Context(), podKey)
	if err != nil || agg.TotalCostUSD == nil {
		return true
	}
	if *agg.TotalCostUSD < maxUSD {
		return true
	}
	c.JSON(http.StatusPaymentRequired, gin.H{
		"error":   "session cost budget exceeded",
		"code":    "cost_budget_exceeded",
		"max_usd": maxUSD,
		"spent":   *agg.TotalCostUSD,
	})
	return false
}

func (d *Deps) handleOrgUsageSummary(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.SessionUsage == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	agg, err := d.SessionUsage.AggregateOrg(c.Request.Context(), tenant.OrganizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "aggregate failed"})
		return
	}
	out := gin.H{"object": "org_usage_summary"}
	if agg.TotalCostUSD != nil {
		out["total_cost_usd"] = *agg.TotalCostUSD
	}
	if len(agg.UsageByModel) > 0 {
		out["usage_by_model"] = agg.UsageByModel
	}
	c.JSON(http.StatusOK, out)
}
