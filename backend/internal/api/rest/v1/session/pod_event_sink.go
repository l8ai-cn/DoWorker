package sessionapi

import (
	"context"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

type PodUpdater interface {
	UpdateExternalSessionID(ctx context.Context, podKey, externalID string) error
}

// PodEventSink receives runner-originated events for agent session streaming.
type PodEventSink interface {
	HandleAcpSession(ctx context.Context, podKey, eventType, payloadJSON string)
	PublishPodStatus(ctx context.Context, podKey, podStatus, agentStatus string)
	HandlePodUsage(ctx context.Context, evt *runnerv1.PodUsageEvent)
	UpdateExternalSessionID(ctx context.Context, podKey, externalID string)
}

var _ PodEventSink = (*SessionStreamPublisher)(nil)
