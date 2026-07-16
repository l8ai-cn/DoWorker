package v1

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/pkg/apierr"
	"github.com/anthropics/agentsmesh/backend/pkg/policy"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

type podWorkspaceSandbox interface {
	IsConnected(runnerID int64) bool
	Exec(
		ctx context.Context,
		runnerID int64,
		command *runnerv1.SandboxFsCommand,
	) (*runnerv1.SandboxFsResultEvent, error)
}

func (h *PodHandler) ListWorkspaceArtifacts(c *gin.Context) {
	pod, ok := h.authorizeReadablePod(c)
	if !ok {
		return
	}
	result, ok := h.execPodWorkspace(c, pod, &runnerv1.SandboxFsCommand{
		PodKey: pod.PodKey,
		Op:     "changes",
	})
	if !ok {
		return
	}
	c.JSON(http.StatusOK, podWorkspaceChangesWire(result.GetChanges()))
}

func (h *PodHandler) ReadWorkspaceArtifact(c *gin.Context) {
	pod, ok := h.authorizeReadablePod(c)
	if !ok {
		return
	}
	path := strings.TrimPrefix(c.Param("filepath"), "/")
	result, ok := h.execPodWorkspace(c, pod, &runnerv1.SandboxFsCommand{
		PodKey: pod.PodKey,
		Op:     "read",
		Path:   path,
	})
	if !ok {
		return
	}
	c.JSON(http.StatusOK, podWorkspaceFileWire(path, result))
}

func (h *PodHandler) authorizeReadablePod(c *gin.Context) (*podDomain.Pod, bool) {
	if h.podService == nil {
		apierr.ResourceNotFound(c, "Pod not found")
		return nil, false
	}
	podKey := c.Param("key")
	pod, err := h.podService.GetPod(c.Request.Context(), podKey)
	if err != nil {
		apierr.ResourceNotFound(c, "Pod not found")
		return nil, false
	}
	tenant := middleware.GetTenant(c)
	if tenant == nil {
		apierr.ForbiddenAccess(c)
		return nil, false
	}
	subject := policy.NewSubject(tenant.OrganizationID, tenant.UserID, tenant.UserRole)
	resource := h.podResourceWithGrants(
		c.Request.Context(),
		podKey,
		pod.OrganizationID,
		pod.CreatedByID,
	)
	if !policy.PodPolicy.AllowRead(subject, resource) {
		apierr.ForbiddenAccess(c)
		return nil, false
	}
	return pod, true
}

func (h *PodHandler) execPodWorkspace(
	c *gin.Context,
	pod *podDomain.Pod,
	command *runnerv1.SandboxFsCommand,
) (*runnerv1.SandboxFsResultEvent, bool) {
	if h.sandboxFs == nil || pod.RunnerID == 0 || !h.sandboxFs.IsConnected(pod.RunnerID) {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": gin.H{"code": "runner_unavailable", "message": "runner unavailable"},
		})
		return nil, false
	}
	result, err := h.sandboxFs.Exec(c.Request.Context(), pod.RunnerID, command)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": gin.H{"message": err.Error()}})
		return nil, false
	}
	if result == nil || result.GetError() == "" {
		return result, true
	}
	message := result.GetError()
	status := http.StatusBadRequest
	if strings.Contains(message, "not found") ||
		strings.Contains(message, "workspace not configured") ||
		strings.Contains(message, "pod not found") {
		status = http.StatusNotFound
	}
	c.JSON(status, gin.H{"error": gin.H{"message": message}})
	return nil, false
}

func podWorkspaceChangesWire(changes []*runnerv1.SandboxFsChange) gin.H {
	data := make([]gin.H, 0, len(changes))
	for _, change := range changes {
		data = append(data, gin.H{
			"path":        change.GetPath(),
			"name":        change.GetName(),
			"status":      change.GetStatus(),
			"bytes":       nullablePodWorkspaceInt(change.GetBytes()),
			"modified_at": nullablePodWorkspaceInt(change.GetModifiedAt()),
		})
	}
	return gin.H{"object": "list", "data": data, "has_more": false}
}

func podWorkspaceFileWire(path string, result *runnerv1.SandboxFsResultEvent) gin.H {
	encoding := result.GetEncoding()
	if encoding == "" {
		encoding = "utf-8"
	}
	return gin.H{
		"object":       "pod.workspace.file_content",
		"path":         path,
		"content_type": result.GetContentType(),
		"encoding":     encoding,
		"content":      result.GetContent(),
		"bytes":        result.GetFileBytes(),
		"truncated":    result.GetTruncated(),
	}
}

func nullablePodWorkspaceInt(value int64) any {
	if value == 0 {
		return nil
	}
	return value
}
