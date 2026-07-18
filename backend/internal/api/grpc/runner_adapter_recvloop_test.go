package grpc

import (
	"context"
	"io"
	"testing"
	"time"

	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/anthropics/agentsmesh/backend/internal/service/runner"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

// ==================== receiveLoop Tests ====================

// mockRecvStream is used for testing receiveLoop
type mockRecvStream struct {
	msgs    []*runnerv1.RunnerMessage
	recvIdx int
	recvErr error
	ctx     context.Context
}

func (m *mockRecvStream) Send(msg *runnerv1.ServerMessage) error {
	return nil
}

func (m *mockRecvStream) Recv() (*runnerv1.RunnerMessage, error) {
	if m.recvErr != nil {
		return nil, m.recvErr
	}
	if m.recvIdx >= len(m.msgs) {
		return nil, io.EOF
	}
	msg := m.msgs[m.recvIdx]
	m.recvIdx++
	return msg, nil
}

func (m *mockRecvStream) Context() context.Context {
	if m.ctx != nil {
		return m.ctx
	}
	return context.Background()
}

func (m *mockRecvStream) SetHeader(metadata.MD) error  { return nil }
func (m *mockRecvStream) SendHeader(metadata.MD) error { return nil }
func (m *mockRecvStream) SetTrailer(metadata.MD)       {}
func (m *mockRecvStream) SendMsg(interface{}) error    { return nil }
func (m *mockRecvStream) RecvMsg(interface{}) error    { return nil }

func TestGRPCRunnerAdapter_ReceiveLoop_EOF(t *testing.T) {
	logger := newTestLogger()
	connMgr := runner.NewRunnerConnectionManager(logger)
	defer connMgr.Close()

	adapter := NewGRPCRunnerAdapter(logger, nil, nil, nil, nil, nil, connMgr, nil)

	// Create connection
	mockStream := &mockRunnerStream{}
	conn := connMgr.AddConnection(1, "test-node", "test-org", mockStream)

	// Create mock stream that returns EOF immediately
	recvStream := &mockRecvStream{
		msgs: []*runnerv1.RunnerMessage{}, // Empty, will return EOF
	}

	// Run receiveLoop
	err := adapter.receiveLoop(context.Background(), 1, conn, recvStream)

	// Should return nil on EOF (graceful disconnect)
	assert.NoError(t, err)
}

func TestGRPCRunnerAdapter_ReceiveLoop_Canceled(t *testing.T) {
	logger := newTestLogger()
	connMgr := runner.NewRunnerConnectionManager(logger)
	defer connMgr.Close()

	adapter := NewGRPCRunnerAdapter(logger, nil, nil, nil, nil, nil, connMgr, nil)

	// Create connection
	mockStream := &mockRunnerStream{}
	conn := connMgr.AddConnection(1, "test-node", "test-org", mockStream)

	// Create mock stream that returns gRPC Canceled status error
	// Note: must use gRPC status.Error, not context.Canceled
	recvStream := &mockRecvStream{
		recvErr: status.Error(codes.Canceled, "context canceled"),
	}

	// Run receiveLoop
	err := adapter.receiveLoop(context.Background(), 1, conn, recvStream)

	// Should return nil on Canceled (graceful disconnect)
	assert.NoError(t, err)
}

func TestGRPCRunnerAdapter_ReceiveLoop_OtherError(t *testing.T) {
	logger := newTestLogger()
	connMgr := runner.NewRunnerConnectionManager(logger)
	defer connMgr.Close()

	adapter := NewGRPCRunnerAdapter(logger, nil, nil, nil, nil, nil, connMgr, nil)

	// Create connection
	mockStream := &mockRunnerStream{}
	conn := connMgr.AddConnection(1, "test-node", "test-org", mockStream)

	// Create mock stream that returns an unexpected error
	recvStream := &mockRecvStream{
		recvErr: context.DeadlineExceeded, // Not EOF or Canceled
	}

	// Run receiveLoop
	err := adapter.receiveLoop(context.Background(), 1, conn, recvStream)

	// Should return the error
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestGRPCRunnerAdapter_ReceiveLoop_ProcessMessages(t *testing.T) {
	logger := newTestLogger()
	connMgr := runner.NewRunnerConnectionManager(logger)
	defer connMgr.Close()

	adapter := NewGRPCRunnerAdapter(logger, nil, nil, nil, nil, nil, connMgr, nil)

	// Create connection
	mockStream := &mockRunnerStream{}
	conn := connMgr.AddConnection(1, "test-node", "test-org", mockStream)

	// Track received heartbeats
	var heartbeatCount int
	connMgr.SetHeartbeatCallback(func(runnerID int64, data *runnerv1.HeartbeatData) {
		heartbeatCount++
	})

	// Create mock stream with messages
	recvStream := &mockRecvStream{
		msgs: []*runnerv1.RunnerMessage{
			{
				Payload: &runnerv1.RunnerMessage_Heartbeat{
					Heartbeat: &runnerv1.HeartbeatData{NodeId: "test"},
				},
			},
			{
				Payload: &runnerv1.RunnerMessage_Heartbeat{
					Heartbeat: &runnerv1.HeartbeatData{NodeId: "test"},
				},
			},
		},
	}

	// Run receiveLoop
	err := adapter.receiveLoop(context.Background(), 1, conn, recvStream)

	// Should return nil after processing all messages and hitting EOF
	assert.NoError(t, err)
	assert.Equal(t, 2, heartbeatCount)
}

type blockingWorkbenchSink struct {
	resultHandled <-chan struct{}
	completed     chan struct{}
}

func (sink *blockingWorkbenchSink) HandleWorkbenchEvents(
	ctx context.Context,
	_ int64,
	_ *agentworkbenchv2.RunnerWorkbenchEventBatch,
) error {
	defer close(sink.completed)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-sink.resultHandled:
		return nil
	}
}

type resultAwareRecvStream struct {
	mockRecvStream
	messages []*runnerv1.RunnerMessage
	done     <-chan struct{}
}

func (stream *resultAwareRecvStream) Recv() (*runnerv1.RunnerMessage, error) {
	if len(stream.messages) > 0 {
		message := stream.messages[0]
		stream.messages = stream.messages[1:]
		return message, nil
	}
	<-stream.done
	return nil, io.EOF
}

func TestGRPCRunnerAdapter_ReceiveLoopHandlesResultWhileWorkbenchWaits(
	t *testing.T,
) {
	logger := newTestLogger()
	connManager := runner.NewRunnerConnectionManager(logger)
	defer connManager.Close()
	resultHandled := make(chan struct{})
	completed := make(chan struct{})
	connManager.SetSandboxFsResultCallback(
		func(_ int64, _ *runnerv1.SandboxFsResultEvent) {
			close(resultHandled)
		},
	)
	adapter := NewGRPCRunnerAdapter(
		logger, nil, nil, nil, nil, nil, connManager, nil,
	)
	adapter.SetWorkbenchEventSink(&blockingWorkbenchSink{
		resultHandled: resultHandled,
		completed:     completed,
	})
	conn := connManager.AddConnection(
		1, "test-node", "test-org", &mockRunnerStream{},
	)
	stream := &resultAwareRecvStream{
		done: completed,
		messages: []*runnerv1.RunnerMessage{
			{
				Payload: &runnerv1.RunnerMessage_WorkbenchEvents{
					WorkbenchEvents: &agentworkbenchv2.RunnerWorkbenchEventBatch{
						PodKey: "pod-1",
					},
				},
			},
			{
				Payload: &runnerv1.RunnerMessage_SandboxFsResult{
					SandboxFsResult: &runnerv1.SandboxFsResultEvent{
						RequestId: "request-1",
					},
				},
			},
		},
	}

	finished := make(chan error, 1)
	go func() {
		finished <- adapter.receiveLoop(
			context.Background(), 1, conn, stream,
		)
	}()

	select {
	case err := <-finished:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("receive loop blocked behind workbench artifact handling")
	}
}
