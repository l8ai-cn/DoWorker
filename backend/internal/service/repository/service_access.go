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
	repos, err := s.repo.ListByOrgSlug(ctx, orgID, slug)
	if err != nil {
		return nil, err
	}

	var accessible *gitprovider.Repository
	for _, repo := range repos {
		if !isRepositoryAccessible(repo, orgID, userID) {
			continue
		}
		if accessible != nil {
			return nil, ErrAmbiguousRepositorySlug
		}
		accessible = repo
	}
	if accessible == nil {
		return nil, ErrNoPermission
	}
	return accessible, nil
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
