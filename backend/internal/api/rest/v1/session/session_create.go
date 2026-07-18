package sessionapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/anthropics/agentsmesh/agentfile"
	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
	runnerservice "github.com/anthropics/agentsmesh/backend/internal/service/runner"
	"github.com/gin-gonic/gin"
)

type createSessionBody struct {
	AgentID         string            `json:"agent_id"`
	HostID          string            `json:"host_id"`
	Workspace       string            `json:"workspace"`
	InitialItems    []json.RawMessage `json:"initial_items"`
	ParentSessionID *string           `json:"parent_session_id"`
	SubAgentName    *string           `json:"sub_agent_name"`
	Title           *string           `json:"title"`
	Scenario        *string           `json:"scenario"`
	PTYOnly         *bool             `json:"pty_only"`

	ModelResourceID *int64 `json:"model_resource_id"`
	TokenBudget     *int64 `json:"token_budget"`
}

func (d *Deps) handleCreateSession(c *gin.Context) {
	if d.PodOrchestrator == nil || d.PodCoordinator == nil ||
		d.DeferredCommitter == nil || d.DispatchQueue == nil ||
		d.Sessions == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "session service unavailable", "code": "unavailable"})
		return
	}
	tenant := middleware.GetTenant(c)
	if tenant == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var body createSessionBody
	raw, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "validation_failed"})
		return
	}
	if field, ok := legacySessionCreateModelField(raw); ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": field + " is no longer supported; use model_resource_id", "code": "validation_failed"})
		return
	}
	if err := json.Unmarshal(raw, &body); err != nil || body.AgentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent_id is required", "code": "validation_failed"})
		return
	}
	if initialItemsContainAttachments(body.InitialItems) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "attachments must be sent after the session is created",
			"code":  "validation_failed",
		})
		return
	}
	sessionID, err := sessionsvc.NewID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "id generation failed"})
		return
	}
	ptyOnly := body.PTYOnly != nil && *body.PTYOnly
	var layerExtras []string
	if body.Scenario != nil && strings.TrimSpace(*body.Scenario) != "" {
		layerExtras = append(layerExtras, `CONFIG scenario = "`+strings.TrimSpace(*body.Scenario)+`"`)
	}
	if pl := promptLayerFromItems(body.InitialItems); pl != nil {
		layerExtras = append(layerExtras, *pl)
	}
	layer := sessionAgentfileLayer(body.AgentID, ptyOnly, layerExtras...)
	workspace := strings.TrimSpace(body.Workspace)
	title := body.Title
	if sessionTitleEmpty(title) {
		if seeded := promptTextFromInitialItems(body.InitialItems); seeded != "" {
			title = deriveSessionTitleFromPrompt(seeded)
		}
	}

	orchReq := sessionCreatePodRequest(tenant.UserID, tenant.OrganizationID, body, layer, workspace)
	orchReq.DeferRunnerDispatch = true
	// pty_only sessions must stay on PTY. Default automation (autonomous)
	// appends MODE acp and would override the MODE pty layer above.
	if ptyOnly {
		orchReq.AutomationLevel = podDomain.AutomationLevelInteractive
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
		ID:              sessionID,
		OrganizationID:  tenant.OrganizationID,
		UserID:          tenant.UserID,
		PodKey:          result.Pod.PodKey,
		AgentSlug:       body.AgentID,
		Title:           title,
		ParentSessionID: body.ParentSessionID,
		Status:          "idle",
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
	var persistedInputs []persistedSessionInput
	err = d.DeferredCommitter.CommitCreate(
		c.Request.Context(),
		row,
		command,
		d.DispatchQueue.MaxPerRunner(),
		func(store *itemsvc.Service) error {
			var writeErr error
			persistedInputs, writeErr = persistInitialUserItems(
				c.Request.Context(),
				store,
				row.ID,
				body.InitialItems,
			)
			return writeErr
		},
	)
	if err != nil {
		cleanupErr := d.terminateCreatedSessionPod(c.Request.Context(), result.Pod.PodKey)
		writeSessionCreationCommitFailure(c, "failed to persist session", err, cleanupErr)
		return
	}
	d.completeCreatedSession(c, row, result, persistedInputs, layer)
}

func (d *Deps) beginAssistantTurn(sessionID string) {
	if d.Stream != nil {
		d.Stream.StartAssistantTurn(sessionID)
	}
}

func promptLayerFromItems(items []json.RawMessage) *string {
	text := promptTextFromInitialItems(items)
	if text == "" {
		return nil
	}
	layer := "PROMPT " + agentfile.FormatStringLiteral(text)
	return &layer
}

func promptTextFromInitialItems(items []json.RawMessage) string {
	for _, raw := range items {
		var evt struct {
			Type string `json:"type"`
			Data struct {
				Role    string `json:"role"`
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
			} `json:"data"`
		}
		if json.Unmarshal(raw, &evt) != nil || evt.Type != "message" {
			continue
		}
		var parts []string
		for _, block := range evt.Data.Content {
			if (block.Type == "text" || block.Type == "input_text") && block.Text != "" {
				parts = append(parts, block.Text)
			}
		}
		if len(parts) == 0 {
			continue
		}
		return strings.Join(parts, "\n")
	}
	return ""
}
