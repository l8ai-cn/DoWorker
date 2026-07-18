package infra

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	"github.com/anthropics/agentsmesh/backend/migrations"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const agentWorkbenchDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestAgentWorkbenchRepository(t *testing.T) {
	db, repo := agentWorkbenchPostgresRepository(t)
	t.Run("AppendAndReplay", func(t *testing.T) {
		testAgentWorkbenchAppendAndReplay(t, db, repo)
	})
	t.Run("RevisionConflictRollsBack", func(t *testing.T) {
		testAgentWorkbenchRevisionConflictRollsBack(t, db, repo)
	})
	t.Run("EventConflictRollsBack", func(t *testing.T) {
		testAgentWorkbenchEventConflictRollsBack(t, db, repo)
	})
	t.Run("SerializesInitialAppend", func(t *testing.T) {
		testAgentWorkbenchSerializesInitialAppend(t, db, repo)
	})
	t.Run("PreservesUint64AndCascadeDelete", func(t *testing.T) {
		testAgentWorkbenchUint64AndCascadeDelete(t, db, repo)
	})
	t.Run("CommandReceipts", func(t *testing.T) {
		testAgentWorkbenchCommandReceipts(t, db, repo)
	})
	t.Run("AppendRollsBackReceiptOnEventConflict", func(t *testing.T) {
		testAgentWorkbenchAppendRollsBackReceiptOnEventConflict(t, db, repo)
	})
	t.Run("SourceIngressConcurrentReplay", func(t *testing.T) {
		testAgentWorkbenchSourceIngressConcurrentReplay(t, db, repo)
	})
	t.Run("SourceIngressConflicts", func(t *testing.T) {
		testAgentWorkbenchSourceIngressConflicts(t, db, repo)
	})
	t.Run("EnsureSnapshot", func(t *testing.T) {
		testAgentWorkbenchEnsureSnapshot(t, db, repo)
	})
}

func testAgentWorkbenchAppendAndReplay(
	t *testing.T,
	db *gorm.DB,
	repo agentworkbench.Repository,
) {
	sessionID := insertAgentWorkbenchSession(t, db)
	ctx := context.Background()
	projection := []byte{0x0a, 0x03, 0x00, 0xff}
	events := []agentworkbench.Event{
		newAgentWorkbenchEvent(sessionID, "epoch-a", 1, 1, []byte{0x08, 0x01}),
		newAgentWorkbenchEvent(sessionID, "epoch-a", 1, 2, []byte{0x08, 0x02, 0xff}),
	}

	require.NoError(t, appendAgentWorkbench(ctx, repo, sessionID, 0, events, agentworkbench.SessionState{
		SessionID: sessionID, StreamEpoch: "epoch-a", Revision: 1,
		LatestSequence: 2, Projection: projection, Digest: agentWorkbenchDigest,
	}))
	state, err := repo.GetSnapshot(ctx, sessionID)
	require.NoError(t, err)
	require.Equal(t, uint64(1), state.Revision)
	require.Equal(t, uint64(2), state.LatestSequence)
	require.Equal(t, projection, state.Projection)

	replayed, err := repo.ListAfter(ctx, sessionID, "epoch-a", 0, 10)
	require.NoError(t, err)
	require.Len(t, replayed, 2)
	require.Equal(t, events[0].Payload, replayed[0].Payload)
	require.Equal(t, events[1].Payload, replayed[1].Payload)
	require.Equal(t, uint64(2), replayed[1].Sequence)

	afterFirst, err := repo.ListAfter(ctx, sessionID, "epoch-a", 1, 10)
	require.NoError(t, err)
	require.Len(t, afterFirst, 1)
	require.Equal(t, uint64(2), afterFirst[0].Sequence)
}

func testAgentWorkbenchRevisionConflictRollsBack(
	t *testing.T,
	db *gorm.DB,
	repo agentworkbench.Repository,
) {
	sessionID := insertAgentWorkbenchSession(t, db)
	ctx := context.Background()
	initial := newAgentWorkbenchEvent(sessionID, "epoch-a", 1, 1, []byte("first"))
	require.NoError(t, appendAgentWorkbench(ctx, repo, sessionID, 0, []agentworkbench.Event{initial},
		agentworkbench.SessionState{
			SessionID: sessionID, StreamEpoch: "epoch-a", Revision: 1,
			LatestSequence: 1, Projection: []byte("projection-1"), Digest: agentWorkbenchDigest,
		}))

	stale := newAgentWorkbenchEvent(sessionID, "epoch-a", 1, 1, []byte("stale"))
	err := appendAgentWorkbench(ctx, repo, sessionID, 0, []agentworkbench.Event{stale},
		agentworkbench.SessionState{
			SessionID: sessionID, StreamEpoch: "epoch-a", Revision: 1,
			LatestSequence: 1, Projection: []byte("stale-projection"), Digest: agentWorkbenchDigest,
		})
	require.ErrorIs(t, err, agentworkbench.ErrRevisionConflict)
	assertAgentWorkbenchState(t, repo, sessionID, 1, 1, []byte("projection-1"))
}

