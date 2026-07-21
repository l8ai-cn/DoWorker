package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/l8ai-cn/agentcloud/backend/internal/service/runner"
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
)

func TestGRPCRunnerAdapter_PersistsSuccessfulTunnelConnectionResult(t *testing.T) {
	logger := newTestLogger()
	runnerSvc := newMockRunnerService()
	connMgr := runner.NewRunnerConnectionManager(logger)
	defer connMgr.Close()

	adapter := NewGRPCRunnerAdapter(logger, nil, runnerSvc, nil, nil, nil, connMgr, nil)
	conn := connMgr.AddConnection(7, "node-7", "org", &mockRunnerStream{})
	message := &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_TunnelConnectionResult{
			TunnelConnectionResult: &runnerv1.TunnelConnectionResultEvent{
				CommandId: "tunnel-7",
				Success:   true,
			},
		},
	}

	adapter.handleProtoMessage(context.Background(), 7, conn, message)

	result, recorded := runnerSvc.TunnelResult(7)
	require.True(t, recorded)
	assert.True(t, result.connected)
	assert.Empty(t, result.errorCode)
}

func TestGRPCRunnerAdapter_PersistsFailedTunnelConnectionResult(t *testing.T) {
	logger := newTestLogger()
	runnerSvc := newMockRunnerService()
	connMgr := runner.NewRunnerConnectionManager(logger)
	defer connMgr.Close()

	adapter := NewGRPCRunnerAdapter(logger, nil, runnerSvc, nil, nil, nil, connMgr, nil)
	conn := connMgr.AddConnection(9, "node-9", "org", &mockRunnerStream{})
	message := &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_TunnelConnectionResult{
			TunnelConnectionResult: &runnerv1.TunnelConnectionResultEvent{
				CommandId: "tunnel-9",
				Success:   false,
				ErrorCode: "gateway_unavailable",
				Message:   "gateway access token must not be persisted",
			},
		},
	}

	adapter.handleProtoMessage(context.Background(), 9, conn, message)

	result, recorded := runnerSvc.TunnelResult(9)
	require.True(t, recorded)
	assert.False(t, result.connected)
	assert.Equal(t, "gateway_unavailable", result.errorCode)
}

func TestGRPCRunnerAdapter_FailsTunnelReadinessWhenStatePersistenceFails(t *testing.T) {
	logger := newTestLogger()
	runnerSvc := newMockRunnerService()
	runnerSvc.err = errors.New("database unavailable")
	connMgr := runner.NewRunnerConnectionManager(logger)
	defer connMgr.Close()

	adapter := NewGRPCRunnerAdapter(logger, nil, runnerSvc, nil, nil, nil, connMgr, nil)
	conn := connMgr.AddConnection(11, "node-11", "org", &mockRunnerStream{})
	conn.SetProtocolVersion(runnerReadyProtocolVersion)
	result := make(chan error, 1)
	go func() {
		result <- adapter.SendConnectTunnel(context.Background(), 11, "wss://relay/runner/tunnel", "token")
	}()

	command := receiveTunnelCommand(t, conn)
	adapter.handleProtoMessage(context.Background(), 11, conn, &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_TunnelConnectionResult{
			TunnelConnectionResult: &runnerv1.TunnelConnectionResultEvent{
				CommandId: command.GetCommandId(),
				Success:   true,
				Message:   "source result must not be returned",
			},
		},
	})

	err := receiveReadyError(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tunnel_state_persistence_failed")
	assert.NotContains(t, err.Error(), "source result must not be returned")
}
