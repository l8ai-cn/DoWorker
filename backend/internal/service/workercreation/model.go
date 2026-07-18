package workercreation

import (
	"context"
	"errors"
	"fmt"
	"strings"

	resourcedomain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	resourceservice "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type ModelResourceResolver interface {
	ResolveExact(
		context.Context,
		resourceservice.Actor,
		int64,
		int64,
		resourceservice.ResolutionRequirements,
	) (*resourceservice.ResolvedResource, error)
	ResolveMetadata(
		context.Context,
		resourceservice.Actor,
		int64,
		int64,
		resourceservice.ResolutionRequirements,
	) (*resourceservice.ResolvedResource, error)
}

type modelResolver struct {
	resources ModelResourceResolver
}

func newModelResolver(resources ModelResourceResolver) *modelResolver {
	return &modelResolver{resources: resources}
}

func (resolver *modelResolver) ResolveModel(
	ctx context.Context,
	scope specservice.Scope,
	requirement specdomain.ModelRequirement,
	resourceID int64,
) (specdomain.ModelBinding, error) {
	if resolver == nil || resolver.resources == nil {
		return specdomain.ModelBinding{}, specservice.ErrResolverUnavailable
	}
	requirements, err := modelRequirements(requirement)
	if err != nil {
		return specdomain.ModelBinding{}, err
	}
	resolved, err := resolver.resources.ResolveMetadata(
		ctx,
		resourceservice.Actor{UserID: scope.UserID},
		scope.OrgID,
		resourceID,
		requirements,
	)
	if err != nil {
		if isModelSelectionError(err) {
			return specdomain.ModelBinding{}, &specservice.InvalidDraftFieldError{
				Field:  "model_resource_id",
				Reason: "selected model resource does not satisfy the selected worker type",
				Cause:  err,
			}
		}
		return specdomain.ModelBinding{}, err
	}
	if err := validateResolvedModel(resolved, resourceID); err != nil {
		return specdomain.ModelBinding{}, err
	}
	protocolAdapter, err := slugkit.NewFromTrusted(
		resolved.Provider.ProtocolAdapter,
	)
	if err != nil {
		return specdomain.ModelBinding{}, invalidResolvedModel(
			"provider protocol adapter is invalid",
		)
	}
	return specdomain.ModelBinding{
		ResourceID:         resolved.Resource.ID,
		ResourceRevision:   resolved.Resource.Revision,
		ConnectionID:       resolved.Connection.ID,
		ConnectionRevision: resolved.Connection.Revision,
		ProviderKey:        resolved.Connection.ProviderKey,
		ProtocolAdapter:    protocolAdapter,
		ModelID:            strings.TrimSpace(resolved.Resource.ModelID),
	}, nil
}

func modelRequirements(
	requirement specdomain.ModelRequirement,
) (resourceservice.ResolutionRequirements, error) {
	if !requirement.Required || len(requirement.ProtocolAdapters) == 0 {
		return resourceservice.ResolutionRequirements{}, fmt.Errorf(
			"%w: model resource is not required by the selected worker type",
			specservice.ErrInvalidDraft,
		)
	}
	adapters := make([]string, len(requirement.ProtocolAdapters))
	for index, adapter := range requirement.ProtocolAdapters {
		if err := slugkit.Validate(adapter.String()); err != nil {
			return resourceservice.ResolutionRequirements{}, fmt.Errorf(
				"%w: model resource: invalid protocol adapter: %w",
				specservice.ErrInvalidDraft,
				err,
			)
		}
		adapters[index] = adapter.String()
	}
	return resourceservice.ResolutionRequirements{
		Modality:                resourcedomain.ModalityChat,
		Capability:              resourcedomain.CapabilityTextGeneration,
		AllowedProtocolAdapters: append([]string{}, adapters...),
	}, nil
}

func validateResolvedModel(
	resolved *resourceservice.ResolvedResource,
	expectedResourceID int64,
) error {
	if resolved == nil {
		return invalidResolvedModel("selection is missing")
	}
	switch {
	case resolved.Resource.ID != expectedResourceID:
		return invalidResolvedModel("resource ID was substituted")
	case resolved.Resource.Revision <= 0:
		return invalidResolvedModel("resource revision is missing")
	case resolved.Connection.ID <= 0:
		return invalidResolvedModel("connection ID is missing")
	case resolved.Connection.ID != resolved.Resource.ProviderConnectionID:
		return invalidResolvedModel("connection ID does not match the resource")
	case resolved.Connection.Revision <= 0:
		return invalidResolvedModel("connection revision is missing")
	case slugkit.Validate(resolved.Connection.ProviderKey.String()) != nil:
		return invalidResolvedModel("provider key is invalid")
	case strings.TrimSpace(resolved.Resource.ModelID) == "":
		return invalidResolvedModel("provider model ID is missing")
	}
	return nil
}

func invalidResolvedModel(reason string) error {
	return fmt.Errorf("%w: model resource: %s", specservice.ErrInvalidDraft, reason)
}

func isModelSelectionError(err error) bool {
	expected := []error{
		resourceservice.ErrNotFound,
		resourceservice.ErrForbidden,
		resourceservice.ErrInvalidOwner,
		resourceservice.ErrInvalidProvider,
		resourceservice.ErrInvalidCredentials,
		resourceservice.ErrInvalidEndpoint,
		resourceservice.ErrDisabled,
		resourceservice.ErrUnhealthy,
		resourceservice.ErrUnchecked,
		resourceservice.ErrIncompatibleModality,
		resourceservice.ErrIncompatibleCapability,
		resourceservice.ErrIncompatibleProtocolAdapter,
	}
	for _, candidate := range expected {
		if errors.Is(err, candidate) {
			return true
		}
	}
	return false
}
