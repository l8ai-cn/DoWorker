package sessionapi

import (
	"net/http"

	domainrunner "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

type runnerWire struct {
	RunnerID        string   `json:"runner_id"`
	Online          bool     `json:"online"`
	Harnesses       []string `json:"harnesses"`
	TunnelState     string   `json:"tunnel_state"`
	TunnelLastError *string  `json:"tunnel_last_error,omitempty"`
}

func (d *Deps) handleListRunners(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant required"})
		return
	}
	runners, err := d.Runner.ListRunners(c.Request.Context(), tenant.OrganizationID, tenant.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list runners"})
		return
	}
	rows := make([]runnerWire, 0, len(runners))
	for _, r := range runners {
		if !r.IsEnabled {
			continue
		}
		rows = append(rows, runnerWire{
			RunnerID:        r.NodeID,
			Online:          r.Status == domainrunner.RunnerStatusOnline,
			Harnesses:       []string(r.AvailableAgents),
			TunnelState:     r.TunnelState,
			TunnelLastError: r.TunnelLastError,
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}
