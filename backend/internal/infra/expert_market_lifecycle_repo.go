package infra

import (
	"context"

	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (repo *expertMarketRepo) UpdateReleaseLifecycle(
	ctx context.Context,
	releaseID int64,
	update expertmarket.LifecycleUpdate,
) error {
	if err := update.Validate(); err != nil {
		return err
	}
	if update.Status == expertmarket.ReleaseStatusPublished {
		return expertmarket.ErrPublicationRequiresLatestUpdate
	}
	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var release expertmarket.Release
		if err := tx.Select("id", "application_id").
			First(&release, releaseID).Error; err != nil {
			return releaseNotFoundError(err)
		}
		application, err := lockMarketApplication(tx, release.ApplicationID)
		if err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id", "status").
			Where("id = ? AND application_id = ?", releaseID, release.ApplicationID).
			First(&release).Error; err != nil {
			return releaseNotFoundError(err)
		}
		if err := ensureExpectedReleaseStatus(release.Status, update); err != nil {
			return err
		}
		if application.LatestPublishedReleaseID != nil &&
			*application.LatestPublishedReleaseID == releaseID {
			return expertmarket.ErrLatestReleaseStatusConflict
		}
		result := updateReleaseLifecycle(tx, "id = ?", []any{releaseID}, update)
		if result.Error != nil {
			return expertMarketConflictError(result.Error)
		}
		return nil
	})
}

func (repo *expertMarketRepo) UpdateReleaseLifecycleAndLatest(
	ctx context.Context,
	applicationID, releaseID int64,
	update expertmarket.LifecycleUpdate,
) error {
	if err := update.Validate(); err != nil {
		return err
	}
	if update.Status != expertmarket.ReleaseStatusPublished {
		return expertmarket.ErrInvalidLatestReleaseStatus
	}
	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		application, err := lockMarketApplication(tx, applicationID)
		if err != nil {
			return err
		}

		var release expertmarket.Release
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id", "version", "status").
			Where("id = ? AND application_id = ?", releaseID, applicationID).
			First(&release).Error; err != nil {
			return releaseApplicationError(tx, releaseID, err)
		}
		if err := ensureExpectedReleaseStatus(release.Status, update); err != nil {
			return err
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
		if application.LatestPublishedReleaseID == nil {
			return setLatestPublishedRelease(tx, applicationID, releaseID)
		}

		var latest expertmarket.Release
		if err := tx.Select("version").
			First(&latest, *application.LatestPublishedReleaseID).Error; err != nil {
			return err
		}
		if release.Version <= latest.Version {
			return nil
		}
		return setLatestPublishedRelease(tx, applicationID, releaseID)
	})
}

func ensureExpectedReleaseStatus(
	current expertmarket.ReleaseStatus,
	update expertmarket.LifecycleUpdate,
) error {
	if update.ExpectedStatus != nil && current != *update.ExpectedStatus {
		return expertmarket.ErrLifecycleStatusConflict
	}
	return nil
}

func updateReleaseLifecycle(
	db *gorm.DB,
	where string,
	args []any,
	update expertmarket.LifecycleUpdate,
) *gorm.DB {
	return db.Model(&expertmarket.Release{}).
		Where(where, args...).
		Updates(lifecycleUpdates(update))
}

func lifecycleUpdates(update expertmarket.LifecycleUpdate) map[string]any {
	updates := map[string]any{"status": update.Status}
	if update.ReviewerUserID != nil {
		updates["reviewer_user_id"] = *update.ReviewerUserID
	}
	if update.RejectionReason != nil {
		updates["rejection_reason"] = *update.RejectionReason
	}
	if update.SubmittedAt != nil {
		updates["submitted_at"] = *update.SubmittedAt
	}
	if update.ReviewedAt != nil {
		updates["reviewed_at"] = *update.ReviewedAt
	}
	if update.PublishedAt != nil {
		updates["published_at"] = *update.PublishedAt
	}
	if update.RejectedAt != nil {
		updates["rejected_at"] = *update.RejectedAt
	}
	if update.WithdrawnAt != nil {
		updates["withdrawn_at"] = *update.WithdrawnAt
	}
	return updates
}
