package omnigent

import (
	"net/http"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domainrunner "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
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
		hosts = append(hosts, map[string]any{
			"host_id": "host_" + r.NodeID,
			"name":    r.NodeID,
			"owner":   tenant.OrganizationSlug,
			"status":  status,
		})
	}
	c.JSON(http.StatusOK, gin.H{"hosts": hosts})
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
	layer := compatAgentfileLayer()
	if body.Workspace != "" {
		layer = compatAgentfileLayer(`LOCAL_PATH "` + body.Workspace + `"`)
	}
	orchReq := &agentpod.OrchestrateCreatePodRequest{
		OrganizationID: tenant.OrganizationID,
		UserID:         tenant.UserID,
		RunnerID:       r.ID,
		AgentSlug:      row.AgentSlug,
		AgentfileLayer: layer,
	}
	result, err := d.PodOrchestrator.CreatePod(c.Request.Context(), orchReq)
	if err != nil {
		writeOrchestratorError(c, err)
		return
	}
	_ = d.Sessions.UpdateRunner(c.Request.Context(), row.ID, r.NodeID)
	_ = d.Sessions.UpdatePodKey(c.Request.Context(), row.ID, result.Pod.PodKey)
	c.JSON(http.StatusOK, gin.H{"runner_id": r.NodeID})
}
