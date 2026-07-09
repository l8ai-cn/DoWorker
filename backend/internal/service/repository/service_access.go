package repository

import (
	"context"

	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
)

func (s *Service) GetAccessibleByID(ctx context.Context, id, orgID, userID int64) (*gitprovider.Repository, error) {
	repo, err := s.repo.GetByID(ctx, id)
	return accessibleRepository(repo, err, orgID, userID)
}

func (s *Service) FindAccessibleByOrgSlug(ctx context.Context, orgID, userID int64, slug string) (*gitprovider.Repository, error) {
	repo, err := s.repo.FindByOrgSlug(ctx, orgID, slug)
	return accessibleRepository(repo, err, orgID, userID)
}

func accessibleRepository(repo *gitprovider.Repository, err error, orgID, userID int64) (*gitprovider.Repository, error) {
	if err != nil {
		return nil, err
	}
	if !isRepositoryAccessible(repo, orgID, userID) {
		return nil, ErrNoPermission
	}
	return repo, nil
}

func isRepositoryAccessible(repo *gitprovider.Repository, orgID, userID int64) bool {
	if repo == nil || repo.OrganizationID != orgID {
		return false
	}
	if repo.Visibility == "organization" {
		return true
	}
	return repo.Visibility == "private" && repo.ImportedByUserID != nil && *repo.ImportedByUserID == userID
}
