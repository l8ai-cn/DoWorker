-- session_read_states.session_id must match agent_sessions.id (^conv_[a-z0-9]+$),
-- not slugkit — 167 shipped the wrong CHECK.

ALTER TABLE session_read_states DROP CONSTRAINT IF EXISTS session_read_states_session_id_check;

ALTER TABLE session_read_states ADD CONSTRAINT session_read_states_session_id_check
    CHECK (session_id ~ '^conv_[a-z0-9]+$' AND char_length(session_id) BETWEEN 8 AND 100);
