package sessionapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	workbenchdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
)

var artifactSHA256 = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)
var errArtifactIntegrityMismatch = errors.New("artifact integrity mismatch")

type artifactSnapshotRepository interface {
	GetSnapshot(
		context.Context,
		string,
	) (*workbenchdomain.SessionState, error)
}

type verifiedArtifactRead struct {
	ArtifactID       string `json:"artifact_id"`
	Digest           string `json:"digest"`
	FileBytes        uint64 `json:"file_bytes"`
	RepresentationID string `json:"representation_id"`
	Revision         uint64 `json:"revision"`
}

func (d *Deps) resolveVerifiedArtifactRead(
	c *gin.Context,
	sessionID string,
	path string,
) (verifiedArtifactRead, bool) {
	artifactID := strings.TrimSpace(c.Query("artifact_id"))
	representationID := strings.TrimSpace(c.Query("representation_id"))
	revision, err := strconv.ParseUint(c.Query("revision"), 10, 64)
	if artifactID == "" || representationID == "" || err != nil || revision == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"message": "artifact identity is required",
		}})
		return verifiedArtifactRead{}, false
	}
	if d.ArtifactSnapshots == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{
			"message": "artifact verification is unavailable",
		}})
		return verifiedArtifactRead{}, false
	}
	state, err := d.ArtifactSnapshots.GetSnapshot(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"message": "artifact verification failed",
		}})
		return verifiedArtifactRead{}, false
	}
	snapshot := &agentworkbenchv2.SessionSnapshot{}
	if state == nil || proto.Unmarshal(state.Projection, snapshot) != nil ||
		snapshot.GetSessionId() != sessionID {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{
			"message": "artifact snapshot is unavailable",
		}})
		return verifiedArtifactRead{}, false
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
				representation.GetTransport().GetResourceId() != "workspace:"+path ||
				representation.ByteSize == nil ||
				!artifactSHA256.MatchString(representation.GetDigest()) {
				continue
			}
			return verifiedArtifactRead{
				ArtifactID: artifactID, Digest: representation.GetDigest(),
				FileBytes:        representation.GetByteSize(),
				RepresentationID: representationID, Revision: revision,
			}, true
		}
	}
	c.JSON(http.StatusConflict, gin.H{"error": gin.H{
		"message": "artifact identity does not match the published result",
	}})
	return verifiedArtifactRead{}, false
}

func (r verifiedArtifactRead) payload() string {
	raw, _ := json.Marshal(r)
	return string(raw)
}
