DROP INDEX IF EXISTS marketplace.idx_marketplace_listings_featured;
ALTER TABLE marketplace.marketplace_listings DROP COLUMN IF EXISTS featured_rank;
DROP TABLE IF EXISTS marketplace.marketplace_listing_version_tags;
DROP TABLE IF EXISTS marketplace.marketplace_taxonomy_tags;
