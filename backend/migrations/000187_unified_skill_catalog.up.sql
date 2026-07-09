-- Unified skill catalog: one implementation for skill sourcing.
--
-- authored_skills becomes the single `skills` catalog (git-backed, one
-- internal repo per skill). External import (single repo or collection)
-- fans out into per-skill rows carrying upstream provenance. The
-- registry-based import pipeline (skill_registries + skill_market_items +
-- skill_registry_overrides) is retired; installed_skills now references the
-- catalog directly via skill_id.

ALTER TABLE authored_skills RENAME TO skills;
ALTER INDEX idx_authored_skills_org_slug RENAME TO idx_skills_org_slug;
ALTER INDEX idx_authored_skills_org_updated RENAME TO idx_skills_org_updated;

ALTER TABLE skills
    ALTER COLUMN organization_id DROP NOT NULL,
    ALTER COLUMN created_by_id DROP NOT NULL,
    ADD COLUMN category VARCHAR(50) NOT NULL DEFAULT '',
    ADD COLUMN compatibility VARCHAR(500) NOT NULL DEFAULT '',
    ADD COLUMN allowed_tools TEXT NOT NULL DEFAULT '',
    ADD COLUMN agent_filter JSONB NOT NULL DEFAULT '[]',
    ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN upstream_url VARCHAR(500) NOT NULL DEFAULT '',
    ADD COLUMN upstream_subdir VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN upstream_commit_sha VARCHAR(40) NOT NULL DEFAULT '';

-- organization_id is now nullable (NULL = platform-level); the plain unique
-- index would treat NULLs as distinct, so key on COALESCE instead.
DROP INDEX idx_skills_org_slug;
CREATE UNIQUE INDEX idx_skills_org_slug ON skills ((COALESCE(organization_id, 0)), slug);

COMMENT ON TABLE skills IS 'Unified skill catalog. Git is the source of truth (one am-skills repo per skill); rows index the packaged artifact (content_sha/storage_key) plus upstream provenance for imported skills.';
COMMENT ON COLUMN skills.upstream_url IS 'External git repo this skill was imported from (empty for platform-authored skills).';
COMMENT ON COLUMN skills.upstream_subdir IS 'Subdirectory inside upstream_url holding this skill (empty when the repo root is the skill).';

ALTER TABLE installed_skills
    ADD COLUMN skill_id BIGINT REFERENCES skills(id) ON DELETE SET NULL;
CREATE INDEX idx_installed_skills_skill ON installed_skills(skill_id);

-- Retire the registry pipeline. Existing market installs keep functioning:
-- their content_sha/storage_key copies stay valid; only live-follow is lost.
ALTER TABLE installed_skills DROP COLUMN IF EXISTS market_item_id;
DROP TABLE IF EXISTS skill_market_items;
DROP TABLE IF EXISTS skill_registry_overrides;
DROP TABLE IF EXISTS skill_registries;
