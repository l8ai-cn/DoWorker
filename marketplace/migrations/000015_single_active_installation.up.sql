DO $$
DECLARE
    conflict RECORD;
BEGIN
    SELECT marketplace_id, listing_id, target_platform_org_id,
        array_agg(id::text ORDER BY created_at, id) AS installation_ids
    INTO conflict
    FROM marketplace.marketplace_installations
    WHERE status IN ('installing', 'verifying', 'active', 'suspended')
    GROUP BY marketplace_id, listing_id, target_platform_org_id
    HAVING COUNT(*) > 1
    LIMIT 1;

    IF FOUND THEN
        RAISE EXCEPTION
            'single active installation conflict: marketplace %, listing %, organization %, installation_ids %',
            conflict.marketplace_id,
            conflict.listing_id,
            conflict.target_platform_org_id,
            conflict.installation_ids
            USING ERRCODE = '23505';
    END IF;
END;
$$;

CREATE UNIQUE INDEX idx_marketplace_installations_single_active
    ON marketplace.marketplace_installations (
        marketplace_id, listing_id, target_platform_org_id
    )
    WHERE status IN ('installing', 'verifying', 'active', 'suspended');
