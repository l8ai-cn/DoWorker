CREATE TABLE marketplace.marketplace_taxonomy_tags (
    id BIGSERIAL PRIMARY KEY,
    marketplace_id BIGINT NOT NULL REFERENCES marketplace.marketplaces(id),
    slug VARCHAR(100) NOT NULL
        CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND char_length(slug) BETWEEN 2 AND 100),
    display_name TEXT NOT NULL CHECK (char_length(display_name) > 0),
    kind VARCHAR(20) NOT NULL
        CHECK (kind IN ('scene', 'industry', 'audience', 'capability', 'integration', 'readiness')),
    parent_tag_id BIGINT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (marketplace_id, slug),
    UNIQUE (marketplace_id, id),
    CHECK (parent_tag_id IS NULL OR parent_tag_id <> id),
    FOREIGN KEY (marketplace_id, parent_tag_id)
        REFERENCES marketplace.marketplace_taxonomy_tags(marketplace_id, id)
);

CREATE TABLE marketplace.marketplace_listing_version_tags (
    marketplace_id BIGINT NOT NULL,
    listing_id BIGINT NOT NULL,
    listing_version_id BIGINT NOT NULL,
    taxonomy_tag_id BIGINT NOT NULL,
    PRIMARY KEY (listing_version_id, taxonomy_tag_id),
    FOREIGN KEY (marketplace_id, listing_id)
        REFERENCES marketplace.marketplace_listings(marketplace_id, id),
    FOREIGN KEY (listing_id, listing_version_id)
        REFERENCES marketplace.marketplace_listing_versions(listing_id, id),
    FOREIGN KEY (marketplace_id, taxonomy_tag_id)
        REFERENCES marketplace.marketplace_taxonomy_tags(marketplace_id, id)
);

CREATE INDEX idx_marketplace_taxonomy_tags_filter
    ON marketplace.marketplace_taxonomy_tags (marketplace_id, kind, slug);
CREATE INDEX idx_marketplace_listing_version_tags_filter
    ON marketplace.marketplace_listing_version_tags (marketplace_id, taxonomy_tag_id, listing_version_id);

WITH legacy_tags AS (
    SELECT l.marketplace_id, lv.listing_id, lv.id AS listing_version_id, BTRIM(tag) AS display_name
    FROM marketplace.marketplace_listing_versions lv
    JOIN marketplace.marketplace_listings l ON l.id = lv.listing_id
    CROSS JOIN LATERAL UNNEST(lv.tags) AS tag
    WHERE BTRIM(tag) <> ''
)
INSERT INTO marketplace.marketplace_taxonomy_tags
    (marketplace_id, slug, display_name, kind)
SELECT DISTINCT marketplace_id, 'legacy-' || MD5(display_name), display_name, 'capability'
FROM legacy_tags
ON CONFLICT (marketplace_id, slug) DO NOTHING;

WITH legacy_tags AS (
    SELECT l.marketplace_id, lv.listing_id, lv.id AS listing_version_id, BTRIM(tag) AS display_name
    FROM marketplace.marketplace_listing_versions lv
    JOIN marketplace.marketplace_listings l ON l.id = lv.listing_id
    CROSS JOIN LATERAL UNNEST(lv.tags) AS tag
    WHERE BTRIM(tag) <> ''
)
INSERT INTO marketplace.marketplace_listing_version_tags
    (marketplace_id, listing_id, listing_version_id, taxonomy_tag_id)
SELECT legacy_tags.marketplace_id, legacy_tags.listing_id, legacy_tags.listing_version_id, tags.id
FROM legacy_tags
JOIN marketplace.marketplace_taxonomy_tags tags
  ON tags.marketplace_id = legacy_tags.marketplace_id
 AND tags.slug = 'legacy-' || MD5(legacy_tags.display_name)
ON CONFLICT DO NOTHING;

CREATE OR REPLACE FUNCTION marketplace.prevent_submitted_listing_version_update() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        IF OLD.review_status <> 'draft' THEN
            RAISE EXCEPTION 'submitted listing versions are immutable';
        END IF;
        RETURN OLD;
    END IF;
    IF (OLD.review_status <> 'draft' OR NEW.review_status <> 'draft') AND
       (NEW.listing_id, NEW.catalog_item_id, NEW.catalog_item_version_id, NEW.revision,
        NEW.display_name, NEW.tagline, NEW.description, NEW.outcomes, NEW.use_cases,
        NEW.target_audience, NEW.requirements, NEW.quota_plan_id, NEW.release_notes, NEW.created_at)
       IS DISTINCT FROM
       (OLD.listing_id, OLD.catalog_item_id, OLD.catalog_item_version_id, OLD.revision,
        OLD.display_name, OLD.tagline, OLD.description, OLD.outcomes, OLD.use_cases,
        OLD.target_audience, OLD.requirements, OLD.quota_plan_id, OLD.release_notes, OLD.created_at) THEN
        RAISE EXCEPTION 'submitted listing version payload is immutable';
    END IF;
    IF NOT (
        NEW.review_status = OLD.review_status
        OR (OLD.review_status = 'draft' AND NEW.review_status = 'submitted')
        OR (OLD.review_status = 'submitted' AND NEW.review_status IN ('approved', 'rejected'))
    ) THEN
        RAISE EXCEPTION 'invalid listing review transition';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

ALTER TABLE marketplace.marketplace_listing_versions DROP COLUMN tags;
ALTER TABLE marketplace.marketplace_listings
    ADD COLUMN featured_rank INTEGER NOT NULL DEFAULT 0 CHECK (featured_rank >= 0);
CREATE INDEX idx_marketplace_listings_featured
    ON marketplace.marketplace_listings (marketplace_id, featured_rank DESC, published_at DESC, id DESC)
    WHERE status = 'published' AND visibility = 'public';
