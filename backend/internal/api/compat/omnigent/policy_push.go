package omnigent

import (
	"log/slog"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func (d *Deps) pushPolicyToActivePods(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.Policies == nil || d.Pod == nil || d.CommandSender == nil {
		return
	}
	rules, err := d.Policies.SnapshotForPodCreate(c.Request.Context(), tenant.OrganizationID, "")
	if err != nil {
		slog.WarnContext(c.Request.Context(), "policy push snapshot failed", "error", err)
		return
	}
	pods, _, err := d.Pod.ListPods(c.Request.Context(), tenant.OrganizationID, podDomain.PodListQuery{
		Statuses: []string{podDomain.StatusRunning, podDomain.StatusInitializing},
		Limit:    200,
	})
	if err != nil {
		return
	}
	for _, pod := range pods {
		if pod == nil || pod.RunnerID == 0 {
			continue
		}
		_ = d.CommandSender.SendUpdatePodPolicyRules(c.Request.Context(), pod.RunnerID, pod.PodKey, rules)
	}
}
