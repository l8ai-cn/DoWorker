package sessionapi

import (
	"encoding/json"
	"net/http"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var updatesUpgrader = websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}

func (d *Deps) handleSessionUpdates(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	userID := middleware.GetUserID(c)
	if tenant == nil || userID == 0 || d.Updates == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	ws, err := updatesUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	peer := d.Updates.Register(userID, tenant.OrganizationID)
	defer d.Updates.Unregister(peer)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, data, err := ws.ReadMessage()
			if err != nil {
				return
			}
			var msg struct {
				Type       string   `json:"type"`
				SessionIDs []string `json:"session_ids"`
			}
			if json.Unmarshal(data, &msg) != nil || msg.Type != "watch" {
				continue
			}
			peer.SetWatch(msg.SessionIDs)
		}
	}()
	for {
		select {
		case <-done:
			return
		case body, ok := <-peer.Out():
			if !ok {
				return
			}
			if err := ws.WriteMessage(websocket.TextMessage, body); err != nil {
				return
			}
		}
	}
}
