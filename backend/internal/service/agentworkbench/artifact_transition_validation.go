package agentworkbench

import agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"

func validateArtifactTransition(
	artifacts []*agentworkbenchv2.ArtifactDescriptor,
	next *agentworkbenchv2.ArtifactDescriptor,
) error {
	if next == nil || next.GetArtifactId() == "" || next.GetRevision() == 0 {
		return ErrInvalidBatch
	}
	for _, current := range artifacts {
		if current.GetArtifactId() != next.GetArtifactId() {
			continue
		}
		if !sameArtifactProducer(current.GetProvenance(), next.GetProvenance()) {
			return ErrInvalidBatch
		}
		if next.GetRevision() == current.GetRevision() {
			if !sameRevisionArtifactEnrichment(current, next) {
				return ErrInvalidBatch
			}
			return nil
		}
		if current.GetRevision() == ^uint64(0) ||
			next.GetRevision() != current.GetRevision()+1 {
			return ErrInvalidBatch
		}
		return nil
	}
	return nil
}

func sameArtifactProducer(
	current *agentworkbenchv2.ArtifactProvenance,
	next *agentworkbenchv2.ArtifactProvenance,
) bool {
	return current.GetProducerNamespace() == next.GetProducerNamespace() &&
		current.GetProducerType() == next.GetProducerType() &&
		current.GetProducerId() == next.GetProducerId() &&
		current.GetCommandId() == next.GetCommandId()
}
