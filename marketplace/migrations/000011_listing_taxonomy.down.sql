ALTER TABLE marketplace.marketplace_listing_versions
    ADD COLUMN tags TEXT[] NOT NULL DEFAULT '{}';

UPDATE marketplace.marketplace_listing_versions lv
SET tags = COALESCE(tag_names.tags, '{}')
FROM (
    SELECT lvt.listing_version_id, ARRAY_AGG(tags.display_name ORDER BY tags.kind, tags.slug) AS tags
    FROM marketplace.marketplace_listing_version_tags lvt
    JOIN marketplace.marketplace_taxonomy_tags tags ON tags.id = lvt.taxonomy_tag_id
    GROUP BY lvt.listing_version_id
) AS tag_names
WHERE tag_names.listing_version_id = lv.id;

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
        NEW.target_audience, NEW.requirements, NEW.tags, NEW.quota_plan_id,
        NEW.release_notes, NEW.created_at)
       IS DISTINCT FROM
       (OLD.listing_id, OLD.catalog_item_id, OLD.catalog_item_version_id, OLD.revision,
        OLD.display_name, OLD.tagline, OLD.description, OLD.outcomes, OLD.use_cases,
        OLD.target_audience, OLD.requirements, OLD.tags, OLD.quota_plan_id,
        OLD.release_notes, OLD.created_at) THEN
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

DROP INDEX IF EXISTS marketplace.idx_marketplace_listings_featured;
ALTER TABLE marketplace.marketplace_listings DROP COLUMN IF EXISTS featured_rank;
DROP TABLE IF EXISTS marketplace.marketplace_listing_version_tags;
DROP TABLE IF EXISTS marketplace.marketplace_taxonomy_tags;
