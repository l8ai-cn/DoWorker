package agentworkbench

import (
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"google.golang.org/protobuf/proto"
)

func sameRevisionArtifactEnrichment(
	current *agentworkbenchv2.ArtifactDescriptor,
	next *agentworkbenchv2.ArtifactDescriptor,
) bool {
	if !revisionRepresentationIDsEnrich(current, next) {
		return false
	}
	if !artifactGrantRepresentationIDsEnrich(current, next) {
		return false
	}
	currentEnvelope := proto.Clone(current).(*agentworkbenchv2.ArtifactDescriptor)
	nextEnvelope := proto.Clone(next).(*agentworkbenchv2.ArtifactDescriptor)
	currentEnvelope.Representations = nil
	nextEnvelope.Representations = nil
	clearCurrentRevisionRepresentationIDs(currentEnvelope)
	clearCurrentRevisionRepresentationIDs(nextEnvelope)
	clearArtifactGrantRepresentationIDs(currentEnvelope)
	clearArtifactGrantRepresentationIDs(nextEnvelope)
	if !proto.Equal(currentEnvelope, nextEnvelope) {
		return false
	}
	nextByID, valid := representationsByID(next.GetRepresentations(), next.GetRevision())
	if !valid {
		return false
	}
	for _, existing := range current.GetRepresentations() {
		enriched, exists := nextByID[existing.GetRepresentationId()]
		if !exists || !representationEnriches(existing, enriched) {
			return false
		}
		delete(nextByID, existing.GetRepresentationId())
	}
	for _, added := range nextByID {
		if added.GetStatus() != agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_QUEUED &&
			added.GetStatus() != agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_PROCESSING {
			return false
		}
	}
	return true
}

func revisionRepresentationIDsEnrich(
	current *agentworkbenchv2.ArtifactDescriptor,
	next *agentworkbenchv2.ArtifactDescriptor,
) bool {
	currentIDs, currentValid := currentRevisionRepresentationIDs(current)
	nextIDs, nextValid := currentRevisionRepresentationIDs(next)
	if !currentValid || !nextValid {
		return false
	}
	for id := range currentIDs {
		if _, exists := nextIDs[id]; !exists {
			return false
		}
	}
	return true
}

func currentRevisionRepresentationIDs(
	artifact *agentworkbenchv2.ArtifactDescriptor,
) (map[string]struct{}, bool) {
	var revision *agentworkbenchv2.ArtifactRevision
	for _, candidate := range artifact.GetRevisions() {
		if candidate.GetRevision() == artifact.GetRevision() {
			if revision != nil {
				return nil, false
			}
			revision = candidate
		}
	}
	if revision == nil {
		return nil, false
	}
	ids := make(map[string]struct{}, len(revision.GetRepresentationIds()))
	for _, id := range revision.GetRepresentationIds() {
		if id == "" {
			return nil, false
		}
		if _, exists := ids[id]; exists {
			return nil, false
		}
		ids[id] = struct{}{}
	}
	if len(ids) != len(artifact.GetRepresentations()) {
		return nil, false
	}
	for _, representation := range artifact.GetRepresentations() {
		if _, exists := ids[representation.GetRepresentationId()]; !exists {
			return nil, false
		}
	}
	return ids, true
}

func clearCurrentRevisionRepresentationIDs(
	artifact *agentworkbenchv2.ArtifactDescriptor,
) {
	for _, revision := range artifact.GetRevisions() {
		if revision.GetRevision() == artifact.GetRevision() {
			revision.RepresentationIds = nil
		}
	}
}

func representationsByID(
	representations []*agentworkbenchv2.ArtifactRepresentation,
	revision uint64,
) (map[string]*agentworkbenchv2.ArtifactRepresentation, bool) {
	byID := make(map[string]*agentworkbenchv2.ArtifactRepresentation, len(representations))
	for _, representation := range representations {
		id := representation.GetRepresentationId()
		if id == "" || representation.GetRevision() != revision ||
			representation.GetMediaType() == "" ||
			representation.GetStatus() == agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_UNSPECIFIED {
			return nil, false
		}
		if _, exists := byID[id]; exists {
			return nil, false
		}
		byID[id] = representation
	}
	return byID, true
}

func representationEnriches(
	current *agentworkbenchv2.ArtifactRepresentation,
	next *agentworkbenchv2.ArtifactRepresentation,
) bool {
	if current.GetRevision() != next.GetRevision() ||
		current.GetMediaType() != next.GetMediaType() ||
		!sameOptionalString(current.Role, next.Role) ||
		!sameOptionalString(current.Filename, next.Filename) ||
		!validRepresentationStatusTransition(current.GetStatus(), next.GetStatus()) {
		return false
	}
	return optionalUint64Enriches(current.ByteSize, next.ByteSize) &&
		optionalUint64Enriches(current.DurationMillis, next.DurationMillis) &&
		optionalStringEnriches(current.Digest, next.Digest) &&
		messageEnriches(current.GetDimensions(), next.GetDimensions()) &&
		messageEnriches(current.GetTransport(), next.GetTransport()) &&
		validRepresentationState(next)
}

func validRepresentationState(
	representation *agentworkbenchv2.ArtifactRepresentation,
) bool {
	switch representation.GetStatus() {
	case agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_QUEUED,
		agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_PROCESSING,
		agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_FAILED:
		return representation.GetTransport().GetTransport() == nil
	case agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY:
		return representation.ByteSize != nil &&
			representation.Digest != nil &&
			representation.GetDigest() != "" &&
			representation.GetTransport().GetTransport() != nil
	default:
		return false
	}
}

func validRepresentationStatusTransition(
	current agentworkbenchv2.ArtifactStatus,
	next agentworkbenchv2.ArtifactStatus,
) bool {
	if current == next {
		return true
	}
	switch current {
	case agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_QUEUED:
		return next == agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_PROCESSING ||
			next == agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY ||
			next == agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_FAILED
	case agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_PROCESSING:
		return next == agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY ||
			next == agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_FAILED
	default:
		return false
	}
}
