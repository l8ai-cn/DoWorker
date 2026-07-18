package sessionapi

import (
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
)

func (d *Deps) handleGetSessionArtifactRepresentation(c *gin.Context) {
	session, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	if embedClaims(c) == nil &&
		!d.requireSessionLevel(c, session, levelOwner) {
		return
	}
	if d.WorkbenchRepo == nil || d.SessionFiles == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "artifact service unavailable"})
		return
	}
	revision, err := strconv.ParseUint(c.Query("revision"), 10, 64)
	if err != nil || revision == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid artifact revision"})
		return
	}
	artifactID := strings.TrimSpace(c.Query("artifact_id"))
	representationID := strings.TrimSpace(c.Query("representation_id"))
	digest := strings.TrimSpace(c.Query("digest"))
	if artifactID == "" || representationID == "" || digest == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "artifact identity is required"})
		return
	}
	representation, fileID, ok := d.authorizeArtifactRepresentation(
		c,
		session.ID,
		artifactID,
		representationID,
		revision,
		digest,
	)
	if !ok {
		return
	}
	fileRow, err := d.SessionFiles.GetForSession(c.Request.Context(), session.ID, fileID)
	if err != nil ||
		fileRow.Bytes != int64(representation.GetByteSize()) ||
		fileRow.ContentType != representation.GetMediaType() {
		c.JSON(http.StatusNotFound, gin.H{"error": "artifact content not found"})
		return
	}
	reader, size, err := d.SessionFiles.Open(c.Request.Context(), fileRow)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "open failed"})
		return
	}
	defer reader.Close()
	if size >= 0 && size != fileRow.Bytes {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "artifact size mismatch"})
		return
	}
	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.DataFromReader(
		http.StatusOK,
		fileRow.Bytes,
		fileRow.ContentType,
		reader,
		map[string]string{
			"Content-Disposition": mime.FormatMediaType(
				"inline",
				map[string]string{"filename": fileRow.Filename},
			),
		},
	)
}

func (d *Deps) authorizeArtifactRepresentation(
	c *gin.Context,
	sessionID string,
	artifactID string,
	representationID string,
	revision uint64,
	digest string,
) (*agentworkbenchv2.ArtifactRepresentation, string, bool) {
	stored, err := d.WorkbenchRepo.GetSnapshot(c.Request.Context(), sessionID)
	if err != nil || stored == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "artifact snapshot not found"})
		return nil, "", false
	}
	snapshot := &agentworkbenchv2.SessionSnapshot{}
	if err := proto.Unmarshal(stored.Projection, snapshot); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "artifact snapshot invalid"})
		return nil, "", false
	}
	if !supportsArtifactDownload(snapshot.GetCapabilities()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "artifact download is not supported"})
		return nil, "", false
	}
	for _, artifact := range snapshot.GetArtifacts() {
		if artifact.GetArtifactId() != artifactID ||
			artifact.GetRevision() != revision {
			continue
		}
		for _, representation := range artifact.GetRepresentations() {
			if representation.GetRepresentationId() != representationID ||
				representation.GetRevision() != revision ||
				representation.GetStatus() !=
					agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY ||
				representation.GetDigest() != digest ||
				representation.ByteSize == nil {
				continue
			}
			resourceID := representation.GetTransport().GetResourceId()
			if !strings.HasPrefix(resourceID, "session-file:") {
				break
			}
			if !artifactDownloadGranted(
				artifact,
				representationID,
				time.Now().UTC(),
			) {
				break
			}
			fileID := strings.TrimPrefix(resourceID, "session-file:")
			if fileID != "" {
				return representation, fileID, true
			}
		}
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "artifact download is not authorized"})
	return nil, "", false
}

func artifactDownloadGranted(
	artifact *agentworkbenchv2.ArtifactDescriptor,
	representationID string,
	now time.Time,
) bool {
	for _, grant := range artifact.GetGrants() {
		if grant.GetGrantId() == "" ||
			!containsString(grant.GetActions(), "artifact.download") ||
			(len(grant.GetRepresentationIds()) > 0 &&
				!containsString(grant.GetRepresentationIds(), representationID)) ||
			(grant.MinimumRevision != nil &&
				artifact.GetRevision() < grant.GetMinimumRevision()) ||
			(grant.MaximumRevision != nil &&
				artifact.GetRevision() > grant.GetMaximumRevision()) {
			continue
		}
		if grant.ExpiresAt == nil {
			return true
		}
		expiresAt, err := time.Parse(time.RFC3339Nano, grant.GetExpiresAt())
		if err == nil && expiresAt.After(now) {
			return true
		}
	}
	return false
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func supportsArtifactDownload(
	capabilities *agentworkbenchv2.SupportCapabilities,
) bool {
	for _, action := range capabilities.GetArtifactOperations() {
		if action == "artifact.download" {
			return true
		}
	}
	return false
}
