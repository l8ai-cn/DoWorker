package infra

import (
	"bytes"
	"encoding/json"
	"sort"

	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	orchestrationservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
)

type applyReferenceEntry struct {
	key       string
	reference orchestrationcontrol.ResolvedReference
}

func validatePlanMutationBinding(
	state orchestrationservice.LockedApplyState,
	mutation orchestrationservice.ApplyMutation,
) error {
	if mutation.ArtifactDigest != state.Plan.ArtifactDigest {
		return orchestrationcontrol.ErrInvalid
	}
	planned, spec, err := plannedApplyDocuments(state.Plan.CanonicalManifest)
	if err != nil {
		return orchestrationcontrol.ErrCorrupt
	}
	applied, err := appliedAuthoringManifest(mutation.Revision.CanonicalManifest)
	if err != nil {
		return orchestrationcontrol.ErrInvalid
	}
	if !bytes.Equal(planned, applied) ||
		!bytes.Equal(mutation.Revision.CanonicalSpec, spec) {
		return orchestrationcontrol.ErrInvalid
	}
	plannedRefs, err := canonicalApplyReferences(
		state.Plan.Scope,
		state.Plan.ResolvedReferences,
	)
	if err != nil {
		return orchestrationcontrol.ErrCorrupt
	}
	appliedRefs, err := canonicalApplyReferences(
		state.Plan.Scope,
		mutation.Revision.ResolvedReferences,
	)
	if err != nil || !bytes.Equal(plannedRefs, appliedRefs) {
		return orchestrationcontrol.ErrInvalid
	}
	return nil
}

func plannedApplyDocuments(
	raw json.RawMessage,
) ([]byte, []byte, error) {
	var manifest orchestrationresource.Manifest
	if err := decodeStrictJSON(raw, &manifest); err != nil {
		return nil, nil, err
	}
	canonical, err := orchestrationcontrol.CanonicalJSONObject(manifest)
	if err != nil {
		return nil, nil, err
	}
	spec, err := orchestrationcontrol.CanonicalJSONObject(manifest.Spec)
	if err != nil {
		return nil, nil, err
	}
	return canonical, spec, nil
}

func appliedAuthoringManifest(raw json.RawMessage) ([]byte, error) {
	var manifest orchestrationresource.Manifest
	if err := decodeStrictJSON(raw, &manifest); err != nil {
		return nil, err
	}
	manifest.Metadata.UID = ""
	manifest.Metadata.ResourceVersion = ""
	manifest.Metadata.Generation = 0
	manifest.Status = nil
	return orchestrationcontrol.CanonicalJSONObject(manifest)
}

func canonicalApplyReferences(
	scope orchestrationcontrol.Scope,
	references []orchestrationcontrol.ResolvedReference,
) ([]byte, error) {
	entries := make([]applyReferenceEntry, 0, len(references))
	for _, reference := range references {
		if err := reference.Validate(scope); err != nil {
			return nil, err
		}
		canonical, err := orchestrationcontrol.CanonicalJSONObject(reference)
		if err != nil {
			return nil, err
		}
		entries = append(entries, applyReferenceEntry{
			key: string(canonical), reference: reference,
		})
	}
	sort.Slice(entries, func(left, right int) bool {
		return entries[left].key < entries[right].key
	})
	sorted := make([]orchestrationcontrol.ResolvedReference, 0, len(entries))
	for index, entry := range entries {
		if index > 0 && entries[index-1].key == entry.key {
			return nil, orchestrationcontrol.ErrInvalid
		}
		sorted = append(sorted, entry.reference)
	}
	return orchestrationcontrol.CanonicalJSONArray(sorted)
}
