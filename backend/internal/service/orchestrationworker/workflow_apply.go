package orchestrationworker

import (
	"context"
	"fmt"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
)

type WorkflowApplyService struct {
	registry   *resource.Registry
	repository WorkflowApplyRepository
	resolver   DefinitionResolver
}

func NewWorkflowApplyService(
	registry *resource.Registry,
	repository WorkflowApplyRepository,
	resolver DefinitionResolver,
) (*WorkflowApplyService, error) {
	if registry == nil || repository == nil || resolver == nil ||
		!registry.Has(resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       resource.KindWorkflow,
		}) {
		return nil, fmt.Errorf(
			"%w: workflow apply dependencies are incomplete",
			controlservice.ErrUnavailable,
		)
	}
	return &WorkflowApplyService{
		registry: registry, repository: repository, resolver: resolver,
	}, nil
}

func (service *WorkflowApplyService) Apply(
	ctx context.Context,
	scope control.Scope,
	planID string,
) (AppliedWorkflow, error) {
	if service == nil || service.registry == nil ||
		service.repository == nil || service.resolver == nil {
		return AppliedWorkflow{}, controlservice.ErrUnavailable
	}
	return service.repository.RunWorkflowApplyTransaction(
		ctx,
		scope,
		planID,
		func(state controlservice.LockedApplyState) (
			WorkflowApplyMutation,
			error,
		) {
			return buildWorkflowApplyMutation(
				ctx,
				service.registry,
				service.resolver,
				state,
			)
		},
	)
}
