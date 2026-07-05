package omnigent

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	domainitem "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
	runnerservice "github.com/anthropics/agentsmesh/backend/internal/service/runner"
	"github.com/gin-gonic/gin"
)

func (d *Deps) handlePostEvent(c *gin.Context) {
	row, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	var evt struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}
	if err := c.ShouldBindJSON(&evt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event"})
		return
	}
	switch evt.Type {
	case "message":
		d.postMessageEvent(c, row, pod, evt.Data)
	default:
		c.JSON(http.StatusAccepted, gin.H{"queued": false})
	}
}

func (d *Deps) postMessageEvent(c *gin.Context, row *domain.Session, pod *podDomain.Pod, data json.RawMessage) {
	if d.CommandSender == nil || pod == nil || d.Items == nil || d.Hub == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "unavailable"})
		return
	}
	content, prompt := parseMessageContent(data)
	if !messageHasContent(content) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty message"})
		return
	}
	itemID, err := itemsvc.NewItemID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "id failed"})
		return
	}
	respID, err := itemsvc.NewResponseID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "id failed"})
		return
	}
	pos, err := d.Items.NextPosition(c.Request.Context(), row.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "persist failed"})
		return
	}
	payload, _ := json.Marshal(map[string]any{
		"id": itemID, "type": "message", "response_id": respID, "status": "completed",
		"role": "user", "content": content,
	})
	if err := d.Items.Append(c.Request.Context(), &domainitem.Item{
		ID: itemID, SessionID: row.ID, ItemType: "message", ResponseID: respID,
		Status: "completed", Position: pos, Payload: payload, CreatedAt: time.Now(),
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "persist failed"})
		return
	}
	_ = d.Sessions.TouchUpdatedAt(c.Request.Context(), row.ID)
	if d.Updates != nil {
		d.Updates.NotifyChanged(row.ID)
	}
	author, _ := c.Get("email")
	authorStr, _ := author.(string)
	d.Hub.StartTurn(row.ID, respID)
	d.Hub.Publish(row.ID, formatSSE("session.input.consumed", map[string]any{
		"type": "session.input.consumed",
		"data": map[string]any{
			"item_id": itemID, "type": "message", "created_by": authorStr,
			"data": map[string]any{"role": "user", "content": content},
		},
	}))
	now := time.Now().Unix()
	d.Hub.Publish(row.ID, formatSSE("response.created", map[string]any{
		"id": respID, "status": "in_progress", "model": "", "created_at": now,
		"conversation": map[string]any{"id": row.ID},
	}))
	d.Hub.Publish(row.ID, formatSSE("response.in_progress", map[string]any{
		"id": respID, "status": "in_progress", "model": "", "created_at": now,
	}))
	if !d.checkCostBudget(c, pod.PodKey) {
		return
	}
	if err := d.CommandSender.SendPrompt(c.Request.Context(), pod.RunnerID, pod.PodKey, prompt); err != nil {
		if errors.Is(err, runnerservice.ErrRunnerNotConnected) || errors.Is(err, runnerservice.ErrRunnerOffline) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "runner unavailable", "code": "runner_unavailable"})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": "send failed", "code": "runner_unreachable"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"queued": true, "item_id": itemID})
}
