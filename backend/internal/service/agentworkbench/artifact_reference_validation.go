package agentworkbench

import agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"

func validateProjectedArtifactReferences(
	snapshot *agentworkbenchv2.SessionSnapshot,
) error {
	catalog := make(
		map[string]*agentworkbenchv2.ArtifactDescriptor,
		len(snapshot.GetArtifacts()),
	)
	for _, artifact := range snapshot.GetArtifacts() {
		if artifact == nil || artifact.GetArtifactId() == "" ||
			artifact.GetRevision() == 0 {
			return ErrInvalidBatch
		}
		if _, exists := catalog[artifact.GetArtifactId()]; exists {
			return ErrInvalidBatch
		}
		catalog[artifact.GetArtifactId()] = artifact
	}
	for _, item := range snapshot.GetHistory() {
		execution := item.GetContent().GetToolExecution()
		if execution == nil {
			continue
		}
		if execution.GetExecutionId() == "" {
			return ErrInvalidBatch
		}
		if err := validateExecutionArtifactReferences(execution, catalog); err != nil {
			return err
		}
	}
	return nil
}

func validateExecutionArtifactReferences(
	execution *agentworkbenchv2.ToolExecution,
	catalog map[string]*agentworkbenchv2.ArtifactDescriptor,
) error {
	for _, reference := range execution.GetArtifacts() {
		if err := validateToolArtifactReference(
			execution.GetExecutionId(),
			reference,
			catalog,
		); err != nil {
			return err
		}
	}
	for _, result := range execution.GetResults() {
		for _, reference := range result.GetArtifacts() {
			if err := validateToolArtifactReference(
				execution.GetExecutionId(),
				reference,
				catalog,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateToolArtifactReference(
	executionID string,
	reference *agentworkbenchv2.ArtifactReference,
	catalog map[string]*agentworkbenchv2.ArtifactDescriptor,
) error {
	if reference == nil || reference.GetArtifactId() == "" ||
		reference.Revision == nil {
		return ErrInvalidBatch
	}
	artifact := catalog[reference.GetArtifactId()]
	revision := findArtifactRevision(artifact, reference.GetRevision())
	if revision == nil ||
		revision.GetProvenance().GetToolExecutionId() != executionID {
		return ErrInvalidBatch
	}
	if reference.RepresentationId == nil {
		return nil
	}
	for _, representationID := range revision.GetRepresentationIds() {
		if representationID == reference.GetRepresentationId() {
			return nil
		}
	}
	return ErrInvalidBatch
}

func findArtifactRevision(
	artifact *agentworkbenchv2.ArtifactDescriptor,
	revision uint64,
) *agentworkbenchv2.ArtifactRevision {
	if artifact == nil {
		return nil
	}
	for _, candidate := range artifact.GetRevisions() {
		if candidate.GetRevision() == revision {
			return candidate
		}
	}
	return nil
}
