package orchestrationworker

import (
	"context"
	"fmt"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
)

type BindingApplyService struct {
	registry   *resource.Registry
	repository BindingApplyRepository
}

func NewBindingApplyService(
	registry *resource.Registry,
	repository BindingApplyRepository,
) (*BindingApplyService, error) {
	if registry == nil || repository == nil {
		return nil, fmt.Errorf(
			"%w: binding apply dependencies are incomplete",
			controlservice.ErrUnavailable,
		)
	}
	for _, kind := range resourceBindingKinds() {
		if !registry.Has(resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       kind,
		}) {
			return nil, fmt.Errorf(
				"%w: binding schema %s is not registered",
				controlservice.ErrUnavailable,
				kind,
			)
		}
	}
	return &BindingApplyService{
		registry: registry, repository: repository,
	}, nil
}

func (service *BindingApplyService) Apply(
	ctx context.Context,
	scope control.Scope,
	planID string,
) (control.ResourceHead, error) {
	if service == nil || service.registry == nil || service.repository == nil {
		return control.ResourceHead{}, controlservice.ErrUnavailable
	}
	return service.repository.RunBindingApplyTransaction(
		ctx,
		scope,
		planID,
		func(state controlservice.LockedApplyState) (
			controlservice.ApplyMutation,
			error,
		) {
			return buildBindingApplyMutation(service.registry, state)
		},
	)
}
