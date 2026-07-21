package workbench

import (
	"path/filepath"
	"strings"

	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
	"google.golang.org/protobuf/proto"
)

const OfficePreviewRepresentationID = "preview-pdf"

type OfficePreviewSource struct {
	Path     string
	Digest   string
	Filename string
}

func ResolveOfficePreviewSource(
	artifact *ArtifactDescriptor,
) (OfficePreviewSource, bool) {
	if artifact.GetStatus() != agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY ||
		!officePreviewMediaType(artifact.GetMediaType()) {
		return OfficePreviewSource{}, false
	}
	for _, representation := range artifact.GetRepresentations() {
		resourceID := representation.GetTransport().GetResourceId()
		if representation.GetRepresentationId() != "original" ||
			representation.GetStatus() != agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY ||
			!strings.HasPrefix(resourceID, "workspace:") {
			continue
		}
		path := strings.TrimPrefix(resourceID, "workspace:")
		if path == "" || representation.GetDigest() == "" {
			return OfficePreviewSource{}, false
		}
		filename := strings.TrimSuffix(
			representation.GetFilename(),
			filepath.Ext(representation.GetFilename()),
		) + ".pdf"
		return OfficePreviewSource{
			Path: path, Digest: representation.GetDigest(), Filename: filename,
		}, true
	}
	return OfficePreviewSource{}, false
}

func OfficePreviewProcessing(
	artifact *ArtifactDescriptor,
	source OfficePreviewSource,
) *ArtifactDescriptor {
	enriched := proto.Clone(artifact).(*agentworkbenchv2.ArtifactDescriptor)
	role := "primary"
	enriched.Representations = append(
		enriched.Representations,
		&agentworkbenchv2.ArtifactRepresentation{
			RepresentationId: OfficePreviewRepresentationID,
			Revision:         artifact.GetRevision(),
			MediaType:        "application/pdf",
			Role:             &role,
			Filename:         &source.Filename,
			Status:           agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_PROCESSING,
		},
	)
	appendCurrentRevisionRepresentationID(
		enriched,
		OfficePreviewRepresentationID,
	)
	appendDownloadGrantRepresentationID(
		enriched,
		OfficePreviewRepresentationID,
	)
	return enriched
}

func OfficePreviewReady(
	processing *ArtifactDescriptor,
	resourceID, digest string,
	byteSize uint64,
) *ArtifactDescriptor {
	enriched := proto.Clone(processing).(*agentworkbenchv2.ArtifactDescriptor)
	representation := officePreviewRepresentation(enriched)
	representation.Status = agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY
	representation.ByteSize = &byteSize
	representation.Digest = &digest
	representation.Transport = &agentworkbenchv2.ArtifactTransport{
		Transport: &agentworkbenchv2.ArtifactTransport_ResourceId{
			ResourceId: resourceID,
		},
	}
	return enriched
}

func OfficePreviewFailed(processing *ArtifactDescriptor) *ArtifactDescriptor {
	enriched := proto.Clone(processing).(*agentworkbenchv2.ArtifactDescriptor)
	officePreviewRepresentation(enriched).Status =
		agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_FAILED
	return enriched
}

func officePreviewRepresentation(
	artifact *ArtifactDescriptor,
) *agentworkbenchv2.ArtifactRepresentation {
	for _, representation := range artifact.GetRepresentations() {
		if representation.GetRepresentationId() == OfficePreviewRepresentationID {
			return representation
		}
	}
	panic("office preview representation is missing")
}

func appendCurrentRevisionRepresentationID(
	artifact *ArtifactDescriptor,
	id string,
) {
	for _, revision := range artifact.GetRevisions() {
		if revision.GetRevision() == artifact.GetRevision() {
			revision.RepresentationIds = append(revision.RepresentationIds, id)
			return
		}
	}
}

func appendDownloadGrantRepresentationID(
	artifact *ArtifactDescriptor,
	representationID string,
) {
	for _, grant := range artifact.GetGrants() {
		for _, action := range grant.GetActions() {
			if action == "artifact.download" {
				grant.RepresentationIds = append(
					grant.RepresentationIds,
					representationID,
				)
				return
			}
		}
	}
}

func officePreviewMediaType(mediaType string) bool {
	switch mediaType {
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return true
	default:
		return false
	}
}
