package agentworkbench

import (
	"testing"

	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestValidateArtifactTransitionRequiresContinuousStableProducer(t *testing.T) {
	current := transitionArtifact(3, "tool-one")
	valid := proto.Clone(current).(*agentworkbenchv2.ArtifactDescriptor)
	valid.Revision = 4
	valid.Revisions = []*agentworkbenchv2.ArtifactRevision{
		artifactRevision(4, "tool-two"),
	}
	require.NoError(t, validateArtifactTransition(
		[]*agentworkbenchv2.ArtifactDescriptor{current},
		valid,
	))

	for name, mutate := range map[string]func(*agentworkbenchv2.ArtifactDescriptor){
		"skipped revision": func(value *agentworkbenchv2.ArtifactDescriptor) {
			value.Revision = 5
		},
		"changed producer": func(value *agentworkbenchv2.ArtifactDescriptor) {
			value.Revision = 4
			value.Provenance.ProducerType = stringPointer("image.generate")
		},
	} {
		t.Run(name, func(t *testing.T) {
			next := proto.Clone(current).(*agentworkbenchv2.ArtifactDescriptor)
			mutate(next)
			require.ErrorIs(t,
				validateArtifactTransition(
					[]*agentworkbenchv2.ArtifactDescriptor{current},
					next,
				),
				ErrInvalidBatch,
			)
		})
	}
}

func transitionArtifact(
	revision uint64,
	toolExecutionID string,
) *agentworkbenchv2.ArtifactDescriptor {
	return &agentworkbenchv2.ArtifactDescriptor{
		ArtifactId: "edited-image",
		Revision:   revision,
		Provenance: &agentworkbenchv2.ArtifactProvenance{
			ProducerNamespace: stringPointer("openai"),
			ProducerType:      stringPointer("image.edit"),
			ProducerId:        stringPointer("editor"),
			CommandId:         stringPointer("command-one"),
		},
		Revisions: []*agentworkbenchv2.ArtifactRevision{
			artifactRevision(revision, toolExecutionID),
		},
	}
}

func artifactRevision(
	revision uint64,
	toolExecutionID string,
) *agentworkbenchv2.ArtifactRevision {
	return &agentworkbenchv2.ArtifactRevision{
		Revision: revision,
		Provenance: &agentworkbenchv2.ArtifactProvenance{
			ProducerNamespace: stringPointer("openai"),
			ProducerType:      stringPointer("image.edit"),
			ProducerId:        stringPointer("editor"),
			CommandId:         stringPointer("command-one"),
			ToolExecutionId:   stringPointer(toolExecutionID),
		},
	}
}
