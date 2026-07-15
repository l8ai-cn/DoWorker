package orchestrationworker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"strconv"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
)

func buildApplyMutation(
	registry *resource.Registry,
	state controlservice.LockedApplyState,
	snapshotID int64,
) (controlservice.ApplyMutation, error) {
	if registry == nil || state.Plan.ArtifactDigest == "" {
		return controlservice.ApplyMutation{}, control.ErrCorrupt
	}
	manifest, spec, err := plannedApplyManifest(registry, state)
	if err != nil {
		return controlservice.ApplyMutation{}, err
	}
	head, revision, generation, identity, err := nextApplyState(state, spec)
	if err != nil {
		return controlservice.ApplyMutation{}, err
	}
	status := json.RawMessage(`{}`)
	createdAt := state.AppliedAt
	createdByID := state.Plan.ActorID
	if state.Head != nil {
		status = bytes.Clone(state.Head.Status)
		createdAt = state.Head.CreatedAt
		createdByID = state.Head.CreatedByID
	}
	manifest.Metadata.UID = identity.UID
	manifest.Metadata.Generation = generation
	manifest.Metadata.ResourceVersion = strconv.FormatInt(revision, 10)
	manifest.Status = bytes.Clone(status)
	canonicalManifest, err := control.CanonicalJSONObject(manifest)
	if err != nil {
		return controlservice.ApplyMutation{}, control.ErrCorrupt
	}
	digest, err := control.DigestCanonicalJSON(canonicalManifest)
	if err != nil {
		return controlservice.ApplyMutation{}, control.ErrCorrupt
	}
	headValue := control.ResourceHead{
		ID: head, OrganizationID: state.Plan.Scope.OrganizationID,
		Identity: identity, DisplayName: manifest.Metadata.DisplayName,
		Labels: maps.Clone(manifest.Metadata.Labels), Status: status,
		Revision: revision, Generation: generation, ResourceVersion: revision,
		CreatedByID: createdByID, UpdatedByID: state.Plan.ActorID,
		CreatedAt: createdAt, UpdatedAt: state.AppliedAt,
	}
	if state.Head != nil {
		headValue.ResourceVersion = state.Head.ResourceVersion + 1
		manifest.Metadata.ResourceVersion = strconv.FormatInt(
			headValue.ResourceVersion,
			10,
		)
		canonicalManifest, err = control.CanonicalJSONObject(manifest)
		if err != nil {
			return controlservice.ApplyMutation{}, control.ErrCorrupt
		}
		digest, err = control.DigestCanonicalJSON(canonicalManifest)
		if err != nil {
			return controlservice.ApplyMutation{}, control.ErrCorrupt
		}
	}
	revisionValue := control.ResourceRevision{
		OrganizationID: state.Plan.Scope.OrganizationID,
		ResourceID:     head, Identity: identity,
		Revision: revision, Generation: generation,
		ResourceVersion:   headValue.ResourceVersion,
		CanonicalManifest: canonicalManifest, CanonicalSpec: spec,
		ResolvedReferences: append(
			[]control.ResolvedReference{},
			state.Plan.ResolvedReferences...,
		),
		Digest: digest, WorkerSpecSnapshotID: snapshotID,
		ActorID: state.Plan.ActorID, CreatedAt: state.AppliedAt,
	}
	if err := headValue.Validate(state.Plan.Scope); err != nil {
		return controlservice.ApplyMutation{}, fmt.Errorf(
			"%w: applied resource head: %v",
			control.ErrCorrupt,
			err,
		)
	}
	if err := revisionValue.Validate(state.Plan.Scope); err != nil {
		return controlservice.ApplyMutation{}, fmt.Errorf(
			"%w: applied resource revision: %v",
			control.ErrCorrupt,
			err,
		)
	}
	return controlservice.ApplyMutation{
		Head: headValue, Revision: revisionValue,
		ArtifactDigest: state.Plan.ArtifactDigest,
	}, nil
}

func plannedApplyManifest(
	registry *resource.Registry,
	state controlservice.LockedApplyState,
) (resource.Manifest, json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(state.Plan.CanonicalManifest))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	var manifest resource.Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return resource.Manifest{}, nil, control.ErrCorrupt
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return resource.Manifest{}, nil, control.ErrCorrupt
	}
	canonical, err := control.CanonicalJSONObject(manifest)
	if err != nil || !bytes.Equal(canonical, state.Plan.CanonicalManifest) {
		return resource.Manifest{}, nil, control.ErrCorrupt
	}
	if err := manifest.ValidateSubmission(); err != nil ||
		manifest.TypeMeta != state.Plan.Target.TypeMeta ||
		manifest.Metadata.Namespace != state.Plan.Target.Namespace ||
		manifest.Metadata.Name != state.Plan.Target.Name {
		return resource.Manifest{}, nil, control.ErrCorrupt
	}
	if _, err := registry.DecodeAndValidate(manifest); err != nil {
		return resource.Manifest{}, nil, control.ErrCorrupt
	}
	spec, err := control.CanonicalJSONObject(manifest.Spec)
	if err != nil {
		return resource.Manifest{}, nil, control.ErrCorrupt
	}
	return manifest, spec, nil
}

func nextApplyState(
	state controlservice.LockedApplyState,
	spec json.RawMessage,
) (int64, int64, int64, control.ResourceIdentity, error) {
	switch state.Plan.Operation {
	case control.PlanOperationCreate:
		if state.Head != nil || state.CurrentRevision != nil ||
			state.ResultResourceID <= 0 ||
			state.ResultIdentity.ResourceTarget != state.Plan.Target {
			return 0, 0, 0, control.ResourceIdentity{}, control.ErrCorrupt
		}
		return state.ResultResourceID, 1, 1, state.ResultIdentity, nil
	case control.PlanOperationUpdate:
		if state.Head == nil || state.CurrentRevision == nil ||
			state.Head.Identity.ResourceTarget != state.Plan.Target ||
			state.CurrentRevision.ResourceID != state.Head.ID {
			return 0, 0, 0, control.ResourceIdentity{}, control.ErrCorrupt
		}
		generation := state.Head.Generation
		if !bytes.Equal(state.CurrentRevision.CanonicalSpec, spec) {
			generation++
		}
		return state.Head.ID, state.Head.Revision + 1, generation,
			state.Head.Identity, nil
	default:
		return 0, 0, 0, control.ResourceIdentity{}, control.ErrCorrupt
	}
}
