DO $$
DECLARE
    market_id BIGINT;
    space_id BIGINT;
    publisher_id BIGINT;
    catalog_id BIGINT;
    catalog_version_id BIGINT;
    quota_plan_id BIGINT;
    listing_id BIGINT;
    listing_version_id BIGINT;
BEGIN
    INSERT INTO marketplace.marketplaces
        (slug, name, summary, description, status, visibility, template_key,
         registration_mode, owner_platform_org_id, created_by_platform_user_id,
         published_at)
    VALUES
        ('do-worker-market', 'Do Worker 专家应用市场',
         '面向真实工作的开箱即用 AI 专家应用',
         '汇集经过验证的专家应用、Skill、系统连接与资源。',
         'published', 'public', 'enterprise', 'public', 1, 1, NOW())
    RETURNING id INTO market_id;

    INSERT INTO marketplace.marketplace_domains
        (marketplace_id, host, kind, status, verification_token, is_primary,
         verified_at)
    VALUES
        (market_id, 'market.l8ai.cn', 'platform', 'active',
         'platform-market-l8ai', TRUE, NOW()),
        (market_id, 'dowork.l8ai.cn', 'platform', 'active',
         'platform-market-core', FALSE, NOW()),
        (market_id, 'localhost', 'platform', 'active',
         'platform-market-local', FALSE, NOW());

    INSERT INTO marketplace.marketplace_spaces
        (marketplace_id, slug, name, summary, description, status, sort_order,
         created_by_platform_user_id, published_at)
    VALUES
        (market_id, 'software-delivery', '软件交付',
         '从需求到合并的可验证交付能力',
         '面向研发团队的开发、测试、评审和代码交付场景。',
         'published', 10, 1, NOW())
    RETURNING id INTO space_id;

    INSERT INTO marketplace.marketplace_publishers
        (slug, publisher_type, display_name, summary, verification_status,
         verified_at)
    VALUES
        ('do-worker', 'platform', 'Do Worker',
         'Do Worker 平台维护的专家应用', 'verified', NOW())
    RETURNING id INTO publisher_id;

    INSERT INTO marketplace.marketplace_catalog_items
        (publisher_id, slug, resource_type, name, summary,
         platform_resource_type, platform_resource_id, status,
         created_by_platform_user_id)
    VALUES
        (publisher_id, 'software-delivery-expert', 'application',
         '软件交付专家',
         '从需求拆解、隔离开发到测试与合并，完成可验证的软件交付。',
         'expert', NULL, 'active', 1)
    RETURNING id INTO catalog_id;

    INSERT INTO marketplace.marketplace_catalog_item_versions
        (catalog_item_id, version, source_revision, content_digest, manifest,
         permissions, compatibility, dependency_lock, validation_status,
         created_by_platform_user_id)
    VALUES
        (catalog_id, '1.0.0', 'backend-expert-template-v1',
         '59b22b8e258e1055aa6eb8e61b734f1bcd3e59953458f5acd2aaa22e85cd8595',
         '{"installation_credits":"20","runtime_snapshot":{"market_application_slug":"software-delivery-expert"}}',
         '["workspace.execute","repository.write","pull_request.create"]',
         '{"agents":["codex-cli"],"locale":"zh-CN"}',
         '{"skills":["worktree","e2e","gh-merge","merge"]}',
         'passed', 1)
    RETURNING id INTO catalog_version_id;

    UPDATE marketplace.marketplace_catalog_items
    SET latest_version_id = catalog_version_id
    WHERE id = catalog_id;

    INSERT INTO marketplace.marketplace_quota_plans
        (marketplace_id, slug, name, description, period, grant_credits,
         charge_scope, status)
    VALUES
        (market_id, 'organization-starter', '组织起步额度',
         '用于专家应用启用和运行的市场额度。', 'total', 100,
         'organization', 'active')
    RETURNING id INTO quota_plan_id;

    UPDATE marketplace.marketplaces
    SET default_quota_plan_id = quota_plan_id
    WHERE id = market_id;

    INSERT INTO marketplace.marketplace_quota_accounts
        (id, marketplace_id, subject_type, subject_ref, quota_plan_id, status,
         period_start, period_end)
    VALUES
        ('d0000000-0000-4000-8000-000000000001', market_id, 'organization',
         1, quota_plan_id, 'active', TIMESTAMPTZ '2026-01-01 00:00:00+00',
         TIMESTAMPTZ '2099-01-01 00:00:00+00');

    INSERT INTO marketplace.marketplace_quota_ledger_entries
        (id, marketplace_id, quota_account_id, entry_type, available_delta,
         period_start, reason, created_by_platform_user_id)
    VALUES
        ('d0000000-0000-4000-8000-000000000002', market_id,
         'd0000000-0000-4000-8000-000000000001', 'grant', 100,
         TIMESTAMPTZ '2026-01-01 00:00:00+00', 'default_marketplace_seed', 1);

    INSERT INTO marketplace.marketplace_listings
        (marketplace_id, catalog_item_id, slug, status, visibility, access_mode)
    VALUES
        (market_id, catalog_id, 'software-delivery-expert', 'approved',
         'public', 'direct')
    RETURNING id INTO listing_id;

    INSERT INTO marketplace.marketplace_listing_versions
        (listing_id, catalog_item_id, catalog_item_version_id, revision,
         display_name, tagline, description, outcomes, use_cases,
         target_audience, requirements, tags, quota_plan_id, release_notes,
         review_status)
    VALUES
        (listing_id, catalog_id, catalog_version_id, 1, '软件交付专家',
         '把需求变成经过测试、评审并可合并的代码交付',
         '适用于功能开发、缺陷修复和版本交付。专家会建立隔离工作区，完成代码修改与测试，并通过适用的 PR 或 MR 流程交付。',
         '["隔离完成代码修改","执行关键路径验证","提交并推动代码合并"]',
         '["功能开发","缺陷修复","版本交付"]',
         '["研发团队","技术负责人","交付工程师"]',
         '["目标组织已配置可用 Runner","代码仓库允许创建分支和合并请求"]',
         ARRAY['开发','测试','代码合并'], quota_plan_id,
         '首个 Marketplace MVP 版本。', 'approved')
    RETURNING id INTO listing_version_id;

    INSERT INTO marketplace.marketplace_listing_spaces
        (marketplace_id, listing_id, space_id, is_primary, sort_order)
    VALUES (market_id, listing_id, space_id, TRUE, 10);

    UPDATE marketplace.marketplace_listings
    SET status = 'published', current_version_id = listing_version_id,
        published_at = NOW()
    WHERE id = listing_id;
END
$$;
