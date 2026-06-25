-- Coordinator: task-source-driven scheduling engine ported from auto-harness.
--
-- A coordinator_project is bound to one repository and periodically scans the
-- external platform (CNB issues today) for tasks that match its claim policy.
-- Matched issues are synced into the existing tickets table (deduped via
-- ticket_external_links), claimed with a marker comment, and dispatched as
-- do-agent pods. coordinator_executions records each claim→dispatch→feedback
-- cycle so the Web execution board can render progress per project.

CREATE TABLE coordinator_projects (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL,
    repository_id BIGINT NOT NULL,
    slug VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    platform_type VARCHAR(32) NOT NULL DEFAULT 'cnb',
    source_type VARCHAR(32) NOT NULL DEFAULT 'issues',
    label_filter TEXT[] NOT NULL DEFAULT '{}',
    claim_policy JSONB NOT NULL DEFAULT '{}'::jsonb,
    agent_slug VARCHAR(100) NOT NULL DEFAULT 'do-agent',
    scan_interval_seconds INTEGER NOT NULL DEFAULT 300,
    max_concurrent INTEGER NOT NULL DEFAULT 1,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_by_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (organization_id, slug)
);

CREATE INDEX coordinator_projects_org ON coordinator_projects (organization_id);
CREATE INDEX coordinator_projects_repository ON coordinator_projects (repository_id);
CREATE INDEX coordinator_projects_enabled ON coordinator_projects (enabled);

CREATE TRIGGER update_coordinator_projects_updated_at
    BEFORE UPDATE ON coordinator_projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE ticket_external_links (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL,
    ticket_id BIGINT NOT NULL,
    platform_type VARCHAR(32) NOT NULL,
    source_id VARCHAR(255),
    external_id VARCHAR(255) NOT NULL,
    external_url VARCHAR(1000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (organization_id, platform_type, external_id)
);

CREATE INDEX ticket_external_links_ticket ON ticket_external_links (ticket_id);

CREATE TRIGGER update_ticket_external_links_updated_at
    BEFORE UPDATE ON ticket_external_links
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE coordinator_executions (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL,
    project_id BIGINT NOT NULL REFERENCES coordinator_projects(id) ON DELETE CASCADE,
    ticket_id BIGINT NOT NULL,
    pod_id BIGINT,
    pod_key VARCHAR(100),
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    stage VARCHAR(64),
    claim_marker TEXT,
    external_id VARCHAR(255),
    summary TEXT,
    feedback_status VARCHAR(32),
    error TEXT,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX coordinator_executions_org ON coordinator_executions (organization_id);
CREATE INDEX coordinator_executions_project ON coordinator_executions (project_id);
CREATE INDEX coordinator_executions_ticket ON coordinator_executions (ticket_id);
CREATE INDEX coordinator_executions_pod_key ON coordinator_executions (pod_key);
CREATE INDEX coordinator_executions_status ON coordinator_executions (status);

CREATE TRIGGER update_coordinator_executions_updated_at
    BEFORE UPDATE ON coordinator_executions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE coordinator_projects IS 'Org-scoped auto-harness coordinator config: one repository, a claim policy, and a dispatch agent.';
COMMENT ON TABLE ticket_external_links IS 'Idempotency map from external platform issues to AgentsMesh tickets; UNIQUE(org,platform,external_id) prevents duplicate sync.';
COMMENT ON TABLE coordinator_executions IS 'One claim→dispatch→feedback cycle, linking a coordinator project to its ticket and the pod that ran it.';
