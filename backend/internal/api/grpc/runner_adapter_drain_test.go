package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anthropics/agentsmesh/backend/internal/service/runner"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

// Transport half of the reconnect-drain path (RFC-006 §5.5): the drainer's
// ServerMessageSender must deliver queued create_pod payloads over a real
// stream in FIFO order, and fail fast when the runner is not connected so the
// drainer keeps rows for the next attempt. The queue state machine itself is
// covered in service/runner drain tests.
func TestRunnerReconnect_ServerMessagesArriveFIFO(t *testing.T) {
	logger := newTestLogger()
	runnerSvc := newMockRunnerService()
	orgSvc := newMockOrgService()
	connMgr := runner.NewRunnerConnectionManager(logger)
	defer connMgr.Close()

	runnerSvc.AddRunner("drain-node", RunnerInfo{
		ID: 7, NodeID: "drain-node", OrganizationID: 100, IsEnabled: true,
	})
	orgSvc.AddOrg("test-org", OrganizationInfo{ID: 100, Slug: "test-org"})

	adapter := NewGRPCRunnerAdapter(logger, nil, runnerSvc, orgSvc, nil, nil, connMgr, nil)
	addr, cleanup := setupTestServer(t, adapter)
	defer cleanup()

	sender := NewGRPCCommandSender(adapter)
	ctx := context.Background()

	err := sender.SendServerMessage(ctx, 7, &runnerv1.ServerMessage{
		Payload: &runnerv1.ServerMessage_CreatePod{CreatePod: &runnerv1.CreatePodCommand{PodKey: "pd-early"}},
	})
	require.Error(t, err, "send before connect must fail so the drainer retains the row")

	stream, conn, cancel := connectRunner(t, addr, "drain-node", "test-org")
	defer cancel()
	defer conn.Close()
	completeHandshake(t, stream, []string{"claude-code"})
	time.Sleep(50 * time.Millisecond)
	require.True(t, connMgr.IsConnected(7))

	for _, key := range []string{"pd-1", "pd-2"} {
		require.NoError(t, sender.SendServerMessage(ctx, 7, &runnerv1.ServerMessage{
			Payload: &runnerv1.ServerMessage_CreatePod{CreatePod: &runnerv1.CreatePodCommand{PodKey: key}},
		}))
	}

	var got []string
	for len(got) < 2 {
		msg, err := stream.Recv()
		require.NoError(t, err)
		if cp := msg.GetCreatePod(); cp != nil {
			got = append(got, cp.PodKey)
		}
	}
	assert.Equal(t, []string{"pd-1", "pd-2"}, got)

	_ = stream.CloseSend()
}
