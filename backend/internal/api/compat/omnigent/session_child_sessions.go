package omnigent

import (
	"encoding/json"
	"net/http"
	"strings"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	"github.com/gin-gonic/gin"
)

func parseChildTitle(title *string) (tool, sessionName string) {
	if title == nil {
		return "", ""
	}
	parts := strings.SplitN(strings.TrimSpace(*title), ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", *title
}

func (d *Deps) handleListChildSessions(c *gin.Context) {
	row, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok || d.Sessions == nil {
		return
	}
	children, err := d.Sessions.ListChildren(c.Request.Context(), row.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list failed"})
		return
	}
	data := make([]map[string]any, 0, len(children))
	for i := range children {
		data = append(data, d.childSessionWire(c, &children[i]))
	}
	c.JSON(http.StatusOK, gin.H{"object": "list", "data": data})
}

func (d *Deps) childSessionWire(c *gin.Context, row *domain.Session) map[string]any {
	tool, sessionName := parseChildTitle(row.Title)
	pod := d.loadPod(c, row.PodKey)
	status := mapSessionStatus(pod)
	busy := status == "running" || status == "waiting" || status == "launching"
	pending := 0
	if d.Elicitations != nil {
		pending = len(d.Elicitations.PendingPayloads(row.ID))
	}
	return map[string]any{
		"id": row.ID, "title": row.Title, "tool": strOrNull(tool),
		"session_name": strOrNull(sessionName), "labels": map[string]string{},
		"current_task_status": statusOrNull(status), "last_task_error": nil,
		"busy": busy, "last_message_preview": d.lastMessagePreview(c, row.ID),
		"pending_elicitations_count": pending,
	}
}

func (d *Deps) lastMessagePreview(c *gin.Context, sessionID string) any {
	if d.Items == nil {
		return nil
	}
	page, err := d.Items.ListPage(c.Request.Context(), sessionID, 1, "", true)
	if err != nil || len(page.Items) == 0 {
		return nil
	}
	var payload struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if json.Unmarshal(page.Items[0].Payload, &payload) != nil {
		return nil
	}
	var parts []string
	for _, b := range payload.Content {
		switch b.Type {
		case "text", "input_text", "output_text":
			if t := strings.TrimSpace(b.Text); t != "" {
				parts = append(parts, t)
			}
		}
	}
	if len(parts) == 0 {
		return nil
	}
	text := strings.Join(parts, " ")
	const maxLen = 150
	if len(text) > maxLen {
		return text[:maxLen] + "…"
	}
	return text
}

func strOrNull(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func statusOrNull(s string) any {
	if s == "" || s == "idle" {
		return nil
	}
	return s
}
