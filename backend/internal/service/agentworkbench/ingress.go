package agentworkbench

import (
	"context"
	"errors"
	"fmt"

	sessiondomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	domainworkbench "github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"google.golang.org/protobuf/proto"
)

var ErrIngressConfiguration = errors.New("invalid agent workbench ingress configuration")

type SessionResolver interface {
	GetByPodKey(context.Context, string) (*sessiondomain.Session, error)
}

type DeltaPublisher interface {
	Publish(string, *agentworkbenchv2.SessionDeltaBatch)
}

type Ingress struct {
	sessions   SessionResolver
	repository domainworkbench.Repository
	publisher  DeltaPublisher
	newEpoch   func() string
}

func NewIngress(
	sessions SessionResolver,
	repository domainworkbench.Repository,
	publisher DeltaPublisher,
	newEpoch func() string,
) (*Ingress, error) {
	if sessions == nil || repository == nil || publisher == nil || newEpoch == nil {
		return nil, ErrIngressConfiguration
	}
	return &Ingress{
		sessions: sessions, repository: repository,
		publisher: publisher, newEpoch: newEpoch,
	}, nil
}

func (ingress *Ingress) HandleWorkbenchEvents(
	ctx context.Context,
	runnerID int64,
	batch *agentworkbenchv2.RunnerWorkbenchEventBatch,
) error {
	return ingress.Ingest(ctx, runnerID, batch)
}

func (ingress *Ingress) Ingest(
	ctx context.Context,
	runnerID int64,
	batch *agentworkbenchv2.RunnerWorkbenchEventBatch,
) error {
	if runnerID <= 0 || batch == nil || batch.PodKey == "" {
		return ErrInvalidBatch
	}
	session, err := ingress.sessions.GetByPodKey(ctx, batch.PodKey)
	if err != nil {
		return fmt.Errorf("resolve agent session for pod %q: %w", batch.PodKey, err)
	}
	if session == nil || session.ID == "" || session.PodKey != batch.PodKey {
		return ErrInvalidBatch
	}
	stored, err := ingress.repository.GetSnapshot(ctx, session.ID)
	if err != nil {
		return fmt.Errorf("load agent workbench snapshot: %w", err)
	}
	current, streamEpoch, err := ingress.decodeSnapshot(session.ID, stored)
	if err != nil {
		return err
	}
	projected, err := ProjectRunnerBatch(current, session.ID, streamEpoch, batch)
	if err != nil {
		return err
	}
	request, err := appendRequest(session.ID, stored, batch, projected)
	if err != nil {
		return err
	}
	result, err := ingress.repository.Append(ctx, request)
	if err != nil {
		return fmt.Errorf("append agent workbench batch: %w", err)
	}
	if result.Applied {
		ingress.publisher.Publish(session.ID, projected.Delta)
	}
	return nil
}

func (ingress *Ingress) decodeSnapshot(
	sessionID string,
	stored *domainworkbench.SessionState,
) (*agentworkbenchv2.SessionSnapshot, string, error) {
	if stored == nil {
		streamEpoch := ingress.newEpoch()
		if streamEpoch == "" {
			return nil, "", ErrIngressConfiguration
		}
		return nil, streamEpoch, nil
	}
	snapshot := &agentworkbenchv2.SessionSnapshot{}
	if err := proto.Unmarshal(stored.Projection, snapshot); err != nil {
		return nil, "", fmt.Errorf("decode agent workbench snapshot: %w", err)
	}
	if stored.SessionID != sessionID ||
		stored.StreamEpoch != snapshot.StreamEpoch ||
		stored.Revision != snapshot.Revision ||
		stored.LatestSequence != snapshot.LatestSequence ||
		stored.Digest != snapshot.GetDigest() {
		return nil, "", ErrInvalidBatch
	}
	return snapshot, stored.StreamEpoch, nil
}
