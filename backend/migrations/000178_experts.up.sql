CREATE TABLE experts (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL,
    slug VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,

    agent_slug VARCHAR(100) NOT NULL,
    runner_id BIGINT,
    repository_id BIGINT,
    branch_name VARCHAR(255),

    prompt TEXT,
    interaction_mode VARCHAR(20) NOT NULL DEFAULT 'pty',
    perpetual BOOLEAN NOT NULL DEFAULT false,

    used_env_bundles TEXT[] NOT NULL DEFAULT '{}',
    skill_slugs TEXT[] NOT NULL DEFAULT '{}',
    knowledge_mounts JSONB NOT NULL DEFAULT '[]',
    config_overrides JSONB NOT NULL DEFAULT '{}',
    agentfile_layer TEXT,

    source_pod_key VARCHAR(100),

    created_by_id BIGINT NOT NULL,
    run_count INT NOT NULL DEFAULT 0,
    last_run_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND char_length(slug) BETWEEN 2 AND 100)
);

CREATE UNIQUE INDEX idx_experts_org_slug ON experts(organization_id, slug);
CREATE INDEX idx_experts_org_updated ON experts(organization_id, updated_at DESC);