func testAgentWorkbenchEventConflictRollsBack(
	t *testing.T,
	db *gorm.DB,
	repo agentworkbench.Repository,
) {
	sessionID := insertAgentWorkbenchSession(t, db)
	ctx := context.Background()
	first := newAgentWorkbenchEvent(sessionID, "epoch-a", 1, 1, []byte("first"))
	require.NoError(t, appendAgentWorkbench(ctx, repo, sessionID, 0, []agentworkbench.Event{first},
		agentworkbench.SessionState{
			SessionID: sessionID, StreamEpoch: "epoch-a", Revision: 1,
			LatestSequence: 1, Projection: []byte("projection-1"), Digest: agentWorkbenchDigest,
		}))
	require.NoError(t, db.Exec(`
INSERT INTO agent_workbench_events
	(session_id, stream_epoch, sequence, revision, payload, digest, created_at)
VALUES (?, 'epoch-a', 2, 2, ?, ?, NOW())`,
		sessionID, []byte("preexisting"), agentWorkbenchDigest).Error)

	second := newAgentWorkbenchEvent(sessionID, "epoch-a", 2, 2, []byte("second"))
	err := appendAgentWorkbench(ctx, repo, sessionID, 1, []agentworkbench.Event{second},
		agentworkbench.SessionState{
			SessionID: sessionID, StreamEpoch: "epoch-a", Revision: 2,
			LatestSequence: 2, Projection: []byte("projection-2"), Digest: agentWorkbenchDigest,
		})
	require.ErrorIs(t, err, agentworkbench.ErrEventConflict)
	assertAgentWorkbenchStateRevision(t, repo, sessionID, 1, 1)
	replayed, listErr := repo.ListAfter(ctx, sessionID, "epoch-a", 1, 10)
	require.NoError(t, listErr)
	require.Len(t, replayed, 1)
	require.Equal(t, []byte("preexisting"), replayed[0].Payload)
}

func testAgentWorkbenchSerializesInitialAppend(
	t *testing.T,
	db *gorm.DB,
	repo agentworkbench.Repository,
) {
	sessionID := insertAgentWorkbenchSession(t, db)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	start := make(chan struct{})
	errs := make(chan error, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	for i := 0; i < 2; i++ {
		go func(value byte) {
			ready.Done()
			<-start
			event := newAgentWorkbenchEvent(sessionID, "epoch-a", 1, 1, []byte{value})
			errs <- appendAgentWorkbench(ctx, repo, sessionID, 0, []agentworkbench.Event{event},
				agentworkbench.SessionState{
					SessionID: sessionID, StreamEpoch: "epoch-a", Revision: 1,
					LatestSequence: 1, Projection: []byte{value}, Digest: agentWorkbenchDigest,
				})
		}(byte(i + 1))
	}
	ready.Wait()
	close(start)
	results := []error{<-errs, <-errs}
	require.Equal(t, 1, countAgentWorkbenchErrors(results, nil))
	require.Equal(t, 1, countAgentWorkbenchErrors(results, agentworkbench.ErrRevisionConflict))
	assertAgentWorkbenchStateRevision(t, repo, sessionID, 1, 1)
}

func testAgentWorkbenchUint64AndCascadeDelete(
	t *testing.T,
	db *gorm.DB,
	repo agentworkbench.Repository,
) {
	sessionID := insertAgentWorkbenchSession(t, db)
	projection := []byte{0x00, 0xff, 0x7f}
	require.NoError(t, db.Exec(`
INSERT INTO agent_workbench_session_states
	(session_id, stream_epoch, revision, latest_sequence, projection, digest)
VALUES (?, 'epoch-max', ?, ?, ?, ?)`,
		sessionID, uint64(math.MaxUint64), uint64(math.MaxUint64),
		projection, agentWorkbenchDigest).Error)
	require.NoError(t, db.Exec(`
INSERT INTO agent_workbench_events
	(session_id, stream_epoch, sequence, revision, payload, digest)
VALUES (?, 'epoch-max', ?, ?, ?, ?)`,
		sessionID, uint64(math.MaxUint64), uint64(math.MaxUint64),
		[]byte{0xff, 0x00}, agentWorkbenchDigest).Error)
	require.NoError(t, db.Exec(`
INSERT INTO agent_workbench_source_events
	(session_id, stable_event_id, runner_session_epoch, source_sequence, payload_digest)
VALUES (?, 'epoch-max:18446744073709551615', 'epoch-max', ?, ?)`,
		sessionID, uint64(math.MaxUint64), agentWorkbenchDigest).Error)
	require.NoError(t, db.Exec(`
INSERT INTO agent_workbench_command_receipts
	(session_id, command_id, payload_digest, state, receipt)
VALUES (?, 'command-max', ?, 1, ?)`,
		sessionID, agentWorkbenchDigest, []byte{0x08, 0x01}).Error)

	state, err := repo.GetSnapshot(context.Background(), sessionID)
	require.NoError(t, err)
	require.Equal(t, uint64(math.MaxUint64), state.Revision)
	require.Equal(t, uint64(math.MaxUint64), state.LatestSequence)
	require.Equal(t, projection, state.Projection)
	events, err := repo.ListAfter(
		context.Background(),
		sessionID,
		"epoch-max",
		uint64(math.MaxUint64-1),
		1,
	)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, uint64(math.MaxUint64), events[0].Sequence)
	require.Equal(t, []byte{0xff, 0x00}, events[0].Payload)

	require.NoError(t, db.Exec(
		`DELETE FROM agent_sessions WHERE id = ?`,
		sessionID,
	).Error)
	for _, table := range []string{
		"agent_workbench_session_states",
		"agent_workbench_events",
		"agent_workbench_source_events",
		"agent_workbench_command_receipts",
	} {
		var count int64
		require.NoError(t, db.Table(table).
			Where("session_id = ?", sessionID).Count(&count).Error)
		require.Zero(t, count)
	}
}

