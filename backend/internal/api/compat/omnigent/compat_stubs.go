package omnigent

import (
	"net/http"
	"strings"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
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

func (d *Deps) handleCreateHostDirectory(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.Runner == nil || d.SandboxFs == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	nodeID := strings.TrimPrefix(c.Param("id"), "host_")
	r, err := d.Runner.GetByNodeIDAndOrgID(c.Request.Context(), nodeID, tenant.OrganizationID)
	if err != nil || r == nil || !r.IsEnabled {
		c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
		return
	}
	var body struct {
		Path string `json:"path"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Path) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path required"})
		return
	}
	if !d.SandboxFs.IsConnected(r.ID) {
		writeRunnerUnavailable(c)
		return
	}
	res, err := d.SandboxFs.Exec(c.Request.Context(), r.ID, &runnerv1.SandboxFsCommand{
		Op: "mkdir", Path: body.Path,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"detail": err.Error()})
		return
	}
	if res.GetError() != "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": res.GetError()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"path": body.Path})
}
