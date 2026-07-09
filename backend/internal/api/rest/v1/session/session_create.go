package sessionapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/anthropics/agentsmesh/agentfile"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
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

	// Model pool selection: mount a configured model (provider key + model id)
	// into the Worker. ModelConfigID references an ai_models row; Model
	// optionally overrides the row's default model id; TokenBudget caps the
	// Worker's token usage.
	ModelConfigID *int64  `json:"model_config_id"`
	Model         *string `json:"model"`
	TokenBudget   *int64  `json:"token_budget"`
}

const workerModelBundleName = "worker-model"

func (d *Deps) handleCreateSession(c *gin.Context) {
	if d.PodOrchestrator == nil || d.Sessions == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "session service unavailable", "code": "unavailable"})
		return
	}
	tenant := middleware.GetTenant(c)
	if tenant == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var body createSessionBody
	if err := c.ShouldBindJSON(&body); err != nil || body.AgentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent_id is required", "code": "validation_failed"})
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

	// Mount a model from the pool: resolve credentials → do-agent settings.json
	// injected as an ephemeral config bundle + a CONFIG model line, so the
	// Worker launches with a working provider instead of exiting on a missing
	// key. Falls back to the caller's default model when none is specified.
	mount, modelErr := d.resolveWorkerModel(c, tenant.UserID, tenant.OrganizationID, body, &layer)
	if modelErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": modelErr.Error(), "code": "model_unavailable"})
		return
	}
	if mount == nil || !mount.mounted() {
		d.resolvePrimaryEnvBundle(c.Request.Context(), tenant.UserID, tenant.OrganizationID, body.AgentID, &layer)
	}

	orchReq := &agentpod.OrchestrateCreatePodRequest{
		OrganizationID:       tenant.OrganizationID,
		UserID:               tenant.UserID,
		AgentSlug:            body.AgentID,
		AgentfileLayer:       layer,
		LocalPath:            workspace,
		SessionConfigBundles: configBundlesFromMount(mount),
		SessionEnvBundles:    envBundlesFromMount(mount),
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
	title := body.Title
	if sessionTitleEmpty(title) {
		if seeded := promptTextFromInitialItems(body.InitialItems); seeded != "" {
			title = deriveSessionTitleFromPrompt(seeded)
		}
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
	if err := d.Sessions.Create(c.Request.Context(), row); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist session"})
		return
	}
	d.persistInitialUserItems(c.Request.Context(), row.ID, body.InitialItems)
	if layer != nil && strings.Contains(*layer, "PROMPT ") {
		d.beginAssistantTurn(row.ID)
	}
	if d.Stream != nil {
		d.Stream.PublishPodStatus(c.Request.Context(), result.Pod.PodKey, result.Pod.Status, result.Pod.AgentStatus)
	}
	c.JSON(http.StatusOK, d.sessionWire(row, result.Pod, nil))
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
