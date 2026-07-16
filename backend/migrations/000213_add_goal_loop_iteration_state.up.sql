ALTER TABLE goal_loops
    ADD COLUMN current_iteration INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN no_progress_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN same_error_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN last_progress_fingerprint VARCHAR(64),
    ADD COLUMN last_error_fingerprint VARCHAR(64),
    ADD CONSTRAINT chk_goal_loops_iteration_state CHECK (
        current_iteration >= 0
        AND no_progress_count >= 0
        AND same_error_count >= 0
    );
