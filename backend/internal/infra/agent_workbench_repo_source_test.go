package infra

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentworkbench"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func testAgentWorkbenchSourceIngressConcurrentReplay(
	t *testing.T,
	db *gorm.DB,
	repo agentworkbench.Repository,
) {
	sessionID := insertAgentWorkbenchSession(t, db)
	request := agentWorkbenchSourceAppend(sessionID, agentWorkbenchDigest)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	start := make(chan struct{})
	results := make(chan agentworkbench.AppendResult, 2)
	errs := make(chan error, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	for i := 0; i < 2; i++ {
		go func() {
			ready.Done()
			<-start
			result, err := repo.Append(ctx, request)
			results <- result
			errs <- err
		}()
	}
	ready.Wait()
	close(start)

	firstErr, secondErr := <-errs, <-errs
	require.NoError(t, firstErr)
	require.NoError(t, secondErr)
	first, second := <-results, <-results
	require.NotEqual(t, first.Applied, second.Applied)
	state, err := repo.GetSnapshot(ctx, sessionID)
	require.NoError(t, err)
	require.Equal(t, uint64(1), state.Revision)
	events, err := repo.ListAfter(ctx, sessionID, "canonical-epoch", 0, 10)
	require.NoError(t, err)
	require.Len(t, events, 1)
	var sourceCount int64
	require.NoError(t, db.Table("agent_workbench_source_events").
		Where("session_id = ?", sessionID).Count(&sourceCount).Error)
	require.Equal(t, int64(1), sourceCount)
}

func testAgentWorkbenchSourceIngressConflicts(
	t *testing.T,
	db *gorm.DB,
	repo agentworkbench.Repository,
) {
	sessionID := insertAgentWorkbenchSession(t, db)
	ctx := context.Background()
	request := agentWorkbenchSourceAppend(sessionID, agentWorkbenchDigest)
	result, err := repo.Append(ctx, request)
	require.NoError(t, err)
	require.True(t, result.Applied)

	digestConflict := request
	digestConflict.Sources = append([]agentworkbench.SourceEvent(nil), request.Sources...)
	digestConflict.Sources[0].PayloadDigest = agentWorkbenchDigestB
	_, err = repo.Append(ctx, digestConflict)
	require.ErrorIs(t, err, agentworkbench.ErrSourceEventConflict)

	stableIDConflict := request
	stableIDConflict.Sources = append([]agentworkbench.SourceEvent(nil), request.Sources...)
	stableIDConflict.Sources[0].SourceSequence = 2
	_, err = repo.Append(ctx, stableIDConflict)
	require.ErrorIs(t, err, agentworkbench.ErrSourceEventConflict)

	sequenceConflict := request
	sequenceConflict.Sources = append([]agentworkbench.SourceEvent(nil), request.Sources...)
	sequenceConflict.Sources[0].StableEventID = "different-stable-id"
	_, err = repo.Append(ctx, sequenceConflict)
	require.ErrorIs(t, err, agentworkbench.ErrSourceEventConflict)
	assertAgentWorkbenchStateRevision(t, repo, sessionID, 1, 1)
}

func agentWorkbenchSourceAppend(
	sessionID string,
	sourceDigest string,
) agentworkbench.AppendRequest {
	return agentworkbench.AppendRequest{
		SessionID: sessionID, ExpectedRevision: 0,
		Sources: []agentworkbench.SourceEvent{{
			SessionID: sessionID, StableEventID: "runner-epoch:1",
			RunnerSessionEpoch: "runner-epoch", SourceSequence: 1,
			PayloadDigest: sourceDigest,
		}},
		Events: []agentworkbench.Event{
			newAgentWorkbenchEvent(
				sessionID,
				"canonical-epoch",
				1,
				1,
				[]byte("canonical-event"),
			),
		},
		Projection: agentworkbench.SessionState{
			SessionID: sessionID, StreamEpoch: "canonical-epoch",
			Revision: 1, LatestSequence: 1,
			Projection: []byte("canonical-projection"), Digest: agentWorkbenchDigest,
		},
	}
}
