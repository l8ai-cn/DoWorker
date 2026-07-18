package sessionapi

import (
	"context"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

type PodUpdater interface {
	GetByKey(ctx context.Context, podKey string) (*podDomain.Pod, error)
	GetByKeyAndRunner(
		ctx context.Context,
		podKey string,
		runnerID int64,
	) (*podDomain.Pod, error)
	UpdateExternalSessionID(ctx context.Context, podKey, externalID string) error
}

// PodEventSink receives runner-originated events for agent session streaming.
type PodEventSink interface {
	HandleAcpSession(ctx context.Context, runnerID int64, podKey, eventType, payloadJSON string)
	PublishPodStatus(ctx context.Context, podKey, podStatus, agentStatus string)
	HandlePodUsage(ctx context.Context, runnerID int64, evt *runnerv1.PodUsageEvent)
	UpdateExternalSessionID(ctx context.Context, runnerID int64, podKey, externalID string)
}

var _ PodEventSink = (*SessionStreamPublisher)(nil)
