package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestAgentWorkbenchStreamMigrationContract(t *testing.T) {
	up := readMigrationForTest(t, "000225_agent_workbench_stream.up.sql")
	for _, fragment := range []string{
		"CREATE TABLE agent_workbench_session_states",
		"session_id VARCHAR(100) PRIMARY KEY",
		"stream_epoch VARCHAR(100) NOT NULL",
		"revision NUMERIC(20, 0) NOT NULL",
		"latest_sequence NUMERIC(20, 0) NOT NULL",
		"projection BYTEA NOT NULL",
		"CREATE TABLE agent_workbench_events",
		"payload BYTEA NOT NULL",
		"PRIMARY KEY (session_id, stream_epoch, sequence)",
		"causation_command_id VARCHAR(100)",
		"CREATE TABLE agent_workbench_source_events",
		"stable_event_id VARCHAR(200) NOT NULL",
		"runner_session_epoch VARCHAR(100) NOT NULL",
		"source_sequence NUMERIC(20, 0) NOT NULL",
		"PRIMARY KEY (session_id, stable_event_id)",
		"UNIQUE (session_id, runner_session_epoch, source_sequence)",
		"CREATE TABLE agent_workbench_command_receipts",
		"PRIMARY KEY (session_id, command_id)",
		"receipt BYTEA NOT NULL",
		"REFERENCES agent_sessions(id) ON DELETE CASCADE",
		"prevent_agent_workbench_append_only_mutation",
		"BEFORE UPDATE OR DELETE ON agent_workbench_events",
		"BEFORE UPDATE OR DELETE ON agent_workbench_source_events",
		"18446744073709551615",
	} {
		require.Contains(t, up, fragment)
	}
	require.NotContains(t, strings.ToUpper(up), "JSON")

	down := readMigrationForTest(t, "000225_agent_workbench_stream.down.sql")
	ordered := []string{
		"DROP TRIGGER IF EXISTS agent_workbench_events_immutable",
		"DROP TRIGGER IF EXISTS agent_workbench_source_events_immutable",
		"DROP FUNCTION IF EXISTS prevent_agent_workbench_append_only_mutation",
		"DROP TABLE IF EXISTS agent_workbench_command_receipts",
		"DROP TABLE IF EXISTS agent_workbench_source_events",
		"DROP TABLE IF EXISTS agent_workbench_events",
		"DROP TABLE IF EXISTS agent_workbench_session_states",
	}
	previous := -1
	for _, fragment := range ordered {
		index := strings.Index(down, fragment)
		require.Greater(t, index, previous, fragment)
		previous = index
	}
}

func TestAgentWorkbenchStreamMigrationPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		var err error
		dsn, err = migrationPostgresDSN()
		require.NoError(t, err)
	}
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN and MIGRATIONS_POSTGRES_TEST_DSN are not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer conn.Close()

	schema := fmt.Sprintf("agent_workbench_stream_%d", time.Now().UnixNano())
	require.NoError(t, execSQL(ctx, conn, `CREATE SCHEMA `+schema))
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		_, _ = db.ExecContext(cleanupCtx, `DROP SCHEMA IF EXISTS `+schema+` CASCADE`)
	})
	require.NoError(t, execSQL(ctx, conn, `SET search_path TO `+schema+`, public`))
	require.NoError(t, execSQL(ctx, conn, `
CREATE TABLE agent_sessions (id VARCHAR(100) PRIMARY KEY);
INSERT INTO agent_sessions (id) VALUES ('conv_0123456789abcdef')`))

	up := readMigrationForTest(t, "000225_agent_workbench_stream.up.sql")
	_, err = conn.ExecContext(ctx, up)
	require.NoError(t, err)
	maxUint64 := "18446744073709551615"
	digest := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	_, err = conn.ExecContext(ctx, `
INSERT INTO agent_workbench_session_states
	(session_id, stream_epoch, revision, latest_sequence, projection, digest)
VALUES ($1, 'epoch-a', $2, $2, decode('00ff', 'hex'), $3)`,
		"conv_0123456789abcdef", maxUint64, digest)
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, `
INSERT INTO agent_workbench_events
	(session_id, stream_epoch, sequence, revision, payload, digest)
VALUES ($1, 'epoch-a', $2, $2, decode('fe00', 'hex'), $3)`,
		"conv_0123456789abcdef", maxUint64, digest)
	require.NoError(t, err)

	var revision, sequence, projection, payload string
	require.NoError(t, conn.QueryRowContext(ctx, `
SELECT revision::text, latest_sequence::text, encode(projection, 'hex')
FROM agent_workbench_session_states WHERE session_id = $1`,
		"conv_0123456789abcdef").Scan(&revision, &sequence, &projection))
	require.Equal(t, maxUint64, revision)
	require.Equal(t, maxUint64, sequence)
	require.Equal(t, "00ff", projection)
	require.NoError(t, conn.QueryRowContext(ctx, `
SELECT encode(payload, 'hex') FROM agent_workbench_events
WHERE session_id = $1 AND stream_epoch = 'epoch-a' AND sequence = $2`,
		"conv_0123456789abcdef", maxUint64).Scan(&payload))
	require.Equal(t, "fe00", payload)

	_, err = conn.ExecContext(ctx, `
INSERT INTO agent_workbench_events
	(session_id, stream_epoch, sequence, revision, payload, digest)
VALUES ($1, 'epoch-a', $2, $2, decode('01', 'hex'), $3)`,
		"conv_0123456789abcdef", maxUint64, digest)
	require.Error(t, err)
	_, err = conn.ExecContext(ctx, `
UPDATE agent_workbench_events SET digest = $1 WHERE session_id = $2`,
		digest, "conv_0123456789abcdef")
	require.Error(t, err)

	_, err = conn.ExecContext(ctx, `
INSERT INTO agent_workbench_source_events
	(session_id, stable_event_id, runner_session_epoch, source_sequence, payload_digest)
VALUES ($1, 'epoch-a:1', 'epoch-a', 1, $2)`,
		"conv_0123456789abcdef", digest)
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, `
INSERT INTO agent_workbench_source_events
	(session_id, stable_event_id, runner_session_epoch, source_sequence, payload_digest)
VALUES ($1, 'epoch-a:1', 'epoch-a', 2, $2)`,
		"conv_0123456789abcdef", digest)
	require.Error(t, err)
	_, err = conn.ExecContext(ctx, `
INSERT INTO agent_workbench_source_events
	(session_id, stable_event_id, runner_session_epoch, source_sequence, payload_digest)
VALUES ($1, 'different-id', 'epoch-a', 1, $2)`,
		"conv_0123456789abcdef", digest)
	require.Error(t, err)
	_, err = conn.ExecContext(ctx, `
UPDATE agent_workbench_source_events SET payload_digest = $1 WHERE session_id = $2`,
		digest, "conv_0123456789abcdef")
	require.Error(t, err)

	_, err = conn.ExecContext(ctx, `
INSERT INTO agent_workbench_command_receipts
	(session_id, command_id, payload_digest, state, receipt)
VALUES ($1, 'command-1', $2, 1, decode('0001', 'hex'))`,
		"conv_0123456789abcdef", digest)
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, `
INSERT INTO agent_workbench_command_receipts
	(session_id, command_id, payload_digest, state, receipt)
VALUES ($1, 'command-1', $2, 1, decode('02', 'hex'))`,
		"conv_0123456789abcdef", digest)
	require.Error(t, err)

	down := readMigrationForTest(t, "000225_agent_workbench_stream.down.sql")
	_, err = conn.ExecContext(ctx, down)
	require.NoError(t, err)
	require.False(t, postgresTableExists(ctx, t, conn, "agent_workbench_session_states"))
	require.False(t, postgresTableExists(ctx, t, conn, "agent_workbench_events"))
	require.False(t, postgresTableExists(ctx, t, conn, "agent_workbench_source_events"))
	require.False(t, postgresTableExists(ctx, t, conn, "agent_workbench_command_receipts"))
}
