ALTER TABLE marketplace.marketplace_listings
    DROP CONSTRAINT IF EXISTS fk_marketplace_listing_current_version;

ALTER TABLE marketplace.marketplace_catalog_items
    DROP CONSTRAINT IF EXISTS fk_marketplace_catalog_latest_version;

DROP TABLE IF EXISTS marketplace.marketplace_listing_spaces;
DROP TABLE IF EXISTS marketplace.marketplace_listing_versions;
DROP TABLE IF EXISTS marketplace.marketplace_listings;
DROP FUNCTION IF EXISTS marketplace.validate_listing_publication();
DROP FUNCTION IF EXISTS marketplace.assert_listing_is_publishable(BIGINT);
DROP TABLE IF EXISTS marketplace.marketplace_catalog_item_versions;
DROP TABLE IF EXISTS marketplace.marketplace_catalog_items;
