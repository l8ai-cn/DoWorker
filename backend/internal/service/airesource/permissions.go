package airesource

import (
	"context"
	"errors"
	"fmt"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/organization"
)

func validateActor(actor Actor) error {
	if actor.UserID <= 0 {
		return fmt.Errorf("%w: actor ID must be positive", ErrInvalidOwner)
	}
	return nil
}

func (s *Service) authorizeOwner(ctx context.Context, actor Actor, scope domain.OwnerScope, ownerID int64, manage bool) (bool, error) {
	if err := validateActor(actor); err != nil {
		return false, err
	}
	if ownerID <= 0 {
		return false, fmt.Errorf("%w: owner ID must be positive", ErrInvalidOwner)
	}
	switch scope {
	case domain.OwnerScopeUser:
		if actor.UserID != ownerID {
			return false, ErrForbidden
		}
		return true, nil
	case domain.OwnerScopeOrg:
		member, err := s.members.GetMember(ctx, ownerID, actor.UserID)
		if errors.Is(err, organization.ErrMemberNotFound) || (err == nil && member == nil) {
			return false, ErrForbidden
		}
		if err != nil {
			return false, err
		}
		canManage := member.Role == organization.RoleOwner || member.Role == organization.RoleAdmin
		if manage && !canManage {
			return false, ErrForbidden
		}
		return canManage, nil
	default:
		return false, fmt.Errorf("%w: unknown owner scope %q", ErrInvalidOwner, scope)
	}
}

func (s *Service) connectionForActor(ctx context.Context, actor Actor, connectionID int64, manage bool) (*domain.Connection, bool, error) {
	if connectionID <= 0 {
		return nil, false, ErrNotFound
	}
	connection, err := s.repository.GetConnectionByID(ctx, connectionID)
	if err != nil {
		return nil, false, err
	}
	if connection == nil {
		return nil, false, ErrNotFound
	}
	canManage, err := s.authorizeOwner(ctx, actor, connection.OwnerScope, connection.OwnerID, manage)
	if err != nil {
		return nil, false, err
	}
	return connection, canManage, nil
}

func (s *Service) resourceForActor(ctx context.Context, actor Actor, resourceID int64, manage bool) (*domain.ModelResource, *domain.Connection, bool, error) {
	if resourceID <= 0 {
		return nil, nil, false, ErrNotFound
	}
	resource, err := s.repository.GetResourceByID(ctx, resourceID)
	if err != nil {
		return nil, nil, false, err
	}
	if resource == nil {
		return nil, nil, false, ErrNotFound
	}
	connection, canManage, err := s.connectionForActor(ctx, actor, resource.ProviderConnectionID, manage)
	if err != nil {
		return nil, nil, false, err
	}
	return resource, connection, canManage, nil
}
