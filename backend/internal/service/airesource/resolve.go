package airesource

import (
	"context"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

func (s *Service) ResolveExact(ctx context.Context, actor Actor, orgID, resourceID int64, required ResolutionRequirements) (*ResolvedResource, error) {
	resolved, err := s.resolveMetadata(ctx, actor, orgID, resourceID, required)
	if err != nil {
		return nil, err
	}
	credentials, err := s.decryptCredentials(&resolved.Connection)
	if err != nil {
		return nil, err
	}
	resolved.Credentials = credentials
	return resolved, nil
}

func (s *Service) ResolveMetadata(ctx context.Context, actor Actor, orgID, resourceID int64, required ResolutionRequirements) (*ResolvedResource, error) {
	return s.resolveMetadata(ctx, actor, orgID, resourceID, required)
}

func (s *Service) resolveMetadata(ctx context.Context, actor Actor, orgID, resourceID int64, required ResolutionRequirements) (*ResolvedResource, error) {
	if err := validateResolutionRequirements(required); err != nil {
		return nil, err
	}
	if orgID < 0 {
		return nil, ErrInvalidOwner
	}
	if orgID > 0 {
		if _, err := s.authorizeOwner(ctx, actor, domain.OwnerScopeOrg, orgID, false); err != nil {
			return nil, err
		}
	}
	resource, connection, _, err := s.resourceForActor(ctx, actor, resourceID, false)
	if err != nil {
		return nil, err
	}
	if connection.OwnerScope == domain.OwnerScopeOrg && orgID != connection.OwnerID {
		return nil, ErrForbidden
	}
	if !connection.IsEnabled || !resource.IsEnabled {
		return nil, ErrDisabled
	}
	if err := healthyStatus(connection.Status); err != nil {
		return nil, err
	}
	if err := healthyStatus(resource.Status); err != nil {
		return nil, err
	}
	if !supportsModality(resource.Modalities, required.Modality) {
		return nil, ErrIncompatibleModality
	}
	if !containsCapability(resource.Capabilities, required.Capability) {
		return nil, ErrIncompatibleCapability
	}
	provider, exists := domain.Provider(connection.ProviderKey.String())
	if !exists {
		return nil, ErrInvalidProvider
	}
	if !supportsModality(provider.Modalities, required.Modality) {
		return nil, ErrIncompatibleModality
	}
	if !containsString(required.AllowedProtocolAdapters, provider.ProtocolAdapter) {
		return nil, ErrIncompatibleProtocolAdapter
	}
	if err := s.endpoints.Validate(ctx, connection.BaseURL); err != nil {
		return nil, ErrInvalidEndpoint
	}
	return &ResolvedResource{Provider: provider, Connection: *connection, Resource: *resource}, nil
}

func (s *Service) EnsureSelectable(ctx context.Context, actor Actor, orgID, resourceID int64) error {
	views, err := s.ListEffective(ctx, actor, orgID, nil)
	if err != nil {
		return err
	}
	for _, view := range views {
		if view.Resource.ID != resourceID {
			continue
		}
		if view.Selectable {
			return nil
		}
		return blockingReasonError(view.BlockingReason)
	}
	return ErrNotFound
}

func blockingReasonError(reason BlockingReason) error {
	switch reason {
	case BlockingConnectionDisabled, BlockingResourceDisabled:
		return ErrDisabled
	case BlockingConnectionUnchecked, BlockingResourceUnchecked:
		return ErrUnchecked
	case BlockingConnectionInvalid, BlockingResourceInvalid:
		return ErrUnhealthy
	default:
		return ErrUnhealthy
	}
}

func validateResolutionRequirements(required ResolutionRequirements) error {
	if !required.Modality.Valid() || !required.Capability.Valid() || len(required.AllowedProtocolAdapters) == 0 {
		return ErrInvalidRequirements
	}
	if !supportsCapability([]domain.Capability{required.Capability}, required.Modality) {
		return ErrInvalidRequirements
	}
	for _, adapter := range required.AllowedProtocolAdapters {
		if slugkit.Validate(adapter) != nil {
			return ErrInvalidRequirements
		}
	}
	return nil
}

func containsCapability(values []domain.Capability, wanted domain.Capability) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func healthyStatus(status domain.ConnectionStatus) error {
	switch status {
	case domain.ConnectionStatusValid:
		return nil
	case domain.ConnectionStatusUnchecked:
		return ErrUnchecked
	default:
		return ErrUnhealthy
	}
}
