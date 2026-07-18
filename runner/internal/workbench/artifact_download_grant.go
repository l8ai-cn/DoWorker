package workbench

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"

	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
)

func artifactDownloadGrant(
	artifactID string,
	revision uint64,
	representationIDs []string,
	issuedAt string,
) *agentworkbenchv2.ArtifactGrant {
	minimumRevision := revision
	maximumRevision := revision
	return &agentworkbenchv2.ArtifactGrant{
		GrantId:           artifactDownloadGrantID(artifactID, revision),
		Issuer:            stringPointer("agentsmesh.runner"),
		Subject:           stringPointer("session.viewer"),
		RepresentationIds: append([]string(nil), representationIDs...),
		Actions:           []string{"artifact.download"},
		MinimumRevision:   &minimumRevision,
		MaximumRevision:   &maximumRevision,
		IssuedAt:          &issuedAt,
	}
}

func artifactDownloadGrantID(artifactID string, revision uint64) string {
	digest := sha256.Sum256([]byte(
		artifactID + "\x00" + strconv.FormatUint(revision, 10),
	))
	return "grant-" + hex.EncodeToString(digest[:16])
}
