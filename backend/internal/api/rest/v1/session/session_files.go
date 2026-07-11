package sessionapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	domainfile "github.com/anthropics/agentsmesh/backend/internal/domain/sessionfile"
	fileservice "github.com/anthropics/agentsmesh/backend/internal/service/file"
	sessionfilesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionfile"
	"github.com/gin-gonic/gin"
)

const maxMultipartMemory = 32 << 20

func (d *Deps) handleUploadSessionFile(c *gin.Context) {
	row, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	if d.SessionFiles == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "file upload unavailable"})
		return
	}
	if err := c.Request.ParseMultipartForm(maxMultipartMemory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart form"})
		return
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file field required"})
		return
	}
	defer file.Close()
	filename := strings.TrimSpace(header.Filename)
	if filename == "" {
		filename = "image.png"
	}
	rowFile, err := d.SessionFiles.Create(c.Request.Context(), sessionfilesvc.CreateInput{
		SessionID: row.ID, Filename: filename,
		ContentType: header.Header.Get("Content-Type"),
		Reader:      file, Size: header.Size,
	})
	if err != nil {
		writeSessionFileError(c, err)
		return
	}
	c.JSON(http.StatusOK, sessionFileWire(rowFile))
}

func (d *Deps) handleGetSessionFileContent(c *gin.Context) {
	row, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	if d.SessionFiles == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "file service unavailable"})
		return
	}
	fileRow, err := d.SessionFiles.GetForSession(c.Request.Context(), row.ID, c.Param("file_id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	reader, size, err := d.SessionFiles.Open(c.Request.Context(), fileRow)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "open failed"})
		return
	}
	defer reader.Close()
	if size < 0 {
		size = fileRow.Bytes
	}
	extra := map[string]string{}
	if size >= 0 {
		extra["Content-Length"] = strconv.FormatInt(size, 10)
	}
	c.DataFromReader(http.StatusOK, size, fileRow.ContentType, reader, extra)
}

func sessionFileWire(row *domainfile.File) map[string]any {
	return map[string]any{
		"id": row.ID, "object": "file", "type": "file", "name": row.Filename,
		"session_id": row.SessionID,
		"metadata": map[string]any{
			"filename": row.Filename, "bytes": row.Bytes, "created_at": row.CreatedAt.Unix(),
		},
	}
}

func writeSessionFileError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, fileservice.ErrFileTooLarge):
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
	case errors.Is(err, fileservice.ErrInvalidFileType):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, sessionfilesvc.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
	}
}
