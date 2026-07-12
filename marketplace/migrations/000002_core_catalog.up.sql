CREATE TABLE marketplace.marketplace_catalog_items (
    id BIGSERIAL PRIMARY KEY,
    publisher_id BIGINT NOT NULL REFERENCES marketplace.marketplace_publishers(id),
    slug VARCHAR(100) NOT NULL
        CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND char_length(slug) BETWEEN 2 AND 100),
    resource_type VARCHAR(20) NOT NULL
        CHECK (resource_type IN ('application', 'skill', 'mcp_connector', 'resource')),
    name VARCHAR(120) NOT NULL,
    summary VARCHAR(240) NOT NULL,
    platform_resource_type VARCHAR(40) NOT NULL,
    platform_resource_id BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'active', 'deprecated', 'blocked')),
    latest_version_id BIGINT,
    created_by_platform_user_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (publisher_id, slug),
    UNIQUE (platform_resource_type, platform_resource_id)
);

CREATE TABLE marketplace.marketplace_catalog_item_versions (
    id BIGSERIAL PRIMARY KEY,
    catalog_item_id BIGINT NOT NULL REFERENCES marketplace.marketplace_catalog_items(id),
    version VARCHAR(50) NOT NULL,
    source_revision VARCHAR(100) NOT NULL,
    content_digest CHAR(64) NOT NULL,
    manifest JSONB NOT NULL,
    permissions JSONB NOT NULL DEFAULT '[]',
    compatibility JSONB NOT NULL DEFAULT '{}',
    dependency_lock JSONB NOT NULL DEFAULT '{}',
    artifact_key VARCHAR(500),
    validation_status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (validation_status IN ('pending', 'passed', 'failed', 'deprecated')),
    created_by_platform_user_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (catalog_item_id, version),
    UNIQUE (catalog_item_id, content_digest)
);

ALTER TABLE marketplace.marketplace_catalog_items
    ADD CONSTRAINT fk_marketplace_catalog_latest_version
    FOREIGN KEY (latest_version_id) REFERENCES marketplace.marketplace_catalog_item_versions(id);

CREATE TABLE marketplace.marketplace_listings (
    id BIGSERIAL PRIMARY KEY,
    marketplace_id BIGINT NOT NULL REFERENCES marketplace.marketplaces(id),
    catalog_item_id BIGINT NOT NULL REFERENCES marketplace.marketplace_catalog_items(id),
    slug VARCHAR(100) NOT NULL
        CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND char_length(slug) BETWEEN 2 AND 100),
    status VARCHAR(24) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'submitted', 'validating', 'needs_changes', 'approved',
            'published', 'suspended', 'deprecated', 'removed')),
    visibility VARCHAR(16) NOT NULL DEFAULT 'hidden'
        CHECK (visibility IN ('public', 'members', 'hidden')),
    access_mode VARCHAR(16) NOT NULL DEFAULT 'direct'
        CHECK (access_mode IN ('direct', 'approval', 'grant_only')),
    current_version_id BIGINT,
    submitted_by_platform_user_id BIGINT,
    published_at TIMESTAMPTZ,
    suspended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (marketplace_id, slug),
    UNIQUE (marketplace_id, catalog_item_id),
    UNIQUE (marketplace_id, id),
    CHECK (status <> 'published' OR (current_version_id IS NOT NULL
        AND published_at IS NOT NULL))
);

CREATE INDEX idx_marketplace_listings_public
    ON marketplace.marketplace_listings (marketplace_id, published_at DESC, id DESC)
    WHERE status = 'published' AND visibility = 'public';

CREATE TABLE marketplace.marketplace_listing_versions (
    id BIGSERIAL PRIMARY KEY,
    listing_id BIGINT NOT NULL REFERENCES marketplace.marketplace_listings(id),
    catalog_item_version_id BIGINT NOT NULL REFERENCES marketplace.marketplace_catalog_item_versions(id),
    revision INTEGER NOT NULL CHECK (revision > 0),
    display_name VARCHAR(120) NOT NULL,
    tagline VARCHAR(160) NOT NULL,
    description TEXT NOT NULL,
    outcomes JSONB NOT NULL DEFAULT '[]',
    use_cases JSONB NOT NULL DEFAULT '[]',
    target_audience JSONB NOT NULL DEFAULT '[]',
    requirements JSONB NOT NULL DEFAULT '[]',
    tags TEXT[] NOT NULL DEFAULT '{}',
    quota_plan_id BIGINT,
    release_notes TEXT NOT NULL DEFAULT '',
    review_status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (review_status IN ('draft', 'submitted', 'approved', 'rejected')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (listing_id, revision),
    UNIQUE (listing_id, id)
);

ALTER TABLE marketplace.marketplace_listings
    ADD CONSTRAINT fk_marketplace_listing_current_version
    FOREIGN KEY (id, current_version_id)
    REFERENCES marketplace.marketplace_listing_versions(listing_id, id);

CREATE TABLE marketplace.marketplace_listing_spaces (
    marketplace_id BIGINT NOT NULL,
    listing_id BIGINT NOT NULL,
    space_id BIGINT NOT NULL,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (listing_id, space_id),
    FOREIGN KEY (marketplace_id, listing_id)
        REFERENCES marketplace.marketplace_listings(marketplace_id, id),
    FOREIGN KEY (marketplace_id, space_id)
        REFERENCES marketplace.marketplace_spaces(marketplace_id, id)
);

CREATE UNIQUE INDEX idx_marketplace_listing_spaces_primary
    ON marketplace.marketplace_listing_spaces (listing_id) WHERE is_primary;

CREATE FUNCTION marketplace.assert_listing_is_publishable(target_listing_id BIGINT) RETURNS VOID AS $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM marketplace.marketplace_listings
        WHERE id = target_listing_id AND status = 'published'
    ) AND NOT EXISTS (
        SELECT 1 FROM marketplace.marketplace_listing_spaces
        WHERE listing_id = target_listing_id AND is_primary
    ) THEN RAISE EXCEPTION 'published listing requires a primary space';
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE FUNCTION marketplace.validate_listing_publication() RETURNS TRIGGER AS $$
BEGIN
    IF TG_TABLE_NAME = 'marketplace_listings' THEN
        PERFORM marketplace.assert_listing_is_publishable(NEW.id);
    ELSIF TG_OP = 'DELETE' THEN
        PERFORM marketplace.assert_listing_is_publishable(OLD.listing_id);
    ELSE
        PERFORM marketplace.assert_listing_is_publishable(NEW.listing_id);
        IF OLD.listing_id IS DISTINCT FROM NEW.listing_id THEN
            PERFORM marketplace.assert_listing_is_publishable(OLD.listing_id);
        END IF;
    END IF;
    IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER marketplace_listing_publication_guard
AFTER INSERT OR UPDATE OF status, current_version_id, published_at
ON marketplace.marketplace_listings DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION marketplace.validate_listing_publication();

CREATE CONSTRAINT TRIGGER marketplace_listing_space_publication_guard
AFTER DELETE OR UPDATE OF listing_id, is_primary
ON marketplace.marketplace_listing_spaces DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION marketplace.validate_listing_publication();
