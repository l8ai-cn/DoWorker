package grpc

import (
	"context"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/service/runner"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/require"
)

type recordingPodEventSink struct {
	acpRunnerID      int64
	usageRunnerID    int64
	externalRunnerID int64
	podKey           string
}

func (s *recordingPodEventSink) HandleAcpSession(
	_ context.Context,
	runnerID int64,
	podKey, _, _ string,
) {
	s.acpRunnerID = runnerID
	s.podKey = podKey
}

func (*recordingPodEventSink) PublishPodStatus(context.Context, string, string, string) {}
func (s *recordingPodEventSink) HandlePodUsage(
	_ context.Context,
	runnerID int64,
	_ *runnerv1.PodUsageEvent,
) {
	s.usageRunnerID = runnerID
}
func (s *recordingPodEventSink) UpdateExternalSessionID(
	_ context.Context,
	runnerID int64,
	_, _ string,
) {
	s.externalRunnerID = runnerID
}

func TestHandleProtoMessageForwardsAuthenticatedRunnerIDToPodEventSink(t *testing.T) {
	logger := newTestLogger()
	connManager := runner.NewRunnerConnectionManager(logger)
	defer connManager.Close()
	sink := &recordingPodEventSink{}
	adapter := NewGRPCRunnerAdapter(
		logger,
		nil,
		newMockRunnerService(),
		newMockOrgService(),
		nil,
		nil,
		connManager,
		nil,
	)
	adapter.podEvents = sink
	conn := connManager.AddConnection(35, "seedance-runner", "dev-org", &mockRunnerStream{})

	adapter.handleProtoMessage(context.Background(), 35, conn, &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_AcpSession{
			AcpSession: &runnerv1.AcpSessionEvent{
				PodKey:      "7-standalone-caa70224",
				EventType:   "contentChunk",
				JsonPayload: `{"role":"assistant","text":"ok"}`,
			},
		},
	})
	adapter.handleProtoMessage(context.Background(), 35, conn, &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_PodUsage{
			PodUsage: &runnerv1.PodUsageEvent{PodKey: "7-standalone-caa70224"},
		},
	})
	adapter.handleProtoMessage(context.Background(), 35, conn, &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_ExternalSessionCaptured{
			ExternalSessionCaptured: &runnerv1.ExternalSessionCapturedEvent{
				PodKey: "7-standalone-caa70224", ExternalSessionId: "ark-session",
			},
		},
	})

	require.Equal(t, int64(35), sink.acpRunnerID)
	require.Equal(t, int64(35), sink.usageRunnerID)
	require.Equal(t, int64(35), sink.externalRunnerID)
	require.Equal(t, "7-standalone-caa70224", sink.podKey)
}
