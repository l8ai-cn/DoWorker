package sessionapi

import (
	"net/http"
	"strings"

	podDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	sessionDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentsession"
	domainrunner "github.com/l8ai-cn/agentcloud/backend/internal/domain/runner"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
	"github.com/gin-gonic/gin"
)

func (d *Deps) handleListHosts(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.Runner == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	runners, err := d.Runner.ListRunners(c.Request.Context(), tenant.OrganizationID, tenant.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list failed"})
		return
	}
	hosts := make([]map[string]any, 0, len(runners))
	for _, r := range runners {
		if !r.IsEnabled {
			continue
		}
		status := "offline"
		if r.Status == domainrunner.RunnerStatusOnline {
			status = "online"
		}
		entry := map[string]any{
			"host_id": "host_" + r.NodeID,
			"name":    r.NodeID,
			"owner":   tenant.OrganizationSlug,
			"status":  status,
		}
		if len(r.AvailableAgents) > 0 {
			harnesses := make(map[string]bool, len(r.AvailableAgents))
			for _, slug := range r.AvailableAgents {
				harnesses[slug] = true
			}
			entry["configured_harnesses"] = harnesses
		}
		hosts = append(hosts, entry)
	}
	c.JSON(http.StatusOK, gin.H{"hosts": hosts})
}

func (d *Deps) runnerForHostID(c *gin.Context, hostID string, orgID int64) (*domainrunner.Runner, bool) {
	if d.Runner == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "runner service unavailable"})
		return nil, false
	}
	nodeID := strings.TrimPrefix(hostID, "host_")
	r, err := d.Runner.GetByNodeIDAndOrgID(c.Request.Context(), nodeID, orgID)
	if err != nil || r == nil || !r.IsEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "host not found", "code": "host_not_found"})
		return nil, false
	}
	if r.Status != domainrunner.RunnerStatusOnline {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "host offline", "code": "runner_unavailable"})
		return nil, false
	}
	return r, true
}

type bindRunnerBody struct {
	SessionID string `json:"session_id"`
	Workspace string `json:"workspace"`
}

func (d *Deps) handleBindHostRunner(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.PodOrchestrator == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "unavailable"})
		return
	}
	var body bindRunnerBody
	if err := c.ShouldBindJSON(&body); err != nil || body.SessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id required"})
		return
	}
	row, err := d.Sessions.Get(c.Request.Context(), body.SessionID)
	if err != nil || row.OrganizationID != tenant.OrganizationID || row.UserID != tenant.UserID {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	hostID := c.Param("id")
	nodeID := hostID
	if len(hostID) > 5 && hostID[:5] == "host_" {
		nodeID = hostID[5:]
	}
	r, err := d.Runner.GetByNodeIDAndOrgID(c.Request.Context(), nodeID, tenant.OrganizationID)
	if err != nil || r == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "host not found"})
		return
	}
	pod, _ := d.Pod.GetPod(c.Request.Context(), row.PodKey)
	if pod != nil && pod.Status == podDomain.StatusRunning {
		_ = d.Sessions.UpdateRunner(c.Request.Context(), row.ID, r.NodeID)
		c.JSON(http.StatusOK, gin.H{"runner_id": r.NodeID})
		return
	}
	snapshotID, err := sessionSnapshotSource(row, pod)
	if err != nil {
		writeOrchestratorError(c, err)
		return
	}
	orchReq := &agentpod.OrchestrateCreatePodRequest{
		OrganizationID:       tenant.OrganizationID,
		UserID:               tenant.UserID,
		RunnerID:             r.ID,
		WorkerSpecSnapshotID: snapshotID,
		LocalPath:            strings.TrimSpace(body.Workspace),
		AgentSessionID:       row.ID,
		SessionProvision: &sessionDomain.ProvisionSpec{
			ID: row.ID, ExpectedPodKey: row.PodKey, UpdateExisting: true,
		},
	}
	_, err = d.PodOrchestrator.CreatePod(c.Request.Context(), orchReq)
	if err != nil {
		writeOrchestratorError(c, err)
		return
	}
	_ = d.Sessions.UpdateRunner(c.Request.Context(), row.ID, r.NodeID)
	c.JSON(http.StatusOK, gin.H{"runner_id": r.NodeID})
}
