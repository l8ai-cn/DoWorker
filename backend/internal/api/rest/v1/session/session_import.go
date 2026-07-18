package sessionapi

import (
	"log/slog"
	"net/http"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	"github.com/anthropics/agentsmesh/backend/internal/service/codeximport"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
	runnerservice "github.com/anthropics/agentsmesh/backend/internal/service/runner"
	"github.com/gin-gonic/gin"
)

type importSessionBody struct {
	SourcePath string  `json:"source_path"`
	AgentID    string  `json:"agent_id"`
	Title      *string `json:"title"`
	HostID     string  `json:"host_id"`
}

func (d *Deps) handleImportSession(c *gin.Context) {
	if d.PodOrchestrator == nil || d.PodCoordinator == nil ||
		d.DeferredCommitter == nil || d.DispatchQueue == nil ||
		d.Sessions == nil || d.Items == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "session service unavailable", "code": "unavailable"})
		return
	}
	tenant := middleware.GetTenant(c)
	if tenant == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var body importSessionBody
	if err := c.ShouldBindJSON(&body); err != nil || body.SourcePath == "" || body.AgentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source_path and agent_id are required", "code": "validation_failed"})
		return
	}

	converted, err := codeximport.Convert(body.SourcePath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "codex_source_invalid"})
		return
	}
	if len(converted.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no conversation items found in source", "code": "codex_source_empty"})
		return
	}

	sessionID, err := sessionsvc.NewID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "id generation failed"})
		return
	}

	title := body.Title
	if sessionTitleEmpty(title) && converted.Title != "" {
		t := converted.Title
		title = &t
	}
	orchReq := &agentpod.OrchestrateCreatePodRequest{
		OrganizationID:      tenant.OrganizationID,
		UserID:              tenant.UserID,
		AgentSlug:           body.AgentID,
		AgentfileLayer:      acpAgentfileLayer(),
		DeferRunnerDispatch: true,
	}
	if body.HostID != "" {
		runner, ok := d.runnerForHostID(c, body.HostID, tenant.OrganizationID)
		if !ok {
			return
		}
		orchReq.RunnerID = runner.ID
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

	row := &domain.Session{
		ID:             sessionID,
		OrganizationID: tenant.OrganizationID,
		UserID:         tenant.UserID,
		PodKey:         result.Pod.PodKey,
		AgentSlug:      body.AgentID,
		Title:          title,
		Status:         "idle",
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
		func(store *itemsvc.Service) error {
			return importConversationItems(
				c.Request.Context(),
				store,
				sessionID,
				converted.Items,
			)
		},
	)
	if err != nil {
		cleanupErr := d.terminateCreatedSessionPod(c.Request.Context(), result.Pod.PodKey)
		writeSessionCreationCommitFailure(c, "failed to import conversation items", err, cleanupErr)
		return
	}

	if row.Title != nil && *row.Title != "" && d.Pod != nil {
		if err := d.Pod.UpdatePodTitle(c.Request.Context(), result.Pod.PodKey, *row.Title); err != nil {
			slog.WarnContext(c.Request.Context(), "import: set pod title failed", "pod_key", result.Pod.PodKey, "error", err)
		} else {
			result.Pod.Title = row.Title
		}
	}

	if d.Stream != nil {
		d.Stream.PublishPodStatus(c.Request.Context(), result.Pod.PodKey, result.Pod.Status, result.Pod.AgentStatus)
	}
	d.DispatchQueue.TriggerDrain(result.Pod.RunnerID)
	c.JSON(http.StatusOK, gin.H{
		"session":     d.sessionWire(row, result.Pod, nil),
		"pod_key":     result.Pod.PodKey,
		"source_kind": string(converted.Kind),
		"source_id":   converted.SourceID,
		"item_count":  len(converted.Items),
	})
}
