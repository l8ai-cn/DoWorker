package airesource

import (
	"context"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/envbundle"
	"gorm.io/gorm"
)

func (m *LegacyMigrator) Check(ctx context.Context) (*MigrationCheckReport, error) {
	report := &MigrationCheckReport{}
	err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := m.checkAIModels(ctx, tx, report); err != nil {
			return err
		}
		if err := m.checkCredentialBundles(ctx, tx, report); err != nil {
			return err
		}
		return m.checkMappingIntegrity(ctx, tx, report)
	})
	if err != nil {
		return nil, err
	}
	return report, nil
}

func (m *LegacyMigrator) checkAIModels(ctx context.Context, tx *gorm.DB, report *MigrationCheckReport) error {
	if !tx.Migrator().HasTable("ai_models") {
		return nil
	}
	var rows []legacyAIModelRow
	if err := tx.WithContext(ctx).Table("ai_models").Order("id").Find(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		credentials, decryptErr := decryptJSONMap(m.cipher, row.EncryptedCredentials)
		if decryptErr != nil {
			report.DecryptFailures++
		}
		if modelResourceID, ok, err := mappedResourceID(ctx, tx, "ai_model", row.ID); err != nil {
			return err
		} else if !ok {
			report.UnmigratedAIModels++
		} else if modelResourceID != row.ID {
			report.BrokenMappings++
		} else if decryptErr == nil {
			if err := m.checkAIModelParity(ctx, tx, row, credentials, report); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *LegacyMigrator) checkCredentialBundles(ctx context.Context, tx *gorm.DB, report *MigrationCheckReport) error {
	if !tx.Migrator().HasTable("env_bundles") {
		return nil
	}
	var rows []legacyEnvBundleRow
	if err := tx.WithContext(ctx).Table("env_bundles").
		Where("kind = ? AND is_active = ?", envbundle.KindCredential, true).
		Order("id").Find(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		values, decryptErr := decryptBundleValues(m.cipher, row.Data)
		if decryptErr != nil {
			report.DecryptFailures++
		}
		if _, ok, err := mappedResourceID(ctx, tx, "env_bundle", row.ID); err != nil {
			return err
		} else if !ok {
			report.UnmigratedEnvBundles++
		} else if decryptErr == nil {
			if err := m.checkCredentialBundleParity(ctx, tx, row, values, report); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *LegacyMigrator) checkMappingIntegrity(ctx context.Context, tx *gorm.DB, report *MigrationCheckReport) error {
	if tx.Migrator().HasTable("virtual_api_keys") {
		unmapped, err := m.countUnmappedVirtualKeys(ctx, tx)
		if err != nil {
			return err
		}
		report.UnmappedVirtualKeys = unmapped
	}
	var broken int64
	err := tx.WithContext(ctx).Raw(
		`SELECT count(*)
		   FROM ai_resource_migration_map migration
		   LEFT JOIN provider_connections connection ON connection.id = migration.provider_connection_id
		   LEFT JOIN model_resources resource ON resource.id = migration.model_resource_id
		  WHERE migration.status = 'migrated'
		    AND (
		      connection.id IS NULL
		      OR resource.id IS NULL
		      OR resource.provider_connection_id <> connection.id
		    )`,
	).Scan(&broken).Error
	report.BrokenMappings += int(broken)
	return err
}

func (m *LegacyMigrator) countUnmappedVirtualKeys(ctx context.Context, tx *gorm.DB) (int, error) {
	if tx.Migrator().HasColumn("virtual_api_keys", "model_resource_id") {
		var unmapped int64
		err := tx.WithContext(ctx).Table("virtual_api_keys").
			Where("model_resource_id IS NULL").Count(&unmapped).Error
		return int(unmapped), err
	}
	if !tx.Migrator().HasColumn("virtual_api_keys", "ai_model_id") {
		return 0, nil
	}
	var unmapped int64
	err := tx.WithContext(ctx).Raw(
		`SELECT count(*)
		   FROM virtual_api_keys key
		   LEFT JOIN ai_resource_migration_map migration
		     ON migration.source_kind = 'ai_model'
		    AND migration.source_id = key.ai_model_id
		    AND migration.status = 'migrated'
		    AND migration.model_resource_id IS NOT NULL
		  WHERE migration.model_resource_id IS NULL`,
	).Scan(&unmapped).Error
	return int(unmapped), err
}

func mappedResourceID(ctx context.Context, tx *gorm.DB, kind string, id int64) (int64, bool, error) {
	var row struct{ ModelResourceID *int64 }
	err := tx.WithContext(ctx).Table("ai_resource_migration_map").
		Select("model_resource_id").
		Where("source_kind = ? AND source_id = ? AND status = ?", kind, id, "migrated").
		Limit(1).Scan(&row).Error
	if err != nil {
		return 0, false, err
	}
	if row.ModelResourceID == nil {
		return 0, false, nil
	}
	return *row.ModelResourceID, true, nil
}
