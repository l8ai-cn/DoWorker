package sessionapi

import (
	"net/http"
	"strings"

	sessionDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentsession"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
	"github.com/gin-gonic/gin"
)

func (d *Deps) completeCreatedSession(
	c *gin.Context,
	row *sessionDomain.Session,
	result *agentpod.OrchestrateCreatePodResult,
	inputs []persistedSessionInput,
	layer *string,
) {
	if d.Stream != nil {
		for _, input := range inputs {
			d.Stream.PublishInputConsumed(row.ID, input.id, "", input.content)
		}
	}
	if layer != nil && strings.Contains(*layer, "PROMPT ") {
		d.beginAssistantTurn(row.ID)
	}
	if d.Stream != nil {
		d.Stream.PublishPodStatus(
			c.Request.Context(),
			result.Pod.PodKey,
			result.Pod.Status,
			result.Pod.AgentStatus,
		)
	}
	d.DispatchQueue.TriggerDrain(result.Pod.RunnerID)
	c.JSON(http.StatusOK, d.sessionWire(row, result.Pod, nil))
}