func appendAgentWorkbench(
	ctx context.Context,
	repo agentworkbench.Repository,
	sessionID string,
	expectedRevision uint64,
	events []agentworkbench.Event,
	projection agentworkbench.SessionState,
) error {
	_, err := repo.Append(ctx, agentworkbench.AppendRequest{
		SessionID: sessionID, ExpectedRevision: expectedRevision,
		Events: events, Projection: projection,
	})
	return err
}

func newAgentWorkbenchEvent(
	sessionID, epoch string,
	revision, sequence uint64,
	payload []byte,
) agentworkbench.Event {
	return agentworkbench.Event{
		SessionID: sessionID, StreamEpoch: epoch, Revision: revision,
		Sequence: sequence, Payload: payload, Digest: agentWorkbenchDigest,
		CreatedAt: time.Now().UTC(),
	}
}

func assertAgentWorkbenchState(
	t *testing.T,
	repo agentworkbench.Repository,
	sessionID string,
	revision, sequence uint64,
	projection []byte,
) {
	t.Helper()
	state, err := repo.GetSnapshot(context.Background(), sessionID)
	require.NoError(t, err)
	require.Equal(t, revision, state.Revision)
	require.Equal(t, sequence, state.LatestSequence)
	require.Equal(t, projection, state.Projection)
	events, err := repo.ListAfter(context.Background(), sessionID, state.StreamEpoch, sequence, 10)
	require.NoError(t, err)
	require.Empty(t, events)
}

func assertAgentWorkbenchStateRevision(
	t *testing.T,
	repo agentworkbench.Repository,
	sessionID string,
	revision, sequence uint64,
) {
	t.Helper()
	state, err := repo.GetSnapshot(context.Background(), sessionID)
	require.NoError(t, err)
	require.Equal(t, revision, state.Revision)
	require.Equal(t, sequence, state.LatestSequence)
}

func countAgentWorkbenchErrors(results []error, target error) int {
	count := 0
	for _, err := range results {
		if target == nil && err == nil {
			count++
		}
		if target != nil && errors.Is(err, target) {
			count++
		}
	}
	return count
}

func agentWorkbenchPostgresRepository(
	t *testing.T,
) (*gorm.DB, agentworkbench.PersistenceRepository) {
	t.Helper()
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://agentsmesh:agentsmesh_dev@localhost:10002/agentsmesh?sslmode=disable"
	}
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("PostgreSQL is unavailable: %v", err)
	}
	sqlDB, err := admin.DB()
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		t.Skipf("PostgreSQL is unavailable: %v", err)
	}
	schema := fmt.Sprintf("agent_workbench_repo_%d", time.Now().UnixNano())
	require.NoError(t, admin.WithContext(ctx).Exec(`CREATE SCHEMA `+schema).Error)
	db, err := gorm.Open(postgres.Open(agentWorkbenchSchemaDSN(dsn, schema)),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	schemaSQL, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = schemaSQL.Close()
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_ = admin.WithContext(cleanupCtx).
			Exec(`DROP SCHEMA IF EXISTS ` + schema + ` CASCADE`).Error
		_ = sqlDB.Close()
	})
	require.NoError(t, db.Exec(
		`CREATE TABLE agent_sessions (id VARCHAR(100) PRIMARY KEY)`,
	).Error)
	sessionFiles, err := migrations.FS.ReadFile("000172_session_files.up.sql")
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(sessionFiles)).Error)
	up, err := migrations.FS.ReadFile("000225_agent_workbench_stream.up.sql")
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(up)).Error)
	return db, NewAgentWorkbenchRepository(db)
}

func agentWorkbenchSchemaDSN(dsn, schema string) string {
	parsed, err := url.Parse(dsn)
	if err == nil && parsed.Scheme != "" {
		query := parsed.Query()
		query.Set("search_path", schema+",public")
		query.Set("default_query_exec_mode", "simple_protocol")
		parsed.RawQuery = query.Encode()
		return parsed.String()
	}
	return dsn + " search_path=" + schema + ",public default_query_exec_mode=simple_protocol"
}

func insertAgentWorkbenchSession(t *testing.T, db *gorm.DB) string {
	t.Helper()
	sessionID := fmt.Sprintf("conv_%016x", time.Now().UnixNano())
	require.NoError(t, db.Exec(
		`INSERT INTO agent_sessions (id) VALUES (?)`,
		sessionID,
	).Error)
	return sessionID
}
