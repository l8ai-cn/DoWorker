package sessionapi

import (
	"net/http"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	itemdomain "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	"github.com/gin-gonic/gin"
)

type forkSessionBody struct {
	Title          *string `json:"title"`
	AgentID        *string `json:"agent_id"`
	UpToResponseID *string `json:"up_to_response_id"`
	ModelOverride  *string `json:"model_override"`
}

func (d *Deps) handleForkSession(c *gin.Context) {
	if d.Sessions == nil || d.Items == nil || d.PodOrchestrator == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "unavailable"})
		return
	}
	source, sourcePod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	var body forkSessionBody
	_ = c.ShouldBindJSON(&body)
	agentSlug := source.AgentSlug
	if body.AgentID != nil && *body.AgentID != "" {
		agentSlug = *body.AgentID
	}
	newID, err := sessionsvc.NewID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "id failed"})
		return
	}
	runnerID := int64(0)
	if sourcePod != nil {
		runnerID = sourcePod.RunnerID
	}
	orchReq := &agentpod.OrchestrateCreatePodRequest{
		OrganizationID: source.OrganizationID,
		UserID:         source.UserID,
		RunnerID:       runnerID,
		AgentSlug:      agentSlug,
		AgentfileLayer: acpAgentfileLayer(),
	}
	if sourcePod != nil && sourcePod.ExternalSessionID != nil {
		orchReq.ResumeExternalSessionID = *sourcePod.ExternalSessionID
	}
	result, err := d.PodOrchestrator.CreatePod(c.Request.Context(), orchReq)
	// Source runner may be at capacity after long smoke suites; fall back to
	// auto-placement rather than 503 the fork.
	if err != nil && runnerID != 0 {
		orchReq.RunnerID = 0
		result, err = d.PodOrchestrator.CreatePod(c.Request.Context(), orchReq)
	}
	if err != nil {
		writeOrchestratorError(c, err)
		return
	}
	parent := source.ID
	row := &domain.Session{
		ID: newID, OrganizationID: source.OrganizationID, UserID: source.UserID,
		PodKey: result.Pod.PodKey, AgentSlug: agentSlug, Title: body.Title,
		ParentSessionID: &parent, Status: "idle",
	}
	if err := d.Sessions.Create(c.Request.Context(), row); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "persist failed"})
		return
	}
	if err := d.copyConversationItems(c, source.ID, newID, body.UpToResponseID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "copy items failed"})
		return
	}
	c.JSON(http.StatusOK, d.sessionWire(row, result.Pod, nil))
}

func (d *Deps) copyConversationItems(c *gin.Context, sourceID, destID string, upToResponseID *string) error {
	page, err := d.Items.ListPage(c.Request.Context(), sourceID, 1000, "", false)
	if err != nil {
		return err
	}
	cutoff := int64(^uint64(0) >> 1)
	if upToResponseID != nil && *upToResponseID != "" {
		for _, it := range page.Items {
			if it.ResponseID == *upToResponseID {
				cutoff = it.Position
				break
			}
		}
	}
	for _, src := range page.Items {
		if src.Position > cutoff {
			break
		}
		id, err := itemsvc.NewItemID()
		if err != nil {
			return err
		}
		row := &itemdomain.Item{
			ID: id, SessionID: destID, ItemType: src.ItemType,
			ResponseID: src.ResponseID, Status: src.Status,
			Position: src.Position, Payload: src.Payload, CreatedAt: src.CreatedAt,
		}
		if err := d.Items.Append(c.Request.Context(), row); err != nil {
			return err
		}
	}
	return nil
}
