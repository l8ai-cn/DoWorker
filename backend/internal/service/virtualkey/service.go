package virtualkey

import (
	"context"
	"errors"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/virtualkey"
	airesourcesvc "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
)

var (
	ErrNotFound = errors.New("virtual api key not found")
	ErrRevoked  = errors.New("virtual api key is revoked")
)

type ModelResourceAccess interface {
	EnsureSelectable(ctx context.Context, actor airesourcesvc.Actor, orgID, resourceID int64) error
}

type Service struct {
	repo      domain.Repository
	resources ModelResourceAccess
}

func NewService(repo domain.Repository, resources ModelResourceAccess) *Service {
	return &Service{repo: repo, resources: resources}
}

type CreateInput struct {
	OrgID           int64
	UserID          int64
	ModelResourceID int64
	Name            string
	TokenBudget     *int64
}

// Created bundles the persisted row with the one-time plaintext token.
type Created struct {
	Key   *domain.VirtualAPIKey
	Token string
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*Created, error) {
	if err := s.resources.EnsureSelectable(ctx, airesourcesvc.Actor{UserID: in.UserID}, in.OrgID, in.ModelResourceID); err != nil {
		return nil, err
	}
	tok, err := newToken()
	if err != nil {
		return nil, err
	}
	k := &domain.VirtualAPIKey{
		OrganizationID:  in.OrgID,
		UserID:          in.UserID,
		ModelResourceID: in.ModelResourceID,
		Name:            in.Name,
		KeyPrefix:       tok.Prefix,
		KeyHash:         tok.Hash,
		TokenBudget:     in.TokenBudget,
		Status:          domain.StatusActive,
	}
	if err := s.repo.Create(ctx, k); err != nil {
		return nil, err
	}
	return &Created{Key: k, Token: tok.Token}, nil
}

func (s *Service) List(ctx context.Context, orgID, userID int64) ([]*domain.VirtualAPIKey, error) {
	return s.repo.ListByScope(ctx, orgID, userID)
}

func (s *Service) Get(ctx context.Context, id int64) (*domain.VirtualAPIKey, error) {
	k, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if k == nil {
		return nil, ErrNotFound
	}
	return k, nil
}

func (s *Service) Revoke(ctx context.Context, id, orgID, userID int64) error {
	updated, err := s.repo.UpdateStatusForScope(ctx, id, orgID, userID, domain.StatusRevoked)
	if err != nil {
		return err
	}
	if !updated {
		return ErrNotFound
	}
	return nil
}

func (s *Service) ResolveResourceForScope(
	ctx context.Context, keyID, orgID, userID int64,
) (int64, *int64, error) {
	k, err := s.repo.GetByIDForScope(ctx, keyID, orgID, userID)
	if err != nil {
		return 0, nil, err
	}
	if k == nil {
		return 0, nil, ErrNotFound
	}
	if k.Status != domain.StatusActive {
		return 0, nil, ErrRevoked
	}
	if err := s.resources.EnsureSelectable(ctx, airesourcesvc.Actor{UserID: userID}, orgID, k.ModelResourceID); err != nil {
		return 0, nil, err
	}
	if err := s.repo.TouchLastUsed(ctx, keyID); err != nil {
		return 0, nil, err
	}
	return k.ModelResourceID, k.TokenBudget, nil
}
