package sessionapi

import (
	"encoding/json"
	"errors"
	"net/http"

	podDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentsession"
	runnerservice "github.com/l8ai-cn/agentcloud/backend/internal/service/runner"
	"github.com/gin-gonic/gin"
)

func (d *Deps) postInterruptEvent(c *gin.Context, row *domain.Session, pod *podDomain.Pod) {
	if d.CommandSender == nil || pod == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "unavailable"})
		return
	}
	payload, _ := json.Marshal(map[string]string{"type": "interrupt"})
	if err := d.CommandSender.SendAcpRelay(c.Request.Context(), pod.RunnerID, pod.PodKey, string(payload)); err != nil {
		if errors.Is(err, runnerservice.ErrRunnerNotConnected) || errors.Is(err, runnerservice.ErrRunnerOffline) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "runner unavailable", "code": "runner_unavailable"})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": "interrupt failed", "code": "runner_unreachable"})
		return
	}
	responseID := ""
	if d.Hub != nil {
		if active, ok := d.Hub.ActiveResponse(row.ID); ok {
			responseID = active
		}
	}
	if d.Stream != nil {
		d.Stream.PublishSessionInterrupted(row.ID, responseID)
		d.Stream.PublishSessionStatus(row.ID, "idle")
	}
	c.JSON(http.StatusAccepted, gin.H{"queued": true})
}
