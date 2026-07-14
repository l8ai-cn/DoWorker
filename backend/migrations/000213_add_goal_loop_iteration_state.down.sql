ALTER TABLE goal_loops
    DROP CONSTRAINT IF EXISTS chk_goal_loops_iteration_state,
    DROP COLUMN IF EXISTS last_error_fingerprint,
    DROP COLUMN IF EXISTS last_progress_fingerprint,
    DROP COLUMN IF EXISTS same_error_count,
    DROP COLUMN IF EXISTS no_progress_count,
    DROP COLUMN IF EXISTS current_iteration;
