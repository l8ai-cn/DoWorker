CREATE TABLE marketplace.marketplace_entitlements (
    id UUID PRIMARY KEY,
    marketplace_id BIGINT NOT NULL REFERENCES marketplace.marketplaces(id),
    listing_id BIGINT NOT NULL,
    subject_type VARCHAR(16) NOT NULL CHECK (subject_type IN ('user', 'organization')),
    subject_platform_id BIGINT NOT NULL,
    target_platform_org_id BIGINT NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'suspended', 'revoked', 'expired')),
    source VARCHAR(16) NOT NULL CHECK (source IN ('direct', 'approval', 'grant')),
    source_request_id UUID,
    starts_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    granted_by_platform_user_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (marketplace_id, id),
    CHECK (expires_at IS NULL OR expires_at > starts_at),
    FOREIGN KEY (marketplace_id, listing_id)
        REFERENCES marketplace.marketplace_listings(marketplace_id, id)
);

CREATE UNIQUE INDEX idx_marketplace_entitlements_active_direct
    ON marketplace.marketplace_entitlements (
        marketplace_id, listing_id, subject_type, subject_platform_id, target_platform_org_id
    )
    WHERE source = 'direct' AND status = 'active';

CREATE TABLE marketplace.marketplace_installations (
    id UUID PRIMARY KEY,
    marketplace_id BIGINT NOT NULL REFERENCES marketplace.marketplaces(id),
    listing_id BIGINT NOT NULL,
    listing_version_id BIGINT NOT NULL,
    entitlement_id UUID NOT NULL,
    target_platform_org_id BIGINT NOT NULL,
    quota_charge_scope VARCHAR(16) NOT NULL
        CHECK (quota_charge_scope IN ('marketplace', 'organization', 'group', 'user')),
    quota_account_id UUID,
    installed_by_platform_user_id BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'planning'
        CHECK (status IN (
            'planning', 'installing', 'verifying', 'active', 'failed', 'suspended', 'uninstalled'
        )),
    runtime_ref VARCHAR(200),
    config_snapshot JSONB NOT NULL DEFAULT '{}',
    plan_digest CHAR(64) NOT NULL CHECK (plan_digest ~ '^[0-9a-f]{64}$'),
    current_operation_id UUID,
    last_verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (marketplace_id, id),
    CHECK (
        (quota_charge_scope = 'user' AND quota_account_id IS NULL)
        OR (quota_charge_scope <> 'user' AND quota_account_id IS NOT NULL)
    ),
    FOREIGN KEY (marketplace_id, listing_id)
        REFERENCES marketplace.marketplace_listings(marketplace_id, id),
    FOREIGN KEY (listing_id, listing_version_id)
        REFERENCES marketplace.marketplace_listing_versions(listing_id, id),
    FOREIGN KEY (marketplace_id, entitlement_id)
        REFERENCES marketplace.marketplace_entitlements(marketplace_id, id),
    FOREIGN KEY (marketplace_id, quota_account_id)
        REFERENCES marketplace.marketplace_quota_accounts(marketplace_id, id)
);

CREATE TABLE marketplace.marketplace_installation_operations (
    id UUID PRIMARY KEY,
    marketplace_id BIGINT NOT NULL REFERENCES marketplace.marketplaces(id),
    installation_id UUID NOT NULL,
    operation_type VARCHAR(16) NOT NULL
        CHECK (operation_type IN ('install', 'upgrade', 'suspend', 'resume', 'uninstall')),
    idempotency_key UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'planned'
        CHECK (status IN ('planned', 'running', 'succeeded', 'failed', 'compensating', 'compensated')),
    stage VARCHAR(40) NOT NULL
        CHECK (stage IN ('entitlement', 'quota', 'runtime', 'dependencies', 'create', 'verify', 'settle')),
    plan JSONB NOT NULL,
    result JSONB,
    error_code VARCHAR(80),
    error_message VARCHAR(500),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_marketplace_installation_operations_idempotency UNIQUE (idempotency_key),
    UNIQUE (marketplace_id, id),
    UNIQUE (installation_id, id),
    FOREIGN KEY (marketplace_id, installation_id)
        REFERENCES marketplace.marketplace_installations(marketplace_id, id)
);

ALTER TABLE marketplace.marketplace_installations
    ADD CONSTRAINT fk_marketplace_installations_current_operation
    FOREIGN KEY (id, current_operation_id)
    REFERENCES marketplace.marketplace_installation_operations(installation_id, id);
