package infra

import (
	"context"

	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
)

func (r *gitProviderRepo) FindByOrgSlug(ctx context.Context, orgID int64, slug string) (*gitprovider.Repository, error) {
	var repo gitprovider.Repository
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND slug = ? AND deleted_at IS NULL", orgID, slug).
		First(&repo).Error
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &repo, nil
}

func (r *gitProviderRepo) ListByOrgSlug(ctx context.Context, orgID int64, slug string) ([]*gitprovider.Repository, error) {
	var repos []*gitprovider.Repository
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND slug = ? AND deleted_at IS NULL", orgID, slug).
		Order("id ASC").
		Find(&repos).Error
	return repos, err
}
