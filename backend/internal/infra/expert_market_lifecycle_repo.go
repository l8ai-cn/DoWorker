package infra

import (
	"context"

	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	"gorm.io/gorm"
)

func (repo *expertMarketRepo) UpdateReleaseLifecycle(
	ctx context.Context,
	releaseID int64,
	update expertmarket.LifecycleUpdate,
) error {
	if err := update.Validate(); err != nil {
		return err
	}
	result := updateReleaseLifecycle(
		repo.db.WithContext(ctx),
		"id = ?",
		[]any{releaseID},
		update,
	)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return expertmarket.ErrNotFound
	}
	return nil
}

func (repo *expertMarketRepo) UpdateReleaseLifecycleAndLatest(
	ctx context.Context,
	applicationID, releaseID int64,
	update expertmarket.LifecycleUpdate,
) error {
	if err := update.Validate(); err != nil {
		return err
	}
	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := updateReleaseLifecycle(
			tx,
			"id = ? AND application_id = ?",
			[]any{releaseID, applicationID},
			update,
		)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			var count int64
			if err := tx.Model(&expertmarket.Release{}).
				Where("id = ?", releaseID).
				Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return expertmarket.ErrConflict
			}
			return expertmarket.ErrNotFound
		}
		result = tx.Model(&expertmarket.Application{}).
			Where("id = ?", applicationID).
			Updates(map[string]any{
				"latest_published_release_id": releaseID,
				"updated_at":                  gorm.Expr("CURRENT_TIMESTAMP"),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return expertmarket.ErrNotFound
		}
		return nil
	})
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
