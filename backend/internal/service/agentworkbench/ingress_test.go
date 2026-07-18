package agentworkbench

import (
	"context"
	"testing"

	poddomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	sessiondomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	domainworkbench "github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"github.com/stretchr/testify/require"
)

type ingressSessionResolver struct {
	session *sessiondomain.Session
	err     error
}

type ingressPodResolver struct {
	pod *poddomain.Pod
	err error
}

func (resolver ingressPodResolver) GetByKeyAndRunner(
	context.Context,
	string,
	int64,
) (*poddomain.Pod, error) {
	return resolver.pod, resolver.err
}

func (resolver ingressSessionResolver) GetByPodKey(
	context.Context,
	string,
) (*sessiondomain.Session, error) {
	return resolver.session, resolver.err
}

type ingressRepository struct {
	state              *domainworkbench.SessionState
	stateAfterConflict *domainworkbench.SessionState
	request            domainworkbench.AppendRequest
	requests           []domainworkbench.AppendRequest
	result             domainworkbench.AppendResult
	err                error
	appendErrors       []error
}

func (repo *ingressRepository) Append(
	_ context.Context,
	request domainworkbench.AppendRequest,
) (domainworkbench.AppendResult, error) {
	repo.request = request
	repo.requests = append(repo.requests, request)
	if len(repo.appendErrors) > 0 {
		err := repo.appendErrors[0]
		repo.appendErrors = repo.appendErrors[1:]
		if err == domainworkbench.ErrRevisionConflict &&
			repo.stateAfterConflict != nil {
			repo.state = repo.stateAfterConflict
		}
		return domainworkbench.AppendResult{}, err
	}
	return repo.result, repo.err
}

func (repo *ingressRepository) GetSnapshot(
	context.Context,
	string,
) (*domainworkbench.SessionState, error) {
	return repo.state, nil
}

func (*ingressRepository) ListAfter(
	context.Context,
	string,
	string,
	uint64,
	int,
) ([]domainworkbench.Event, error) {
	return nil, nil
}

func (*ingressRepository) PutCommandReceipt(
	context.Context,
	domainworkbench.CommandReceipt,
) (*domainworkbench.CommandReceipt, error) {
	return nil, nil
}

func (*ingressRepository) GetCommandReceipt(
	context.Context,
	string,
	string,
) (*domainworkbench.CommandReceipt, error) {
	return nil, nil
}

type ingressPublisher struct {
	sessionID string
	delta     *agentworkbenchv2.SessionDeltaBatch
	count     int
}

type ingressArtifactMaterializer struct {
	count *int
}

func (materializer ingressArtifactMaterializer) Materialize(
	_ context.Context,
	_ int64,
	_ string,
	_ string,
	batch *agentworkbenchv2.RunnerWorkbenchEventBatch,
) (*ArtifactMaterialization, error) {
	if materializer.count != nil {
		(*materializer.count)++
	}
	return &ArtifactMaterialization{Batch: batch}, nil
}

func (publisher *ingressPublisher) Publish(
	sessionID string,
	delta *agentworkbenchv2.SessionDeltaBatch,
) {
	publisher.sessionID = sessionID
	publisher.delta = delta
	publisher.count++
}

