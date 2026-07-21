package infra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"gorm.io/gorm"
)

func (repo *aiResourceRepo) GetConnectionByID(ctx context.Context, id int64) (*airesource.Connection, error) {
	var row providerConnectionRow
	err := repo.db.WithContext(ctx).First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return row.domain(), nil
}

func (repo *aiResourceRepo) CreateConnection(ctx context.Context, connection *airesource.Connection) error {
	if err := connection.ValidateIdentifiers(); err != nil {
		return err
	}
	row := connectionRow(connection)
	if err := repo.db.WithContext(ctx).Create(row).Error; err != nil {
		if isUniqueViolation(err) {
			return airesource.ErrConflict
		}
		return err
	}
	*connection = *row.domain()
	return nil
}

func (repo *aiResourceRepo) SaveConnection(ctx context.Context, connection *airesource.Connection) error {
	if err := connection.ValidateIdentifiers(); err != nil {
		return err
	}
	if connection.Revision <= 0 {
		return airesource.ErrConflict
	}
	var saved *airesource.Connection
	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row := connectionRow(connection)
		result := tx.Model(&providerConnectionRow{}).
			Where(
				"id = ? AND revision = ? AND updated_at = ? AND owner_scope = ? AND owner_id = ?",
				connection.ID,
				connection.Revision,
				connection.UpdatedAt,
				connection.OwnerScope,
				connection.OwnerID,
			).
			Updates(map[string]any{
				"identifier": row.Identifier, "provider_key": row.ProviderKey, "name": row.Name,
				"base_url": row.BaseURL, "credentials_encrypted": row.CredentialsEncrypted,
				"configured_fields": row.ConfiguredFields, "status": row.Status,
				"is_enabled": row.IsEnabled, "last_validated_at": row.LastValidatedAt,
				"validation_error": row.ValidationError, "revision": gorm.Expr("revision + 1"),
				"updated_at": time.Now().UTC(),
			})
		if result.Error != nil {
			if isUniqueViolation(result.Error) {
				return airesource.ErrConflict
			}
			return result.Error
		}
		if result.RowsAffected == 0 {
			var stored providerConnectionRow
			if err := tx.First(&stored, connection.ID).Error; err != nil {
				return err
			}
			if stored.OwnerScope != connection.OwnerScope || stored.OwnerID != connection.OwnerID {
				return fmt.Errorf("provider connection %d owner is immutable", connection.ID)
			}
			return airesource.ErrConflict
		}
		var stored providerConnectionRow
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

func (repo *aiResourceRepo) DeleteConnection(ctx context.Context, id, expectedRevision int64, expectedUpdatedAt time.Time) error {
	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		resourceIDs := tx.Model(&modelResourceRow{}).
			Select("id").Where("provider_connection_id = ?", id)
		if err := tx.Where("model_resource_id IN (?)", resourceIDs).
			Delete(&modelResourceDefaultRow{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&aiResourceMigrationRow{}).
			Where("model_resource_id IN (?)", resourceIDs).Update("model_resource_id", nil).Error; err != nil {
			return err
		}
		if err := tx.Model(&aiResourceMigrationRow{}).
			Where("provider_connection_id = ?", id).Update("provider_connection_id", nil).Error; err != nil {
			return err
		}
		if err := tx.Where("provider_connection_id = ?", id).Delete(&modelResourceRow{}).Error; err != nil {
			return err
		}
		result := tx.Where(
			"id = ? AND revision = ? AND updated_at = ?",
			id,
			expectedRevision,
			expectedUpdatedAt,
		).Delete(&providerConnectionRow{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			var count int64
			if err := tx.Model(&providerConnectionRow{}).Where("id = ?", id).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				return gorm.ErrRecordNotFound
			}
			return airesource.ErrConflict
		}
		return nil
	})
}

func (repo *aiResourceRepo) ListConnectionsByOwner(ctx context.Context, scope airesource.OwnerScope, ownerID int64) ([]*airesource.Connection, error) {
	var rows []providerConnectionRow
	if err := repo.db.WithContext(ctx).Where("owner_scope = ? AND owner_id = ?", scope, ownerID).
		Order("identifier, id").Find(&rows).Error; err != nil {
		return nil, err
	}
	connections := make([]*airesource.Connection, len(rows))
	for index := range rows {
		connections[index] = rows[index].domain()
	}
	return connections, nil
}

type aiResourceMigrationRow struct {
	ID                   int64
	ProviderConnectionID *int64
	ModelResourceID      *int64
}

func (aiResourceMigrationRow) TableName() string { return "ai_resource_migration_map" }
