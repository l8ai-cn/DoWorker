package infra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"gorm.io/gorm"
)

func (repo *aiResourceRepo) GetResourceByID(ctx context.Context, id int64) (*airesource.ModelResource, error) {
	var row modelResourceRow
	err := repo.db.WithContext(ctx).First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	resource := row.domain()
	defaults, err := repo.loadDefaultRows(ctx, []int64{id})
	if err != nil {
		return nil, err
	}
	resource.DefaultModalities = defaultModalities(defaults[id], nil)
	return resource, nil
}

func (repo *aiResourceRepo) CreateResource(ctx context.Context, resource *airesource.ModelResource) error {
	if err := airesource.ValidateModelResource(*resource); err != nil {
		return err
	}
	row := resourceRow(resource)
	if err := repo.db.WithContext(ctx).Create(row).Error; err != nil {
		if isUniqueViolation(err) {
			return airesource.ErrConflict
		}
		return err
	}
	*resource = *row.domain()
	return nil
}

func (repo *aiResourceRepo) SaveResource(ctx context.Context, resource *airesource.ModelResource) error {
	candidate := *resource
	candidate.DefaultModalities = nil
	if err := airesource.ValidateModelResource(candidate); err != nil {
		return err
	}
	if resource.Revision <= 0 {
		return airesource.ErrConflict
	}
	var saved *airesource.ModelResource
	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row := resourceRow(&candidate)
		result := tx.Model(&modelResourceRow{}).
			Where(
				"id = ? AND revision = ? AND updated_at = ? AND provider_connection_id = ?",
				resource.ID,
				resource.Revision,
				resource.UpdatedAt,
				resource.ProviderConnectionID,
			).
			Updates(map[string]any{
				"identifier": row.Identifier, "model_id": row.ModelID, "display_name": row.DisplayName,
				"modalities": row.Modalities, "capabilities": row.Capabilities, "status": row.Status,
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
			var stored modelResourceRow
			if err := tx.First(&stored, resource.ID).Error; err != nil {
				return err
			}
			if stored.ProviderConnectionID != resource.ProviderConnectionID {
				return fmt.Errorf("model resource %d connection is immutable", resource.ID)
			}
			return airesource.ErrConflict
		}
		var stored modelResourceRow
		if err := tx.First(&stored, resource.ID).Error; err != nil {
			return err
		}
		var defaults []modelResourceDefaultRow
		if err := tx.Where("model_resource_id = ?", resource.ID).Order("modality").Find(&defaults).Error; err != nil {
			return err
		}
		saved = stored.domain()
		saved.DefaultModalities = defaultModalities(defaults, nil)
		return nil
	})
	if err != nil {
		return err
	}
	*resource = *saved
	return nil
}

func (repo *aiResourceRepo) DeleteResource(ctx context.Context, id, expectedRevision int64, expectedUpdatedAt time.Time) error {
	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("model_resource_id = ?", id).Delete(&modelResourceDefaultRow{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&aiResourceMigrationRow{}).
			Where("model_resource_id = ?", id).Update("model_resource_id", nil).Error; err != nil {
			return err
		}
		result := tx.Where(
			"id = ? AND revision = ? AND updated_at = ?",
			id,
			expectedRevision,
			expectedUpdatedAt,
		).Delete(&modelResourceRow{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			var count int64
			if err := tx.Model(&modelResourceRow{}).Where("id = ?", id).Count(&count).Error; err != nil {
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

func (repo *aiResourceRepo) ListResourcesByConnection(ctx context.Context, connectionID int64) ([]*airesource.ModelResource, error) {
	var rows []modelResourceRow
	if err := repo.db.WithContext(ctx).Where("provider_connection_id = ?", connectionID).
		Order("identifier, id").Find(&rows).Error; err != nil {
		return nil, err
	}
	return repo.resourcesWithDefaults(ctx, rows, nil, false)
}

func (repo *aiResourceRepo) resourcesWithDefaults(
	ctx context.Context,
	rows []modelResourceRow,
	wanted map[airesource.Modality]struct{},
	personalPrecedence bool,
) ([]*airesource.ModelResource, error) {
	ids := make([]int64, len(rows))
	for index := range rows {
		ids[index] = rows[index].ID
	}
	defaults, err := repo.loadDefaultRows(ctx, ids)
	if err != nil {
		return nil, err
	}
	personal := make(map[airesource.Modality]struct{})
	if personalPrecedence {
		for _, rows := range defaults {
			for _, row := range rows {
				if row.OwnerScope == airesource.OwnerScopeUser {
					personal[row.Modality] = struct{}{}
				}
			}
		}
	}
	resources := make([]*airesource.ModelResource, len(rows))
	for index := range rows {
		resources[index] = rows[index].domain()
		resources[index].DefaultModalities = effectiveDefaultModalities(defaults[rows[index].ID], wanted, personal)
	}
	return resources, nil
}
