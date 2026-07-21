package sessionapi

import (
	"errors"
	"net/http"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentsession"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
	sessionsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/agentsession"
	itemsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/conversationitem"
	runnerservice "github.com/l8ai-cn/agentcloud/backend/internal/service/runner"
	"github.com/gin-gonic/gin"
)

var errForkResponseNotFound = errors.New("fork response not found")

type forkSessionBody struct {
	Title           *string                `json:"title"`
	AgentID         *string                `json:"agent_id"`
	UpToResponseID  *string                `json:"up_to_response_id"`
	ModelOverride   *string                `json:"model_override"`
	ModelResourceID *int64                 `json:"model_resource_id"`
	WorkerSpec      *sessionWorkerSpecBody `json:"worker_spec"`
	AutomationLevel string                 `json:"automation_level"`
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
	tenant := middleware.GetTenant(c)
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
	var orchReq *agentpod.OrchestrateCreatePodRequest
	if agentSlug == source.AgentSlug {
		if hasSessionWorkerConfigChange(
			body.ModelResourceID,
			body.WorkerSpec,
			body.AutomationLevel,
		) || body.ModelOverride != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": rejectSameAgentWorkerConfigChangeMessage(),
				"code":  "validation_failed",
			})
			return
		}
		snapshotID, snapshotErr := sessionSnapshotSource(source, sourcePod)
		if snapshotErr != nil {
			writeOrchestratorError(c, snapshotErr)
			return
		}
		orchReq = buildForkSnapshotPodRequest(source, runnerID, snapshotID)
	} else {
		draft, draftErr := d.buildFreshWorkerPlan(
			c.Request.Context(),
			source.OrganizationID,
			source.UserID,
			tenant.OrganizationSlug,
			sessionWorkerPlanInput{
				WorkerSpec:      body.WorkerSpec,
				WorkerTypeSlug:  agentSlug,
				ModelResourceID: body.ModelResourceID,
				AgentfileLayer:  acpAgentfileLayer(),
				AutomationLevel: body.AutomationLevel,
			},
		)
		if draftErr != nil {
			writeOrchestratorError(c, draftErr)
			return
		}
		orchReq = buildForkPlanPodRequest(source, runnerID, draft)
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
	command, err := pendingSessionCreateCommand(
		result,
		d.DispatchQueue,
		d.DispatchQueue.SendPromptTTL(),
	)
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
