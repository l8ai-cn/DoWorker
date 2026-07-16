package workbench

import (
	"fmt"
	"sort"
)

func (o *ArtifactObserver) changedDeclaredArtifacts(
	current map[string]declaredArtifact,
) ([]*ArtifactDescriptor, error) {
	ids := sortedDeclaredArtifactIDs(current)
	changed := make([]declaredArtifact, 0, len(ids))
	for _, id := range ids {
		artifact := current[id]
		previous, wasEmitted := o.declaredEmitted[id]
		baseline, wasBaseline := o.declaredBaseline[id]
		isChanged, err := validateDeclaredArtifactRevision(
			artifact,
			previous,
			wasEmitted,
			baseline,
			wasBaseline,
		)
		if err != nil {
			return nil, fmt.Errorf("artifact %q: %w", id, err)
		}
		if isChanged {
			changed = append(changed, artifact)
		}
	}
	descriptors := make([]*ArtifactDescriptor, 0, len(changed))
	for _, artifact := range changed {
		descriptors = append(descriptors, readyDeclaredArtifactDescriptor(artifact))
		o.declaredEmitted[artifact.artifactID] = emittedDeclaredArtifact{
			artifact: artifact,
			revision: artifact.revision,
		}
	}
	return descriptors, nil
}

func validateDeclaredArtifactRevision(
	current declaredArtifact,
	emitted emittedDeclaredArtifact,
	wasEmitted bool,
	baseline declaredArtifact,
	wasBaseline bool,
) (bool, error) {
	previousRevision := uint64(0)
	previousFingerprint := ""
	var previous *declaredArtifact
	if wasEmitted {
		previousRevision = emitted.revision
		previousFingerprint = emitted.artifact.fingerprint
		previous = &emitted.artifact
	} else if wasBaseline {
		previousRevision = baseline.revision
		previousFingerprint = baseline.fingerprint
		previous = &baseline
	}
	if previousRevision == 0 {
		if current.revision != 1 {
			return false, fmt.Errorf("first revision must be 1")
		}
		return true, nil
	}
	if current.fingerprint == previousFingerprint {
		if current.revision != previousRevision {
			return false, fmt.Errorf(
				"unchanged artifact revision must remain %d",
				previousRevision,
			)
		}
		return false, nil
	}
	if previous != nil && previous.producer != current.producer {
		return false, fmt.Errorf("producer must remain stable across revisions")
	}
	if current.revision != previousRevision+1 {
		return false, fmt.Errorf(
			"changed artifact revision must be %d",
			previousRevision+1,
		)
	}
	return true, nil
}

func sortedDeclaredArtifactIDs(files map[string]declaredArtifact) []string {
	ids := make([]string, 0, len(files))
	for id := range files {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
