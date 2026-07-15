package orchestrationworker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
)

const (
	workerLaunchLeaseDuration = 2 * time.Minute
	workerDispatchTTL         = 24 * time.Hour
	maxWorkerCommandBytes     = 4 << 20
)

type WorkerApplyService struct {
	registry   *resource.Registry
	repository WorkerApplyRepository
	resolver   DefinitionResolver
	launcher   WorkerPodLauncher
	notifier   WorkerDispatchNotifier
}

func NewWorkerApplyService(
	registry *resource.Registry,
	repository WorkerApplyRepository,
	resolver DefinitionResolver,
	launcher WorkerPodLauncher,
	notifier WorkerDispatchNotifier,
) (*WorkerApplyService, error) {
	if registry == nil || repository == nil || resolver == nil ||
		launcher == nil || notifier == nil ||
		!registry.Has(resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       resource.KindWorker,
		}) {
		return nil, fmt.Errorf(
			"%w: worker apply dependencies are incomplete",
			controlservice.ErrUnavailable,
		)
	}
	return &WorkerApplyService{
		registry: registry, repository: repository,
		resolver: resolver, launcher: launcher,
		notifier: notifier,
	}, nil
}

func (service *WorkerApplyService) Apply(
	ctx context.Context,
	scope control.Scope,
	planID string,
) (AppliedWorker, error) {
	if service == nil || service.registry == nil ||
		service.repository == nil || service.resolver == nil ||
		service.launcher == nil || service.notifier == nil {
		return AppliedWorker{}, controlservice.ErrUnavailable
	}
	applied, err := service.repository.RunWorkerApplyTransaction(
		ctx,
		scope,
		planID,
		func(state controlservice.LockedApplyState) (
			WorkerApplyMutation,
			error,
		) {
			return buildWorkerApplyMutation(
				ctx,
				service.registry,
				service.resolver,
				state,
			)
		},
	)
	if err != nil {
		return AppliedWorker{}, err
	}
	if applied.PodKey != "" {
		if err := validateDispatchedWorker(applied); err != nil {
			return AppliedWorker{}, err
		}
		service.notifier.TriggerWorkerDispatch(applied.RunnerID)
		return applied, nil
	}
	return service.materializeWorkerLaunch(ctx, scope, applied)
}

func (service *WorkerApplyService) materializeWorkerLaunch(
	ctx context.Context,
	scope control.Scope,
	applied AppliedWorker,
) (AppliedWorker, error) {
	claim, err := service.repository.ClaimWorkerLaunch(
		ctx,
		scope,
		applied.LaunchID,
		workerLaunchLeaseDuration,
		uuid.NewString(),
	)
	if err != nil {
		return AppliedWorker{}, err
	}
	if err := validateWorkerLaunchClaim(scope, applied, claim); err != nil {
		return AppliedWorker{}, err
	}
	launch, err := service.launcher.MaterializeWorkerPod(ctx, claim)
	if err != nil {
		releaseErr := service.repository.ReleaseWorkerLaunch(
			ctx,
			scope,
			claim,
			"materialization failed",
		)
		return AppliedWorker{}, errors.Join(err, releaseErr)
	}
	if err := validateWorkerPodLaunch(launch); err != nil {
		releaseErr := service.repository.ReleaseWorkerLaunch(
			ctx,
			scope,
			claim,
			"materialization result invalid",
		)
		return AppliedWorker{}, errors.Join(err, releaseErr)
	}
	result, err := service.repository.CompleteWorkerLaunch(
		ctx,
		scope,
		claim,
		launch,
		workerDispatchTTL,
	)
	if err != nil {
		return AppliedWorker{}, err
	}
	if err := validateDispatchedWorker(result); err != nil {
		return AppliedWorker{}, err
	}
	service.notifier.TriggerWorkerDispatch(result.RunnerID)
	return result, nil
}

func validateWorkerLaunchClaim(
	scope control.Scope,
	applied AppliedWorker,
	claim WorkerLaunchClaim,
) error {
	switch {
	case claim.LaunchID != applied.LaunchID:
		return fmt.Errorf("%w: worker launch claim id", control.ErrCorrupt)
	case claim.OrganizationID != scope.OrganizationID ||
		claim.ActorID != scope.ActorID:
		return fmt.Errorf("%w: worker launch claim scope", control.ErrCorrupt)
	case claim.ResourceID != applied.Head.ID ||
		claim.ResourceRevision != applied.ResourceRevision:
		return fmt.Errorf("%w: worker launch claim resource", control.ErrCorrupt)
	case claim.WorkerSpecSnapshotID != applied.WorkerSpecSnapshotID:
		return fmt.Errorf("%w: worker launch claim snapshot", control.ErrCorrupt)
	case claim.ClaimToken == "" || claim.LeaseExpiresAt.IsZero():
		return fmt.Errorf("%w: worker launch claim lease", control.ErrCorrupt)
	}
	return nil
}

func validateWorkerPodLaunch(launch WorkerPodLaunch) error {
	if launch.PodID <= 0 || launch.PodKey == "" ||
		launch.RunnerID <= 0 || len(launch.CommandPayload) == 0 ||
		len(launch.CommandPayload) > maxWorkerCommandBytes {
		return control.ErrCorrupt
	}
	return nil
}

func validateDispatchedWorker(applied AppliedWorker) error {
	if applied.Head.ID <= 0 || applied.LaunchID <= 0 ||
		applied.WorkerSpecSnapshotID <= 0 ||
		applied.ResourceRevision <= 0 ||
		applied.PodID <= 0 || applied.PodKey == "" ||
		applied.RunnerID <= 0 {
		return control.ErrCorrupt
	}
	return nil
}
