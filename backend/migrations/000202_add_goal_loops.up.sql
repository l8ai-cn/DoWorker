CREATE TABLE goal_loops (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_by_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    worker_spec_snapshot_id BIGINT NOT NULL REFERENCES worker_spec_snapshots(id) ON DELETE RESTRICT,
    objective TEXT NOT NULL,
    acceptance_criteria JSONB NOT NULL DEFAULT '[]'::jsonb,
    verification_command TEXT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'draft',
    pod_key VARCHAR(100),
    autopilot_controller_key VARCHAR(255),
    max_iterations INTEGER NOT NULL DEFAULT 10,
    token_budget BIGINT,
    timeout_minutes INTEGER NOT NULL DEFAULT 60,
    no_progress_limit INTEGER NOT NULL DEFAULT 3,
    same_error_limit INTEGER NOT NULL DEFAULT 2,
    escalation_policy VARCHAR(20) NOT NULL DEFAULT 'pause',
    verification_request_id VARCHAR(100),
    verification_exit_code INTEGER,
    verification_output TEXT,
    verification_output_truncated BOOLEAN NOT NULL DEFAULT FALSE,
    verification_error TEXT,
    started_at TIMESTAMPTZ,
    verified_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_goal_loops_organization_slug UNIQUE (organization_id, slug),
    CONSTRAINT chk_goal_loops_slug CHECK (
        slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'
        AND char_length(slug) BETWEEN 2 AND 100
    ),
    CONSTRAINT chk_goal_loops_status CHECK (
        status IN ('draft', 'active', 'paused', 'verifying', 'completed', 'failed', 'cancelled')
    ),
    CONSTRAINT chk_goal_loops_acceptance_criteria CHECK (
        jsonb_typeof(acceptance_criteria) = 'array'
    ),
    CONSTRAINT chk_goal_loops_bounds CHECK (
        max_iterations BETWEEN 1 AND 100
        AND timeout_minutes BETWEEN 1 AND 1440
        AND no_progress_limit BETWEEN 1 AND 20
        AND same_error_limit BETWEEN 1 AND 20
    ),
    CONSTRAINT chk_goal_loops_escalation_policy CHECK (
        escalation_policy IN ('pause', 'fail')
    )
);

CREATE INDEX idx_goal_loops_organization_created_at
    ON goal_loops (organization_id, created_at DESC);
CREATE INDEX idx_goal_loops_pod_key
    ON goal_loops (pod_key)
    WHERE pod_key IS NOT NULL;
CREATE UNIQUE INDEX idx_goal_loops_verification_request_id
    ON goal_loops (verification_request_id)
    WHERE verification_request_id IS NOT NULL;
