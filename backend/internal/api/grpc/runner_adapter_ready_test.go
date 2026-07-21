package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/l8ai-cn/agentcloud/backend/internal/service/runner"
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
)

func TestSendSubscribePodWaitsForCorrelatedReadyResult(t *testing.T) {
	adapter, conn := newReadyTestAdapter(t, 7)
	result := make(chan error, 1)
	go func() {
		result <- adapter.SendSubscribePod(
			context.Background(), 7, "pod-1", "wss://relay/relay", "runner-token", true, 1000,
		)
	}()

	cmd := receiveSubscribeCommand(t, conn)
	require.NotEmpty(t, cmd.GetCommandId())
	assertNoReadyResult(t, result)

	adapter.handleProtoMessage(context.Background(), 7, conn, relayReadyResult(cmd.GetCommandId(), true))
	require.NoError(t, receiveReadyError(t, result))
	assert.Equal(t, "RelaySubscriptionResult", extractMessageType(relayReadyResult("id", true)))
}

func TestSendConnectTunnelReturnsRunnerFailure(t *testing.T) {
	adapter, conn := newReadyTestAdapter(t, 7)
	result := make(chan error, 1)
	go func() {
		result <- adapter.SendConnectTunnel(context.Background(), 7, "wss://relay/runner/tunnel", "tunnel-token")
	}()

	cmd := receiveTunnelCommand(t, conn)
	adapter.handleProtoMessage(
		context.Background(),
		7,
		conn,
		tunnelReadyResult(cmd.GetCommandId(), false, "gateway_unavailable"),
	)

	err := receiveReadyError(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gateway_unavailable")
	assert.NotContains(t, err.Error(), "tunnel-token")
	assert.Equal(t, "TunnelConnectionResult", extractMessageType(
		tunnelReadyResult("id", true, ""),
	))
}

func TestSendSubscribePodHonorsContextTimeout(t *testing.T) {
	adapter, conn := newReadyTestAdapter(t, 7)
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	result := make(chan error, 1)
	go func() {
		result <- adapter.SendSubscribePod(
			ctx, 7, "pod-1", "wss://relay/relay", "runner-token", true, 1000,
		)
	}()
	receiveSubscribeCommand(t, conn)

	err := receiveReadyError(t, result)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))
}

func TestReadyCommandDoesNotDispatchCanceledContext(t *testing.T) {
	adapter, conn := newReadyTestAdapter(t, 7)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := adapter.SendConnectTunnel(ctx, 7, "wss://relay/runner/tunnel", "token")

	require.ErrorIs(t, err, context.Canceled)
	select {
	case msg := <-conn.Send:
		t.Fatalf("canceled command was dispatched: %T", msg.GetPayload())
	default:
	}
}

func TestReadyCommandRejectsLegacyRunnerBeforeDispatch(t *testing.T) {
	adapter, conn := newReadyTestAdapter(t, 7)
	conn.SetProtocolVersion(2)

	err := adapter.SendConnectTunnel(
		context.Background(),
		7,
		"wss://relay/runner/tunnel",
		"token",
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "upgrade to protocol 3")
	select {
	case msg := <-conn.Send:
		t.Fatalf("legacy runner received readiness command: %T", msg.GetPayload())
	default:
	}
}

func TestSendConnectTunnelFailsWhenRunnerDisconnects(t *testing.T) {
	adapter, conn := newReadyTestAdapter(t, 7)
	result := make(chan error, 1)
	go func() {
		result <- adapter.SendConnectTunnel(context.Background(), 7, "wss://relay/runner/tunnel", "token")
	}()
	receiveTunnelCommand(t, conn)

	conn.Close()

	err := receiveReadyError(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "disconnected")
}

