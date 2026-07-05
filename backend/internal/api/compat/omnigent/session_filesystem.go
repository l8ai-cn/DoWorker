package omnigent

import (
	"net/http"
	"strings"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/gin-gonic/gin"
)

func (d *Deps) handleSessionEnvironment(c *gin.Context) {
	_, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	res, ok := d.execSandboxFs(c, pod, &runnerv1.SandboxFsCommand{Op: "list", Path: ""})
	if !ok {
		return
	}
	if res.GetError() != "" || res.GetWorkspaceRoot() == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "no filesystem environment"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":     c.Param("env"),
		"object": "session.environment",
		"metadata": gin.H{
			"root": res.GetWorkspaceRoot(),
			"home": firstNonEmpty(res.GetWorkspaceHome(), res.GetWorkspaceRoot()),
		},
	})
}

func (d *Deps) handleSessionFilesystemList(c *gin.Context) {
	_, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	path := strings.TrimPrefix(c.Param("filepath"), "/")
	res, ok := d.execSandboxFs(c, pod, &runnerv1.SandboxFsCommand{Op: "list", Path: path, PodKey: podKeyOf(pod)})
	if !ok {
		return
	}
	if fsAPIError(c, res) {
		return
	}
	if res.GetContent() != "" {
		c.JSON(http.StatusOK, fileContentWire(path, res))
		return
	}
	c.JSON(http.StatusOK, listWire(res.GetEntries()))
}

func (d *Deps) handleSessionFilesystemWrite(c *gin.Context) {
	_, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	var body struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Encoding != "" && body.Encoding != "utf-8" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	path := strings.TrimPrefix(c.Param("filepath"), "/")
	res, ok := d.execSandboxFs(c, pod, &runnerv1.SandboxFsCommand{
		Op: "write", Path: path, Payload: body.Content, PodKey: podKeyOf(pod),
	})
		if !ok || fsAPIError(c, res) {
		return
	}
	c.Status(http.StatusNoContent)
}

func (d *Deps) handleSessionFilesystemChanges(c *gin.Context) {
	_, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	res, ok := d.execSandboxFs(c, pod, &runnerv1.SandboxFsCommand{Op: "changes", PodKey: podKeyOf(pod)})
	if !ok || fsAPIError(c, res) {
		return
	}
	c.JSON(http.StatusOK, changesWire(res.GetChanges()))
}

func (d *Deps) handleSessionFilesystemDiff(c *gin.Context) {
	_, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	path := strings.TrimPrefix(c.Param("filepath"), "/")
	res, ok := d.execSandboxFs(c, pod, &runnerv1.SandboxFsCommand{Op: "diff", Path: path, PodKey: podKeyOf(pod)})
	if !ok || fsAPIError(c, res) {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"object": "session.environment.filesystem.file_diff",
		"path":   path,
		"before": nullIfEmpty(res.GetBefore()),
		"after":  nullIfEmpty(res.GetAfter()),
	})
}

func (d *Deps) handleSessionFilesystemSearch(c *gin.Context) {
	_, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	res, ok := d.execSandboxFs(c, pod, &runnerv1.SandboxFsCommand{
		Op: "search", Payload: c.Query("q"),
		IncludeGlob: c.Query("include"), ExcludeGlob: c.Query("exclude"),
		PodKey: podKeyOf(pod),
	})
	if !ok || fsAPIError(c, res) {
		return
	}
	c.JSON(http.StatusOK, listWire(res.GetEntries()))
}

func (d *Deps) execSandboxFs(c *gin.Context, pod *podDomain.Pod, cmd *runnerv1.SandboxFsCommand) (*runnerv1.SandboxFsResultEvent, bool) {
	if d.SandboxFs == nil || pod == nil || pod.RunnerID == 0 || pod.PodKey == "" {
		writeRunnerUnavailable(c)
		return nil, false
	}
	if cmd.PodKey == "" {
		cmd.PodKey = pod.PodKey
	}
	if !d.SandboxFs.IsConnected(pod.RunnerID) {
		writeRunnerUnavailable(c)
		return nil, false
	}
	res, err := d.SandboxFs.Exec(c.Request.Context(), pod.RunnerID, cmd)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": gin.H{"message": err.Error()}})
		return nil, false
	}
	return res, true
}

func writeRunnerUnavailable(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{"code": "runner_unavailable", "message": "runner unavailable"}})
}

func fsAPIError(c *gin.Context, res *runnerv1.SandboxFsResultEvent) bool {
	if res == nil || res.GetError() == "" {
		return false
	}
	msg := res.GetError()
	switch {
	case strings.Contains(msg, "workspace not configured"), strings.Contains(msg, "pod not found"):
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": msg}})
	case strings.Contains(msg, "not found"):
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": msg}})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": msg}})
	}
	return true
}

func podKeyOf(pod *podDomain.Pod) string {
	if pod == nil {
		return ""
	}
	return pod.PodKey
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
