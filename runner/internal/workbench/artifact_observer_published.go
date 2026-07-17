package workbench

import (
	"fmt"

	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"google.golang.org/protobuf/proto"
)

func (o *ArtifactObserver) PublishedArtifact(
	artifactID string,
	executionID string,
) (*ArtifactDescriptor, error) {
	declared, reservedPaths, err := scanArtifactDeclarations(o.root)
	if err != nil {
		return nil, fmt.Errorf("scan published artifact: %w", err)
	}
	artifact, exists := declared[artifactID]
	if !exists {
		return nil, fmt.Errorf("published artifact %q is missing", artifactID)
	}
	for path := range reservedPaths {
		o.reservedPaths[path] = struct{}{}
	}
	previous, wasEmitted := o.declaredEmitted[artifactID]
	baseline, wasBaseline := o.declaredBaseline[artifactID]
	changed, err := validateDeclaredArtifactRevision(
		artifact,
		previous,
		wasEmitted,
		baseline,
		wasBaseline,
	)
	if err != nil {
		return nil, err
	}
	if !changed {
		return nil, fmt.Errorf(
			"artifact %q revision %d was already published",
			artifactID,
			artifact.revision,
		)
	}
	descriptor := readyDeclaredArtifactDescriptor(artifact)
	setRevisionExecutionProvenance(descriptor, executionID)
	o.declaredEmitted[artifactID] = emittedDeclaredArtifact{
		artifact: artifact,
		revision: artifact.revision,
	}
	return descriptor, nil
}

func setRevisionExecutionProvenance(
	artifact *ArtifactDescriptor,
	executionID string,
) {
	for _, revision := range artifact.GetRevisions() {
		if revision.GetRevision() != artifact.GetRevision() {
			continue
		}
		provenance := proto.Clone(
			artifact.GetProvenance(),
		).(*agentworkbenchv2.ArtifactProvenance)
		provenance.ToolExecutionId = &executionID
		revision.Provenance = provenance
		return
	}
}
