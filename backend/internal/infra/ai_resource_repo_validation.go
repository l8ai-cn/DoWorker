package infra

import (
	"context"
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (repo *aiResourceRepo) SetValidationState(
	ctx context.Context,
	connectionID, expectedRevision int64,
	expectedCredentialsEncrypted string,
	status airesource.ConnectionStatus,
	at time.Time,
	validationError string,
) (int64, error) {
	if status != airesource.ConnectionStatusValid && status != airesource.ConnectionStatusInvalid && status != airesource.ConnectionStatusUnchecked {
		return 0, fmt.Errorf("invalid persisted AI resource validation status %q", status)
	}
	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var stored providerConnectionRow
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id", "revision", "credentials_encrypted").
			First(&stored, connectionID).Error; err != nil {
			return err
		}
		if stored.Revision != expectedRevision ||
			subtle.ConstantTimeCompare([]byte(stored.CredentialsEncrypted), []byte(expectedCredentialsEncrypted)) != 1 {
			return airesource.ErrConflict
		}
		updates := map[string]any{
			"status": status, "last_validated_at": at,
			"validation_error": validationError, "updated_at": time.Now().UTC(),
		}
		result := tx.Model(&providerConnectionRow{}).
			Where("id = ? AND revision = ?", connectionID, expectedRevision).
			Updates(updates)
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
	return expectedRevision, nil
}
