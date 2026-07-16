package workercreation

import (
	"context"
	"fmt"
	"strings"

	resourcedomain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	resourceservice "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func (resolver *modelResolver) ResolveToolModel(
	ctx context.Context,
	scope specservice.Scope,
	requirement specdomain.ToolModelRequirement,
	resourceID int64,
) (specdomain.ToolModelBinding, error) {
	if resolver == nil || resolver.resources == nil {
		return specdomain.ToolModelBinding{}, specservice.ErrResolverUnavailable
	}
	requirements, err := toolModelRequirements(requirement)
	if err != nil {
		return specdomain.ToolModelBinding{}, err
	}
	resolved, err := resolver.resources.ResolveExact(
		ctx,
		resourceservice.Actor{UserID: scope.UserID},
		scope.OrgID,
		resourceID,
		requirements,
	)
	if err != nil {
		if isModelSelectionError(err) {
			return specdomain.ToolModelBinding{}, fmt.Errorf(
				"%w: tool model %q: %w",
				specservice.ErrInvalidDraft,
				requirement.Role,
				err,
			)
		}
		return specdomain.ToolModelBinding{}, err
	}
	if err := validateResolvedModel(resolved, resourceID); err != nil {
		return specdomain.ToolModelBinding{}, err
	}
	if resolved.Provider.Key != resolved.Connection.ProviderKey ||
		!containsSlug(requirement.ProviderKeys, resolved.Connection.ProviderKey) {
		return specdomain.ToolModelBinding{}, invalidResolvedModel("provider was substituted")
	}
	if err := validateToolModelFamily(requirement, resolved); err != nil {
		return specdomain.ToolModelBinding{}, err
	}
	adapter, err := slugkit.NewFromTrusted(resolved.Provider.ProtocolAdapter)
	if err != nil {
		return specdomain.ToolModelBinding{}, invalidResolvedModel(
			"provider protocol adapter is invalid",
		)
	}
	return specdomain.ToolModelBinding{
		Role: requirement.Role,
		ModelBinding: specdomain.ModelBinding{
			ResourceID:         resolved.Resource.ID,
			ResourceRevision:   resolved.Resource.Revision,
			ConnectionID:       resolved.Connection.ID,
			ConnectionRevision: resolved.Connection.Revision,
			ProviderKey:        resolved.Connection.ProviderKey,
			ProtocolAdapter:    adapter,
			ModelID:            strings.TrimSpace(resolved.Resource.ModelID),
		},
		Modality:    requirement.Modality,
		Capability:  requirement.Capability,
		Environment: requirement.Environment,
	}, nil
}

func validateToolModelFamily(
	requirement specdomain.ToolModelRequirement,
	resolved *resourceservice.ResolvedResource,
) error {
	if err := resourcedomain.ValidateProviderModelCapability(
		resolved.Connection.ProviderKey,
		resolved.Resource.ModelID,
		requirement.Capability,
	); err != nil {
		return invalidResolvedModel(err.Error())
	}
	return nil
}

func toolModelRequirements(
	requirement specdomain.ToolModelRequirement,
) (resourceservice.ResolutionRequirements, error) {
	if err := slugkit.Validate(requirement.Role.String()); err != nil ||
		!requirement.Modality.Valid() ||
		!requirement.Capability.Valid() ||
		len(requirement.ProviderKeys) == 0 ||
		len(requirement.ProtocolAdapters) == 0 {
		return resourceservice.ResolutionRequirements{}, fmt.Errorf(
			"%w: invalid tool model requirement %q",
			specservice.ErrInvalidDraft,
			requirement.Role,
		)
	}
	adapters := make([]string, len(requirement.ProtocolAdapters))
	for index, adapter := range requirement.ProtocolAdapters {
		if err := slugkit.Validate(adapter.String()); err != nil {
			return resourceservice.ResolutionRequirements{}, fmt.Errorf(
				"%w: invalid tool model protocol adapter: %v",
				specservice.ErrInvalidDraft,
				err,
			)
		}
		adapters[index] = adapter.String()
	}
	return resourceservice.ResolutionRequirements{
		Modality: requirement.Modality, Capability: requirement.Capability,
		AllowedProtocolAdapters: adapters,
	}, nil
}

func containsSlug(values []slugkit.Slug, wanted slugkit.Slug) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
