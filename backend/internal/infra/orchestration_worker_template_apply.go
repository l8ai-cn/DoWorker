package infra

import (
	"context"
	"fmt"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdependencyartifact"
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
			snapshot, err := persistPlannedWorkerBuild(tx, state)
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

func persistPlannedWorkerBuild(
	tx *gorm.DB,
	state controlservice.LockedApplyState,
) (workerspec.Snapshot, error) {
	if state.Plan.Target.Kind != resource.KindWorkerTemplate ||
		state.Plan.ArtifactKind != workerdependencyartifact.PlanArtifactKind {
		return workerspec.Snapshot{}, control.ErrInvalid
	}
	artifact, err := workerdependencyartifact.DecodeApplyPlan(state.Plan)
	if err != nil {
		return workerspec.Snapshot{}, fmt.Errorf(
			"%w: invalid WorkerTemplate build artifact: %v",
			control.ErrCorrupt,
			err,
		)
	}
	spec, err := workerspec.DecodeSpec(artifact.WorkerSpecJSON())
	if err != nil {
		return workerspec.Snapshot{}, fmt.Errorf(
			"%w: invalid planned workerspec: %v",
			control.ErrCorrupt,
			err,
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
	snapshot, err := createWorkerSpecSnapshot(tx, resolved)
	if err != nil {
		return workerspec.Snapshot{}, err
	}
	err = createWorkerSpecDependencyArtifact(
		tx,
		state.Plan.Scope.OrganizationID,
		snapshot.ID,
		artifact.DependenciesJSON(),
		artifact.DependenciesDigest(),
	)
	if err != nil {
		return workerspec.Snapshot{}, err
	}
	return snapshot, nil
}