func TestIngressPersistsAndPublishesRichArtifactBatch(t *testing.T) {
	repo := &ingressRepository{
		result: domainworkbench.AppendResult{Applied: true},
	}
	publisher := &ingressPublisher{}
	service, err := NewIngress(
		ingressSessionResolver{session: &sessiondomain.Session{
			ID: "conv_1", PodKey: "pod-1", OrganizationID: 9,
		}},
		ingressPodResolver{pod: &poddomain.Pod{
			PodKey: "pod-1", RunnerID: 17, OrganizationID: 9,
		}},
		repo,
		publisher,
		ingressArtifactMaterializer{},
		func() string { return "stream-1" },
	)
	require.NoError(t, err)

	err = service.Ingest(context.Background(), 17, runnerBatch(
		artifactMutation("source:1", 1, imageArtifact("source", "result")),
		artifactMutation("source:2", 2, videoArtifact()),
	))
	require.NoError(t, err)
	require.Equal(t, "conv_1", repo.request.SessionID)
	require.Equal(t, uint64(0), repo.request.ExpectedRevision)
	require.Len(t, repo.request.Sources, 2)
	require.Len(t, repo.request.Events, 2)
	require.Equal(t, "runner-epoch-1", repo.request.Sources[0].RunnerSessionEpoch)
	require.Equal(t, uint64(1), repo.request.Sources[0].SourceSequence)
	require.NotEmpty(t, repo.request.Sources[0].PayloadDigest)
	require.NotEmpty(t, repo.request.Events[0].Payload)
	require.NotEmpty(t, repo.request.Projection.Projection)
	require.Equal(t, "conv_1", publisher.sessionID)
	require.Equal(t, uint64(1), publisher.delta.Revision)
}

func TestIngressDoesNotRepublishDuplicateSourceBatch(t *testing.T) {
	repo := &ingressRepository{result: domainworkbench.AppendResult{Applied: false}}
	publisher := &ingressPublisher{}
	service, err := NewIngress(
		ingressSessionResolver{session: &sessiondomain.Session{
			ID: "conv_1", PodKey: "pod-1", OrganizationID: 9,
		}},
		ingressPodResolver{pod: &poddomain.Pod{
			PodKey: "pod-1", RunnerID: 17, OrganizationID: 9,
		}},
		repo,
		publisher,
		ingressArtifactMaterializer{},
		func() string { return "stream-1" },
	)
	require.NoError(t, err)

	err = service.Ingest(context.Background(), 17, runnerBatch(
		statusMutation("source:1", 1, agentworkbenchv2.SessionStatus_SESSION_STATUS_RUNNING),
	))
	require.NoError(t, err)
	require.Nil(t, publisher.delta)
}

func TestIngressRetriesRevisionConflictBeforePublishing(t *testing.T) {
	concurrent, err := ProjectRunnerBatch(
		initialSnapshot("conv_1", "stream-1"),
		"conv_1",
		"stream-1",
		runnerBatch(
			statusMutation("source:1", 1, agentworkbenchv2.SessionStatus_SESSION_STATUS_RUNNING),
		),
	)
	require.NoError(t, err)
	repo := &ingressRepository{
		state:              storedInitialSnapshot(t),
		stateAfterConflict: storedProjectedSnapshot(t, concurrent),
		result:             domainworkbench.AppendResult{Applied: true},
		appendErrors:       []error{domainworkbench.ErrRevisionConflict},
	}
	publisher := &ingressPublisher{}
	service, err := NewIngress(
		ingressSessionResolver{session: &sessiondomain.Session{
			ID: "conv_1", PodKey: "pod-1", OrganizationID: 9,
		}},
		ingressPodResolver{pod: &poddomain.Pod{
			PodKey: "pod-1", RunnerID: 17, OrganizationID: 9,
		}},
		repo,
		publisher,
		ingressArtifactMaterializer{},
		func() string { return "stream-1" },
	)
	require.NoError(t, err)

	err = service.Ingest(context.Background(), 17, runnerBatch(
		statusMutation("source:2", 2, agentworkbenchv2.SessionStatus_SESSION_STATUS_IDLE),
	))
	require.NoError(t, err)
	require.Len(t, repo.requests, 2)
	require.Equal(t, uint64(0), repo.requests[0].ExpectedRevision)
	require.Equal(t, uint64(1), repo.requests[1].ExpectedRevision)
	require.Equal(t, uint64(2), repo.requests[1].Projection.Revision)
	require.Equal(t, "conv_1", publisher.sessionID)
	require.Equal(t, uint64(2), publisher.delta.Revision)
	require.Equal(t, 1, publisher.count)
}

