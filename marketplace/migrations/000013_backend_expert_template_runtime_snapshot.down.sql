UPDATE marketplace.marketplace_catalog_item_versions civ
SET source_revision = 'inline-expert-v1',
    content_digest = 'b8d0d5d9c77698ff343a2663009261f64f98b9452aeca9c13a143ed57bf8382f',
    manifest = jsonb_set(
        civ.manifest,
        '{runtime_snapshot}',
        '{"name":"软件交付专家","description":"适用于功能开发、缺陷修复和版本交付。","agent_slug":"codex-cli","prompt":"负责把明确需求转化为经过测试、评审并可合并的代码交付。先理解现有代码与约束，再在隔离工作区完成最小修改、关键验证和交付。","interaction_mode":"pty","automation_level":"autonomous","perpetual":false,"used_env_bundles":[],"skill_slugs":["worktree","e2e","gh-merge","merge"],"knowledge_mounts":[],"config_overrides":{}}'::jsonb,
        true
    )
FROM marketplace.marketplace_catalog_items ci
WHERE ci.id = civ.catalog_item_id
  AND ci.slug = 'software-delivery-expert'
  AND ci.platform_resource_type = 'expert'
  AND civ.version = '1.0.0';
