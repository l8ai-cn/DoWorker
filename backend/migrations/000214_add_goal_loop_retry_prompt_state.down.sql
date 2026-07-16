DROP INDEX IF EXISTS idx_goal_loops_retry_prompt_pending;

ALTER TABLE goal_loops
    DROP CONSTRAINT IF EXISTS chk_goal_loops_retry_prompt_state,
    DROP COLUMN IF EXISTS retry_prompt_created_at,
    DROP COLUMN IF EXISTS retry_prompt_command_id;
