CREATE TABLE agent_workbench_session_states (
    session_id VARCHAR(100) PRIMARY KEY
        REFERENCES agent_sessions(id) ON DELETE CASCADE,
    stream_epoch VARCHAR(100) NOT NULL
        CHECK (stream_epoch = btrim(stream_epoch) AND char_length(stream_epoch) BETWEEN 1 AND 100),
    revision NUMERIC(20, 0) NOT NULL
        CHECK (revision BETWEEN 0 AND 18446744073709551615),
    latest_sequence NUMERIC(20, 0) NOT NULL
        CHECK (latest_sequence BETWEEN 0 AND 18446744073709551615),
    projection BYTEA NOT NULL,
    digest VARCHAR(71) NOT NULL
        CHECK (digest ~ '^sha256:[0-9a-f]{64}$'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE agent_workbench_events (
    session_id VARCHAR(100) NOT NULL
        REFERENCES agent_sessions(id) ON DELETE CASCADE,
    stream_epoch VARCHAR(100) NOT NULL
        CHECK (stream_epoch = btrim(stream_epoch) AND char_length(stream_epoch) BETWEEN 1 AND 100),
    sequence NUMERIC(20, 0) NOT NULL
        CHECK (sequence BETWEEN 1 AND 18446744073709551615),
    revision NUMERIC(20, 0) NOT NULL
        CHECK (revision BETWEEN 1 AND 18446744073709551615),
    payload BYTEA NOT NULL,
    digest VARCHAR(71) NOT NULL
        CHECK (digest ~ '^sha256:[0-9a-f]{64}$'),
    causation_command_id VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (session_id, stream_epoch, sequence),
    CHECK (
        causation_command_id IS NULL
        OR (
            causation_command_id = btrim(causation_command_id)
            AND char_length(causation_command_id) BETWEEN 1 AND 100
        )
    )
);

CREATE TABLE agent_workbench_source_events (
    session_id VARCHAR(100) NOT NULL
        REFERENCES agent_sessions(id) ON DELETE CASCADE,
    stable_event_id VARCHAR(200) NOT NULL
        CHECK (stable_event_id = btrim(stable_event_id) AND char_length(stable_event_id) BETWEEN 1 AND 200),
    runner_session_epoch VARCHAR(100) NOT NULL
        CHECK (
            runner_session_epoch = btrim(runner_session_epoch)
            AND char_length(runner_session_epoch) BETWEEN 1 AND 100
        ),
    source_sequence NUMERIC(20, 0) NOT NULL
        CHECK (source_sequence BETWEEN 1 AND 18446744073709551615),
    payload_digest VARCHAR(71) NOT NULL
        CHECK (payload_digest ~ '^sha256:[0-9a-f]{64}$'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (session_id, stable_event_id),
    UNIQUE (session_id, runner_session_epoch, source_sequence)
);

CREATE TABLE agent_workbench_command_receipts (
    session_id VARCHAR(100) NOT NULL
        REFERENCES agent_sessions(id) ON DELETE CASCADE,
    command_id VARCHAR(100) NOT NULL
        CHECK (command_id = btrim(command_id) AND char_length(command_id) BETWEEN 1 AND 100),
    payload_digest VARCHAR(71) NOT NULL
        CHECK (payload_digest ~ '^sha256:[0-9a-f]{64}$'),
    state SMALLINT NOT NULL CHECK (state BETWEEN 1 AND 7),
    receipt BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (session_id, command_id)
);

CREATE FUNCTION prevent_agent_workbench_append_only_mutation()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE'
        AND pg_trigger_depth() > 1
        AND NOT EXISTS (
            SELECT 1 FROM agent_sessions WHERE id = OLD.session_id
        )
    THEN
        RETURN OLD;
    END IF;
    RAISE EXCEPTION '% is append-only', TG_TABLE_NAME;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER agent_workbench_events_immutable
    BEFORE UPDATE OR DELETE ON agent_workbench_events
    FOR EACH ROW
    EXECUTE FUNCTION prevent_agent_workbench_append_only_mutation();

CREATE TRIGGER agent_workbench_source_events_immutable
    BEFORE UPDATE OR DELETE ON agent_workbench_source_events
    FOR EACH ROW
    EXECUTE FUNCTION prevent_agent_workbench_append_only_mutation();
