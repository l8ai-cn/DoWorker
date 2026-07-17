package infra

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func testAgentWorkbenchEnsureSnapshot(
	t *testing.T,
	db *gorm.DB,
	repo agentworkbench.PersistenceRepository,
) {
	sessionID := insertAgentWorkbenchSession(t, db)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	candidates := []agentworkbench.SessionState{
		initialAgentWorkbenchSnapshot(sessionID, "epoch-initial-a", []byte{0x08, 0x01}, agentWorkbenchDigest),
		initialAgentWorkbenchSnapshot(sessionID, "epoch-initial-b", []byte{0x08, 0x02}, agentWorkbenchDigestB),
	}
	start := make(chan struct{})
	results := make(chan *agentworkbench.SessionState, 2)
	errs := make(chan error, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	for _, candidate := range candidates {
		go func(state agentworkbench.SessionState) {
			ready.Done()
			<-start
			stored, err := repo.EnsureSnapshot(ctx, state)
			results <- stored
			errs <- err
		}(candidate)
	}
	ready.Wait()
	close(start)

	require.NoError(t, <-errs)
	require.NoError(t, <-errs)
	first, second := <-results, <-results
	requireSameAgentWorkbenchSnapshot(t, first, second)
	require.Contains(t, []string{"epoch-initial-a", "epoch-initial-b"}, first.StreamEpoch)
	require.Zero(t, first.Revision)
	require.Zero(t, first.LatestSequence)

	var count int64
	require.NoError(t, db.Table("agent_workbench_session_states").
		Where("session_id = ?", sessionID).Count(&count).Error)
	require.Equal(t, int64(1), count)

	existing, err := repo.EnsureSnapshot(ctx, initialAgentWorkbenchSnapshot(
		sessionID,
		"epoch-ignored",
		[]byte("ignored"),
		agentWorkbenchDigest,
	))
	require.NoError(t, err)
	requireSameAgentWorkbenchSnapshot(t, first, existing)

	for _, invalid := range invalidAgentWorkbenchInitialSnapshots(sessionID) {
		_, err := repo.EnsureSnapshot(ctx, invalid)
		require.ErrorIs(t, err, agentworkbench.ErrInvalidArgument)
	}

	event := newAgentWorkbenchEvent(
		sessionID,
		first.StreamEpoch,
		1,
		1,
		[]byte("first-runner-event"),
	)
	require.NoError(t, appendAgentWorkbench(
		ctx,
		repo,
		sessionID,
		0,
		[]agentworkbench.Event{event},
		agentworkbench.SessionState{
			SessionID: sessionID, StreamEpoch: first.StreamEpoch,
			Revision: 1, LatestSequence: 1,
			Projection: []byte("revision-1"), Digest: agentWorkbenchDigest,
		},
	))
	assertAgentWorkbenchStateRevision(t, repo, sessionID, 1, 1)

	advanced, err := repo.GetSnapshot(ctx, sessionID)
	require.NoError(t, err)
	ensured, err := repo.EnsureSnapshot(ctx, initialAgentWorkbenchSnapshot(
		sessionID,
		"epoch-still-ignored",
		[]byte("still-ignored"),
		agentWorkbenchDigestB,
	))
	require.NoError(t, err)
	requireSameAgentWorkbenchSnapshot(t, advanced, ensured)
}

func initialAgentWorkbenchSnapshot(
	sessionID string,
	streamEpoch string,
	projection []byte,
	digest string,
) agentworkbench.SessionState {
	return agentworkbench.SessionState{
		SessionID: sessionID, StreamEpoch: streamEpoch,
		Projection: projection, Digest: digest,
	}
}

func invalidAgentWorkbenchInitialSnapshots(
	sessionID string,
) []agentworkbench.SessionState {
	valid := initialAgentWorkbenchSnapshot(
		sessionID,
		"epoch-valid",
		[]byte("snapshot"),
		agentWorkbenchDigest,
	)
	invalidRevision := valid
	invalidRevision.Revision = 1
	invalidSequence := valid
	invalidSequence.LatestSequence = 1
	invalidDigest := valid
	invalidDigest.Digest = "sha256:invalid"
	invalidEpoch := valid
	invalidEpoch.StreamEpoch = ""
	return []agentworkbench.SessionState{
		invalidRevision,
		invalidSequence,
		invalidDigest,
		invalidEpoch,
	}
}

func requireSameAgentWorkbenchSnapshot(
	t *testing.T,
	expected *agentworkbench.SessionState,
	actual *agentworkbench.SessionState,
) {
	t.Helper()
	require.NotNil(t, expected)
	require.NotNil(t, actual)
	require.Equal(t, expected.SessionID, actual.SessionID)
	require.Equal(t, expected.StreamEpoch, actual.StreamEpoch)
	require.Equal(t, expected.Revision, actual.Revision)
	require.Equal(t, expected.LatestSequence, actual.LatestSequence)
	require.Equal(t, expected.Projection, actual.Projection)
	require.Equal(t, expected.Digest, actual.Digest)
}
