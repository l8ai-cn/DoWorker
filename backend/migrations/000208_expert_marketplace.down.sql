DROP INDEX IF EXISTS idx_experts_org_market_application;

ALTER TABLE experts
  DROP CONSTRAINT IF EXISTS experts_market_release_fkey,
  DROP CONSTRAINT IF EXISTS experts_market_source_pair_check,
  DROP COLUMN IF EXISTS source_market_release_id,
  DROP COLUMN IF EXISTS source_market_application_id;

ALTER TABLE expert_market_applications
  DROP CONSTRAINT IF EXISTS expert_market_applications_latest_release_fkey;

DROP TRIGGER IF EXISTS expert_market_releases_immutable
  ON expert_market_releases;
DROP FUNCTION IF EXISTS prevent_expert_market_release_immutable_update;
DROP TRIGGER IF EXISTS expert_market_releases_validate_source
  ON expert_market_releases;
DROP FUNCTION IF EXISTS validate_expert_market_release_source;

DROP TABLE IF EXISTS expert_market_releases;
DROP TABLE IF EXISTS expert_market_applications;
