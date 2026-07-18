package sessionapi

import (
	"encoding/json"
	"errors"
	"net/http"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	itemdomain "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
	runnerservice "github.com/anthropics/agentsmesh/backend/internal/service/runner"
	"github.com/gin-gonic/gin"
)

var errForkResponseNotFound = errors.New("fork response not found")

type forkSessionBody struct {
	Title          *string `json:"title"`
	AgentID        *string `json:"agent_id"`
	UpToResponseID *string `json:"up_to_response_id"`
	ModelOverride  *string `json:"model_override"`
}

func (d *Deps) handleForkSession(c *gin.Context) {
	if d.Sessions == nil || d.Items == nil || d.PodOrchestrator == nil ||
		d.PodCoordinator == nil || d.DeferredCommitter == nil ||
		d.DispatchQueue == nil {
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
		OrganizationID:      source.OrganizationID,
		UserID:              source.UserID,
		RunnerID:            runnerID,
		AgentSlug:           agentSlug,
		AgentfileLayer:      acpAgentfileLayer(),
		DeferRunnerDispatch: true,
	}
	if sourcePod != nil && sourcePod.ExternalSessionID != nil {
		orchReq.ResumeExternalSessionID = *sourcePod.ExternalSessionID
	}
	result, err := d.PodOrchestrator.CreatePod(c.Request.Context(), orchReq)
	if err != nil {
		writeOrchestratorError(c, err)
		return
	}
	if !d.DispatchQueue.AllowsDurableCommand(result.Pod.RunnerID) {
		cleanupErr := d.terminateCreatedSessionPod(c.Request.Context(), result.Pod.PodKey)
		writeSessionCreationCommitFailure(
			c,
			"runner unavailable",
			runnerservice.ErrRunnerNotConnected,
			cleanupErr,
		)
		return
	}
	parent := source.ID
	row := &domain.Session{
		ID: newID, OrganizationID: source.OrganizationID, UserID: source.UserID,
		PodKey: result.Pod.PodKey, AgentSlug: agentSlug, Title: body.Title,
		ParentSessionID: &parent, Status: "idle",
	}
	command, err := pendingSessionCreateCommand(result, d.DispatchQueue.SendPromptTTL())
	if err != nil {
		cleanupErr := d.terminateCreatedSessionPod(c.Request.Context(), result.Pod.PodKey)
		writeSessionCreationFailure(c, "failed to prepare session dispatch", cleanupErr)
		return
	}
	err = d.DeferredCommitter.CommitCreate(
		c.Request.Context(),
		row,
		command,
		d.DispatchQueue.MaxPerRunner(),
		func(writer *itemsvc.Service) error {
			return d.copyConversationItems(
				c,
				writer,
				source.ID,
				newID,
				body.UpToResponseID,
			)
		},
	)
	if err != nil {
		cleanupErr := d.terminateCreatedSessionPod(c.Request.Context(), result.Pod.PodKey)
		if cleanupErr == nil && errors.Is(err, errForkResponseNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "up_to_response_id was not found",
				"code":  "validation_failed",
			})
			return
		}
		writeSessionCreationCommitFailure(c, "copy items failed", err, cleanupErr)
		return
	}
	d.DispatchQueue.TriggerDrain(result.Pod.RunnerID)
	c.JSON(http.StatusOK, d.sessionWire(row, result.Pod, nil))
}

func (d *Deps) copyConversationItems(
	c *gin.Context,
	writer *itemsvc.Service,
	sourceID, destID string,
	upToResponseID *string,
) error {
	targetResponseID := ""
	if upToResponseID != nil && *upToResponseID != "" {
		targetResponseID = *upToResponseID
	}
	afterID := ""
	targetFound := false
	for {
		page, err := d.Items.ListPage(c.Request.Context(), sourceID, 100, afterID, false)
		if err != nil {
			return err
		}
		for _, src := range page.Items {
			if targetFound && src.ResponseID != targetResponseID {
				return nil
			}
			if err := d.copyConversationItem(c, writer, destID, src); err != nil {
				return err
			}
			if targetResponseID != "" && src.ResponseID == targetResponseID {
				targetFound = true
			}
		}
		if !page.HasMore {
			if targetResponseID != "" && !targetFound {
				return errForkResponseNotFound
			}
			return nil
		}
		if len(page.Items) == 0 {
			return errors.New("conversation item pagination stalled")
		}
		afterID = page.Items[len(page.Items)-1].ID
	}
}

func (d *Deps) copyConversationItem(
	c *gin.Context,
	writer *itemsvc.Service,
	destID string,
	src itemdomain.Item,
) error {
	id, err := itemsvc.NewItemID()
	if err != nil {
		return err
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(src.Payload, &payload); err != nil || payload == nil {
		return errors.New("conversation item payload is invalid")
	}
	encodedID, err := json.Marshal(id)
	if err != nil {
		return err
	}
	payload["id"] = encodedID
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return writer.Append(c.Request.Context(), &itemdomain.Item{
		ID: id, SessionID: destID, ItemType: src.ItemType,
		ResponseID: src.ResponseID, Status: src.Status,
		Position: src.Position, Payload: encodedPayload, CreatedAt: src.CreatedAt,
	})
}
