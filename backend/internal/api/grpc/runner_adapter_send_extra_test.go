package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anthropics/agentsmesh/backend/internal/service/runner"
)

// ==================== Additional Send Tests ====================

func TestGRPCRunnerAdapter_SendSubscribePod(t *testing.T) {
	logger := newTestLogger()
	connMgr := runner.NewRunnerConnectionManager(logger)
	defer connMgr.Close()

	adapter := NewGRPCRunnerAdapter(logger, nil, nil, nil, nil, nil, connMgr, nil)

	t.Run("runner not connected", func(t *testing.T) {
		err := adapter.SendSubscribePod(context.Background(), 999, "pod-1", "ws://relay", "token", true, 100)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("successful send", func(t *testing.T) {
		mockStream := &mockRunnerStream{}
		conn := connMgr.AddConnection(2, "test-node", "test-org", mockStream)
		conn.SetProtocolVersion(runnerReadyProtocolVersion)

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		err := adapter.SendSubscribePod(ctx, 2, "pod-1", "ws://relay", "token", true, 100)
		require.ErrorIs(t, err, context.DeadlineExceeded)
	})
}

func TestGRPCRunnerAdapter_SendConnectTunnel(t *testing.T) {
	logger := newTestLogger()
	connMgr := runner.NewRunnerConnectionManager(logger)
	defer connMgr.Close()

	adapter := NewGRPCRunnerAdapter(logger, nil, nil, nil, nil, nil, connMgr, nil)

	t.Run("runner not connected", func(t *testing.T) {
		err := adapter.SendConnectTunnel(context.Background(), 999, "wss://d/relay", "tok")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("successful send", func(t *testing.T) {
		mockStream := &mockRunnerStream{}
		conn := connMgr.AddConnection(1, "n", "o", mockStream)
		conn.SetProtocolVersion(runnerReadyProtocolVersion)

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		err := adapter.SendConnectTunnel(ctx, 1, "wss://d/relay", "tok")
		require.ErrorIs(t, err, context.DeadlineExceeded)
	})
}

func TestGRPCRunnerAdapter_SendUnsubscribePod(t *testing.T) {
	logger := newTestLogger()
	connMgr := runner.NewRunnerConnectionManager(logger)
	defer connMgr.Close()

	adapter := NewGRPCRunnerAdapter(logger, nil, nil, nil, nil, nil, connMgr, nil)

	t.Run("runner not connected", func(t *testing.T) {
		err := adapter.SendUnsubscribePod(999, "pod-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("successful send", func(t *testing.T) {
		mockStream := &mockRunnerStream{}
		connMgr.AddConnection(3, "test-node", "test-org", mockStream)

		err := adapter.SendUnsubscribePod(3, "pod-1")
		require.NoError(t, err)
	})
}

func TestGRPCRunnerAdapter_SendQuerySandboxes(t *testing.T) {
	logger := newTestLogger()
	connMgr := runner.NewRunnerConnectionManager(logger)
	defer connMgr.Close()

	adapter := NewGRPCRunnerAdapter(logger, nil, nil, nil, nil, nil, connMgr, nil)

	t.Run("runner not connected", func(t *testing.T) {
		err := adapter.SendQuerySandboxes(999, "req-1", []string{"pod-1", "pod-2"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("successful send", func(t *testing.T) {
		mockStream := &mockRunnerStream{}
		connMgr.AddConnection(4, "test-node", "test-org", mockStream)

		err := adapter.SendQuerySandboxes(4, "req-1", []string{"pod-1", "pod-2"})
		require.NoError(t, err)
	})
}
