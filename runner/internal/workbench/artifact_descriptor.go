package workbench

import (
	"time"

	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
)

type ArtifactDescriptor = agentworkbenchv2.ArtifactDescriptor

func readyArtifactDescriptor(
	file artifactFile,
	revision uint64,
) *agentworkbenchv2.ArtifactDescriptor {
	status := agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY
	return artifactDescriptor(file, revision, status)
}

func deletedArtifactDescriptor(
	file artifactFile,
	revision uint64,
) *agentworkbenchv2.ArtifactDescriptor {
	return artifactDescriptor(
		file,
		revision,
		agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_DELETED,
	)
}

func artifactDescriptor(
	file artifactFile,
	revision uint64,
	status agentworkbenchv2.ArtifactStatus,
) *agentworkbenchv2.ArtifactDescriptor {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	artifactID := "workspace:" + file.path
	role := "preview"
	representation := &agentworkbenchv2.ArtifactRepresentation{
		RepresentationId: "original",
		Revision:         revision,
		MediaType:        file.mediaType,
		Role:             &role,
		Filename:         &file.filename,
		Status:           status,
		ByteSize:         &file.byteSize,
		Digest:           &file.digest,
	}
	if status != agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_DELETED {
		representation.Transport = &agentworkbenchv2.ArtifactTransport{
			Transport: &agentworkbenchv2.ArtifactTransport_ResourceId{
				ResourceId: "workspace:" + file.path,
			},
		}
	}
	descriptor := &agentworkbenchv2.ArtifactDescriptor{
		ArtifactId: artifactID,
		Revision:   revision,
		Filename:   file.filename,
		MediaType:  file.mediaType,
		Role:       &role,
		Status:     status,
		ByteSize:   &file.byteSize,
		Provenance: &agentworkbenchv2.ArtifactProvenance{
			ProducerNamespace: stringPointer("agentsmesh.runner"),
			ProducerType:      stringPointer("workspace.scan"),
			CreatedAt:         &now,
		},
		Representations: []*agentworkbenchv2.ArtifactRepresentation{
			representation,
		},
		Revisions: []*agentworkbenchv2.ArtifactRevision{{
			Revision:          revision,
			RepresentationIds: []string{"original"},
			Digest:            &file.digest,
			CreatedAt:         &now,
		}},
	}
	if status == agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY {
		descriptor.Grants = []*agentworkbenchv2.ArtifactGrant{
			artifactDownloadGrant(
				artifactID,
				revision,
				[]string{"original"},
				now,
			),
		}
	}
	return descriptor
}
