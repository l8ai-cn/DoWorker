package sessionapi

import (
	"net/http"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/gin-gonic/gin"
)

const terminalMainID = "terminal_tui_main"

func terminalResourceWire(sessionID, terminalID string) map[string]any {
	return map[string]any{
		"id":         terminalID,
		"object":     "session.resource",
		"type":       "terminal",
		"session_id": sessionID,
		"name":       "main:tui",
		"metadata": map[string]any{
			"terminal_name": "tui",
			"session_key":   "main",
			"running":       true,
		},
	}
}

func (d *Deps) handleListTerminals(c *gin.Context) {
	row, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	if pod == nil || !pod.IsActive() || pod.InteractionMode != podDomain.InteractionModePTY {
		c.JSON(http.StatusOK, gin.H{"data": []any{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": []any{terminalResourceWire(row.ID, terminalMainID)},
	})
}

func (d *Deps) handleCreateTerminal(c *gin.Context) {
	row, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	c.JSON(http.StatusOK, terminalResourceWire(row.ID, terminalMainID))
}
