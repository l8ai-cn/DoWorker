package agentworkbench

import (
	"context"
	"testing"

	sessiondomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	domainworkbench "github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"github.com/stretchr/testify/require"
)

type ingressSessionResolver struct {
	session *sessiondomain.Session
	err     error
}

func (resolver ingressSessionResolver) GetByPodKey(
	context.Context,
	string,
) (*sessiondomain.Session, error) {
	return resolver.session, resolver.err
}

type ingressRepository struct {
	state   *domainworkbench.SessionState
	request domainworkbench.AppendRequest
	result  domainworkbench.AppendResult
	err     error
}

func (repo *ingressRepository) Append(
	_ context.Context,
	request domainworkbench.AppendRequest,
) (domainworkbench.AppendResult, error) {
	repo.request = request
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
}

func (publisher *ingressPublisher) Publish(
	sessionID string,
	delta *agentworkbenchv2.SessionDeltaBatch,
) {
	publisher.sessionID = sessionID
	publisher.delta = delta
}

func TestIngressPersistsAndPublishesRichArtifactBatch(t *testing.T) {
	repo := &ingressRepository{
		result: domainworkbench.AppendResult{Applied: true},
	}
	publisher := &ingressPublisher{}
	service, err := NewIngress(
		ingressSessionResolver{session: &sessiondomain.Session{
			ID: "conv_1", PodKey: "pod-1",
		}},
		repo,
		publisher,
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
			ID: "conv_1", PodKey: "pod-1",
		}},
		repo,
		publisher,
		func() string { return "stream-1" },
	)
	require.NoError(t, err)

	err = service.Ingest(context.Background(), 17, runnerBatch(
		statusMutation("source:1", 1, agentworkbenchv2.SessionStatus_SESSION_STATUS_RUNNING),
	))
	require.NoError(t, err)
	require.Nil(t, publisher.delta)
}
