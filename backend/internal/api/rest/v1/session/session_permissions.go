package sessionapi

import (
	"net/http"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentsession"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func (d *Deps) handleListPermissions(c *gin.Context) {
	row, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	if !d.requireSessionLevel(c, row, levelOwner) {
		return
	}
	if d.SessionPermissions == nil {
		c.JSON(http.StatusOK, []gin.H{})
		return
	}
	grants, err := d.SessionPermissions.List(c.Request.Context(), row.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list failed"})
		return
	}
	out := make([]gin.H, 0, len(grants))
	for _, g := range grants {
		out = append(out, gin.H{
			"user_id": g.UserID, "conversation_id": row.ID, "level": g.Level,
		})
	}
	c.JSON(http.StatusOK, out)
}

func (d *Deps) handlePutPermission(c *gin.Context) {
	row, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	if !d.requireSessionLevel(c, row, levelOwner) {
		return
	}
	if d.SessionPermissions == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "unavailable"})
		return
	}
	var body struct {
		UserID string `json:"user_id"`
		Level  int    `json:"level"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.UserID == "" || body.Level < levelRead || body.Level > levelOwner {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	grant, err := d.SessionPermissions.Upsert(c.Request.Context(), row.ID, body.UserID, body.Level)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "grant failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user_id": grant.UserID, "conversation_id": row.ID, "level": grant.Level,
	})
}

func (d *Deps) handleDeletePermission(c *gin.Context) {
	row, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	if !d.requireSessionLevel(c, row, levelOwner) {
		return
	}
	if d.SessionPermissions == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "unavailable"})
		return
	}
	if err := d.SessionPermissions.Delete(c.Request.Context(), row.ID, c.Param("user_id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (d *Deps) handleGetSessionOwner(c *gin.Context) {
	row, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	owner := d.ownerLabel(c, row.UserID)
	c.JSON(http.StatusOK, gin.H{"owner": owner})
}

func (d *Deps) ownerLabel(c *gin.Context, userID int64) *string {
	if d.User == nil {
		return nil
	}
	u, err := d.User.GetByID(c.Request.Context(), userID)
	if err != nil || u == nil {
		return nil
	}
	if u.Email != "" {
		return &u.Email
	}
	if u.Username != "" {
		return &u.Username
	}
	return nil
}

func (d *Deps) enrichOwnership(c *gin.Context, row *domain.Session, item *conversationListItem) {
	if item == nil || row == nil {
		return
	}
	level := d.sessionAccessLevel(c, row)
	item.PermissionLevel = &level
	item.Owner = d.ownerLabel(c, row.UserID)
	tenant := middleware.GetTenant(c)
	if tenant != nil {
		d.enrichReadState(tenant.UserID, item.ID, item)
	}
}
