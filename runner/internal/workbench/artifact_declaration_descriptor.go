package workbench

import (
	"time"

	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"google.golang.org/protobuf/proto"
)

func readyDeclaredArtifactDescriptor(
	artifact declaredArtifact,
) *agentworkbenchv2.ArtifactDescriptor {
	revision := artifact.revision
	now := time.Now().UTC().Format(time.RFC3339Nano)
	representations := orderedDeclaredRepresentations(artifact)
	protoRepresentations := make(
		[]*agentworkbenchv2.ArtifactRepresentation,
		0,
		len(representations),
	)
	representationIDs := make([]string, 0, len(representations))
	for _, representation := range representations {
		role := representation.role
		filename := representation.file.filename
		byteSize := representation.file.byteSize
		digest := representation.file.digest
		protoRepresentations = append(
			protoRepresentations,
			&agentworkbenchv2.ArtifactRepresentation{
				RepresentationId: representation.representationID,
				Revision:         revision,
				MediaType:        representation.file.mediaType,
				Role:             optionalNonEmptyString(role),
				Filename:         &filename,
				Status:           agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY,
				ByteSize:         &byteSize,
				Dimensions:       cloneDimensions(representation.dimensions),
				DurationMillis:   representation.durationMillis,
				Digest:           &digest,
				Transport: &agentworkbenchv2.ArtifactTransport{
					Transport: &agentworkbenchv2.ArtifactTransport_ResourceId{
						ResourceId: "workspace:" + representation.file.path,
					},
				},
			},
		)
		representationIDs = append(representationIDs, representation.representationID)
	}
	primary := representations[0]
	role := artifact.role
	byteSize := primary.file.byteSize
	return &agentworkbenchv2.ArtifactDescriptor{
		ArtifactId:      artifact.artifactID,
		Revision:        revision,
		Filename:        primary.file.filename,
		MediaType:       primary.file.mediaType,
		Role:            &role,
		Status:          agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY,
		ByteSize:        &byteSize,
		Dimensions:      cloneDimensions(primary.dimensions),
		DurationMillis:  primary.durationMillis,
		Provenance:      declaredArtifactProvenance(artifact.producer, now),
		Representations: protoRepresentations,
		Revisions: []*agentworkbenchv2.ArtifactRevision{{
			Revision: revision, RepresentationIds: representationIDs,
			Digest: &artifact.fingerprint, CreatedAt: &now,
		}},
		Manifest: cloneManifest(artifact.manifest),
	}
}

func orderedDeclaredRepresentations(
	artifact declaredArtifact,
) []declaredArtifactRepresentation {
	ordered := make(
		[]declaredArtifactRepresentation,
		0,
		len(artifact.representations),
	)
	for _, representation := range artifact.representations {
		if representation.representationID == artifact.primaryRepresentationID {
			ordered = append(ordered, representation)
			break
		}
	}
	for _, representation := range artifact.representations {
		if representation.representationID != artifact.primaryRepresentationID {
			ordered = append(ordered, representation)
		}
	}
	return ordered
}

func declaredArtifactProvenance(
	producer artifactDeclarationProducer,
	createdAt string,
) *agentworkbenchv2.ArtifactProvenance {
	return &agentworkbenchv2.ArtifactProvenance{
		ProducerNamespace: optionalNonEmptyString(producer.Namespace),
		ProducerType:      optionalNonEmptyString(producer.Type),
		ProducerId:        optionalNonEmptyString(producer.ID),
		CommandId:         optionalNonEmptyString(producer.CommandID),
		ToolExecutionId:   optionalNonEmptyString(producer.ToolExecutionID),
		CreatedAt:         &createdAt,
	}
}

func optionalNonEmptyString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func cloneDimensions(
	value *agentworkbenchv2.ArtifactDimensions,
) *agentworkbenchv2.ArtifactDimensions {
	if value == nil {
		return nil
	}
	return proto.Clone(value).(*agentworkbenchv2.ArtifactDimensions)
}

func cloneManifest(
	value *agentworkbenchv2.ArtifactManifest,
) *agentworkbenchv2.ArtifactManifest {
	if value == nil {
		return nil
	}
	return proto.Clone(value).(*agentworkbenchv2.ArtifactManifest)
}
