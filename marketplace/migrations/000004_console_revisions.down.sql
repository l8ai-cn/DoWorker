ALTER TABLE marketplace.marketplace_listings
    DROP COLUMN IF EXISTS revision;

ALTER TABLE marketplace.marketplace_catalog_items
    DROP COLUMN IF EXISTS revision;

ALTER TABLE marketplace.marketplace_spaces
    DROP COLUMN IF EXISTS revision;

ALTER TABLE marketplace.marketplaces
    DROP COLUMN IF EXISTS revision;
