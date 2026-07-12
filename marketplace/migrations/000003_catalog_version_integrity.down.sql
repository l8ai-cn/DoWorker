ALTER TABLE marketplace.marketplace_listing_versions
    DROP CONSTRAINT IF EXISTS fk_marketplace_listing_versions_catalog_item;

ALTER TABLE marketplace.marketplace_listing_versions
    DROP CONSTRAINT IF EXISTS fk_marketplace_listing_versions_listing_item;

ALTER TABLE marketplace.marketplace_listings
    DROP CONSTRAINT IF EXISTS uq_marketplace_listings_id_item;

ALTER TABLE marketplace.marketplace_catalog_item_versions
    DROP CONSTRAINT IF EXISTS uq_marketplace_catalog_versions_id_item;

ALTER TABLE marketplace.marketplace_listing_versions
    DROP COLUMN IF EXISTS catalog_item_id;
