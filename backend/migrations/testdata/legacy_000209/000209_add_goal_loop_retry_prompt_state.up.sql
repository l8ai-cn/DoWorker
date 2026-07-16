ALTER TABLE goal_loops
    ADD COLUMN retry_prompt_command_id VARCHAR(64),
    ADD COLUMN retry_prompt_created_at TIMESTAMPTZ,
    ADD CONSTRAINT chk_goal_loops_retry_prompt_state CHECK (
        (retry_prompt_command_id IS NULL AND retry_prompt_created_at IS NULL)
        OR (retry_prompt_command_id IS NOT NULL AND retry_prompt_created_at IS NOT NULL)
    );

CREATE INDEX idx_goal_loops_retry_prompt_pending
    ON goal_loops (id)
    WHERE status = 'verifying' AND retry_prompt_command_id IS NOT NULL;
