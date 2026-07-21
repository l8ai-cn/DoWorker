package agentworkbench

import (
	"testing"

	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
	"github.com/stretchr/testify/require"
)

func TestUpsertArtifactPreservesRevisionHistory(t *testing.T) {
	current := transitionArtifact(1, "tool-one")
	next := transitionArtifact(2, "tool-two")

	artifacts := upsertArtifact(
		[]*agentworkbenchv2.ArtifactDescriptor{current},
		next,
	)

	require.Len(t, artifacts, 1)
	require.Equal(t, uint64(2), artifacts[0].GetRevision())
	require.Len(t, artifacts[0].GetRevisions(), 2)
	require.Equal(t,
		"tool-one",
		artifacts[0].GetRevisions()[0].GetProvenance().GetToolExecutionId(),
	)
	require.Equal(t,
		"tool-two",
		artifacts[0].GetRevisions()[1].GetProvenance().GetToolExecutionId(),
	)
}

func TestToolArtifactReferenceUsesRevisionProvenance(t *testing.T) {
	artifact := transitionArtifact(2, "tool-two")
	artifact.Revisions = append(
		[]*agentworkbenchv2.ArtifactRevision{artifactRevision(1, "tool-one")},
		artifact.Revisions...,
	)
	catalog := map[string]*agentworkbenchv2.ArtifactDescriptor{
		artifact.GetArtifactId(): artifact,
	}

	require.NoError(t, validateToolArtifactReference(
		"tool-one",
		&agentworkbenchv2.ArtifactReference{
			ArtifactId: artifact.GetArtifactId(),
			Revision:   uint64Pointer(1),
		},
		catalog,
	))
	require.NoError(t, validateToolArtifactReference(
		"tool-two",
		&agentworkbenchv2.ArtifactReference{
			ArtifactId: artifact.GetArtifactId(),
			Revision:   uint64Pointer(2),
		},
		catalog,
	))
}
