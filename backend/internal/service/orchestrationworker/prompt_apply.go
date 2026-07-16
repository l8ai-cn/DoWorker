package orchestrationworker

import (
	"context"
	"fmt"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
)

type PromptApplyService struct {
	registry   *resource.Registry
	repository PromptApplyRepository
}

func NewPromptApplyService(
	registry *resource.Registry,
	repository PromptApplyRepository,
) (*PromptApplyService, error) {
	if registry == nil || repository == nil ||
		!registry.Has(resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       resource.KindPrompt,
		}) {
		return nil, fmt.Errorf(
			"%w: prompt apply dependencies are incomplete",
			controlservice.ErrUnavailable,
		)
	}
	return &PromptApplyService{
		registry: registry, repository: repository,
	}, nil
}

func (service *PromptApplyService) Apply(
	ctx context.Context,
	scope control.Scope,
	planID string,
) (control.ResourceHead, error) {
	if service == nil || service.registry == nil || service.repository == nil {
		return control.ResourceHead{}, controlservice.ErrUnavailable
	}
	return service.repository.RunPromptApplyTransaction(
		ctx,
		scope,
		planID,
		func(state controlservice.LockedApplyState) (
			controlservice.ApplyMutation,
			error,
		) {
			return buildPromptApplyMutation(service.registry, state)
		},
	)
}
