package omnigent

import (
	"net/http"
	"strings"

	domainrunner "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func (d *Deps) handleSessionHealth(c *gin.Context) {
	idsParam := c.Query("session_ids")
	if idsParam == "" {
		c.JSON(http.StatusOK, gin.H{"sessions": map[string]any{}})
		return
	}
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	onlineRunners := d.onlineRunnerNodes(c)
	out := make(map[string]gin.H)
	for _, raw := range strings.Split(idsParam, ",") {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		row, err := d.Sessions.Get(c.Request.Context(), id)
		if err != nil || row == nil || row.UserID != userID {
			continue
		}
		runnerOnline := false
		if row.RunnerNodeID != nil {
			runnerOnline = onlineRunners[*row.RunnerNodeID]
		} else if d.Pod != nil && row.PodKey != "" {
			if pod, _ := d.Pod.GetByKey(c.Request.Context(), row.PodKey); pod != nil && d.Runner != nil {
				if r, _ := d.Runner.GetRunner(c.Request.Context(), pod.RunnerID); r != nil {
					runnerOnline = r.Status == domainrunner.RunnerStatusOnline
				}
			}
		}
		out[id] = gin.H{
			"runner_online": runnerOnline,
			"host_online":   nil,
			"host_version":  nil,
		}
	}
	c.JSON(http.StatusOK, gin.H{"sessions": out})
}

func (d *Deps) onlineRunnerNodes(c *gin.Context) map[string]bool {
	out := make(map[string]bool)
	if d.Runner == nil {
		return out
	}
	tenant := middleware.GetTenant(c)
	if tenant == nil {
		return out
	}
	runners, err := d.Runner.ListRunners(c.Request.Context(), tenant.OrganizationID, tenant.UserID)
	if err != nil {
		return out
	}
	for _, r := range runners {
		if r.IsEnabled && r.Status == domainrunner.RunnerStatusOnline {
			out[r.NodeID] = true
		}
	}
	return out
}
