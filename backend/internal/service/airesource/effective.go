package airesource

import (
	"context"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
)

func (s *Service) ListOwnerConnections(ctx context.Context, actor Actor, scope domain.OwnerScope, ownerID int64) ([]ConnectionView, error) {
	canManage, err := s.authorizeOwner(ctx, actor, scope, ownerID, false)
	if err != nil {
		return nil, err
	}
	connections, err := s.repository.ListConnectionsByOwner(ctx, scope, ownerID)
	if err != nil {
		return nil, err
	}
	resources, err := s.repository.ListResourcesByOwner(ctx, scope, ownerID)
	if err != nil {
		return nil, err
	}
	byConnection := resourcesByConnection(resources)
	views := make([]ConnectionView, 0, len(connections))
	for _, connection := range connections {
		views = append(views, connectionView(connection, canManage, byConnection[connection.ID]))
	}
	return views, nil
}

func (s *Service) ListEffective(ctx context.Context, actor Actor, orgID int64, modalities []domain.Modality) ([]EffectiveResourceView, error) {
	if err := validateActor(actor); err != nil {
		return nil, err
	}
	if orgID < 0 {
		return nil, ErrInvalidOwner
	}
	for _, modality := range modalities {
		if !modality.Valid() {
			return nil, ErrIncompatibleModality
		}
	}
	orgCanManage := false
	if orgID > 0 {
		var err error
		orgCanManage, err = s.authorizeOwner(ctx, actor, domain.OwnerScopeOrg, orgID, false)
		if err != nil {
			return nil, err
		}
	}
	enabled, err := s.repository.ListEffective(ctx, actor.UserID, orgID, modalities)
	if err != nil {
		return nil, err
	}
	enabledByID := make(map[int64]*domain.ModelResource, len(enabled))
	for _, resource := range enabled {
		enabledByID[resource.ID] = resource
	}
	views, err := s.ownerEffectiveViews(ctx, domain.OwnerScopeUser, actor.UserID, true, modalities, enabledByID)
	if err != nil {
		return nil, err
	}
	if orgID > 0 {
		orgViews, listErr := s.ownerEffectiveViews(ctx, domain.OwnerScopeOrg, orgID, orgCanManage, modalities, enabledByID)
		if listErr != nil {
			return nil, listErr
		}
		views = append(views, orgViews...)
	}
	return views, nil
}

func (s *Service) ownerEffectiveViews(ctx context.Context, scope domain.OwnerScope, ownerID int64, canManage bool, modalities []domain.Modality, enabled map[int64]*domain.ModelResource) ([]EffectiveResourceView, error) {
	connections, err := s.repository.ListConnectionsByOwner(ctx, scope, ownerID)
	if err != nil {
		return nil, err
	}
	resources, err := s.repository.ListResourcesByOwner(ctx, scope, ownerID)
	if err != nil {
		return nil, err
	}
	byConnection := resourcesByConnection(resources)
	views := make([]EffectiveResourceView, 0)
	for _, connection := range connections {
		for _, resource := range byConnection[connection.ID] {
			if !matchesModalities(resource.Modalities, modalities) {
				continue
			}
			if effective := enabled[resource.ID]; effective != nil {
				resource = effective
			}
			reason := resourceBlockingReason(connection, resource)
			view := resourceView(resource)
			if reason != "" {
				view.DefaultModalities = nil
			}
			views = append(views, EffectiveResourceView{Connection: connectionView(connection, canManage, nil), Resource: view, Selectable: reason == "", BlockingReason: reason})
		}
	}
	return views, nil
}

func resourcesByConnection(resources []*domain.ModelResource) map[int64][]*domain.ModelResource {
	grouped := make(map[int64][]*domain.ModelResource)
	for _, resource := range resources {
		grouped[resource.ProviderConnectionID] = append(grouped[resource.ProviderConnectionID], resource)
	}
	return grouped
}

func matchesModalities(supported, wanted []domain.Modality) bool {
	if len(wanted) == 0 {
		return true
	}
	for _, modality := range wanted {
		if supportsModality(supported, modality) {
			return true
		}
	}
	return false
}

func resourceBlockingReason(connection *domain.Connection, resource *domain.ModelResource) BlockingReason {
	if !connection.IsEnabled {
		return BlockingConnectionDisabled
	}
	if !resource.IsEnabled {
		return BlockingResourceDisabled
	}
	if connection.Status == domain.ConnectionStatusUnchecked {
		return BlockingConnectionUnchecked
	}
	if connection.Status != domain.ConnectionStatusValid {
		return BlockingConnectionInvalid
	}
	if resource.Status == domain.ConnectionStatusUnchecked {
		return BlockingResourceUnchecked
	}
	if resource.Status != domain.ConnectionStatusValid {
		return BlockingResourceInvalid
	}
	return ""
}
