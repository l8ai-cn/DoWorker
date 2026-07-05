package omnigent

import (
	"net/http"
	"strings"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/gin-gonic/gin"
)

func (d *Deps) handleHostFilesystem(c *gin.Context) {
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
	if !d.SandboxFs.IsConnected(r.ID) {
		writeRunnerUnavailable(c)
		return
	}
	path := normalizeCompatFSPath(c.Param("filepath"))
	res, err := d.SandboxFs.Exec(c.Request.Context(), r.ID, &runnerv1.SandboxFsCommand{
		Op: "list_host", Path: path,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	if res.GetError() != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": res.GetError()})
		return
	}
	c.JSON(http.StatusOK, listWire(res.GetEntries(), res.GetWorkspaceRoot()))
}

func normalizeCompatFSPath(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "/" {
		return ""
	}
	if !strings.HasPrefix(raw, "/") {
		raw = "/" + raw
	}
	return strings.TrimSuffix(raw, "/")
}
