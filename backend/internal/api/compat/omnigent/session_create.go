package omnigent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
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
}

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
	layer := compatAgentfileLayer()
	if body.PTYOnly != nil && *body.PTYOnly {
		layer = compatPTYLayer()
	}
	if body.Scenario != nil && strings.TrimSpace(*body.Scenario) != "" {
		scenarioLayer := `CONFIG scenario = "` + strings.TrimSpace(*body.Scenario) + `"`
		if body.PTYOnly != nil && *body.PTYOnly {
			layer = compatPTYLayer(scenarioLayer)
		} else {
			layer = compatAgentfileLayer(scenarioLayer)
		}
	}
	if pl := promptLayerFromItems(body.InitialItems); pl != nil {
		if body.PTYOnly != nil && *body.PTYOnly {
			layer = compatPTYLayer(*pl)
		} else {
			layer = compatAgentfileLayer(*pl)
		}
	}
	workspace := strings.TrimSpace(body.Workspace)
	orchReq := &agentpod.OrchestrateCreatePodRequest{
		OrganizationID: tenant.OrganizationID,
		UserID:         tenant.UserID,
		AgentSlug:      body.AgentID,
		AgentfileLayer: layer,
		LocalPath:      workspace,
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
	row := &domain.Session{
		ID:              sessionID,
		OrganizationID:  tenant.OrganizationID,
		UserID:          tenant.UserID,
		PodKey:          result.Pod.PodKey,
		AgentSlug:       body.AgentID,
		Title:           body.Title,
		ParentSessionID: body.ParentSessionID,
		Status:          "idle",
	}
	if err := d.Sessions.Create(c.Request.Context(), row); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist session"})
		return
	}
	if layer != nil && strings.Contains(*layer, "PROMPT ") {
		d.beginAssistantTurn(row.ID)
	}
	ForwardPodStatus(c.Request.Context(), result.Pod.PodKey, result.Pod.Status, result.Pod.AgentStatus)
	c.JSON(http.StatusOK, d.sessionWire(row, result.Pod, nil))
}

func (d *Deps) beginAssistantTurn(sessionID string) {
	if d.Hub == nil {
		return
	}
	respID, err := itemsvc.NewResponseID()
	if err != nil {
		return
	}
	d.Hub.StartTurn(sessionID, respID)
	now := time.Now().Unix()
	d.Hub.Publish(sessionID, formatSSE("response.created", map[string]any{
		"id": respID, "status": "in_progress", "model": "", "created_at": now,
		"conversation": map[string]any{"id": sessionID},
	}))
	d.Hub.Publish(sessionID, formatSSE("response.in_progress", map[string]any{
		"id": respID, "status": "in_progress", "model": "", "created_at": now,
	}))
}

func promptLayerFromItems(items []json.RawMessage) *string {
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
			if block.Type == "text" && block.Text != "" {
				parts = append(parts, block.Text)
			}
		}
		if len(parts) == 0 {
			continue
		}
		text := strings.Join(parts, "\n")
		escaped := strings.ReplaceAll(text, `"`, `\"`)
		layer := fmt.Sprintf("PROMPT \"%s\"", escaped)
		return &layer
	}
	return nil
}