func TestIngressStopsAfterRevisionConflictLimit(t *testing.T) {
	repo := &ingressRepository{
		state: storedInitialSnapshot(t),
		err:   domainworkbench.ErrRevisionConflict,
	}
	publisher := &ingressPublisher{}
	service, err := NewIngress(
		ingressSessionResolver{session: &sessiondomain.Session{
			ID: "conv_1", PodKey: "pod-1", OrganizationID: 9,
		}},
		ingressPodResolver{pod: &poddomain.Pod{
			PodKey: "pod-1", RunnerID: 17, OrganizationID: 9,
		}},
		repo,
		publisher,
		ingressArtifactMaterializer{},
		func() string { return "stream-1" },
	)
	require.NoError(t, err)

	err = service.Ingest(context.Background(), 17, runnerBatch(
		statusMutation("source:1", 1, agentworkbenchv2.SessionStatus_SESSION_STATUS_RUNNING),
	))
	require.ErrorIs(t, err, domainworkbench.ErrRevisionConflict)
	require.Len(t, repo.requests, ingressAppendAttempts)
	require.Zero(t, publisher.count)
}

func TestIngressRejectsRunnerThatDoesNotOwnPod(t *testing.T) {
	repo := &ingressRepository{}
	publisher := &ingressPublisher{}
	materialized := 0
	service, err := NewIngress(
		ingressSessionResolver{session: &sessiondomain.Session{
			ID: "conv_1", PodKey: "pod-1", OrganizationID: 9,
		}},
		ingressPodResolver{pod: &poddomain.Pod{
			PodKey: "pod-1", RunnerID: 18, OrganizationID: 9,
		}},
		repo,
		publisher,
		ingressArtifactMaterializer{count: &materialized},
		func() string { return "stream-1" },
	)
	require.NoError(t, err)

	err = service.Ingest(context.Background(), 17, runnerBatch(
		statusMutation("source:1", 1, agentworkbenchv2.SessionStatus_SESSION_STATUS_RUNNING),
	))

	require.ErrorIs(t, err, ErrInvalidBatch)
	require.Zero(t, materialized)
	require.Empty(t, repo.requests)
	require.Zero(t, publisher.count)
}

func TestIngressRejectsPodFromAnotherOrganization(t *testing.T) {
	repo := &ingressRepository{}
	materialized := 0
	service, err := NewIngress(
		ingressSessionResolver{session: &sessiondomain.Session{
			ID: "conv_1", PodKey: "pod-1", OrganizationID: 9,
		}},
		ingressPodResolver{pod: &poddomain.Pod{
			PodKey: "pod-1", RunnerID: 17, OrganizationID: 10,
		}},
		repo,
		&ingressPublisher{},
		ingressArtifactMaterializer{count: &materialized},
		func() string { return "stream-1" },
	)
	require.NoError(t, err)

	err = service.Ingest(context.Background(), 17, runnerBatch(
		statusMutation("source:1", 1, agentworkbenchv2.SessionStatus_SESSION_STATUS_RUNNING),
	))

	require.ErrorIs(t, err, ErrInvalidBatch)
	require.Zero(t, materialized)
	require.Empty(t, repo.requests)
}

func storedProjectedSnapshot(
	t *testing.T,
	projected *ProjectedBatch,
) *domainworkbench.SessionState {
	t.Helper()
	encoded, err := marshalDeterministic(projected.Snapshot)
	require.NoError(t, err)
	return &domainworkbench.SessionState{
		SessionID:      projected.Snapshot.SessionId,
		StreamEpoch:    projected.Snapshot.StreamEpoch,
		Revision:       projected.Snapshot.Revision,
		LatestSequence: projected.Snapshot.LatestSequence,
		Projection:     encoded,
		Digest:         projected.Snapshot.GetDigest(),
	}
}
