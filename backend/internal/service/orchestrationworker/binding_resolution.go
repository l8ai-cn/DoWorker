package orchestrationworker

import (
	"context"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
)

func resolveEntityID(
	ctx context.Context,
	scope control.Scope,
	reference resource.Reference,
	pins pinnedReferenceIndex,
	bindings BindingResolver,
) (int64, error) {
	pinned, err := pins.resolve(reference)
	if err != nil {
		return 0, err
	}
	id, err := bindings.ResolveEntityID(ctx, scope, pinned)
	if err != nil {
		return 0, err
	}
	if id <= 0 {
		return 0, control.ErrCorrupt
	}
	return id, nil
}

func resolveOptionalEntityID(
	ctx context.Context,
	scope control.Scope,
	reference *resource.Reference,
	pins pinnedReferenceIndex,
	bindings BindingResolver,
) (int64, error) {
	if reference == nil {
		return 0, nil
	}
	return resolveEntityID(ctx, scope, *reference, pins, bindings)
}
