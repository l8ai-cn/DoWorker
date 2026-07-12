DROP TRIGGER IF EXISTS marketplace_space_publication_guard
    ON marketplace.marketplace_spaces;
DROP FUNCTION IF EXISTS marketplace.validate_space_publication();

CREATE OR REPLACE FUNCTION marketplace.assert_listing_is_publishable(
    target_listing_id BIGINT
) RETURNS VOID AS $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM marketplace.marketplace_listings
        WHERE id = target_listing_id AND status = 'published'
    ) AND NOT EXISTS (
        SELECT 1 FROM marketplace.marketplace_listing_spaces
        WHERE listing_id = target_listing_id AND is_primary
    ) THEN
        RAISE EXCEPTION 'published listing requires a primary space';
    END IF;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS marketplace_listing_version_immutable
    ON marketplace.marketplace_listing_versions;
DROP FUNCTION IF EXISTS marketplace.prevent_submitted_listing_version_update();

DROP TRIGGER IF EXISTS marketplace_catalog_version_immutable
    ON marketplace.marketplace_catalog_item_versions;
DROP FUNCTION IF EXISTS marketplace.prevent_catalog_version_payload_update();

ALTER TABLE marketplace.marketplace_catalog_items
    DROP CONSTRAINT fk_marketplace_catalog_latest_version;

ALTER TABLE marketplace.marketplace_catalog_items
    ADD CONSTRAINT fk_marketplace_catalog_latest_version
    FOREIGN KEY (latest_version_id)
    REFERENCES marketplace.marketplace_catalog_item_versions(id);
