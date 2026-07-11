ALTER TABLE marketplace.marketplace_listing_versions
    ADD COLUMN catalog_item_id BIGINT;

UPDATE marketplace.marketplace_listing_versions lv
SET catalog_item_id = l.catalog_item_id
FROM marketplace.marketplace_listings l
WHERE l.id = lv.listing_id;

ALTER TABLE marketplace.marketplace_listing_versions
    ALTER COLUMN catalog_item_id SET NOT NULL;

ALTER TABLE marketplace.marketplace_catalog_item_versions
    ADD CONSTRAINT uq_marketplace_catalog_versions_id_item
    UNIQUE (id, catalog_item_id);

ALTER TABLE marketplace.marketplace_listings
    ADD CONSTRAINT uq_marketplace_listings_id_item
    UNIQUE (id, catalog_item_id);

ALTER TABLE marketplace.marketplace_listing_versions
    ADD CONSTRAINT fk_marketplace_listing_versions_listing_item
    FOREIGN KEY (listing_id, catalog_item_id)
    REFERENCES marketplace.marketplace_listings(id, catalog_item_id);

ALTER TABLE marketplace.marketplace_listing_versions
    ADD CONSTRAINT fk_marketplace_listing_versions_catalog_item
    FOREIGN KEY (catalog_item_version_id, catalog_item_id)
    REFERENCES marketplace.marketplace_catalog_item_versions(id, catalog_item_id);
