ALTER TABLE marketplace.marketplaces
    ADD COLUMN revision BIGINT NOT NULL DEFAULT 1;

ALTER TABLE marketplace.marketplace_spaces
    ADD COLUMN revision BIGINT NOT NULL DEFAULT 1;

ALTER TABLE marketplace.marketplace_catalog_items
    ADD COLUMN revision BIGINT NOT NULL DEFAULT 1;

ALTER TABLE marketplace.marketplace_listings
    ADD COLUMN revision BIGINT NOT NULL DEFAULT 1;
