package client

import (
	"errors"
	"testing"
	"time"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type readyResultHandler struct {
	mockHandler
	subscribeErr error
	tunnelErr    error
}

func (h *readyResultHandler) OnSubscribePod(SubscribePodRequest) error {
	return h.subscribeErr
}

func (h *readyResultHandler) OnConnectTunnel(ConnectTunnelRequest) error {
	return h.tunnelErr
}

func TestHandleSubscribePodReportsExactlyOneResult(t *testing.T) {
	tests := []struct {
		name        string
		handler     MessageHandler
		wantSuccess bool
		wantCode    string
	}{
		{name: "ready", handler: &readyResultHandler{}, wantSuccess: true},
		{
			name:     "handler failure",
			handler:  &readyResultHandler{subscribeErr: errors.New("relay rejected publisher")},
			wantCode: "relay_subscription_failed",
		},
		{name: "missing handler", wantCode: "handler_unavailable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := NewGRPCConnection("localhost:9443", "node", "org", "", "", "")
			setMockStream(conn)
			conn.handler = tt.handler

			conn.handleSubscribePod(&runnerv1.SubscribePodCommand{
				CommandId: "subscribe-1",
				PodKey:    "pod-1",
			})

			result := receiveRelaySubscriptionResult(t, conn)
			assert.Equal(t, "subscribe-1", result.GetCommandId())
			assert.Equal(t, tt.wantSuccess, result.GetSuccess())
			assert.Equal(t, tt.wantCode, result.GetErrorCode())
			assertNoControlResult(t, conn)
		})
	}
}

func TestHandleConnectTunnelReportsExactlyOneResult(t *testing.T) {
	tests := []struct {
		name        string
		handler     MessageHandler
		wantSuccess bool
		wantCode    string
	}{
		{name: "ready", handler: &readyResultHandler{}, wantSuccess: true},
		{
			name:     "handler failure",
			handler:  &readyResultHandler{tunnelErr: errors.New("gateway rejected tunnel")},
			wantCode: "tunnel_connection_failed",
		},
		{name: "missing handler", wantCode: "handler_unavailable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := NewGRPCConnection("localhost:9443", "node", "org", "", "", "")
			setMockStream(conn)
			conn.handler = tt.handler

			conn.handleConnectTunnel(&runnerv1.ConnectTunnelCommand{
				CommandId: "tunnel-1",
			})

			result := receiveTunnelConnectionResult(t, conn)
			assert.Equal(t, "tunnel-1", result.GetCommandId())
			assert.Equal(t, tt.wantSuccess, result.GetSuccess())
			assert.Equal(t, tt.wantCode, result.GetErrorCode())
			assertNoControlResult(t, conn)
		})
	}
}

func TestReadyResultDoesNotShareSaturatedControlBuffer(t *testing.T) {
	conn := NewGRPCConnection("localhost:9443", "node", "org", "", "", "")
	setMockStream(conn)
	conn.handler = &readyResultHandler{}
	for i := 0; i < cap(conn.controlCh); i++ {
		conn.controlCh <- &runnerv1.RunnerMessage{}
	}

	conn.handleSubscribePod(&runnerv1.SubscribePodCommand{
		CommandId: "subscribe-1",
		PodKey:    "pod-1",
	})

	result := receiveRelaySubscriptionResult(t, conn)
	assert.True(t, result.GetSuccess())
	assert.Equal(t, "subscribe-1", result.GetCommandId())
}

func TestCorrelatedResultsDoNotShareSaturatedControlBuffer(t *testing.T) {
	conn := NewGRPCConnection("localhost:9443", "node", "org", "", "", "")
	setMockStream(conn)
	for i := 0; i < cap(conn.controlCh); i++ {
		conn.controlCh <- &runnerv1.RunnerMessage{}
	}

	require.NoError(t, conn.SendSandboxFsResult(
		&runnerv1.SandboxFsResultEvent{RequestId: "fs-1"},
	))
	require.Equal(t, "fs-1", (<-conn.readyCh).GetSandboxFsResult().GetRequestId())
	require.NoError(t, conn.SendVerificationResult(
		&runnerv1.VerificationResultEvent{RequestId: "verify-1"},
	))
	require.Equal(
		t,
		"verify-1",
		(<-conn.readyCh).GetVerificationResult().GetRequestId(),
	)
}

func TestReadyCommandsHaveTracingTypes(t *testing.T) {
	assert.Equal(t, "SubscribePod", extractServerMessageType(&runnerv1.ServerMessage{
		Payload: &runnerv1.ServerMessage_SubscribePod{},
	}))
	assert.Equal(t, "ConnectTunnel", extractServerMessageType(&runnerv1.ServerMessage{
		Payload: &runnerv1.ServerMessage_ConnectTunnel{},
	}))
}

func TestSubscribePodCommandPreservesHistoricalReservedTags(t *testing.T) {
	descriptor := (&runnerv1.SubscribePodCommand{}).ProtoReflect().Descriptor()

	assert.Nil(t, descriptor.Fields().ByNumber(protoreflect.FieldNumber(3)))
	require.NotNil(t, descriptor.Fields().ByName("command_id"))
	assert.Equal(
		t,
		protoreflect.FieldNumber(8),
		descriptor.Fields().ByName("command_id").Number(),
	)
	assert.True(t, descriptor.ReservedNames().Has("public_relay_url"))
}

func receiveRelaySubscriptionResult(
	t *testing.T,
	conn *GRPCConnection,
) *runnerv1.RelaySubscriptionResultEvent {
	t.Helper()
	select {
	case msg := <-conn.readyCh:
		require.NotNil(t, msg.GetRelaySubscriptionResult())
		return msg.GetRelaySubscriptionResult()
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for relay subscription result")
		return nil
	}
}

func receiveTunnelConnectionResult(
	t *testing.T,
	conn *GRPCConnection,
) *runnerv1.TunnelConnectionResultEvent {
	t.Helper()
	select {
	case msg := <-conn.readyCh:
		require.NotNil(t, msg.GetTunnelConnectionResult())
		return msg.GetTunnelConnectionResult()
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for tunnel connection result")
		return nil
	}
}

func assertNoControlResult(t *testing.T, conn *GRPCConnection) {
	t.Helper()
	select {
	case msg := <-conn.readyCh:
		t.Fatalf("unexpected duplicate result: %T", msg.GetPayload())
	default:
	}
}
