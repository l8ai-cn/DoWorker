package v1

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	expertSvc "github.com/l8ai-cn/agentcloud/backend/internal/service/expert"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/gitops"
	"github.com/l8ai-cn/agentcloud/backend/pkg/apierr"
)

const (
	// maxAvatarBytes caps the decoded avatar size to guard against
	// decode-bomb / oversized-blob uploads committed into the repo.
	maxAvatarBytes = 2 * 1024 * 1024 // 2 MB
)

// avatarMIMEToExt is the allow-list of image types (sniffed via magic bytes),
// mapped to the platform-controlled file extension.
var avatarMIMEToExt = map[string]string{
	"image/png":  "png",
	"image/jpeg": "jpg",
	"image/webp": "webp",
	"image/gif":  "gif",
}

// validateAvatarInput decodes + validates a base64 avatar upload and returns
// the service-facing AvatarInput. A nil/empty input yields (nil, nil) so avatar
// is optional. The client filename is intentionally ignored; the stored path is
// always assets/avatar.<sniffed-ext>.
func validateAvatarInput(in *avatarInput) (*expertSvc.AvatarInput, error) {
	if in == nil || strings.TrimSpace(in.ContentBase64) == "" {
		return nil, nil
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(in.ContentBase64))
	if err != nil {
		return nil, fmt.Errorf("avatar: invalid base64 content: %w", err)
	}
	if len(data) == 0 {
		return nil, errors.New("avatar: empty content")
	}
	if len(data) > maxAvatarBytes {
		return nil, fmt.Errorf("avatar: exceeds max size of %d bytes", maxAvatarBytes)
	}
	// Sniff the real content type from magic bytes; never trust the client.
	sniffed := http.DetectContentType(data)
	if i := strings.IndexByte(sniffed, ';'); i >= 0 {
		sniffed = strings.TrimSpace(sniffed[:i])
	}
	ext, ok := avatarMIMEToExt[sniffed]
	if !ok {
		return nil, fmt.Errorf("avatar: unsupported type %q (allowed: png, jpeg, webp, gif)", sniffed)
	}
	return &expertSvc.AvatarInput{Data: data, Ext: ext}, nil
}

// sanitizeRepoPath neutralizes path traversal on repo-content read routes. It
// trims a leading slash, path.Cleans, and rejects absolute paths, "..", any
// ".." segment, and control/NUL bytes. Returns the cleaned repo-relative path.
func sanitizeRepoPath(raw string) (string, error) {
	p := strings.TrimPrefix(raw, "/")
	if p == "" {
		return "", errors.New("path is required")
	}
	for _, r := range p {
		if r == 0 || r < 0x20 {
			return "", errors.New("path contains control characters")
		}
	}
	cleaned := path.Clean(p)
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", errors.New("path escapes repository root")
	}
	if cleaned == "." || cleaned == "/" || strings.HasPrefix(cleaned, "/") {
		return "", errors.New("invalid path")
	}
	// Defense in depth: reject any residual ".." segment.
	for _, seg := range strings.Split(cleaned, "/") {
		if seg == ".." {
			return "", errors.New("path escapes repository root")
		}
	}
	return cleaned, nil
}

// GetExpertFile returns a single file from the expert's backing repo.
func (h *ExpertHandler) GetExpertFile(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	rel, err := sanitizeRepoPath(c.Param("path"))
	if err != nil {
		apierr.BadRequest(c, apierr.VALIDATION_FAILED, err.Error())
		return
	}
	data, entry, err := h.service.ReadExpertFile(c.Request.Context(), tenant.OrganizationID, c.Param("expertSlug"), rel)
	if err != nil {
		h.gitReadError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"path":    entry.Path,
		"name":    entry.Name,
		"size":    entry.Size,
		"sha":     entry.SHA,
		"content": base64.StdEncoding.EncodeToString(data),
	})
}

// GetExpertTree returns the file tree of the expert's backing repo.
func (h *ExpertHandler) GetExpertTree(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	entries, err := h.service.ListExpertTree(c.Request.Context(), tenant.OrganizationID, c.Param("expertSlug"))
	if err != nil {
		h.gitReadError(c, err)
		return
	}
	out := make([]gin.H, 0, len(entries))
	for _, e := range entries {
		out = append(out, gin.H{"path": e.Path, "name": e.Name, "type": e.Type, "size": e.Size, "sha": e.SHA})
	}
	c.JSON(http.StatusOK, gin.H{"entries": out})
}

func (h *ExpertHandler) gitReadError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, expertSvc.ErrGitBackingDisabled),
		errors.Is(err, gitops.ErrNotFound):
		apierr.ResourceNotFound(c, "Not found")
	default:
		h.notFoundOrInternal(c, err)
	}
}
