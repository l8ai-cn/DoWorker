package orchestrationcontrol

import (
	"context"
	"errors"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
)

func (service *Service) GetResourceCapabilities(
	ctx context.Context,
	scope control.Scope,
	target control.ResourceTarget,
) (ResourceCapabilities, error) {
	if service == nil || service.repository == nil || service.authorizer == nil {
		return ResourceCapabilities{}, ErrUnavailable
	}
	if err := target.Validate(scope); err != nil {
		return ResourceCapabilities{}, err
	}
	head, err := service.repository.GetResource(ctx, scope, target)
	if errors.Is(err, control.ErrNotFound) {
		canPlan, permissionErr := permissionCapability(
			service.authorizer.AuthorizeCreate(ctx, scope, target),
		)
		return ResourceCapabilities{CanPlan: canPlan}, permissionErr
	}
	if err != nil {
		return ResourceCapabilities{}, err
	}
	if err := validateQueriedResource(scope, target, head); err != nil {
		return ResourceCapabilities{}, err
	}
	canReference, err := permissionCapability(
		service.authorizer.AuthorizeReference(ctx, scope, head),
	)
	if err != nil {
		return ResourceCapabilities{}, err
	}
	canPlan, err := permissionCapability(
		service.authorizer.AuthorizeUpdate(ctx, scope, head),
	)
	if err != nil {
		return ResourceCapabilities{}, err
	}
	return ResourceCapabilities{
		Exists:        true,
		CanViewSource: canReference,
		CanReference:  canReference,
		CanPlan:       canPlan,
	}, nil
}

func permissionCapability(err error) (bool, error) {
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ErrForbidden) {
		return false, nil
	}
	return false, err
}
