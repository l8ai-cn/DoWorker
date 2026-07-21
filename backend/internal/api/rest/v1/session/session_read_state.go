package sessionapi

import (
	"net/http"

	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func (d *Deps) handleGetReadState(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.ReadState == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	row, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	entry, ok := d.ReadState.Get(tenant.UserID, row.ID)
	if !ok {
		c.JSON(http.StatusOK, gin.H{"last_seen": nil, "unread": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"last_seen": entry.LastSeen, "unread": entry.Unread})
}

func (d *Deps) handlePutReadState(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.ReadState == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	row, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	var body struct {
		LastSeen *int64 `json:"last_seen"`
		Unread   *bool  `json:"unread"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	entry, _ := d.ReadState.Get(tenant.UserID, row.ID)
	if body.LastSeen != nil {
		entry.LastSeen = *body.LastSeen
	}
	if body.Unread != nil {
		entry.Unread = *body.Unread
	}
	d.ReadState.Put(tenant.UserID, row.ID, entry)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (d *Deps) enrichReadState(userID int64, sessionID string, item *conversationListItem) {
	if d.ReadState == nil || item == nil {
		return
	}
	entry, ok := d.ReadState.Get(userID, sessionID)
	if !ok {
		return
	}
	item.ViewerLastSeen = &entry.LastSeen
	item.ViewerUnread = entry.Unread
}
