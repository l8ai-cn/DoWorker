package agentworkbench

import (
	"context"
	"errors"
	"fmt"

	poddomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	sessiondomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentsession"
	domainworkbench "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentworkbench"
	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
	"google.golang.org/protobuf/proto"
)

var ErrIngressConfiguration = errors.New("invalid agent workbench ingress configuration")

const ingressAppendAttempts = 8

type SessionResolver interface {
	GetByPodKey(context.Context, string) (*sessiondomain.Session, error)
}

type PodResolver interface {
	GetByKeyAndRunner(context.Context, string, int64) (*poddomain.Pod, error)
}

type DeltaPublisher interface {
	Publish(string, *agentworkbenchv2.SessionDeltaBatch)
}

type ArtifactMaterializer interface {
	Materialize(
		context.Context,
		int64,
		string,
		string,
		*agentworkbenchv2.RunnerWorkbenchEventBatch,
	) (*ArtifactMaterialization, error)
}

type Ingress struct {
	sessions     SessionResolver
	pods         PodResolver
	repository   domainworkbench.Repository
	publisher    DeltaPublisher
	materializer ArtifactMaterializer
	newEpoch     func() string
}

func NewIngress(
	sessions SessionResolver,
	pods PodResolver,
	repository domainworkbench.Repository,
	publisher DeltaPublisher,
	materializer ArtifactMaterializer,
	newEpoch func() string,
) (*Ingress, error) {
	if sessions == nil || pods == nil || repository == nil || publisher == nil ||
		materializer == nil || newEpoch == nil {
		return nil, ErrIngressConfiguration
	}
	return &Ingress{
		sessions: sessions, pods: pods, repository: repository,
		publisher: publisher, materializer: materializer, newEpoch: newEpoch,
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
	pod, err := ingress.pods.GetByKeyAndRunner(ctx, batch.PodKey, runnerID)
	if err != nil {
		return fmt.Errorf("authorize agent workbench runner: %w", err)
	}
	if pod == nil || pod.PodKey != batch.PodKey ||
		pod.RunnerID != runnerID ||
		pod.OrganizationID != session.OrganizationID {
		return ErrInvalidBatch
	}
	materialized, err := ingress.materializer.Materialize(
		ctx,
		runnerID,
		session.ID,
		session.PodKey,
		batch,
	)
	if err != nil {
		return fmt.Errorf("materialize agent workbench artifacts: %w", err)
	}
	err = ingress.appendMaterialized(ctx, session.ID, materialized)
	if err != nil {
		return errors.Join(
			err,
			materialized.Abort(context.WithoutCancel(ctx)),
		)
	}
	if err := materialized.Reconcile(context.WithoutCancel(ctx)); err != nil {
		return fmt.Errorf("reconcile agent workbench artifacts: %w", err)
	}
	return nil
}

func (ingress *Ingress) appendMaterialized(
	ctx context.Context,
	sessionID string,
	materialized *ArtifactMaterialization,
) error {
	for attempt := 0; attempt < ingressAppendAttempts; attempt++ {
		stored, err := ingress.repository.GetSnapshot(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("load agent workbench snapshot: %w", err)
		}
		current, streamEpoch, err := ingress.decodeSnapshot(sessionID, stored)
		if err != nil {
			return err
		}
		projected, err := ProjectRunnerBatch(
			current,
			sessionID,
			streamEpoch,
			materialized.Batch,
		)
		if err != nil {
			return err
		}
		request, err := appendRequest(
			sessionID,
			stored,
			materialized.Batch,
			projected,
		)
		if err != nil {
			return err
		}
		request.ArtifactFiles = materialized.Files
		result, err := ingress.repository.Append(ctx, request)
		if errors.Is(err, domainworkbench.ErrRevisionConflict) {
			continue
		}
		if err != nil {
			return fmt.Errorf("append agent workbench batch: %w", err)
		}
		if result.Applied {
			ingress.publisher.Publish(sessionID, projected.Delta)
		}
		return nil
	}
	return domainworkbench.ErrRevisionConflict
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
