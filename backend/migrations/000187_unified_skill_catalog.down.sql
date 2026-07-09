-- Recreate the retired registry pipeline tables (shape as of 000134).
CREATE TABLE IF NOT EXISTS skill_registries (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT REFERENCES organizations(id) ON DELETE CASCADE,
    repository_url VARCHAR(500) NOT NULL,
    branch VARCHAR(100) DEFAULT 'main',
    source_type VARCHAR(20) DEFAULT 'auto',
    detected_type VARCHAR(20),
    compatible_agents JSONB DEFAULT '["claude-code"]',
    auth_type VARCHAR(20) DEFAULT 'none',
    auth_credential TEXT,
    last_synced_at TIMESTAMP WITH TIME ZONE,
    last_commit_sha VARCHAR(40),
    sync_status VARCHAR(20) DEFAULT 'pending',
    sync_error TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX idx_skill_registries_org_id ON skill_registries(organization_id);
CREATE INDEX idx_skill_registries_active ON skill_registries(is_active);
CREATE UNIQUE INDEX idx_skill_registries_unique_url
    ON skill_registries(organization_id, repository_url)
    WHERE organization_id IS NOT NULL;
CREATE UNIQUE INDEX idx_skill_registries_unique_url_platform
    ON skill_registries(repository_url)
    WHERE organization_id IS NULL;

CREATE TABLE IF NOT EXISTS skill_registry_overrides (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    registry_id BIGINT NOT NULL REFERENCES skill_registries(id) ON DELETE CASCADE,
    is_disabled BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(organization_id, registry_id)
);
CREATE INDEX idx_skill_registry_overrides_org ON skill_registry_overrides(organization_id);

CREATE TABLE IF NOT EXISTS skill_market_items (
    id BIGSERIAL PRIMARY KEY,
    registry_id BIGINT NOT NULL REFERENCES skill_registries(id) ON DELETE CASCADE,
    slug VARCHAR(100) NOT NULL,
    display_name VARCHAR(100),
    description VARCHAR(1024),
    license VARCHAR(100),
    compatibility VARCHAR(500),
    allowed_tools TEXT,
    metadata JSONB DEFAULT '{}',
    category VARCHAR(50),
    content_sha VARCHAR(64) NOT NULL,
    storage_key VARCHAR(500) NOT NULL,
    package_size BIGINT,
    version INTEGER DEFAULT 1,
    agent_filter JSONB DEFAULT '["claude-code"]',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(registry_id, slug)
);
CREATE INDEX idx_skill_market_items_registry ON skill_market_items(registry_id);
CREATE INDEX idx_skill_market_items_category ON skill_market_items(category);
CREATE INDEX idx_skill_market_items_active ON skill_market_items(is_active);

ALTER TABLE installed_skills
    ADD COLUMN market_item_id BIGINT REFERENCES skill_market_items(id) ON DELETE SET NULL;
DROP INDEX IF EXISTS idx_installed_skills_skill;
ALTER TABLE installed_skills DROP COLUMN IF EXISTS skill_id;

-- Revert skills -> authored_skills. Imported (upstream) rows have NULL
-- created_by_id / possibly NULL organization_id; give them defaults so the
-- NOT NULL constraints can be restored.
UPDATE skills SET organization_id = 0 WHERE organization_id IS NULL;
UPDATE skills SET created_by_id = 0 WHERE created_by_id IS NULL;

DROP INDEX idx_skills_org_slug;
ALTER TABLE skills
    DROP COLUMN category,
    DROP COLUMN compatibility,
    DROP COLUMN allowed_tools,
    DROP COLUMN agent_filter,
    DROP COLUMN is_active,
    DROP COLUMN upstream_url,
    DROP COLUMN upstream_subdir,
    DROP COLUMN upstream_commit_sha,
    ALTER COLUMN organization_id SET NOT NULL,
    ALTER COLUMN created_by_id SET NOT NULL;

ALTER TABLE skills RENAME TO authored_skills;
ALTER INDEX idx_skills_org_updated RENAME TO idx_authored_skills_org_updated;
CREATE UNIQUE INDEX idx_authored_skills_org_slug ON authored_skills(organization_id, slug);
