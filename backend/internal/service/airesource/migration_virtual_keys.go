package airesource

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func (m *LegacyMigrator) remapVirtualKeys(ctx context.Context, tx *gorm.DB, report *MigrationReport) error {
	if !tx.Migrator().HasTable("virtual_api_keys") ||
		!tx.Migrator().HasColumn("virtual_api_keys", "model_resource_id") {
		return nil
	}
	if !tx.Migrator().HasColumn("virtual_api_keys", "ai_model_id") {
		return ensureVirtualKeysMapped(ctx, tx)
	}
	result := tx.WithContext(ctx).Exec(
		`UPDATE virtual_api_keys
		    SET model_resource_id = (
		      SELECT model_resource_id FROM ai_resource_migration_map
		       WHERE source_kind = 'ai_model' AND source_id = virtual_api_keys.ai_model_id
		    )
		  WHERE model_resource_id IS NULL`,
	)
	if result.Error != nil {
		return result.Error
	}
	report.VirtualKeysRemapped = int(result.RowsAffected)
	return ensureVirtualKeysMapped(ctx, tx)
}

func ensureVirtualKeysMapped(ctx context.Context, tx *gorm.DB) error {
	var unmapped int64
	if err := tx.WithContext(ctx).Table("virtual_api_keys").
		Where("model_resource_id IS NULL").Count(&unmapped).Error; err != nil {
		return err
	}
	if unmapped > 0 {
		return fmt.Errorf("%d virtual API keys have no model resource mapping", unmapped)
	}
	return nil
}
