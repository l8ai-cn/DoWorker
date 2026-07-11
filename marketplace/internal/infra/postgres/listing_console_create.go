package postgres

import (
	"context"
	"errors"

	"github.com/anthropics/agentsmesh/marketplace/internal/domain/listing"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

func (r *ListingConsoleRepository) CreateListingDraft(
	ctx context.Context,
	marketSlug string,
	item *listing.Listing,
	version *listing.Version,
	spaceSlugs []string,
	primarySpaceSlug string,
) error {
	uniqueSpaces := uniqueStrings(spaceSlugs)
	if len(uniqueSpaces) == 0 || !containsString(uniqueSpaces, primarySpaceSlug) {
		return listing.ErrPrimarySpaceRequired
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockListingMarket(tx, marketSlug, item.MarketplaceID); err != nil {
			return err
		}
		if err := lockPassedCatalogVersion(
			tx,
			item.CatalogItemID,
			version.CatalogItemVersionID(),
		); err != nil {
			return err
		}
		spaces, err := lockPublishedSpaces(tx, item.MarketplaceID, uniqueSpaces)
		if err != nil {
			return err
		}
		if len(spaces) != len(uniqueSpaces) {
			return errors.New("listing spaces must be published in the target market")
		}
		var listingRow struct{ ID int64 }
		if err := tx.Raw(`
INSERT INTO marketplace.marketplace_listings
  (marketplace_id, catalog_item_id, slug, status, visibility, access_mode, revision)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING id
`, item.MarketplaceID, item.CatalogItemID, item.Slug().String(), item.Status(),
			item.Visibility(), item.AccessMode(), item.Revision()).Scan(&listingRow).Error; err != nil {
			return err
		}
		item.ID = listingRow.ID
		if err := version.BindListingID(item.ID); err != nil {
			return err
		}
		var versionRow struct{ ID int64 }
		if err := tx.Raw(`
INSERT INTO marketplace.marketplace_listing_versions
  (listing_id, catalog_item_id, catalog_item_version_id, revision, display_name,
   tagline, description, outcomes, use_cases, target_audience, requirements,
   tags, release_notes, review_status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?::jsonb, ?::jsonb, ?::jsonb, ?::jsonb, ?, ?, ?)
RETURNING id
`, item.ID, item.CatalogItemID, version.CatalogItemVersionID(), version.Revision(),
			version.DisplayName(), version.Tagline(), version.Description(),
			string(version.Outcomes()), string(version.UseCases()),
			string(version.TargetAudience()), string(version.Requirements()),
			pq.Array(version.Tags()), version.ReleaseNotes(), version.ReviewStatus()).
			Scan(&versionRow).Error; err != nil {
			return err
		}
		version.AssignID(versionRow.ID)
		for _, space := range spaces {
			if err := tx.Exec(`
INSERT INTO marketplace.marketplace_listing_spaces
  (marketplace_id, listing_id, space_id, is_primary)
VALUES (?, ?, ?, ?)
`, item.MarketplaceID, item.ID, space.ID, space.Slug == primarySpaceSlug).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

type lockedSpace struct {
	ID   int64
	Slug string
}

func lockPublishedSpaces(
	tx *gorm.DB,
	marketplaceID int64,
	slugs []string,
) ([]lockedSpace, error) {
	var rows []lockedSpace
	err := tx.Raw(`
SELECT id, slug
FROM marketplace.marketplace_spaces
WHERE marketplace_id = ? AND status = 'published' AND slug = ANY(?)
ORDER BY id
FOR UPDATE
`, marketplaceID, pq.Array(slugs)).Scan(&rows).Error
	return rows, err
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
