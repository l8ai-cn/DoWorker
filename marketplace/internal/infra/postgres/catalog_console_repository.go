package postgres

import (
	"context"

	"github.com/l8ai-cn/agentcloud/marketplace/internal/domain/catalog"
	"github.com/l8ai-cn/agentcloud/marketplace/internal/service"
	"gorm.io/gorm"
)

type CatalogConsoleRepository struct {
	db *gorm.DB
}

func NewCatalogConsoleRepository(db *gorm.DB) *CatalogConsoleRepository {
	return &CatalogConsoleRepository{db: db}
}

func (r *CatalogConsoleRepository) CreateCatalogItem(
	ctx context.Context,
	item *catalog.Item,
) (int64, error) {
	var row struct{ ID int64 }
	err := r.db.WithContext(ctx).Raw(`
INSERT INTO marketplace.marketplace_catalog_items
  (publisher_id, slug, resource_type, name, summary, platform_resource_type,
   platform_resource_id, status, created_by_platform_user_id, revision)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1)
RETURNING id
`, item.PublisherID(), item.Slug().String(), item.ResourceType(), item.Name(),
		item.Summary(), item.PlatformResourceType(), item.PlatformResourceID(),
		item.Status(), item.CreatedByPlatformUserID()).Scan(&row).Error
	return row.ID, err
}

func (r *CatalogConsoleRepository) CreateCatalogVersion(
	ctx context.Context,
	version *catalog.Version,
) (int64, error) {
	var row struct{ ID int64 }
	err := r.db.WithContext(ctx).Raw(`
INSERT INTO marketplace.marketplace_catalog_item_versions
  (catalog_item_id, version, source_revision, content_digest, manifest,
   compatibility, validation_status, created_by_platform_user_id)
VALUES (?, ?, ?, ?, ?::jsonb, ?::jsonb, ?, ?)
RETURNING id
`, version.CatalogItemID(), version.Version(), version.SourceRevision(),
		version.ContentDigest(), string(version.Manifest()),
		string(version.Compatibility()), version.ValidationStatus(),
		version.CreatedByPlatformUserID()).Scan(&row).Error
	return row.ID, err
}

func (r *CatalogConsoleRepository) GetCatalogItem(
	ctx context.Context,
	id int64,
) (*catalog.Item, error) {
	var row catalog.ItemState
	result := r.db.WithContext(ctx).Raw(`
SELECT id, publisher_id, slug, resource_type, name, summary, platform_resource_type,
  platform_resource_id, created_by_platform_user_id, status,
  COALESCE(latest_version_id, 0) AS latest_version_id
FROM marketplace.marketplace_catalog_items WHERE id = ? LIMIT 1
`, id).Scan(&row)
	if result.Error != nil {
		return nil, result.Error
	}
	if row.ID == 0 {
		return nil, service.ErrCatalogItemNotFound
	}
	return catalog.RestoreItem(row)
}

func (r *CatalogConsoleRepository) GetCatalogVersion(
	ctx context.Context,
	id int64,
) (*catalog.Version, error) {
	var row catalog.VersionState
	result := r.db.WithContext(ctx).Raw(`
SELECT id, catalog_item_id, version, source_revision, content_digest, manifest,
  compatibility, validation_status, created_by_platform_user_id
FROM marketplace.marketplace_catalog_item_versions WHERE id = ? LIMIT 1
`, id).Scan(&row)
	if result.Error != nil {
		return nil, result.Error
	}
	if row.ID == 0 {
		return nil, service.ErrCatalogVersionNotFound
	}
	return catalog.RestoreVersion(row)
}

func (r *CatalogConsoleRepository) ActivateCatalogVersion(
	ctx context.Context,
	item *catalog.Item,
	version *catalog.Version,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var itemRow struct {
			ID              int64
			LatestVersionID int64
		}
		if err := tx.Raw(`
SELECT id, COALESCE(latest_version_id, 0) AS latest_version_id
FROM marketplace.marketplace_catalog_items WHERE id = ? FOR UPDATE
`, item.ID()).Scan(&itemRow).Error; err != nil {
			return err
		}
		if itemRow.ID == 0 {
			return service.ErrCatalogItemNotFound
		}
		var versionRow struct {
			ID               int64
			CatalogItemID    int64
			ValidationStatus catalog.ValidationStatus
		}
		if err := tx.Raw(`
SELECT id, catalog_item_id, validation_status
FROM marketplace.marketplace_catalog_item_versions WHERE id = ? FOR UPDATE
`, version.ID()).Scan(&versionRow).Error; err != nil {
			return err
		}
		if versionRow.ID == 0 {
			return service.ErrCatalogVersionNotFound
		}
		if versionRow.CatalogItemID != item.ID() {
			return catalog.ErrVersionItemMismatch
		}
		if versionRow.ValidationStatus == catalog.ValidationPending {
			if err := tx.Exec(`
UPDATE marketplace.marketplace_catalog_item_versions
SET validation_status = 'passed' WHERE id = ? AND validation_status = 'pending'
`, version.ID()).Error; err != nil {
				return err
			}
		} else if versionRow.ValidationStatus != catalog.ValidationPassed {
			return service.ErrCatalogVersionNotPassed
		}
		return tx.Exec(`
UPDATE marketplace.marketplace_catalog_items
SET status = 'active', latest_version_id = ?, revision = revision + 1, updated_at = NOW()
WHERE id = ?
`, version.ID(), item.ID()).Error
	})
}
