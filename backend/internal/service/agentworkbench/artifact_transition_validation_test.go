package agentworkbench

import (
	"testing"

	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
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

func TestValidateArtifactTransitionAllowsSameRevisionRepresentationEnrichment(t *testing.T) {
	current := transitionArtifact(3, "tool-one")
	current.Representations = []*agentworkbenchv2.ArtifactRepresentation{
		previewRepresentation(
			"pdf-preview",
			agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_QUEUED,
		),
	}
	current.Revisions[0].RepresentationIds = []string{"pdf-preview"}

	processing := proto.Clone(current).(*agentworkbenchv2.ArtifactDescriptor)
	processing.Representations[0].Status =
		agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_PROCESSING
	require.NoError(t, validateArtifactTransition(
		[]*agentworkbenchv2.ArtifactDescriptor{current},
		processing,
	))

	ready := proto.Clone(processing).(*agentworkbenchv2.ArtifactDescriptor)
	ready.Representations[0].Status =
		agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY
	ready.Representations[0].ByteSize = uint64Pointer(2048)
	ready.Representations[0].Digest = stringPointer("sha256:preview")
	ready.Representations[0].Transport = &agentworkbenchv2.ArtifactTransport{
		Transport: &agentworkbenchv2.ArtifactTransport_ResourceId{
			ResourceId: "artifact-cache:pdf-preview",
		},
	}
	require.NoError(t, validateArtifactTransition(
		[]*agentworkbenchv2.ArtifactDescriptor{processing},
		ready,
	))
}

func TestValidateArtifactTransitionAllowsSameRevisionRepresentationAddition(t *testing.T) {
	current := transitionArtifact(3, "tool-one")
	current.Representations = []*agentworkbenchv2.ArtifactRepresentation{
		previewRepresentation(
			"original",
			agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY,
		),
	}
	current.Revisions[0].RepresentationIds = []string{"original"}

	next := proto.Clone(current).(*agentworkbenchv2.ArtifactDescriptor)
	next.Representations = append(
		next.Representations,
		previewRepresentation(
			"pdf-preview",
			agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_QUEUED,
		),
	)
	next.Revisions[0].RepresentationIds = append(
		next.Revisions[0].RepresentationIds,
		"pdf-preview",
	)
	current.Grants = []*agentworkbenchv2.ArtifactGrant{
		artifactGrant("download", "original"),
	}
	next.Grants = []*agentworkbenchv2.ArtifactGrant{
		artifactGrant("download", "original", "pdf-preview"),
	}
	require.NoError(t, validateArtifactTransition(
		[]*agentworkbenchv2.ArtifactDescriptor{current},
		next,
	))
}

func TestValidateArtifactTransitionRejectsGrantRewrites(t *testing.T) {
	current := transitionArtifact(3, "tool-one")
	current.Representations = []*agentworkbenchv2.ArtifactRepresentation{
		previewRepresentation(
			"original",
			agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY,
		),
	}
	current.Revisions[0].RepresentationIds = []string{"original"}
	current.Grants = []*agentworkbenchv2.ArtifactGrant{
		artifactGrant("download", "original"),
	}

	for name, mutate := range map[string]func(*agentworkbenchv2.ArtifactDescriptor){
		"action changed": func(value *agentworkbenchv2.ArtifactDescriptor) {
			value.Grants[0].Actions = []string{"artifact.delete"}
		},
		"existing grant widened": func(value *agentworkbenchv2.ArtifactDescriptor) {
			value.Grants[0].RepresentationIds = append(
				value.Grants[0].RepresentationIds,
				"ungranted-existing",
			)
		},
		"unknown representation": func(value *agentworkbenchv2.ArtifactDescriptor) {
			value.Grants[0].RepresentationIds = append(
				value.Grants[0].RepresentationIds,
				"missing",
			)
		},
		"representation removed": func(value *agentworkbenchv2.ArtifactDescriptor) {
			value.Grants[0].RepresentationIds = nil
		},
	} {
		t.Run(name, func(t *testing.T) {
			baseline := proto.Clone(current).(*agentworkbenchv2.ArtifactDescriptor)
			if name == "existing grant widened" {
				baseline.Representations = append(
					baseline.Representations,
					previewRepresentation(
						"ungranted-existing",
						agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_PROCESSING,
					),
				)
				baseline.Revisions[0].RepresentationIds = append(
					baseline.Revisions[0].RepresentationIds,
					"ungranted-existing",
				)
			}
			next := proto.Clone(baseline).(*agentworkbenchv2.ArtifactDescriptor)
			mutate(next)
			require.ErrorIs(t,
				validateArtifactTransition(
					[]*agentworkbenchv2.ArtifactDescriptor{baseline},
					next,
				),
				ErrInvalidBatch,
			)
		})
	}
}

func TestValidateArtifactTransitionRejectsReadyRepresentationAddition(t *testing.T) {
	current := transitionArtifact(3, "tool-one")
	next := proto.Clone(current).(*agentworkbenchv2.ArtifactDescriptor)
	next.Representations = []*agentworkbenchv2.ArtifactRepresentation{
		previewRepresentation(
			"pdf-preview",
			agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY,
		),
	}
	next.Revisions[0].RepresentationIds = []string{"pdf-preview"}

	require.ErrorIs(t,
		validateArtifactTransition(
			[]*agentworkbenchv2.ArtifactDescriptor{current},
			next,
		),
		ErrInvalidBatch,
	)
}

func TestValidateArtifactTransitionRejectsSameRevisionArtifactRewrites(t *testing.T) {
	current := transitionArtifact(3, "tool-one")
	current.Representations = []*agentworkbenchv2.ArtifactRepresentation{
		previewRepresentation(
			"pdf-preview",
			agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY,
		),
	}
	current.Revisions[0].RepresentationIds = []string{"pdf-preview"}

	for name, mutate := range map[string]func(*agentworkbenchv2.ArtifactDescriptor){
		"descriptor changed": func(value *agentworkbenchv2.ArtifactDescriptor) {
			value.Filename = "rewritten.pdf"
		},
		"representation removed": func(value *agentworkbenchv2.ArtifactDescriptor) {
			value.Representations = nil
		},
		"representation media changed": func(value *agentworkbenchv2.ArtifactDescriptor) {
			value.Representations[0].MediaType = "text/html"
		},
		"representation revision reference missing": func(value *agentworkbenchv2.ArtifactDescriptor) {
			value.Revisions[0].RepresentationIds = nil
		},
		"ready representation regressed": func(value *agentworkbenchv2.ArtifactDescriptor) {
			value.Representations[0].Status =
				agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_PROCESSING
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

func previewRepresentation(
	id string,
	status agentworkbenchv2.ArtifactStatus,
) *agentworkbenchv2.ArtifactRepresentation {
	representation := &agentworkbenchv2.ArtifactRepresentation{
		RepresentationId: id,
		Revision:         3,
		MediaType:        "application/pdf",
		Role:             stringPointer("preview"),
		Filename:         stringPointer("preview.pdf"),
		Status:           status,
	}
	if status == agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY {
		representation.ByteSize = uint64Pointer(2048)
		representation.Digest = stringPointer("sha256:ready")
		representation.Transport = &agentworkbenchv2.ArtifactTransport{
			Transport: &agentworkbenchv2.ArtifactTransport_ResourceId{
				ResourceId: "session-file:file-" + id,
			},
		}
	}
	return representation
}

func artifactGrant(
	id string,
	representationIDs ...string,
) *agentworkbenchv2.ArtifactGrant {
	minimumRevision := uint64(3)
	maximumRevision := uint64(3)
	return &agentworkbenchv2.ArtifactGrant{
		GrantId:           id,
		Issuer:            stringPointer("agentcloud.runner"),
		Subject:           stringPointer("session.viewer"),
		RepresentationIds: representationIDs,
		Actions:           []string{"artifact.download"},
		MinimumRevision:   &minimumRevision,
		MaximumRevision:   &maximumRevision,
	}
}
