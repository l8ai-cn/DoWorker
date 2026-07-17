package sessionapi

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/gin-gonic/gin"
)

const maxSessionArtifactChunkBytes int64 = 4 << 20

func (d *Deps) handleSessionArtifactContent(c *gin.Context) {
	_, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	requestedRange, err := parseSessionArtifactRange(c.GetHeader("Range"))
	if err != nil {
		c.Status(http.StatusRequestedRangeNotSatisfiable)
		return
	}
	path := strings.TrimPrefix(c.Param("filepath"), "/")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "artifact path is required"}})
		return
	}
	stat, ok := d.execSandboxFs(c, pod, &runnerv1.SandboxFsCommand{
		Op: "stat", Path: path, PodKey: podKeyOf(pod),
	})
	if !ok || fsAPIError(c, stat) {
		return
	}
	fileBytes := stat.GetFileBytes()
	if fileBytes < 0 {
		c.JSON(http.StatusBadGateway, gin.H{"error": gin.H{"message": "invalid artifact metadata"}})
		return
	}
	start, end, status, ok := resolveArtifactResponseRange(c, requestedRange, fileBytes)
	if !ok {
		return
	}
	contentLength := end - start + 1
	if fileBytes == 0 {
		contentLength = 0
	}
	var first []byte
	if contentLength > 0 {
		first, err = d.readSessionArtifactChunk(
			c.Request.Context(),
			pod,
			path,
			start,
			min(contentLength, maxSessionArtifactChunkBytes),
			fileBytes,
		)
		if err != nil {
			writeSessionArtifactReadError(c, err)
			return
		}
	}
	writeSessionArtifactHeaders(c, stat.GetContentType(), fileBytes, start, end, contentLength, status)
	if len(first) > 0 {
		if _, err := c.Writer.Write(first); err != nil {
			return
		}
	}
	offset := start + int64(len(first))
	for offset <= end {
		chunk, err := d.readSessionArtifactChunk(
			c.Request.Context(),
			pod,
			path,
			offset,
			min(end-offset+1, maxSessionArtifactChunkBytes),
			fileBytes,
		)
		if err != nil {
			_ = c.Error(err)
			return
		}
		if _, err := c.Writer.Write(chunk); err != nil {
			return
		}
		offset += int64(len(chunk))
	}
}

func resolveArtifactResponseRange(
	c *gin.Context,
	requested *sessionArtifactRange,
	fileBytes int64,
) (int64, int64, int, bool) {
	if requested == nil {
		if fileBytes == 0 {
			return 0, -1, http.StatusOK, true
		}
		return 0, fileBytes - 1, http.StatusOK, true
	}
	start, end, err := requested.resolve(fileBytes)
	if err != nil {
		c.Header("Content-Range", fmt.Sprintf("bytes */%d", fileBytes))
		c.Status(http.StatusRequestedRangeNotSatisfiable)
		return 0, 0, 0, false
	}
	return start, end, http.StatusPartialContent, true
}

func (d *Deps) readSessionArtifactChunk(
	ctx context.Context,
	pod *podDomain.Pod,
	path string,
	offset int64,
	length int64,
	fileBytes int64,
) ([]byte, error) {
	if d.SandboxFs == nil || pod == nil || !d.SandboxFs.IsConnected(pod.RunnerID) {
		return nil, fmt.Errorf("runner unavailable")
	}
	result, err := d.SandboxFs.Exec(ctx, pod.RunnerID, &runnerv1.SandboxFsCommand{
		Op: "read_bytes", Path: path, PodKey: pod.PodKey, Offset: offset, Length: length,
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("artifact read returned no result")
	}
	if result.GetError() != "" {
		return nil, fmt.Errorf("artifact read failed: %s", result.GetError())
	}
	data := result.GetContentBytes()
	if result.GetContentOffset() != offset ||
		result.GetFileBytes() != fileBytes ||
		int64(len(data)) != length {
		return nil, fmt.Errorf("invalid runner artifact response")
	}
	return data, nil
}

func writeSessionArtifactHeaders(
	c *gin.Context,
	contentType string,
	fileBytes int64,
	start int64,
	end int64,
	contentLength int64,
	status int,
) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Length", strconv.FormatInt(contentLength, 10))
	c.Header("Content-Type", contentType)
	c.Header("X-Content-Type-Options", "nosniff")
	if status == http.StatusPartialContent {
		c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileBytes))
	}
	c.Status(status)
}

func writeSessionArtifactReadError(c *gin.Context, err error) {
	status := http.StatusBadGateway
	if strings.Contains(err.Error(), "runner unavailable") {
		status = http.StatusServiceUnavailable
	}
	c.JSON(status, gin.H{"error": gin.H{"message": err.Error()}})
}
