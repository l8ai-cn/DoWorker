UPDATE marketplace.marketplace_catalog_item_versions civ
SET source_revision = 'backend-expert-template-v1',
    content_digest = '59b22b8e258e1055aa6eb8e61b734f1bcd3e59953458f5acd2aaa22e85cd8595',
    manifest = jsonb_set(
        civ.manifest,
        '{runtime_snapshot}',
        '{"market_application_slug":"software-delivery-expert"}'::jsonb,
        true
    )
FROM marketplace.marketplace_catalog_items ci
WHERE ci.id = civ.catalog_item_id
  AND ci.slug = 'software-delivery-expert'
  AND ci.platform_resource_type = 'expert'
  AND civ.version = '1.0.0';
