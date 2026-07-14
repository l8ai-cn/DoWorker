package infra

import (
	"errors"

	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func lockMarketApplication(
	tx *gorm.DB,
	applicationID int64,
) (*expertmarket.Application, error) {
	var application expertmarket.Application
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id", "latest_published_release_id").
		First(&application, applicationID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, expertmarket.ErrNotFound
	}
	if err != nil {
		return nil, expertMarketConflictError(err)
	}
	return &application, nil
}

func releaseNotFoundError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return expertmarket.ErrNotFound
	}
	return expertMarketConflictError(err)
}

func releaseApplicationError(tx *gorm.DB, releaseID int64, err error) error {
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return expertMarketConflictError(err)
	}
	var count int64
	if countErr := tx.Model(&expertmarket.Release{}).
		Where("id = ?", releaseID).
		Count(&count).Error; countErr != nil {
		return expertMarketConflictError(countErr)
	}
	if count > 0 {
		return expertmarket.ErrConflict
	}
	return expertmarket.ErrNotFound
}

func expertMarketConflictError(err error) error {
	if isForeignKeyViolation(err) || isUniqueViolation(err) {
		return expertmarket.ErrConflict
	}
	return err
}

func setLatestPublishedRelease(
	tx *gorm.DB,
	applicationID, releaseID int64,
) error {
	result := tx.Model(&expertmarket.Application{}).
		Where("id = ?", applicationID).
		Updates(map[string]any{
			"latest_published_release_id": releaseID,
			"updated_at":                  gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		return expertMarketConflictError(result.Error)
	}
	if result.RowsAffected == 0 {
		return expertmarket.ErrNotFound
	}
	return nil
}
