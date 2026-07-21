package postgres

import (
	"errors"

	"github.com/l8ai-cn/agentcloud/marketplace/internal/service"
	"gorm.io/gorm"
)

func insertListingTaxonomyTags(
	tx *gorm.DB,
	marketplaceID, listingID, listingVersionID int64,
	tags []service.ListingTaxonomyTag,
) error {
	for _, tag := range tags {
		id, err := upsertTaxonomyTag(tx, marketplaceID, tag)
		if err != nil {
			return err
		}
		if err := tx.Exec(`
INSERT INTO marketplace.marketplace_listing_version_tags
  (marketplace_id, listing_id, listing_version_id, taxonomy_tag_id)
VALUES (?, ?, ?, ?)
`, marketplaceID, listingID, listingVersionID, id).Error; err != nil {
			return err
		}
	}
	return nil
}

func upsertTaxonomyTag(
	tx *gorm.DB,
	marketplaceID int64,
	tag service.ListingTaxonomyTag,
) (int64, error) {
	var row struct{ ID int64 }
	result := tx.Raw(`
INSERT INTO marketplace.marketplace_taxonomy_tags
  (marketplace_id, slug, display_name, kind)
VALUES (?, ?, ?, ?)
ON CONFLICT (marketplace_id, slug) DO UPDATE
SET display_name = EXCLUDED.display_name
WHERE marketplace.marketplace_taxonomy_tags.kind = EXCLUDED.kind
RETURNING id
`, marketplaceID, tag.Slug, tag.DisplayName, tag.Kind).Scan(&row)
	if result.Error != nil {
		return 0, result.Error
	}
	if row.ID == 0 {
		return 0, errors.New("taxonomy slug is already used by another kind")
	}
	return row.ID, nil
}
