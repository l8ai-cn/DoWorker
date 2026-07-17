-- Agent platform foundations: capability declarations, policy, usage pricing, resume id.

-- A4b: claude-code
UPDATE agents SET agentfile_source = agentfile_source || E'\nCAPABILITY resume cli\nCAPABILITY permission acp\nCAPABILITY usage live\nCAPABILITY interrupt true\nCAPABILITY streaming true\nCAPABILITY model_family claude\n', updated_at = NOW()
WHERE slug = 'claude-code' AND is_builtin = true AND agentfile_source NOT LIKE '%CAPABILITY resume%';

-- A4b: codex-cli
UPDATE agents SET agentfile_source = agentfile_source || E'\nCAPABILITY resume cli\nCAPABILITY permission acp\nCAPABILITY usage live\nCAPABILITY interrupt true\nCAPABILITY streaming true\nCAPABILITY model_family gpt\n', updated_at = NOW()
WHERE slug = 'codex-cli' AND is_builtin = true AND agentfile_source NOT LIKE '%CAPABILITY resume%';

-- A4b: gemini-cli
UPDATE agents SET agentfile_source = agentfile_source || E'\nCAPABILITY resume none\nCAPABILITY permission acp\nCAPABILITY usage exit\nCAPABILITY interrupt true\nCAPABILITY streaming true\nCAPABILITY model_family gemini\n', updated_at = NOW()
WHERE slug = 'gemini-cli' AND is_builtin = true AND agentfile_source NOT LIKE '%CAPABILITY resume%';

-- A4b: opencode
UPDATE agents SET agentfile_source = agentfile_source || E'\nCAPABILITY resume none\nCAPABILITY permission acp\nCAPABILITY usage live\nCAPABILITY interrupt true\nCAPABILITY streaming true\nCAPABILITY model_family multi\n', updated_at = NOW()
WHERE slug = 'opencode' AND is_builtin = true AND agentfile_source NOT LIKE '%CAPABILITY resume%';

-- A4b: cursor-cli
UPDATE agents SET agentfile_source = agentfile_source || E'\nCAPABILITY resume none\nCAPABILITY permission acp\nCAPABILITY usage live\nCAPABILITY interrupt true\nCAPABILITY streaming true\nCAPABILITY model_family multi\n', updated_at = NOW()
WHERE slug = 'cursor-cli' AND is_builtin = true AND agentfile_source NOT LIKE '%CAPABILITY resume%';

-- A4b: loopal
UPDATE agents SET agentfile_source = agentfile_source || E'\nCAPABILITY resume none\nCAPABILITY permission acp\nCAPABILITY usage live\nCAPABILITY interrupt true\nCAPABILITY streaming true\nCAPABILITY model_family multi\n', updated_at = NOW()
WHERE slug = 'loopal' AND is_builtin = true AND agentfile_source NOT LIKE '%CAPABILITY resume%';

-- B: org-scoped tool permission policies
CREATE TABLE permission_policies (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    scope VARCHAR(20) NOT NULL DEFAULT 'org' CHECK (scope IN ('org', 'project', 'pod')),
    agent_slug VARCHAR(50),
    tool_pattern TEXT NOT NULL,
    path_pattern TEXT,
    verdict VARCHAR(10) NOT NULL CHECK (verdict IN ('allow', 'deny', 'ask')),
    priority INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_permission_policies_org ON permission_policies(organization_id, priority DESC);

-- C: model pricing for USD aggregation
CREATE TABLE model_prices (
    model VARCHAR(100) PRIMARY KEY,
    input_per_million NUMERIC(12, 6) NOT NULL DEFAULT 0,
    output_per_million NUMERIC(12, 6) NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO model_prices (model, input_per_million, output_per_million) VALUES
    ('claude-sonnet-4-20250514', 3.0, 15.0),
    ('gpt-4o', 2.5, 10.0),
    ('gpt-4o-mini', 0.15, 0.6)
ON CONFLICT (model) DO NOTHING;

-- D: vendor session id capture on pods
ALTER TABLE pods ADD COLUMN IF NOT EXISTS external_session_id VARCHAR(200);
