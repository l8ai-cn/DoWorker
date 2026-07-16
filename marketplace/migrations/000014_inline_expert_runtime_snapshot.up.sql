CREATE FUNCTION marketplace.validate_expert_runtime_compatibility() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'published' AND EXISTS (
        SELECT 1
        FROM marketplace.marketplace_listing_versions lv
        JOIN marketplace.marketplace_catalog_items ci
          ON ci.id = lv.catalog_item_id
        JOIN marketplace.marketplace_catalog_item_versions civ
          ON civ.id = lv.catalog_item_version_id
        WHERE lv.id = NEW.current_version_id
          AND lv.listing_id = NEW.id
          AND ci.platform_resource_type = 'expert'
          AND NOT (
              COALESCE(civ.compatibility->'agents'->>0, '')
                  ~ '^[a-z0-9]+(-[a-z0-9]+)*$'
              AND char_length(civ.compatibility->'agents'->>0)
                  BETWEEN 2 AND 100
          )
    ) THEN
        RAISE EXCEPTION 'published expert listing requires a valid compatible agent identifier';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER marketplace_expert_runtime_compatibility_guard
AFTER INSERT OR UPDATE OF status, current_version_id
ON marketplace.marketplace_listings
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION marketplace.validate_expert_runtime_compatibility();

DO $$
DECLARE
    catalog_id BIGINT;
    source_catalog_version_id BIGINT;
    target_catalog_version_id BIGINT;
    target_listing_id BIGINT;
    source_listing_version_id BIGINT;
    target_listing_version_id BIGINT;
BEGIN
    SELECT ci.id, civ.id
    INTO catalog_id, source_catalog_version_id
    FROM marketplace.marketplace_catalog_items ci
    JOIN marketplace.marketplace_catalog_item_versions civ
      ON civ.catalog_item_id = ci.id
    WHERE ci.slug = 'software-delivery-expert'
      AND ci.platform_resource_type = 'expert'
      AND civ.version = '1.0.0';

    INSERT INTO marketplace.marketplace_catalog_item_versions
        (catalog_item_id, version, source_revision, content_digest, manifest,
         permissions, compatibility, dependency_lock, artifact_key,
         validation_status, created_by_platform_user_id)
    SELECT catalog_id, '1.1.0', 'inline-expert-v3',
        'c6a76da2220f1a13418a43283828d2a6800a48d08e263f703ef6c472ea867e48',
        jsonb_set(
            civ.manifest,
            '{runtime_snapshot}',
            '{"version":1,"expert":{"version":1,"slug":"software-delivery-expert","name":"软件交付专家","description":"适用于功能开发、缺陷修复和版本交付。","agent_slug":"codex-cli","prompt":"负责把明确需求转化为经过测试、评审并可合并的代码交付。先理解现有代码与约束，再在隔离工作区完成最小修改、关键验证和交付。","interaction_mode":"acp","automation_level":"autonomous","perpetual":false,"used_env_bundles":[],"skill_slugs":["delivery-worktree","delivery-e2e","delivery-github-merge","delivery-gitlab-merge"],"knowledge_mounts":[],"config_overrides":{"approval_mode":"never"},"metadata":{"expert_type":"software-delivery"}},"worker_spec":{"version":1,"runtime":{"model_binding":{"resource_id":1,"resource_revision":1,"connection_id":1,"connection_revision":1,"provider_key":"openai","model_id":"market-placeholder"},"worker_type":{"slug":"codex-cli","definition_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},"image":{"id":1,"digest":"sha256:963c99fb047c0a4fed518eb9949e805fd31329a8395526fbb1fe34d8254ebea1"}},"placement":{"policy":"explicit","compute_target":{"id":1,"kind":"runner-pool"},"deployment_mode":"pooled","resource_profile":{"id":1,"resources":{"cpu_request_millicpu":200,"cpu_limit_millicpu":1000,"memory_request_bytes":268435456,"memory_limit_bytes":1073741824}}},"type_config":{"schema_version":1,"values":{"approval_mode":"never"},"secret_refs":{},"interaction_mode":"acp","automation_level":"autonomous"},"workspace":{"branch":"","skill_ids":[],"skill_packages":[],"knowledge_mounts":[],"env_bundle_ids":[],"instructions":"负责把明确需求转化为经过测试、评审并可合并的代码交付。先理解现有代码与约束，再在隔离工作区完成最小修改、关键验证和交付。","initial_task":""},"lifecycle":{"termination_policy":"manual","idle_timeout_minutes":0},"metadata":{"alias":"software-delivery-expert"}}}'::jsonb,
            true
        ),
        civ.permissions,
        civ.compatibility,
        '{"skills":["delivery-worktree","delivery-e2e","delivery-github-merge","delivery-gitlab-merge"]}'::jsonb,
        civ.artifact_key,
        'passed',
        civ.created_by_platform_user_id
    FROM marketplace.marketplace_catalog_item_versions civ
    WHERE civ.id = source_catalog_version_id
      AND NOT EXISTS (
          SELECT 1
          FROM marketplace.marketplace_catalog_item_versions existing
          WHERE existing.catalog_item_id = catalog_id
            AND existing.version = '1.1.0'
      );

    SELECT id INTO target_catalog_version_id
    FROM marketplace.marketplace_catalog_item_versions
    WHERE catalog_item_id = catalog_id AND version = '1.1.0';

    SELECT l.id, lv.id
    INTO target_listing_id, source_listing_version_id
    FROM marketplace.marketplace_listings l
    JOIN marketplace.marketplace_listing_versions lv
      ON lv.listing_id = l.id
    WHERE l.catalog_item_id = catalog_id
      AND lv.revision = 1;

    INSERT INTO marketplace.marketplace_listing_versions
        (listing_id, catalog_item_id, catalog_item_version_id, revision,
         display_name, tagline, description, outcomes, use_cases,
         target_audience, requirements, tags, quota_plan_id, release_notes,
         review_status)
    SELECT target_listing_id, catalog_id, target_catalog_version_id, 2,
        lv.display_name, lv.tagline, lv.description, lv.outcomes, lv.use_cases,
        lv.target_audience, lv.requirements, lv.tags, lv.quota_plan_id,
        '内联完整 Expert 与 WorkerSpec 快照，安装时绑定目标组织模型。',
        'approved'
    FROM marketplace.marketplace_listing_versions lv
    WHERE lv.id = source_listing_version_id
      AND NOT EXISTS (
          SELECT 1
          FROM marketplace.marketplace_listing_versions existing
          WHERE existing.listing_id = target_listing_id AND existing.revision = 2
      );

    SELECT id INTO target_listing_version_id
    FROM marketplace.marketplace_listing_versions
    WHERE listing_id = target_listing_id AND revision = 2;

    UPDATE marketplace.marketplace_catalog_items
    SET latest_version_id = target_catalog_version_id,
        revision = revision + 1,
        updated_at = NOW()
    WHERE id = catalog_id
      AND latest_version_id IS DISTINCT FROM target_catalog_version_id;

    UPDATE marketplace.marketplace_listings
    SET current_version_id = target_listing_version_id,
        revision = revision + 1,
        updated_at = NOW()
    WHERE id = target_listing_id
      AND current_version_id IS DISTINCT FROM target_listing_version_id;
END
$$;
