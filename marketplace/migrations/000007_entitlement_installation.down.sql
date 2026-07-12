ALTER TABLE marketplace.marketplace_installations
    DROP CONSTRAINT IF EXISTS fk_marketplace_installations_current_operation;

DROP TABLE IF EXISTS marketplace.marketplace_installation_operations;
DROP TABLE IF EXISTS marketplace.marketplace_installations;
DROP TABLE IF EXISTS marketplace.marketplace_entitlements;
