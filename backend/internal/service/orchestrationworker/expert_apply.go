package orchestrationworker

import (
	"context"
	"fmt"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
)

type ExpertApplyService struct {
	registry   *resource.Registry
	repository ExpertApplyRepository
	resolver   DefinitionResolver
}

func NewExpertApplyService(
	registry *resource.Registry,
	repository ExpertApplyRepository,
	resolver DefinitionResolver,
) (*ExpertApplyService, error) {
	if registry == nil || repository == nil || resolver == nil ||
		!registry.Has(resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       resource.KindExpert,
		}) {
		return nil, fmt.Errorf(
			"%w: expert apply dependencies are incomplete",
			controlservice.ErrUnavailable,
		)
	}
	return &ExpertApplyService{
		registry: registry, repository: repository, resolver: resolver,
	}, nil
}

func (service *ExpertApplyService) Apply(
	ctx context.Context,
	scope control.Scope,
	planID string,
) (AppliedExpert, error) {
	if service == nil || service.registry == nil ||
		service.repository == nil || service.resolver == nil {
		return AppliedExpert{}, controlservice.ErrUnavailable
	}
	return service.repository.RunExpertApplyTransaction(
		ctx,
		scope,
		planID,
		func(state controlservice.LockedApplyState) (
			ExpertApplyMutation,
			error,
		) {
			return buildExpertApplyMutation(
				ctx,
				service.registry,
				service.resolver,
				state,
			)
		},
	)
}
