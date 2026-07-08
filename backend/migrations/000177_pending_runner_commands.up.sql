CREATE TABLE pending_runner_commands (
    id              BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL,
    runner_id       BIGINT NOT NULL REFERENCES runners(id) ON DELETE CASCADE,
    pod_key         VARCHAR(100) NOT NULL,
    command_type    VARCHAR(20) NOT NULL CHECK (command_type IN ('create_pod', 'send_prompt')),
    command_id      VARCHAR(64) NOT NULL,
    payload         BYTEA NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_pending_cmds_runner_fifo ON pending_runner_commands (runner_id, id);
CREATE INDEX idx_pending_cmds_expiry      ON pending_runner_commands (expires_at);
CREATE UNIQUE INDEX uq_pending_cmds_command ON pending_runner_commands (command_id);
