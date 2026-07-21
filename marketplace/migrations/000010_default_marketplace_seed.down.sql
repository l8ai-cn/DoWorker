DO $$
DECLARE
    seed_market_id BIGINT;
    seed_listing_id BIGINT;
    seed_catalog_id BIGINT;
    seed_publisher_id BIGINT;
BEGIN
    SELECT id INTO seed_market_id
    FROM marketplace.marketplaces
    WHERE slug = 'agent-cloud-market';

    IF seed_market_id IS NULL THEN
        RETURN;
    END IF;

    SELECT id, catalog_item_id INTO seed_listing_id, seed_catalog_id
    FROM marketplace.marketplace_listings
    WHERE marketplace_id = seed_market_id AND slug = 'software-delivery-expert';

    IF EXISTS (
        SELECT 1 FROM marketplace.marketplace_installations
        WHERE marketplace_id = seed_market_id AND listing_id = seed_listing_id
    ) THEN
        RAISE EXCEPTION 'default marketplace seed has active installation history';
    END IF;

    SELECT publisher_id INTO seed_publisher_id
    FROM marketplace.marketplace_catalog_items
    WHERE id = seed_catalog_id;

    ALTER TABLE marketplace.marketplace_quota_ledger_entries
        DISABLE TRIGGER marketplace_quota_ledger_immutable;
    ALTER TABLE marketplace.marketplace_listing_versions
        DISABLE TRIGGER marketplace_listing_version_immutable;
    ALTER TABLE marketplace.marketplace_catalog_item_versions
        DISABLE TRIGGER marketplace_catalog_version_immutable;

    DELETE FROM marketplace.marketplace_quota_ledger_entries
    WHERE marketplace_id = seed_market_id;
    DELETE FROM marketplace.marketplace_quota_accounts
    WHERE marketplace_id = seed_market_id;
    DELETE FROM marketplace.marketplace_listing_spaces
    WHERE marketplace_id = seed_market_id;
    UPDATE marketplace.marketplace_listings
    SET current_version_id = NULL, status = 'removed', published_at = NULL
    WHERE id = seed_listing_id;
    DELETE FROM marketplace.marketplace_listing_versions
    WHERE listing_id = seed_listing_id;
    DELETE FROM marketplace.marketplace_listings
    WHERE id = seed_listing_id;
    UPDATE marketplace.marketplace_catalog_items
    SET latest_version_id = NULL
    WHERE id = seed_catalog_id;
    DELETE FROM marketplace.marketplace_catalog_item_versions
    WHERE catalog_item_id = seed_catalog_id;
    DELETE FROM marketplace.marketplace_catalog_items
    WHERE id = seed_catalog_id;
    UPDATE marketplace.marketplaces
    SET default_quota_plan_id = NULL
    WHERE id = seed_market_id;
    DELETE FROM marketplace.marketplace_quota_plans
    WHERE marketplace_id = seed_market_id;
    DELETE FROM marketplace.marketplace_spaces
    WHERE marketplace_id = seed_market_id;
    DELETE FROM marketplace.marketplace_domains
    WHERE marketplace_id = seed_market_id;
    DELETE FROM marketplace.marketplaces
    WHERE id = seed_market_id;
    DELETE FROM marketplace.marketplace_publishers
    WHERE id = seed_publisher_id;

    ALTER TABLE marketplace.marketplace_catalog_item_versions
        ENABLE TRIGGER marketplace_catalog_version_immutable;
    ALTER TABLE marketplace.marketplace_listing_versions
        ENABLE TRIGGER marketplace_listing_version_immutable;
    ALTER TABLE marketplace.marketplace_quota_ledger_entries
        ENABLE TRIGGER marketplace_quota_ledger_immutable;
END
$$;
