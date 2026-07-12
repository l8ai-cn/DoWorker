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
}

type modelResolver struct {
	resources ModelResourceResolver
}

var workerModelProtocolAdapters = map[string][]string{
	"do-agent":    {"openai-compatible", "anthropic", "minimax"},
	"codex-cli":   {"openai-compatible"},
	"claude-code": {"anthropic"},
	"gemini-cli":  {"gemini"},
	"grok-build":  {"openai-compatible"},
	"minimax-cli": {"minimax"},
	"openclaw":    {"openai-compatible", "anthropic", "gemini"},
	"hermes":      {"openai-compatible", "anthropic", "gemini"},
}

func newModelResolver(resources ModelResourceResolver) *modelResolver {
	return &modelResolver{resources: resources}
}

func (resolver *modelResolver) ResolveModel(
	ctx context.Context,
	scope specservice.Scope,
	workerType slugkit.Slug,
	resourceID int64,
) (specdomain.ModelBinding, error) {
	if resolver == nil || resolver.resources == nil {
		return specdomain.ModelBinding{}, specservice.ErrResolverUnavailable
	}
	requirements, err := modelRequirements(workerType)
	if err != nil {
		return specdomain.ModelBinding{}, err
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
			return specdomain.ModelBinding{}, fmt.Errorf(
				"%w: model resource: %w",
				specservice.ErrInvalidDraft,
				err,
			)
		}
		return specdomain.ModelBinding{}, err
	}
	if err := validateResolvedModel(resolved, resourceID); err != nil {
		return specdomain.ModelBinding{}, err
	}
	if err := validateWorkerModelProvider(workerType, resolved); err != nil {
		return specdomain.ModelBinding{}, err
	}
	return specdomain.ModelBinding{
		ResourceID:         resolved.Resource.ID,
		ResourceRevision:   resolved.Resource.Revision,
		ConnectionID:       resolved.Connection.ID,
		ConnectionRevision: resolved.Connection.Revision,
		ProviderKey:        resolved.Connection.ProviderKey,
		ModelID:            strings.TrimSpace(resolved.Resource.ModelID),
	}, nil
}

func modelRequirements(
	workerType slugkit.Slug,
) (resourceservice.ResolutionRequirements, error) {
	adapters, exists := workerModelProtocolAdapters[workerType.String()]
	if !exists {
		return resourceservice.ResolutionRequirements{}, fmt.Errorf(
			"%w: model resource: worker type %q has no supported model protocol",
			specservice.ErrInvalidDraft,
			workerType,
		)
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

func validateWorkerModelProvider(
	workerType slugkit.Slug,
	resolved *resourceservice.ResolvedResource,
) error {
	if workerType.String() == "grok-build" &&
		resolved.Connection.ProviderKey.String() != "xai" {
		return invalidResolvedModel("grok-build requires an xai provider")
	}
	return nil
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
