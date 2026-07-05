ALTER TABLE session_read_states DROP CONSTRAINT IF EXISTS session_read_states_session_id_check;

ALTER TABLE session_read_states ADD CONSTRAINT session_read_states_session_id_check
    CHECK (session_id ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND char_length(session_id) BETWEEN 2 AND 100);
