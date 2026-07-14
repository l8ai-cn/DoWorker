package infra

import (
	"context"
	"errors"

	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	"gorm.io/gorm"
)

func (repo *expertMarketRepo) CreateApplication(
	ctx context.Context,
	application *expertmarket.Application,
) error {
	err := repo.db.WithContext(ctx).Create(application).Error
	return expertMarketConflictError(err)
}

func (repo *expertMarketRepo) GetApplicationByID(
	ctx context.Context,
	id int64,
) (*expertmarket.Application, error) {
	var application expertmarket.Application
	err := repo.db.WithContext(ctx).First(&application, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, expertmarket.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &application, nil
}

func (repo *expertMarketRepo) GetApplicationBySlug(
	ctx context.Context,
	slug string,
) (*expertmarket.Application, error) {
	var application expertmarket.Application
	err := repo.db.WithContext(ctx).Where("slug = ?", slug).First(&application).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, expertmarket.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &application, nil
}

func (repo *expertMarketRepo) ListApplications(
	ctx context.Context,
	filter expertmarket.ApplicationListFilter,
) ([]expertmarket.Application, int64, error) {
	query := repo.db.WithContext(ctx).Model(&expertmarket.Application{})
	if filter.PublisherOrganizationID != nil {
		query = query.Where(
			"publisher_organization_id = ?",
			*filter.PublisherOrganizationID,
		)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var applications []expertmarket.Application
	err := query.Order("created_at DESC, id DESC").
		Limit(listLimit(filter.Limit)).
		Offset(filter.Offset).
		Find(&applications).Error
	return applications, total, err
}

func listLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	return limit
}
