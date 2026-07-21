WITH target_market AS (
    SELECT id
    FROM marketplace.marketplaces
    WHERE slug = 'agent-cloud-market'
)
INSERT INTO marketplace.marketplace_taxonomy_tags
    (marketplace_id, slug, display_name, kind, sort_order)
SELECT target_market.id, tag.slug, tag.display_name, tag.kind, tag.sort_order
FROM target_market
CROSS JOIN (
    VALUES
        ('software-delivery', '软件交付', 'scene', 10),
        ('enterprise-services', '企业服务', 'industry', 10),
        ('engineering-team', '研发团队', 'audience', 10),
        ('runner-required', '需要 Runner', 'readiness', 10)
) AS tag(slug, display_name, kind, sort_order)
ON CONFLICT (marketplace_id, slug) DO NOTHING;

INSERT INTO marketplace.marketplace_listing_version_tags
    (marketplace_id, listing_id, listing_version_id, taxonomy_tag_id)
SELECT m.id, l.id, lv.id, tag.id
FROM marketplace.marketplaces m
JOIN marketplace.marketplace_listings l
  ON l.marketplace_id = m.id AND l.slug = 'software-delivery-expert'
JOIN marketplace.marketplace_listing_versions lv
  ON lv.id = l.current_version_id
JOIN marketplace.marketplace_taxonomy_tags tag
  ON tag.marketplace_id = m.id
 AND tag.slug IN (
     'software-delivery',
     'enterprise-services',
     'engineering-team',
     'runner-required'
 )
WHERE m.slug = 'agent-cloud-market'
ON CONFLICT DO NOTHING;
