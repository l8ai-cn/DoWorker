package omnigent

import (
	"net/http"
	"strings"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

const compatRunnerWorkspace = "/home/runner/workspace"

func (d *Deps) handleHostFilesystem(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.Runner == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if !d.compatHostExists(c, c.Param("id")) {
		c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
		return
	}
	listed := normalizeCompatFSPath(c.Param("filepath"))
	c.JSON(http.StatusOK, gin.H{
		"object":   "list",
		"data":     compatFilesystemEntries(listed),
		"has_more": false,
	})
}

func (d *Deps) compatHostExists(c *gin.Context, hostID string) bool {
	nodeID := strings.TrimPrefix(hostID, "host_")
	tenant := middleware.GetTenant(c)
	r, err := d.Runner.GetByNodeIDAndOrgID(c.Request.Context(), nodeID, tenant.OrganizationID)
	return err == nil && r != nil && r.IsEnabled
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

func compatFilesystemEntries(path string) []map[string]any {
	now := time.Now().Unix()
	dir := func(name, abs string) map[string]any {
		return map[string]any{
			"name":        name,
			"path":        abs,
			"type":        "directory",
			"bytes":       nil,
			"modified_at": now,
		}
	}
	switch path {
	case "":
		return []map[string]any{dir("workspace", compatRunnerWorkspace)}
	case compatRunnerWorkspace:
		return nil
	default:
		return nil
	}
}
