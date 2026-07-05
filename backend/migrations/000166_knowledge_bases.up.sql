-- Knowledge bases: org-scoped llm-wiki repositories backed by internal Gitea.
--
-- Each knowledge_base row maps 1:1 onto a git repository with the llm-wiki
-- layout (llms.txt index, AGENTS.md schema, raw/ immutable sources, wiki/
-- LLM-maintained pages). Pods mount the repo read-only or read-write;
-- external sources (feishu / dingtalk / google) sync one-way into raw/ via
-- connectors, so the mount pipeline only ever deals with git.

CREATE TABLE knowledge_bases (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    slug VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    git_repo_path VARCHAR(255) NOT NULL,
    http_clone_url VARCHAR(1000) NOT NULL,
    default_branch VARCHAR(255) NOT NULL DEFAULT 'main',
    source_type VARCHAR(32) NOT NULL DEFAULT 'git',
    source_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    sync_status VARCHAR(32) NOT NULL DEFAULT 'idle',
    sync_error TEXT,
    last_synced_at TIMESTAMPTZ,
    created_by_user_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (organization_id, slug),
    CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND char_length(slug) BETWEEN 2 AND 100),
    CHECK (source_type IN ('git', 'feishu', 'dingtalk', 'google')),
    CHECK (sync_status IN ('idle', 'syncing', 'synced', 'failed'))
);

CREATE INDEX knowledge_bases_org ON knowledge_bases (organization_id);
CREATE INDEX knowledge_bases_source_type ON knowledge_bases (source_type);

CREATE TRIGGER update_knowledge_bases_updated_at
    BEFORE UPDATE ON knowledge_bases
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Default mounts: which agents get this KB attached automatically at pod
-- create (merged with per-request selections and Agentfile KNOWLEDGE decls).
CREATE TABLE knowledge_base_agent_mounts (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL,
    knowledge_base_id BIGINT NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    agent_slug VARCHAR(100) NOT NULL,
    mode VARCHAR(8) NOT NULL DEFAULT 'ro',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (knowledge_base_id, agent_slug),
    CHECK (mode IN ('ro', 'rw')),
    CHECK (agent_slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND char_length(agent_slug) BETWEEN 2 AND 100)
);

CREATE INDEX knowledge_base_agent_mounts_org ON knowledge_base_agent_mounts (organization_id);
CREATE INDEX knowledge_base_agent_mounts_agent ON knowledge_base_agent_mounts (organization_id, agent_slug);

COMMENT ON TABLE knowledge_bases IS 'Org-scoped llm-wiki knowledge bases; each row maps to an internal Gitea repository with llms.txt / AGENTS.md / raw/ / wiki/ layout.';
COMMENT ON TABLE knowledge_base_agent_mounts IS 'Default KB→agent mounts applied at pod create; mode ro|rw controls whether the pod may push wiki updates back.';
