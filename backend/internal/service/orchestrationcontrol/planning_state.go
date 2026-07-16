package orchestrationcontrol

import (
	"context"
	"errors"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
)

func (service *Service) loadCurrentRevision(
	ctx context.Context,
	scope control.Scope,
	draft validatedDraft,
) (*control.ResourceRevision, error) {
	if draft.head == nil {
		return nil, nil
	}
	revision, err := service.repository.GetRevision(
		ctx,
		scope,
		draft.head.ID,
		draft.head.Revision,
	)
	if err != nil {
		return nil, err
	}
	return &revision, nil
}

func (service *Service) ensureBaseCurrent(
	ctx context.Context,
	scope control.Scope,
	draft validatedDraft,
) error {
	current, err := service.repository.GetResource(
		ctx,
		scope,
		draft.result.Target,
	)
	if draft.head == nil {
		if errors.Is(err, control.ErrNotFound) {
			return service.authorizer.AuthorizeCreate(
				ctx,
				scope,
				draft.result.Target,
			)
		}
		if err != nil {
			return err
		}
		return control.ErrStale
	}
	if errors.Is(err, control.ErrNotFound) {
		return control.ErrStale
	}
	if err != nil {
		return err
	}
	if current.ID != draft.head.ID ||
		current.Identity != draft.head.Identity ||
		current.ResourceVersion != draft.head.ResourceVersion {
		return control.ErrStale
	}
	return service.authorizer.AuthorizeUpdate(ctx, scope, current)
}
