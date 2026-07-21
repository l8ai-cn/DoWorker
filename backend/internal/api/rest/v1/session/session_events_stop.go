package sessionapi

import (
	"errors"
	"net/http"

	podDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	runnerservice "github.com/l8ai-cn/agentcloud/backend/internal/service/runner"
	"github.com/gin-gonic/gin"
)

func (d *Deps) postStopSessionEvent(c *gin.Context, pod *podDomain.Pod) {
	if pod != nil && d.PodCoordinator != nil && pod.IsActive() {
		err := d.PodCoordinator.TerminatePod(c.Request.Context(), pod.PodKey)
		if err != nil && !errors.Is(err, runnerservice.ErrPodAlreadyTerminated) {
			_ = err
		}
	}
	c.JSON(http.StatusAccepted, gin.H{"queued": false})
}
