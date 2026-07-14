package infra

import (
	"context"
	"errors"

	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (repo *expertMarketRepo) WithdrawReleaseAndRefreshLatest(
	ctx context.Context,
	applicationID, releaseID int64,
	update expertmarket.LifecycleUpdate,
) error {
	if err := update.Validate(); err != nil {
		return err
	}
	if update.Status != expertmarket.ReleaseStatusWithdrawn {
		return expertmarket.ErrInvalidWithdrawalStatus
	}
	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		application, err := lockMarketApplication(tx, applicationID)
		if err != nil {
			return err
		}
		var release expertmarket.Release
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id").
			Where("id = ? AND application_id = ?", releaseID, applicationID).
			First(&release).Error; err != nil {
			return releaseApplicationError(tx, releaseID, err)
		}
		result := updateReleaseLifecycle(
			tx,
			"id = ? AND application_id = ?",
			[]any{releaseID, applicationID},
			update,
		)
		if result.Error != nil {
			return expertMarketConflictError(result.Error)
		}
		if application.LatestPublishedReleaseID == nil ||
			*application.LatestPublishedReleaseID != releaseID {
			return nil
		}
		return refreshLatestPublishedRelease(tx, applicationID, releaseID)
	})
}

func refreshLatestPublishedRelease(
	tx *gorm.DB,
	applicationID, withdrawnReleaseID int64,
) error {
	var previous expertmarket.Release
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id").
		Where(
			"application_id = ? AND id <> ? AND status = ?",
			applicationID,
			withdrawnReleaseID,
			expertmarket.ReleaseStatusPublished,
		).
		Order("version DESC, id DESC").
		First(&previous).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return setLatestPublishedReleaseID(tx, applicationID, nil)
	}
	if err != nil {
		return expertMarketConflictError(err)
	}
	return setLatestPublishedRelease(tx, applicationID, previous.ID)
}
