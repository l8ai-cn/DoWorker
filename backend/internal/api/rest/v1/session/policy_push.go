package sessionapi

import (
	"log/slog"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/gin-gonic/gin"
)

func (d *Deps) pushPolicyToSessionPod(c *gin.Context, row *domain.Session, pod *podDomain.Pod) {
	if row == nil || d.Policies == nil || d.CommandSender == nil {
		return
	}
	if pod == nil {
		pod = d.loadPod(c, row.PodKey)
	}
	if pod == nil || pod.RunnerID == 0 {
		return
	}
	rules, err := d.Policies.SnapshotForSession(c.Request.Context(), row.OrganizationID, row.ID, row.AgentSlug)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "session policy push snapshot failed", "session_id", row.ID, "error", err)
		return
	}
	_ = d.CommandSender.SendUpdatePodPolicyRules(c.Request.Context(), pod.RunnerID, pod.PodKey, rules)
}

func (d *Deps) pushPolicyToActivePods(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.Policies == nil || d.Pod == nil || d.CommandSender == nil {
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
		rules, err := d.policyRulesForPod(c, tenant.OrganizationID, pod)
		if err != nil {
			slog.WarnContext(c.Request.Context(), "policy push snapshot failed", "pod_key", pod.PodKey, "error", err)
			continue
		}
		_ = d.CommandSender.SendUpdatePodPolicyRules(c.Request.Context(), pod.RunnerID, pod.PodKey, rules)
	}
}

func (d *Deps) policyRulesForPod(c *gin.Context, orgID int64, pod *podDomain.Pod) ([]*runnerv1.PolicyRuleSnapshot, error) {
	if d.Sessions != nil {
		if row, err := d.Sessions.GetByPodKey(c.Request.Context(), pod.PodKey); err == nil && row != nil {
			return d.Policies.SnapshotForSession(c.Request.Context(), orgID, row.ID, row.AgentSlug)
		}
	}
	agentSlug := pod.AgentSlug
	return d.Policies.SnapshotForPodCreate(c.Request.Context(), orgID, agentSlug)
}
