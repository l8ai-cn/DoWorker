package sessionapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	itemdomain "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	"github.com/anthropics/agentsmesh/backend/internal/service/codeximport"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
	"github.com/gin-gonic/gin"
)

// importSessionBody is the request for POST /v1/sessions/import — migrating a
// local Codex conversation record into a new Worker session.
type importSessionBody struct {
	// SourcePath is a server-local path to a Codex rollout transcript
	// (rollout-*.jsonl) or a workflow output_* directory. Auto-detected.
	SourcePath string `json:"source_path"`
	// AgentID is the target Worker's agent slug (same contract as create).
	AgentID string `json:"agent_id"`
	// Title overrides the auto-derived conversation title.
	Title *string `json:"title"`
	// HostID optionally pins the Worker to a specific runner host.
	HostID string `json:"host_id"`
}

// handleImportSession migrates a local Codex conversation record into a fresh
// Worker session. It mirrors the fork flow (create an unbound Worker, then bulk
// insert conversation items) but sources the transcript from a local Codex
// rollout or workflow directory instead of an existing session.
func (d *Deps) handleImportSession(c *gin.Context) {
	if d.PodOrchestrator == nil || d.Sessions == nil || d.Items == nil {
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

	orchReq := &agentpod.OrchestrateCreatePodRequest{
		OrganizationID: tenant.OrganizationID,
		UserID:         tenant.UserID,
		AgentSlug:      body.AgentID,
		AgentfileLayer: acpAgentfileLayer(),
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
	if sessionTitleEmpty(title) && converted.Title != "" {
		t := converted.Title
		title = &t
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
	if err := d.Sessions.Create(c.Request.Context(), row); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist session"})
		return
	}

	if err := d.importConversationItems(c.Request.Context(), sessionID, converted.Items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to import conversation items"})
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
	c.JSON(http.StatusOK, gin.H{
		"session":     d.sessionWire(row, result.Pod, nil),
		"pod_key":     result.Pod.PodKey,
		"source_kind": string(converted.Kind),
		"source_id":   converted.SourceID,
		"item_count":  len(converted.Items),
	})
}

// importConversationItems persists the normalized Codex items into the
// destination session, assigning fresh item ids, monotonic positions, and
// turn-grouped response ids (a new response id starts at each user prompt so
// the assistant reply and its tool calls share the turn).
func (d *Deps) importConversationItems(ctx context.Context, sessionID string, items []codeximport.Item) error {
	currentResp, err := itemsvc.NewResponseID()
	if err != nil {
		return err
	}
	now := time.Now()
	for i, src := range items {
		if src.StartsTurn {
			// Reuse the pre-minted id for the very first item so we never emit
			// an empty leading turn.
			if i != 0 {
				currentResp, err = itemsvc.NewResponseID()
				if err != nil {
					return err
				}
			}
		}
		itemID, err := itemsvc.NewItemID()
		if err != nil {
			return err
		}
		status := src.Status
		if status == "" {
			status = "completed"
		}
		payload := src.Payload
		if payload == nil {
			payload = map[string]any{}
		}
		payload["id"] = itemID
		payload["response_id"] = currentResp
		payload["status"] = status
		encoded, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		rowItem := &itemdomain.Item{
			ID:         itemID,
			SessionID:  sessionID,
			ItemType:   src.Type,
			ResponseID: currentResp,
			Status:     status,
			Position:   int64(i + 1),
			Payload:    encoded,
			CreatedAt:  now.Add(time.Duration(i) * time.Millisecond),
		}
		if err := d.Items.Append(ctx, rowItem); err != nil {
			return err
		}
	}
	return nil
}
