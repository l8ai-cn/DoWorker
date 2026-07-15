package infra

import (
	"bytes"
	"context"
	"fmt"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	workerspecservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"gorm.io/gorm"
)

func (repo *orchestrationResourceRepo) RunWorkerTemplateApplyTransaction(
	ctx context.Context,
	scope control.Scope,
	planID string,
	build controlservice.WorkerTemplateApplyBuilder,
) (controlservice.AppliedWorkerTemplate, error) {
	if build == nil {
		return controlservice.AppliedWorkerTemplate{}, control.ErrInvalid
	}
	if err := validateApplyCoordinates(scope, planID); err != nil {
		return controlservice.AppliedWorkerTemplate{}, err
	}
	var snapshotID int64
	head, err := repo.runApplyTransaction(
		ctx,
		scope,
		planID,
		func(
			tx *gorm.DB,
			state controlservice.LockedApplyState,
		) (controlservice.ApplyMutation, error) {
			snapshot, err := persistPlannedWorkerSpec(tx, state)
			if err != nil {
				return controlservice.ApplyMutation{}, err
			}
			snapshotID = snapshot.ID
			return build(state, snapshot.ID)
		},
	)
	if err != nil {
		return controlservice.AppliedWorkerTemplate{}, err
	}
	return controlservice.AppliedWorkerTemplate{
		Head: head, WorkerSpecSnapshotID: snapshotID,
	}, nil
}

func persistPlannedWorkerSpec(
	tx *gorm.DB,
	state controlservice.LockedApplyState,
) (workerspec.Snapshot, error) {
	if state.Plan.Target.Kind != resource.KindWorkerTemplate ||
		state.Plan.ArtifactKind != "WorkerSpec" {
		return workerspec.Snapshot{}, control.ErrInvalid
	}
	spec, err := workerspec.DecodeSpec(state.Plan.ArtifactJSON)
	if err != nil {
		return workerspec.Snapshot{}, fmt.Errorf(
			"%w: invalid planned workerspec: %v",
			control.ErrCorrupt,
			err,
		)
	}
	canonical, err := workerspec.EncodeSpec(spec)
	if err != nil {
		return workerspec.Snapshot{}, fmt.Errorf(
			"%w: encode planned workerspec: %v",
			control.ErrCorrupt,
			err,
		)
	}
	canonical, err = control.CanonicalJSONObject(canonical)
	if err != nil || !bytes.Equal(canonical, state.Plan.ArtifactJSON) {
		return workerspec.Snapshot{}, fmt.Errorf(
			"%w: planned workerspec must be canonical",
			control.ErrCorrupt,
		)
	}
	resolved, err := workerspecservice.NewResolvedSnapshot(
		state.Plan.Scope.OrganizationID,
		spec,
	)
	if err != nil {
		return workerspec.Snapshot{}, fmt.Errorf(
			"%w: resolve planned workerspec: %v",
			control.ErrCorrupt,
			err,
		)
	}
	return createWorkerSpecSnapshot(tx, resolved)
}
