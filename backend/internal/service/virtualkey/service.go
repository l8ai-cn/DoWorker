package virtualkey

import (
	"context"
	"errors"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/virtualkey"
	aimodelsvc "github.com/anthropics/agentsmesh/backend/internal/service/aimodel"
)

var (
	ErrNotFound = errors.New("virtual api key not found")
	ErrRevoked  = errors.New("virtual api key is revoked")
)

// Service mints and resolves virtual API keys. Each key wraps an ai_models
// row (the real provider credential); resolution decrypts that credential so
// the orchestrator can inject it into the Worker pod.
type Service struct {
	repo   domain.Repository
	models *aimodelsvc.Service
}

func NewService(repo domain.Repository, models *aimodelsvc.Service) *Service {
	return &Service{repo: repo, models: models}
}

type CreateInput struct {
	OrgID       int64
	UserID      int64
	AIModelID   int64
	Name        string
	TokenBudget *int64
}

// Created bundles the persisted row with the one-time plaintext token.
type Created struct {
	Key   *domain.VirtualAPIKey
	Token string
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*Created, error) {
	if _, err := s.models.GetVisible(ctx, in.AIModelID, in.UserID, in.OrgID); err != nil {
		return nil, err
	}
	tok, err := newToken()
	if err != nil {
		return nil, err
	}
	k := &domain.VirtualAPIKey{
		OrganizationID: in.OrgID,
		UserID:         in.UserID,
		AIModelID:      in.AIModelID,
		Name:           in.Name,
		KeyPrefix:      tok.Prefix,
		KeyHash:        tok.Hash,
		TokenBudget:    in.TokenBudget,
		Status:         domain.StatusActive,
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

func (s *Service) Revoke(ctx context.Context, id int64) error {
	return s.repo.UpdateStatus(ctx, id, domain.StatusRevoked)
}

func (s *Service) ResolveModelForScope(
	ctx context.Context, keyID, orgID, userID int64,
) (*aimodelsvc.ResolvedModel, *int64, error) {
	k, err := s.repo.GetByIDForScope(ctx, keyID, orgID, userID)
	if err != nil {
		return nil, nil, err
	}
	if k == nil {
		return nil, nil, ErrNotFound
	}
	if k.Status != domain.StatusActive {
		return nil, nil, ErrRevoked
	}
	resolved, err := s.models.ResolveVisible(ctx, k.AIModelID, userID, orgID)
	if err != nil {
		return nil, nil, err
	}
	if err := s.repo.TouchLastUsed(ctx, keyID); err != nil {
		return nil, nil, err
	}
	return resolved, k.TokenBudget, nil
}
