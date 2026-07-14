package infra

import (
	"context"
	"errors"

	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	"gorm.io/gorm"
)

func (repo *expertMarketRepo) CreateRelease(
	ctx context.Context,
	release *expertmarket.Release,
) error {
	if err := release.Validate(); err != nil {
		return err
	}
	err := repo.db.WithContext(ctx).Create(release).Error
	return expertMarketConflictError(err)
}

func (repo *expertMarketRepo) GetReleaseByID(
	ctx context.Context,
	id int64,
) (*expertmarket.Release, error) {
	var release expertmarket.Release
	err := repo.db.WithContext(ctx).First(&release, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, expertmarket.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &release, nil
}

func (repo *expertMarketRepo) ListReleases(
	ctx context.Context,
	filter expertmarket.ReleaseListFilter,
) ([]expertmarket.Release, int64, error) {
	query := repo.db.WithContext(ctx).Model(&expertmarket.Release{})
	if filter.ApplicationID != nil {
		query = query.Where("application_id = ?", *filter.ApplicationID)
	}
	if filter.PublisherOrganizationID != nil {
		query = query.Where(
			"publisher_organization_id = ?",
			*filter.PublisherOrganizationID,
		)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var releases []expertmarket.Release
	err := query.Order("created_at DESC, id DESC").
		Limit(listLimit(filter.Limit)).
		Offset(filter.Offset).
		Find(&releases).Error
	return releases, total, err
}
