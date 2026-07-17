package sessionapi

import (
	"encoding/json"
	"errors"
	"net/http"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/gin-gonic/gin"
)

func (d *Deps) handleGetElicitation(c *gin.Context) {
	_, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok || d.Elicitations == nil {
		return
	}
	rec, found := d.Elicitations.Get(c.Param("id"), c.Param("elicitation_id"))
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "elicitation not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":          rec.Status,
		"message":         rec.Message,
		"phase":           rec.Phase,
		"policy_name":     "tool_call_approval",
		"content_preview": "",
	})
}

func (d *Deps) handleResolveElicitation(c *gin.Context) {
	row, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok || !d.requireSessionLevel(c, row, levelEdit) || d.Elicitations == nil || d.CommandSender == nil {
		return
	}
	elicitID := c.Param("elicitation_id")
	rec, found := d.Elicitations.Get(row.ID, elicitID)
	if !found || rec.Status != "pending" {
		c.JSON(http.StatusNotFound, gin.H{"error": "elicitation not found"})
		return
	}
	var body struct {
		Action  string         `json:"action"`
		Content map[string]any `json:"content"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	approved := body.Action == "accept"
	if err := d.forwardPermissionResponse(c, pod, rec.RequestID, approved, body.Content); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "forward failed", "code": "runner_unreachable"})
		return
	}
	d.Elicitations.Resolve(row.ID, elicitID)
	if d.Stream != nil {
		d.Stream.PublishElicitationResolved(row.ID, elicitID)
	}
	c.JSON(http.StatusAccepted, gin.H{"queued": true})
}

func (d *Deps) forwardPermissionResponse(c *gin.Context, pod *podDomain.Pod, requestID string, approved bool, content map[string]any) error {
	if pod == nil {
		return errors.New("pod unavailable")
	}
	payload, _ := json.Marshal(map[string]any{
		"type": "permission_response", "requestId": requestID, "approved": approved,
		"updatedInput": content,
	})
	return d.CommandSender.SendAcpRelay(c.Request.Context(), pod.RunnerID, pod.PodKey, string(payload))
}
