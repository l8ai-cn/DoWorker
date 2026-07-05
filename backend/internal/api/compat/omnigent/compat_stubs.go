package omnigent

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (d *Deps) handleListSessionPolicies(c *gin.Context) {
	if _, _, ok := d.authorizeSession(c, c.Param("id")); !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{"object": "list", "data": []any{}})
}

func (d *Deps) handleCreateSessionPolicy(c *gin.Context) {
	if _, _, ok := d.authorizeSession(c, c.Param("id")); !ok {
		return
	}
	c.JSON(http.StatusNotImplemented, gin.H{"error": "session-scoped policies use org defaults"})
}

func (d *Deps) handleDeleteSessionPolicy(c *gin.Context) {
	if _, _, ok := d.authorizeSession(c, c.Param("id")); !ok {
		return
	}
	c.Status(http.StatusNoContent)
}

func (d *Deps) handleSwitchAgent(c *gin.Context) {
	if _, _, ok := d.authorizeSession(c, c.Param("id")); !ok {
		return
	}
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{"code": "not_implemented", "message": "switch-agent is not available yet"},
	})
}

func (d *Deps) handleCreateHostDirectory(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"detail": "host directory creation is not available on self-hosted runners",
	})
}

func (d *Deps) handleSessionFilesystemStub(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{"code": "not_implemented", "message": "sandbox filesystem API is not available"},
	})
}
