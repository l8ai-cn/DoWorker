DO $$
DECLARE
    catalog_id BIGINT;
    source_catalog_version_id BIGINT;
    target_listing_id BIGINT;
    source_listing_version_id BIGINT;
BEGIN
    SELECT ci.id, civ.id
    INTO catalog_id, source_catalog_version_id
    FROM marketplace.marketplace_catalog_items ci
    JOIN marketplace.marketplace_catalog_item_versions civ
      ON civ.catalog_item_id = ci.id
    WHERE ci.slug = 'software-delivery-expert'
      AND ci.platform_resource_type = 'expert'
      AND civ.version = '1.0.0';

    SELECT l.id, lv.id
    INTO target_listing_id, source_listing_version_id
    FROM marketplace.marketplace_listings l
    JOIN marketplace.marketplace_listing_versions lv
      ON lv.listing_id = l.id
    WHERE l.catalog_item_id = catalog_id
      AND lv.revision = 1;

    UPDATE marketplace.marketplace_listings
    SET current_version_id = source_listing_version_id,
        revision = revision + 1,
        updated_at = NOW()
    WHERE id = target_listing_id
      AND current_version_id IS DISTINCT FROM source_listing_version_id;

    UPDATE marketplace.marketplace_catalog_items
    SET latest_version_id = source_catalog_version_id,
        revision = revision + 1,
        updated_at = NOW()
    WHERE id = catalog_id
      AND latest_version_id IS DISTINCT FROM source_catalog_version_id;
END
$$;

DROP TRIGGER IF EXISTS marketplace_expert_runtime_compatibility_guard
    ON marketplace.marketplace_listings;
DROP FUNCTION IF EXISTS marketplace.validate_expert_runtime_compatibility();
