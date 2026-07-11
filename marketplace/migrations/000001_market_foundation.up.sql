CREATE SCHEMA IF NOT EXISTS marketplace;

CREATE TABLE marketplace.marketplaces (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(100) NOT NULL UNIQUE
        CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND char_length(slug) BETWEEN 2 AND 100),
    name VARCHAR(120) NOT NULL,
    summary VARCHAR(240) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status VARCHAR(24) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'configuring', 'review', 'published', 'suspended', 'archived')),
    visibility VARCHAR(16) NOT NULL DEFAULT 'private'
        CHECK (visibility IN ('public', 'private')),
    template_key VARCHAR(50) NOT NULL DEFAULT 'blank'
        CHECK (template_key IN ('blank', 'cross-border-commerce', 'higher-education', 'enterprise')),
    default_locale VARCHAR(16) NOT NULL DEFAULT 'zh-CN',
    registration_mode VARCHAR(16) NOT NULL DEFAULT 'invite'
        CHECK (registration_mode IN ('public', 'invite', 'sso')),
    owner_platform_org_id BIGINT NOT NULL,
    default_quota_plan_id BIGINT,
    created_by_platform_user_id BIGINT NOT NULL,
    published_at TIMESTAMPTZ,
    suspended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE marketplace.marketplace_domains (
    id BIGSERIAL PRIMARY KEY,
    marketplace_id BIGINT NOT NULL REFERENCES marketplace.marketplaces(id),
    host VARCHAR(253) NOT NULL UNIQUE CHECK (host = lower(host)),
    kind VARCHAR(16) NOT NULL CHECK (kind IN ('platform', 'custom')),
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'verifying', 'active', 'failed', 'disabled')),
    verification_token VARCHAR(100) NOT NULL,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at TIMESTAMPTZ,
    last_error_code VARCHAR(80),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_marketplace_domains_primary
    ON marketplace.marketplace_domains (marketplace_id) WHERE is_primary;

CREATE TABLE marketplace.marketplace_spaces (
    id BIGSERIAL PRIMARY KEY,
    marketplace_id BIGINT NOT NULL REFERENCES marketplace.marketplaces(id),
    slug VARCHAR(100) NOT NULL
        CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND char_length(slug) BETWEEN 2 AND 100),
    name VARCHAR(80) NOT NULL,
    summary VARCHAR(240) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    icon_asset_key VARCHAR(500),
    status VARCHAR(16) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'published', 'hidden', 'archived')),
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_by_platform_user_id BIGINT NOT NULL,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (marketplace_id, slug),
    UNIQUE (marketplace_id, id)
);

CREATE TABLE marketplace.marketplace_publishers (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(100) NOT NULL UNIQUE
        CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND char_length(slug) BETWEEN 2 AND 100),
    publisher_type VARCHAR(16) NOT NULL
        CHECK (publisher_type IN ('user', 'organization', 'platform')),
    platform_user_id BIGINT,
    platform_org_id BIGINT,
    display_name VARCHAR(120) NOT NULL,
    summary VARCHAR(240),
    logo_asset_key VARCHAR(500),
    verification_status VARCHAR(20) NOT NULL DEFAULT 'unverified'
        CHECK (verification_status IN ('unverified', 'pending', 'verified', 'revoked')),
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (
        (publisher_type = 'user' AND platform_user_id IS NOT NULL AND platform_org_id IS NULL)
        OR (publisher_type = 'organization' AND platform_user_id IS NULL AND platform_org_id IS NOT NULL)
        OR (publisher_type = 'platform' AND platform_user_id IS NULL AND platform_org_id IS NULL)
    )
);
