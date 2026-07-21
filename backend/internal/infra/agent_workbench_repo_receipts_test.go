package infra

import (
	"context"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentworkbench"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const agentWorkbenchDigestB = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

func testAgentWorkbenchCommandReceipts(
	t *testing.T,
	db *gorm.DB,
	repo agentworkbench.Repository,
) {
	sessionID := insertAgentWorkbenchSession(t, db)
	ctx := context.Background()
	received := agentworkbench.CommandReceipt{
		SessionID: sessionID, CommandID: "command-1",
		PayloadDigest: agentWorkbenchDigest,
		State:         agentworkbench.ReceiptStateReceived,
		Receipt:       []byte{0x08, 0x01},
	}

	stored, err := repo.PutCommandReceipt(ctx, received)
	require.NoError(t, err)
	require.Equal(t, received.Receipt, stored.Receipt)
	loaded, err := repo.GetCommandReceipt(ctx, sessionID, received.CommandID)
	require.NoError(t, err)
	require.Equal(t, stored, loaded)

	replayed, err := repo.PutCommandReceipt(ctx, received)
	require.NoError(t, err)
	require.Equal(t, stored, replayed)
	digestConflict := received
	digestConflict.PayloadDigest = agentWorkbenchDigestB
	_, err = repo.PutCommandReceipt(ctx, digestConflict)
	require.ErrorIs(
		t,
		err,
		agentworkbench.ErrCommandIDConflict,
	)

	accepted := received
	accepted.State = agentworkbench.ReceiptStateAccepted
	accepted.Receipt = []byte{0x08, 0x02}
	_, err = repo.PutCommandReceipt(ctx, accepted)
	require.NoError(t, err)
	terminal := accepted
	terminal.State = agentworkbench.ReceiptStateSucceeded
	terminal.Receipt = []byte{0x08, 0x04}
	storedTerminal, err := repo.PutCommandReceipt(ctx, terminal)
	require.NoError(t, err)
	replayedTerminal, err := repo.PutCommandReceipt(ctx, terminal)
	require.NoError(t, err)
	require.Equal(t, storedTerminal, replayedTerminal)

	changedTerminal := terminal
	changedTerminal.Receipt = []byte{0x08, 0x04, 0x01}
	_, err = repo.PutCommandReceipt(ctx, changedTerminal)
	require.ErrorIs(
		t,
		err,
		agentworkbench.ErrReceiptConflict,
	)
	rollback := terminal
	rollback.State = agentworkbench.ReceiptStateRunning
	_, err = repo.PutCommandReceipt(ctx, rollback)
	require.ErrorIs(
		t,
		err,
		agentworkbench.ErrReceiptConflict,
	)

	failed := received
	failed.CommandID = "command-failed"
	_, err = repo.PutCommandReceipt(ctx, failed)
	require.NoError(t, err)
	failed.State = agentworkbench.ReceiptStateFailed
	failed.Receipt = []byte{0x08, 0x05}
	storedFailed, err := repo.PutCommandReceipt(ctx, failed)
	require.NoError(t, err)
	require.Equal(t, agentworkbench.ReceiptStateFailed, storedFailed.State)
}

func testAgentWorkbenchAppendRollsBackReceiptOnEventConflict(
	t *testing.T,
	db *gorm.DB,
	repo agentworkbench.Repository,
) {
	sessionID := insertAgentWorkbenchSession(t, db)
	ctx := context.Background()
	received := agentworkbench.CommandReceipt{
		SessionID: sessionID, CommandID: "command-atomic",
		PayloadDigest: agentWorkbenchDigest,
		State:         agentworkbench.ReceiptStateReceived,
		Receipt:       []byte("received"),
	}
	_, err := repo.PutCommandReceipt(ctx, received)
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
INSERT INTO agent_workbench_events
	(session_id, stream_epoch, sequence, revision, payload, digest, created_at)
VALUES (?, 'epoch-a', 1, 1, ?, ?, NOW())`,
		sessionID, []byte("preexisting"), agentWorkbenchDigest).Error)

	accepted := received
	accepted.State = agentworkbench.ReceiptStateAccepted
	accepted.Receipt = []byte("accepted")
	event := newAgentWorkbenchEvent(sessionID, "epoch-a", 1, 1, []byte("event"))
	_, err = repo.Append(ctx, agentworkbench.AppendRequest{
		SessionID: sessionID, ExpectedRevision: 0,
		Receipts: []agentworkbench.CommandReceipt{accepted},
		Events:   []agentworkbench.Event{event},
		Projection: agentworkbench.SessionState{
			SessionID: sessionID, StreamEpoch: "epoch-a", Revision: 1,
			LatestSequence: 1, Projection: []byte("projection"), Digest: agentWorkbenchDigest,
		},
	})
	require.ErrorIs(t, err, agentworkbench.ErrEventConflict)
	stored, err := repo.GetCommandReceipt(ctx, sessionID, received.CommandID)
	require.NoError(t, err)
	require.Equal(t, agentworkbench.ReceiptStateReceived, stored.State)
	require.Equal(t, received.Receipt, stored.Receipt)
}
