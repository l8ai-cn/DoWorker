ALTER TABLE marketplace.marketplace_catalog_items
    DROP CONSTRAINT fk_marketplace_catalog_latest_version;

ALTER TABLE marketplace.marketplace_catalog_items
    ADD CONSTRAINT fk_marketplace_catalog_latest_version
    FOREIGN KEY (latest_version_id, id)
    REFERENCES marketplace.marketplace_catalog_item_versions(id, catalog_item_id);

CREATE FUNCTION marketplace.prevent_catalog_version_payload_update() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'catalog item versions are immutable';
    END IF;
    IF (NEW.catalog_item_id, NEW.version, NEW.source_revision, NEW.content_digest,
        NEW.manifest, NEW.permissions, NEW.compatibility, NEW.dependency_lock,
        NEW.artifact_key, NEW.created_by_platform_user_id, NEW.created_at)
        IS DISTINCT FROM
       (OLD.catalog_item_id, OLD.version, OLD.source_revision, OLD.content_digest,
        OLD.manifest, OLD.permissions, OLD.compatibility, OLD.dependency_lock,
        OLD.artifact_key, OLD.created_by_platform_user_id, OLD.created_at) THEN
        RAISE EXCEPTION 'catalog item version payload is immutable';
    END IF;
    IF NOT (
        NEW.validation_status = OLD.validation_status
        OR (OLD.validation_status = 'pending' AND NEW.validation_status IN ('passed', 'failed'))
    ) THEN
        RAISE EXCEPTION 'invalid catalog validation transition';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER marketplace_catalog_version_immutable
BEFORE UPDATE OR DELETE ON marketplace.marketplace_catalog_item_versions
FOR EACH ROW EXECUTE FUNCTION marketplace.prevent_catalog_version_payload_update();

CREATE FUNCTION marketplace.prevent_submitted_listing_version_update() RETURNS TRIGGER AS $$
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

CREATE TRIGGER marketplace_listing_version_immutable
BEFORE UPDATE OR DELETE ON marketplace.marketplace_listing_versions
FOR EACH ROW EXECUTE FUNCTION marketplace.prevent_submitted_listing_version_update();

CREATE OR REPLACE FUNCTION marketplace.assert_listing_is_publishable(
    target_listing_id BIGINT
) RETURNS VOID AS $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM marketplace.marketplace_listings
        WHERE id = target_listing_id AND status = 'published'
    ) AND NOT EXISTS (
        SELECT 1
        FROM marketplace.marketplace_listings l
        JOIN marketplace.marketplace_listing_versions lv
          ON lv.id = l.current_version_id AND lv.listing_id = l.id
        JOIN marketplace.marketplace_catalog_item_versions civ
          ON civ.id = lv.catalog_item_version_id
        WHERE l.id = target_listing_id
          AND lv.review_status = 'approved'
          AND civ.validation_status = 'passed'
    ) THEN
        RAISE EXCEPTION 'published listing requires approved validated versions';
    END IF;
    IF EXISTS (
        SELECT 1 FROM marketplace.marketplace_listings
        WHERE id = target_listing_id AND status = 'published'
    ) AND NOT EXISTS (
        SELECT 1
        FROM marketplace.marketplace_listing_spaces ls
        JOIN marketplace.marketplace_spaces s ON s.id = ls.space_id
        WHERE ls.listing_id = target_listing_id
          AND ls.is_primary
          AND s.status = 'published'
    ) THEN
        RAISE EXCEPTION 'published listing requires a published primary space';
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE FUNCTION marketplace.validate_space_publication() RETURNS TRIGGER AS $$
BEGIN
    PERFORM marketplace.assert_listing_is_publishable(ls.listing_id)
    FROM marketplace.marketplace_listing_spaces ls
    WHERE ls.space_id = NEW.id AND ls.is_primary;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER marketplace_space_publication_guard
AFTER UPDATE OF status ON marketplace.marketplace_spaces
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION marketplace.validate_space_publication();
