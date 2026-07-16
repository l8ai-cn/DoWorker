package orchestrationworker

import (
	"context"
	"fmt"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
)

type GoalLoopApplyService struct {
	registry   *resource.Registry
	repository GoalLoopApplyRepository
}

func NewGoalLoopApplyService(
	registry *resource.Registry,
	repository GoalLoopApplyRepository,
) (*GoalLoopApplyService, error) {
	if registry == nil || repository == nil ||
		!registry.Has(resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       resource.KindGoalLoop,
		}) {
		return nil, fmt.Errorf(
			"%w: goal loop apply dependencies are incomplete",
			controlservice.ErrUnavailable,
		)
	}
	return &GoalLoopApplyService{
		registry: registry, repository: repository,
	}, nil
}

func (service *GoalLoopApplyService) Apply(
	ctx context.Context,
	scope control.Scope,
	planID string,
) (AppliedGoalLoop, error) {
	if service == nil || service.registry == nil ||
		service.repository == nil {
		return AppliedGoalLoop{}, controlservice.ErrUnavailable
	}
	return service.repository.RunGoalLoopApplyTransaction(
		ctx,
		scope,
		planID,
		func(state controlservice.LockedApplyState) (
			GoalLoopApplyMutation,
			error,
		) {
			return buildGoalLoopApplyMutation(service.registry, state)
		},
	)
}
