package omnigent

import (
	"net/http"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

const levelOwner = 4

func (d *Deps) handleListPermissions(c *gin.Context) {
	row, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	c.JSON(http.StatusOK, []gin.H{})
	_ = row
}

func (d *Deps) handlePutPermission(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{"code": "not_implemented", "message": "session sharing is not enabled"},
	})
}

func (d *Deps) handleDeletePermission(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{"code": "not_implemented", "message": "session sharing is not enabled"},
	})
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

func (d *Deps) enrichOwnership(c *gin.Context, rowUserID int64, item *conversationListItem) {
	if item == nil {
		return
	}
	level := levelOwner
	item.PermissionLevel = &level
	item.Owner = d.ownerLabel(c, rowUserID)
	tenant := middleware.GetTenant(c)
	if tenant != nil {
		d.enrichReadState(tenant.UserID, item.ID, item)
	}
}
