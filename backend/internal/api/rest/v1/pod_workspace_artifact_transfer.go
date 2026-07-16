package v1

import (
	"context"
	"errors"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	fileservice "github.com/anthropics/agentsmesh/backend/internal/service/file"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/gin-gonic/gin"
)

const workspaceArtifactUploadTimeout = 160 * time.Second

func (h *PodHandler) TransferWorkspaceArtifact(c *gin.Context) {
	pod, ok := h.authorizeReadablePod(c)
	if !ok {
		return
	}
	if h.workspaceArtifacts == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{"message": "artifact transfer unavailable"}})
		return
	}
	if _, active := h.artifactTransfers.LoadOrStore(pod.PodKey, struct{}{}); active {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": gin.H{"message": "workspace artifact transfer already in progress"}})
		return
	}
	defer h.artifactTransfers.Delete(pod.PodKey)
	path := strings.TrimPrefix(c.Param("filepath"), "/")
	stat, ok := h.execPodWorkspace(c, pod, &runnerv1.SandboxFsCommand{
		PodKey: pod.PodKey, Op: "stat", Path: path,
	})
	if !ok {
		return
	}
	filename := filepath.Base(path)
	transfer, err := h.workspaceArtifacts.PrepareWorkspaceArtifactTransfer(
		c.Request.Context(),
		pod.OrganizationID,
		filename,
		stat.GetContentType(),
		stat.GetFileBytes(),
	)
	if err != nil {
		writeWorkspaceArtifactTransferError(c, err)
		return
	}
	defer h.deleteWorkspaceArtifact(transfer)
	uploadCtx, cancelUpload := context.WithTimeout(context.Background(), workspaceArtifactUploadTimeout)
	defer cancelUpload()
	uploaded, ok := h.execPodWorkspaceContext(c, uploadCtx, pod, &runnerv1.SandboxFsCommand{
		PodKey: pod.PodKey, Op: "upload", Path: path, Payload: transfer.PutURL,
	})
	if !ok {
		return
	}
	if uploaded.GetFileBytes() != transfer.Size {
		c.JSON(http.StatusBadGateway, gin.H{"error": gin.H{"message": "workspace artifact changed during transfer"}})
		return
	}
	reader, size, err := h.workspaceArtifacts.OpenWorkspaceArtifact(c.Request.Context(), transfer)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": gin.H{"message": "workspace artifact open failed"}})
		return
	}
	defer reader.Close()
	if size != transfer.Size {
		c.JSON(http.StatusBadGateway, gin.H{"error": gin.H{"message": "workspace artifact size mismatch"}})
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
	c.Header("X-Content-Type-Options", "nosniff")
	c.DataFromReader(http.StatusOK, size, transfer.ContentType, reader, nil)
}

func (h *PodHandler) deleteWorkspaceArtifact(transfer *fileservice.WorkspaceArtifactTransfer) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h.workspaceArtifacts.DeleteWorkspaceArtifact(ctx, transfer); err != nil {
		slog.Warn("workspace artifact cleanup failed", "key", transfer.Key, "error", err)
	}
}

func writeWorkspaceArtifactTransferError(c *gin.Context, err error) {
	if errors.Is(err, fileservice.ErrFileTooLarge) {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{"message": "artifact transfer unavailable"}})
}
