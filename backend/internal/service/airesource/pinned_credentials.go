package airesource

import (
	"context"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
)

func (s *Service) ResolvePinnedCredentials(
	ctx context.Context,
	actor Actor,
	orgID int64,
	resourceID int64,
	connectionID int64,
) (map[string]string, error) {
	if orgID > 0 {
		if _, err := s.authorizeOwner(ctx, actor, domain.OwnerScopeOrg, orgID, false); err != nil {
			return nil, err
		}
	}
	resource, connection, _, err := s.resourceForActor(ctx, actor, resourceID, false)
	if err != nil {
		return nil, err
	}
	if connection.ID != connectionID ||
		resource.ProviderConnectionID != connectionID {
		return nil, ErrForbidden
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
	return s.decryptCredentials(connection)
}
