package postgres

import (
	"github.com/anthropics/agentsmesh/marketplace/internal/service"
	"gorm.io/gorm"
)

func lockListingMarket(tx *gorm.DB, slug string, expectedID int64) error {
	var row struct {
		ID     int64
		Status string
	}
	if err := tx.Raw(`
SELECT id, status FROM marketplace.marketplaces WHERE slug = ? FOR UPDATE
`, slug).Scan(&row).Error; err != nil {
		return err
	}
	if row.ID == 0 || row.ID != expectedID {
		return service.ErrMarketNotFound
	}
	if row.Status != "configuring" && row.Status != "review" && row.Status != "published" {
		return service.ErrMarketNotConfigurable
	}
	return nil
}

func lockPassedCatalogVersion(
	tx *gorm.DB,
	catalogItemID int64,
	catalogItemVersionID int64,
) error {
	var row struct {
		ItemID           int64
		VersionID        int64
		ItemStatus       string
		ValidationStatus string
	}
	if err := tx.Raw(`
SELECT ci.id AS item_id, civ.id AS version_id, ci.status AS item_status,
  civ.validation_status
FROM marketplace.marketplace_catalog_items ci
JOIN marketplace.marketplace_catalog_item_versions civ
  ON civ.catalog_item_id = ci.id
WHERE ci.id = ? AND civ.id = ?
FOR UPDATE OF ci, civ
`, catalogItemID, catalogItemVersionID).Scan(&row).Error; err != nil {
		return err
	}
	if row.ItemID == 0 || row.VersionID == 0 ||
		row.ItemStatus != "active" || row.ValidationStatus != "passed" {
		return service.ErrCatalogVersionNotPassed
	}
	return nil
}
