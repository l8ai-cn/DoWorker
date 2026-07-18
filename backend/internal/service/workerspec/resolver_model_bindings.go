package workerspec

import (
	"context"
	"fmt"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
)

func (resolver *Resolver) resolveModelBinding(
	ctx context.Context,
	scope Scope,
	requirement domain.ModelRequirement,
	resourceID int64,
) (domain.ModelBinding, error) {
	if !requirement.Required {
		if resourceID != 0 {
			return domain.ModelBinding{}, &InvalidDraftFieldError{
				Field:  "model_resource_id",
				Reason: "selected worker type does not use a platform model resource",
			}
		}
		return domain.ModelBinding{}, nil
	}
	if resourceID <= 0 {
		return domain.ModelBinding{}, &InvalidDraftFieldError{
			Field: "model_resource_id", Reason: "is required by the selected worker type",
		}
	}
	if resolver.models == nil {
		return domain.ModelBinding{}, ErrResolverUnavailable
	}
	binding, err := resolver.models.ResolveModel(ctx, scope, requirement, resourceID)
	if err != nil {
		return domain.ModelBinding{}, fmt.Errorf("resolve worker model: %w", err)
	}
	if err := validateModelResolution(resourceID, binding); err != nil {
		return domain.ModelBinding{}, err
	}
	return binding, nil
}

func (resolver *Resolver) resolveToolModelBindings(
	ctx context.Context,
	scope Scope,
	requirements []domain.ToolModelRequirement,
	resourceIDs map[string]int64,
) ([]domain.ToolModelBinding, error) {
	if len(requirements) == 0 {
		if len(resourceIDs) != 0 {
			return nil, &InvalidDraftFieldError{
				Field:  "tool_model_resource_ids",
				Reason: "selected worker type has no tool models",
			}
		}
		return nil, nil
	}
	if resolver.toolModels == nil {
		return nil, ErrResolverUnavailable
	}
	known := make(map[string]struct{}, len(requirements))
	bindings := make([]domain.ToolModelBinding, 0, len(requirements))
	for _, requirement := range requirements {
		role := requirement.Role.String()
		known[role] = struct{}{}
		resourceID := resourceIDs[role]
		if resourceID <= 0 {
			return nil, &InvalidDraftFieldError{
				Field:  "tool_model_resource_ids." + role,
				Reason: "is required by the selected worker type",
			}
		}
		binding, err := resolver.toolModels.ResolveToolModel(
			ctx, scope, requirement, resourceID,
		)
		if err != nil {
			return nil, fmt.Errorf("resolve worker tool model %q: %w", role, err)
		}
		if binding.Role != requirement.Role ||
			binding.ModelBinding.ResourceID != resourceID {
			return nil, fmt.Errorf("tool model resolver substituted %q", role)
		}
		bindings = append(bindings, binding)
	}
	for role := range resourceIDs {
		if _, exists := known[role]; !exists {
			return nil, &InvalidDraftFieldError{
				Field:  "tool_model_resource_ids." + role,
				Reason: "is not declared by the worker type",
			}
		}
	}
	return bindings, nil
}
