package grpc

import (
	"context"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/service/runner"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/require"
)

type recordingWorkbenchEventSink struct {
	runnerID int64
	batch    *agentworkbenchv2.RunnerWorkbenchEventBatch
}

func (sink *recordingWorkbenchEventSink) HandleWorkbenchEvents(
	_ context.Context,
	runnerID int64,
	batch *agentworkbenchv2.RunnerWorkbenchEventBatch,
) error {
	sink.runnerID = runnerID
	sink.batch = batch
	return nil
}

func TestGRPCRunnerAdapterRoutesWorkbenchEvents(t *testing.T) {
	logger := newTestLogger()
	connManager := runner.NewRunnerConnectionManager(logger)
	defer connManager.Close()
	adapter := NewGRPCRunnerAdapter(logger, nil, nil, nil, nil, nil, connManager, nil)
	sink := &recordingWorkbenchEventSink{}
	adapter.SetWorkbenchEventSink(sink)
	conn := connManager.AddConnection(17, "runner-17", "org", &mockRunnerStream{})
	batch := &agentworkbenchv2.RunnerWorkbenchEventBatch{
		PodKey:             "pod-1",
		RunnerSessionEpoch: "runner-epoch-1",
	}

	adapter.handleProtoMessage(context.Background(), 17, conn, &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_WorkbenchEvents{WorkbenchEvents: batch},
	})

	require.Equal(t, int64(17), sink.runnerID)
	require.Same(t, batch, sink.batch)
	require.Equal(
		t,
		"WorkbenchEvents",
		extractMessageType(&runnerv1.RunnerMessage{
			Payload: &runnerv1.RunnerMessage_WorkbenchEvents{WorkbenchEvents: batch},
		}),
	)
}
