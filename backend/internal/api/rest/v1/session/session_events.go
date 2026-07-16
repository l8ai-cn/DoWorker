package sessionapi

import (
	"encoding/json"
	"errors"
	"net/http"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
	sessionmessagesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionmessage"
	"github.com/gin-gonic/gin"
)

func (d *Deps) handlePostEvent(c *gin.Context) {
	row, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok || !d.requireSessionLevel(c, row, levelEdit) {
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
	if !authorizeEmbedEvent(c, evt.Type) {
		return
	}
	switch evt.Type {
	case "message":
		d.postMessageEvent(c, row, pod, evt.Data)
	case "stop_session":
		d.postStopSessionEvent(c, pod)
	case "interrupt":
		d.postInterruptEvent(c, row, pod)
	default:
		c.JSON(http.StatusAccepted, gin.H{"queued": false})
	}
}

func authorizeEmbedEvent(c *gin.Context, eventType string) bool {
	claims := embedClaims(c)
	if claims == nil {
		return true
	}
	capability := "write"
	if eventType == "interrupt" || eventType == "stop_session" {
		capability = "control"
	}
	if hasEmbedCapability(claims, capability) {
		return true
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	return false
}

func (d *Deps) postMessageEvent(c *gin.Context, row *domain.Session, pod *podDomain.Pod, data json.RawMessage) {
	if d.MessageOutbox == nil || pod == nil || d.Hub == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "unavailable"})
		return
	}
	var err error
	pod, err = d.ensureMessagePod(c.Request.Context(), row, pod)
	if err != nil {
		writeSessionPodError(c, err)
		return
	}
	content, _ := parseMessageContent(data)
	if !messageHasContent(content) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty message"})
		return
	}
	attachmentPaths, err := stageMessageAttachments(
		c.Request.Context(),
		d.SessionFiles,
		d.SandboxFs,
		pod,
		row.ID,
		messageAttachments(data),
	)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "attachment delivery failed"})
		return
	}
	prompt := materializedMessagePrompt(data, attachmentPaths)
	if !d.checkCostBudget(c, pod.PodKey) {
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
	item, err := sessionmessagesvc.UserItem(itemID, row.ID, respID, content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "payload failed"})
		return
	}
	err = d.MessageOutbox.PersistAndQueue(c.Request.Context(), sessionmessagesvc.PromptInput{
		OrganizationID: row.OrganizationID,
		RunnerID:       pod.RunnerID,
		PodKey:         pod.PodKey,
		Item:           item,
		Prompt:         prompt,
	})
	if err != nil {
		if errors.Is(err, sessionmessagesvc.ErrUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "runner unavailable", "code": "runner_unavailable"})
			return
		}
		if errors.Is(err, podDomain.ErrQueueFull) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "runner queue full", "code": "runner_queue_full"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "persist failed"})
		return
	}
	d.maybeSeedSessionTitle(c.Request.Context(), row, prompt)
	if d.Sessions != nil {
		_ = d.Sessions.TouchUpdatedAt(c.Request.Context(), row.ID)
	}
	if d.Updates != nil {
		d.Updates.NotifyChanged(row.ID)
	}
	author, _ := c.Get("email")
	authorStr, _ := author.(string)
	if d.Stream != nil {
		d.Stream.publishTurnStarted(row.ID, respID)
		d.Stream.PublishInputConsumed(row.ID, itemID, authorStr, content)
	}
	c.JSON(http.StatusAccepted, gin.H{"queued": true, "item_id": itemID})
}
