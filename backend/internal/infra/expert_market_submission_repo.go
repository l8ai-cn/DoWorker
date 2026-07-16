package infra

import (
	"context"

	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (repo *expertMarketRepo) CreateSubmission(
	ctx context.Context,
	application *expertmarket.Application,
	release *expertmarket.Release,
) error {
	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if application.ID == 0 {
			if err := tx.Create(application).Error; err != nil {
				return expertMarketConflictError(err)
			}
		} else {
			var locked expertmarket.Application
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&locked, application.ID).Error
			if err != nil {
				return releaseNotFoundError(err)
			}
			if locked.PublisherOrganizationID != application.PublisherOrganizationID {
				return expertmarket.ErrConflict
			}
		}
		var pendingCount int64
		if err := tx.Model(&expertmarket.Release{}).
			Where(
				"application_id = ? AND status = ?",
				application.ID,
				expertmarket.ReleaseStatusPendingReview,
			).
			Count(&pendingCount).Error; err != nil {
			return err
		}
		if pendingCount > 0 {
			return expertmarket.ErrPendingReleaseExists
		}
		var latestVersion int
		if err := tx.Model(&expertmarket.Release{}).
			Where("application_id = ?", application.ID).
			Select("COALESCE(MAX(version), 0)").
			Scan(&latestVersion).Error; err != nil {
			return err
		}
		release.ApplicationID = application.ID
		release.Version = latestVersion + 1
		if err := release.Validate(); err != nil {
			return err
		}
		if err := tx.Create(release).Error; err != nil {
			return expertMarketConflictError(err)
		}
		return nil
	})
}
