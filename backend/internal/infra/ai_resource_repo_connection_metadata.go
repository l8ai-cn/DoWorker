package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"gorm.io/gorm"
)

func (repo *aiResourceRepo) SaveConnectionMetadata(ctx context.Context, connection *airesource.Connection) error {
	if connection.Revision <= 0 {
		return airesource.ErrConflict
	}
	var saved *airesource.Connection
	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var stored providerConnectionRow
		if err := tx.First(&stored, connection.ID).Error; err != nil {
			return err
		}
		if stored.OwnerScope != connection.OwnerScope || stored.OwnerID != connection.OwnerID {
			return fmt.Errorf("provider connection %d owner is immutable", connection.ID)
		}
		if stored.Revision != connection.Revision || !stored.UpdatedAt.Equal(connection.UpdatedAt) {
			return airesource.ErrConflict
		}
		if stored.Identifier != connection.Identifier ||
			stored.ProviderKey != connection.ProviderKey ||
			stored.BaseURL != connection.BaseURL {
			return fmt.Errorf("provider connection %d metadata write cannot change runtime configuration", connection.ID)
		}
		row := connectionRow(connection)
		result := tx.Model(&providerConnectionRow{}).
			Where("id = ? AND revision = ? AND updated_at = ?", connection.ID, connection.Revision, connection.UpdatedAt).
			Updates(map[string]any{
				"name": row.Name, "credentials_encrypted": row.CredentialsEncrypted,
				"configured_fields": row.ConfiguredFields, "status": row.Status,
				"is_enabled": row.IsEnabled, "last_validated_at": row.LastValidatedAt,
				"validation_error": row.ValidationError, "updated_at": time.Now().UTC(),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return airesource.ErrConflict
		}
		if err := tx.First(&stored, connection.ID).Error; err != nil {
			return err
		}
		saved = stored.domain()
		return nil
	})
	if err != nil {
		return err
	}
	*connection = *saved
	return nil
}
