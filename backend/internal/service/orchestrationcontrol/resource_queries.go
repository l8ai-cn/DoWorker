package orchestrationcontrol

import (
	"context"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
)

func (service *Service) GetResource(
	ctx context.Context,
	scope control.Scope,
	target control.ResourceTarget,
) (control.ResourceHead, error) {
	if service == nil || service.repository == nil || service.authorizer == nil {
		return control.ResourceHead{}, ErrUnavailable
	}
	if err := target.Validate(scope); err != nil {
		return control.ResourceHead{}, err
	}
	head, err := service.repository.GetResource(ctx, scope, target)
	if err != nil {
		return control.ResourceHead{}, err
	}
	if err := validateQueriedResource(scope, target, head); err != nil {
		return control.ResourceHead{}, err
	}
	if err := service.authorizer.AuthorizeReference(ctx, scope, head); err != nil {
		return control.ResourceHead{}, err
	}
	return head, nil
}

func (service *Service) ListResources(
	ctx context.Context,
	scope control.Scope,
	filter ResourceListFilter,
) (ResourceListPage, error) {
	if service == nil || service.repository == nil || service.authorizer == nil ||
		service.workerDefinitions == nil {
		return ResourceListPage{}, ErrUnavailable
	}
	if err := filter.Validate(scope); err != nil {
		return ResourceListPage{}, err
	}
	if err := service.authorizer.AuthorizeList(ctx, scope); err != nil {
		return ResourceListPage{}, err
	}
	filter, err := service.applyWorkerDefinitionPolicy(filter)
	if err != nil {
		return ResourceListPage{}, err
	}
	if err := filter.Validate(scope); err != nil {
		return ResourceListPage{}, err
	}
	page, err := service.repository.ListResources(ctx, scope, filter)
	if err != nil {
		return ResourceListPage{}, err
	}
	if page.Total < 0 || page.Total < int64(len(page.Items)) {
		return ResourceListPage{}, control.ErrCorrupt
	}
	result := ResourceListPage{
		Items:         append([]control.ResourceHead{}, page.Items...),
		Total:         page.Total,
		AppliedFilter: filter,
	}
	for _, head := range result.Items {
		if err := head.Validate(scope); err != nil ||
			(filter.Kind != "" && head.Identity.Kind != filter.Kind) {
			return ResourceListPage{}, control.ErrCorrupt
		}
		if err := service.authorizer.AuthorizeReference(ctx, scope, head); err != nil {
			return ResourceListPage{}, err
		}
	}
	return result, nil
}

func (service *Service) GetResourcePlan(
	ctx context.Context,
	scope control.Scope,
	planID string,
) (control.Plan, error) {
	if service == nil || service.repository == nil {
		return control.Plan{}, ErrUnavailable
	}
	if err := scope.Validate(); err != nil {
		return control.Plan{}, err
	}
	plan, err := service.repository.GetPlan(ctx, scope, planID)
	if err != nil {
		return control.Plan{}, err
	}
	if err := plan.Validate(); err != nil ||
		plan.Scope != scope ||
		plan.ActorID != scope.ActorID ||
		plan.ID != planID {
		return control.Plan{}, control.ErrCorrupt
	}
	return plan, nil
}

func validateQueriedResource(
	scope control.Scope,
	target control.ResourceTarget,
	head control.ResourceHead,
) error {
	if err := head.Validate(scope); err != nil ||
		head.Identity.ResourceTarget != target {
		return control.ErrCorrupt
	}
	return nil
}
