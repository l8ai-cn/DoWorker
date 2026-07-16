package orchestrationcontrol

import (
	"context"
	"fmt"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
)

func (service *Service) AuthorizeApply(
	ctx context.Context,
	scope control.Scope,
	planID string,
) error {
	if service == nil || service.repository == nil || service.authorizer == nil {
		return ErrUnavailable
	}
	if err := scope.Validate(); err != nil {
		return err
	}
	plan, err := service.repository.GetPlan(ctx, scope, planID)
	if err != nil {
		return err
	}
	if err := plan.Validate(); err != nil {
		return fmt.Errorf("%w: invalid resource plan", control.ErrCorrupt)
	}
	if plan.Scope != scope || plan.ActorID != scope.ActorID {
		return ErrForbidden
	}
	if err := service.authorizeApplyTarget(ctx, scope, plan); err != nil {
		return err
	}
	for _, reference := range plan.ResolvedReferences {
		if err := service.authorizeApplyReference(
			ctx,
			scope,
			reference,
		); err != nil {
			return err
		}
	}
	return nil
}

func (service *Service) authorizeApplyTarget(
	ctx context.Context,
	scope control.Scope,
	plan control.Plan,
) error {
	if plan.Operation == control.PlanOperationCreate {
		return service.authorizer.AuthorizeCreate(ctx, scope, plan.Target)
	}
	head, err := service.repository.GetResource(ctx, scope, plan.Target)
	if err != nil {
		return err
	}
	if head.ID != plan.TargetResourceID ||
		head.Identity.UID != plan.BaseUID {
		return control.ErrCorrupt
	}
	return service.authorizer.AuthorizeUpdate(ctx, scope, head)
}

func (service *Service) authorizeApplyReference(
	ctx context.Context,
	scope control.Scope,
	reference control.ResolvedReference,
) error {
	target := control.ResourceTarget{
		TypeMeta:  reference.TypeMeta,
		Namespace: reference.Namespace,
		Name:      reference.Name,
	}
	head, err := service.repository.GetResource(ctx, scope, target)
	if err != nil {
		return err
	}
	if head.Identity.ResourceTarget != target ||
		head.Identity.UID != reference.UID {
		return control.ErrCorrupt
	}
	if err := service.authorizer.AuthorizeReference(ctx, scope, head); err != nil {
		return err
	}
	revision, err := service.repository.GetRevision(
		ctx,
		scope,
		head.ID,
		reference.Revision,
	)
	if err != nil {
		return err
	}
	if err := validateReferenceRevision(
		scope,
		head,
		revision,
		reference.Revision,
	); err != nil {
		return err
	}
	if revision.Digest != reference.Digest {
		return control.ErrCorrupt
	}
	return nil
}
