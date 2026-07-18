package infra

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"gorm.io/gorm"
)

func (repo *aiResourceRepo) SaveResourceMetadata(ctx context.Context, resource *airesource.ModelResource) error {
	if resource.Revision <= 0 {
		return airesource.ErrConflict
	}
	var saved *airesource.ModelResource
	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var stored modelResourceRow
		if err := tx.First(&stored, resource.ID).Error; err != nil {
			return err
		}
		if stored.ProviderConnectionID != resource.ProviderConnectionID {
			return fmt.Errorf("model resource %d connection is immutable", resource.ID)
		}
		if stored.Revision != resource.Revision || !stored.UpdatedAt.Equal(resource.UpdatedAt) {
			return airesource.ErrConflict
		}
		row := resourceRow(resource)
		if stored.Identifier != resource.Identifier ||
			stored.ModelID != resource.ModelID ||
			!slices.Equal(stored.Modalities, row.Modalities) ||
			!slices.Equal(stored.Capabilities, row.Capabilities) {
			return fmt.Errorf("model resource %d metadata write cannot change runtime configuration", resource.ID)
		}
		result := tx.Model(&modelResourceRow{}).
			Where("id = ? AND revision = ? AND updated_at = ?", resource.ID, resource.Revision, resource.UpdatedAt).
			Updates(map[string]any{
				"display_name": row.DisplayName, "status": row.Status,
				"is_enabled": row.IsEnabled, "last_validated_at": row.LastValidatedAt,
				"validation_error": row.ValidationError, "updated_at": time.Now().UTC(),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return airesource.ErrConflict
		}
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
