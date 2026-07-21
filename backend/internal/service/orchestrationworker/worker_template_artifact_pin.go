package orchestrationworker

import (
	"context"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
)

func resolvePinnedID(
	ctx context.Context,
	scope control.Scope,
	reference resource.Reference,
	pins pinnedReferenceIndex,
	bindings BindingResolver,
) (control.ResolvedReference, int64, error) {
	pinned, err := pins.resolve(reference)
	if err != nil {
		return control.ResolvedReference{}, 0, err
	}
	id, err := bindings.ResolveEntityID(ctx, scope, pinned)
	if err != nil {
		return control.ResolvedReference{}, 0, err
	}
	if id <= 0 {
		return control.ResolvedReference{}, 0, control.ErrCorrupt
	}
	return pinned, id, nil
}
