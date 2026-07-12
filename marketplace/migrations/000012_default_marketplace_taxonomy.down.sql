DELETE FROM marketplace.marketplace_listing_version_tags relation
USING marketplace.marketplaces market,
      marketplace.marketplace_listings listing,
      marketplace.marketplace_taxonomy_tags tag
WHERE market.slug = 'do-worker-market'
  AND listing.marketplace_id = market.id
  AND listing.slug = 'software-delivery-expert'
  AND relation.marketplace_id = market.id
  AND relation.listing_id = listing.id
  AND relation.taxonomy_tag_id = tag.id
  AND tag.slug IN (
      'software-delivery',
      'enterprise-services',
      'engineering-team',
      'runner-required'
  );

DELETE FROM marketplace.marketplace_taxonomy_tags tag
USING marketplace.marketplaces market
WHERE market.slug = 'do-worker-market'
  AND tag.marketplace_id = market.id
  AND tag.slug IN (
      'software-delivery',
      'enterprise-services',
      'engineering-team',
      'runner-required'
  )
  AND NOT EXISTS (
      SELECT 1
      FROM marketplace.marketplace_listing_version_tags relation
      WHERE relation.taxonomy_tag_id = tag.id
  );
