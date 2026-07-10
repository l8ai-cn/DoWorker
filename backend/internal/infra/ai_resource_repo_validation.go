package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"gorm.io/gorm"
)

func (repo *aiResourceRepo) SetValidationState(
	ctx context.Context,
	connectionID, expectedRevision int64,
	status airesource.ConnectionStatus,
	at time.Time,
	validationError string,
) (int64, error) {
	if status != airesource.ConnectionStatusValid && status != airesource.ConnectionStatusInvalid && status != airesource.ConnectionStatusUnchecked {
		return 0, fmt.Errorf("invalid persisted AI resource validation status %q", status)
	}
	newRevision := expectedRevision + 1
	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{
			"status": status, "last_validated_at": at,
			"validation_error": validationError, "revision": gorm.Expr("revision + 1"),
			"updated_at": time.Now().UTC(),
		}
		result := tx.Model(&providerConnectionRow{}).Where("id = ? AND revision = ?", connectionID, expectedRevision).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			var count int64
			if err := tx.Model(&providerConnectionRow{}).Where("id = ?", connectionID).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				return gorm.ErrRecordNotFound
			}
			return airesource.ErrConflict
		}
		return tx.Model(&modelResourceRow{}).Where("provider_connection_id = ?", connectionID).Updates(updates).Error
	})
	if err != nil {
		return 0, err
	}
	return newRevision, nil
}
