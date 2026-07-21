package workbench

import (
	"testing"

	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
	"github.com/stretchr/testify/require"
)

func TestReadyArtifactCarriesRevisionScopedDownloadGrant(t *testing.T) {
	descriptor := readyArtifactDescriptor(artifactFile{
		path:      "outputs/report.pdf",
		filename:  "report.pdf",
		mediaType: "application/pdf",
		digest:    "sha256:report",
		byteSize:  42,
	}, 3)

	require.Len(t, descriptor.GetGrants(), 1)
	grant := descriptor.GetGrants()[0]
	require.NotEmpty(t, grant.GetGrantId())
	require.Equal(t, []string{"artifact.download"}, grant.GetActions())
	require.Equal(t, []string{"original"}, grant.GetRepresentationIds())
	require.Equal(t, uint64(3), grant.GetMinimumRevision())
	require.Equal(t, uint64(3), grant.GetMaximumRevision())
	require.NotEmpty(t, grant.GetIssuedAt())

	deleted := deletedArtifactDescriptor(artifactFile{
		path: "outputs/report.pdf",
	}, 4)
	require.Empty(t, deleted.GetGrants())
	require.Equal(t,
		agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_DELETED,
		deleted.GetStatus(),
	)
}
