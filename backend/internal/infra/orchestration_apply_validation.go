package infra

import (
	"bytes"
	"encoding/json"
	"maps"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	orchestrationservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
)

func validateApplyMutation(
	state orchestrationservice.LockedApplyState,
	mutation orchestrationservice.ApplyMutation,
) error {
	appliedAt := state.AppliedAt
	head := mutation.Head
	revision := mutation.Revision
	if err := head.Validate(state.Plan.Scope); err != nil {
		return err
	}
	if err := revision.Validate(state.Plan.Scope); err != nil {
		return err
	}
	if head.ID != state.ResultResourceID || head.Identity != state.ResultIdentity ||
		revision.ResourceID != head.ID || revision.Identity != head.Identity ||
		revision.Revision != head.Revision ||
		revision.Generation != head.Generation ||
		revision.ResourceVersion != head.ResourceVersion ||
		revision.ActorID != state.Plan.ActorID ||
		!revision.CreatedAt.Equal(appliedAt) ||
		head.UpdatedByID != state.Plan.ActorID ||
		!head.UpdatedAt.Equal(appliedAt) {
		return orchestrationcontrol.ErrInvalid
	}
	if err := validateHeadManifest(head, revision); err != nil {
		return err
	}
	if err := validatePlanMutationBinding(state, mutation); err != nil {
		return err
	}
	if state.Head == nil {
		return validateCreateMutation(state, mutation, appliedAt)
	}
	return validateUpdateMutation(state, mutation)
}

func validateCreateMutation(
	state orchestrationservice.LockedApplyState,
	mutation orchestrationservice.ApplyMutation,
	appliedAt time.Time,
) error {
	head := mutation.Head
	if state.CurrentRevision != nil || head.Revision != 1 ||
		head.Generation != 1 || head.ResourceVersion != 1 ||
		head.CreatedByID != state.Plan.ActorID ||
		!head.CreatedAt.Equal(appliedAt) {
		return orchestrationcontrol.ErrInvalid
	}
	return nil
}

func validateUpdateMutation(
	state orchestrationservice.LockedApplyState,
	mutation orchestrationservice.ApplyMutation,
) error {
	if state.CurrentRevision == nil {
		return orchestrationcontrol.ErrCorrupt
	}
	old := *state.Head
	head := mutation.Head
	if head.ID != old.ID || head.OrganizationID != old.OrganizationID ||
		head.Identity != old.Identity || head.CreatedByID != old.CreatedByID ||
		!head.CreatedAt.Equal(old.CreatedAt) ||
		head.Revision != old.Revision+1 ||
		head.ResourceVersion != old.ResourceVersion+1 {
		return orchestrationcontrol.ErrInvalid
	}
	specChanged := !bytes.Equal(
		state.CurrentRevision.CanonicalSpec,
		mutation.Revision.CanonicalSpec,
	)
	expectedGeneration := old.Generation
	if specChanged {
		expectedGeneration++
	}
	if head.Generation != expectedGeneration {
		return orchestrationcontrol.ErrInvalid
	}
	return nil
}

func validateHeadManifest(
	head orchestrationcontrol.ResourceHead,
	revision orchestrationcontrol.ResourceRevision,
) error {
	var manifest orchestrationresource.Manifest
	if err := json.Unmarshal(revision.CanonicalManifest, &manifest); err != nil {
		return orchestrationcontrol.ErrCorrupt
	}
	status, err := orchestrationcontrol.CanonicalJSONObject(manifest.Status)
	if err != nil {
		return orchestrationcontrol.ErrCorrupt
	}
	if manifest.Metadata.DisplayName != head.DisplayName ||
		!maps.Equal(manifest.Metadata.Labels, head.Labels) ||
		!bytes.Equal(status, head.Status) {
		return orchestrationcontrol.ErrInvalid
	}
	return nil
}