func TestLateAndDuplicateResultsCannotCompleteAnotherCommand(t *testing.T) {
	adapter, conn := newReadyTestAdapter(t, 7)
	firstCtx, cancelFirst := context.WithCancel(context.Background())
	firstResult := make(chan error, 1)
	go func() {
		firstResult <- adapter.SendSubscribePod(
			firstCtx, 7, "pod-1", "wss://relay/relay", "first-token", true, 1000,
		)
	}()
	first := receiveSubscribeCommand(t, conn)
	cancelFirst()
	require.Error(t, receiveReadyError(t, firstResult))

	secondResult := make(chan error, 1)
	go func() {
		secondResult <- adapter.SendSubscribePod(
			context.Background(), 7, "pod-1", "wss://relay/relay", "second-token", true, 1000,
		)
	}()
	second := receiveSubscribeCommand(t, conn)
	require.NotEqual(t, first.GetCommandId(), second.GetCommandId())

	late := relayReadyResult(first.GetCommandId(), true)
	adapter.handleProtoMessage(context.Background(), 7, conn, late)
	adapter.handleProtoMessage(context.Background(), 7, conn, late)
	assertNoReadyResult(t, secondResult)

	adapter.handleProtoMessage(context.Background(), 7, conn, relayReadyResult(second.GetCommandId(), true))
	require.NoError(t, receiveReadyError(t, secondResult))
	adapter.handleProtoMessage(context.Background(), 7, conn, relayReadyResult(second.GetCommandId(), true))
}

func TestResultFromOldConnectionGenerationIsIgnored(t *testing.T) {
	adapter, oldConn := newReadyTestAdapter(t, 7)
	newConn := adapter.connManager.AddConnection(7, "replacement", "org", &mockRunnerStream{})
	newConn.SetProtocolVersion(runnerReadyProtocolVersion)
	result := make(chan error, 1)
	go func() {
		result <- adapter.SendConnectTunnel(context.Background(), 7, "wss://relay/runner/tunnel", "token")
	}()
	cmd := receiveTunnelCommand(t, newConn)

	adapter.handleProtoMessage(
		context.Background(),
		7,
		oldConn,
		tunnelReadyResult(cmd.GetCommandId(), true, ""),
	)
	assertNoReadyResult(t, result)

	adapter.handleProtoMessage(
		context.Background(),
		7,
		newConn,
		tunnelReadyResult(cmd.GetCommandId(), true, ""),
	)
	require.NoError(t, receiveReadyError(t, result))
}

func newReadyTestAdapter(t *testing.T, runnerID int64) (*GRPCRunnerAdapter, *runner.GRPCConnection) {
	t.Helper()
	connMgr := runner.NewRunnerConnectionManager(newTestLogger())
	t.Cleanup(connMgr.Close)
	adapter := NewGRPCRunnerAdapter(newTestLogger(), nil, nil, nil, nil, nil, connMgr, nil)
	conn := connMgr.AddConnection(runnerID, "node", "org", &mockRunnerStream{})
	conn.SetProtocolVersion(runnerReadyProtocolVersion)
	return adapter, conn
}

func receiveSubscribeCommand(t *testing.T, conn *runner.GRPCConnection) *runnerv1.SubscribePodCommand {
	t.Helper()
	select {
	case msg := <-conn.Send:
		require.NotNil(t, msg.GetSubscribePod())
		return msg.GetSubscribePod()
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for subscribe command")
		return nil
	}
}

func receiveTunnelCommand(t *testing.T, conn *runner.GRPCConnection) *runnerv1.ConnectTunnelCommand {
	t.Helper()
	select {
	case msg := <-conn.Send:
		require.NotNil(t, msg.GetConnectTunnel())
		return msg.GetConnectTunnel()
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for tunnel command")
		return nil
	}
}

func relayReadyResult(commandID string, success bool) *runnerv1.RunnerMessage {
	return &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_RelaySubscriptionResult{
			RelaySubscriptionResult: &runnerv1.RelaySubscriptionResultEvent{
				CommandId: commandID,
				Success:   success,
			},
		},
	}
}

func tunnelReadyResult(commandID string, success bool, code string) *runnerv1.RunnerMessage {
	return &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_TunnelConnectionResult{
			TunnelConnectionResult: &runnerv1.TunnelConnectionResultEvent{
				CommandId: commandID,
				Success:   success,
				ErrorCode: code,
			},
		},
	}
}

func receiveReadyError(t *testing.T, result <-chan error) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for ready result")
		return nil
	}
}

func assertNoReadyResult(t *testing.T, result <-chan error) {
	t.Helper()
	select {
	case err := <-result:
		t.Fatalf("command completed before correlated result: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
}
