package sessionapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	sessionfiledomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/sessionfile"
	"github.com/gin-gonic/gin"
)

type sessionArtifactRange struct {
	end    *int64
	start  *int64
	suffix *int64
}

func parseSessionArtifactRange(value string) (*sessionArtifactRange, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	if !strings.HasPrefix(value, "bytes=") {
		return nil, fmt.Errorf("unsupported range unit")
	}
	value = strings.TrimSpace(strings.TrimPrefix(value, "bytes="))
	if value == "" || strings.Contains(value, ",") {
		return nil, fmt.Errorf("exactly one byte range is required")
	}
	parts := strings.Split(value, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid byte range")
	}
	if parts[0] == "" {
		suffix, err := parseRangeNumber(parts[1])
		if err != nil || suffix == 0 {
			return nil, fmt.Errorf("invalid suffix range")
		}
		return &sessionArtifactRange{suffix: &suffix}, nil
	}
	start, err := parseRangeNumber(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid range start")
	}
	parsed := &sessionArtifactRange{start: &start}
	if parts[1] == "" {
		return parsed, nil
	}
	end, err := parseRangeNumber(parts[1])
	if err != nil || end < start {
		return nil, fmt.Errorf("invalid range end")
	}
	parsed.end = &end
	return parsed, nil
}

func (r *sessionArtifactRange) resolve(fileBytes int64) (int64, int64, error) {
	if fileBytes <= 0 {
		return 0, 0, fmt.Errorf("range cannot resolve against empty file")
	}
	if r.suffix != nil {
		length := min(*r.suffix, fileBytes)
		return fileBytes - length, fileBytes - 1, nil
	}
	if r.start == nil || *r.start >= fileBytes {
		return 0, 0, fmt.Errorf("range start exceeds file size")
	}
	end := fileBytes - 1
	if r.end != nil {
		end = min(*r.end, end)
	}
	return *r.start, end, nil
}

func parseRangeNumber(value string) (int64, error) {
	if value == "" || strings.HasPrefix(value, "+") || strings.HasPrefix(value, "-") {
		return 0, fmt.Errorf("invalid byte position")
	}
	return strconv.ParseInt(value, 10, 64)
}

func (d *Deps) serveArtifactRange(
	c *gin.Context,
	fileRow *sessionfiledomain.File,
) bool {
	header := c.GetHeader("Range")
	if strings.TrimSpace(header) == "" {
		return false
	}
	parsed, err := parseSessionArtifactRange(header)
	if err != nil || parsed == nil {
		writeUnsatisfiableArtifactRange(c, fileRow.Bytes)
		return true
	}
	start, end, err := parsed.resolve(fileRow.Bytes)
	if err != nil {
		writeUnsatisfiableArtifactRange(c, fileRow.Bytes)
		return true
	}
	reader, size, err := d.SessionFiles.OpenRange(
		c.Request.Context(),
		fileRow,
		start,
		end,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "open failed"})
		return true
	}
	defer reader.Close()
	expected := end - start + 1
	if size >= 0 && size != expected {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "artifact size mismatch"})
		return true
	}
	c.Header(
		"Content-Range",
		fmt.Sprintf("bytes %d-%d/%d", start, end, fileRow.Bytes),
	)
	c.DataFromReader(
		http.StatusPartialContent,
		expected,
		fileRow.ContentType,
		reader,
		nil,
	)
	return true
}

func writeUnsatisfiableArtifactRange(c *gin.Context, fileBytes int64) {
	c.Header("Content-Range", fmt.Sprintf("bytes */%d", fileBytes))
	c.Status(http.StatusRequestedRangeNotSatisfiable)
}
