package podconnect

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	agentpodservice "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/service/relay"
	"github.com/anthropics/agentsmesh/backend/internal/service/runner"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	podv1 "github.com/anthropics/agentsmesh/proto/gen/go/pod/v1"
)

type connectionCommandSender struct {
	runner.RunnerCommandSender
	err    error
	called bool
}

func (s *connectionCommandSender) SendSubscribePod(
	_ context.Context,
	_ int64,
	_, _, _ string,
	_ bool,
	_ int32,
) error {
	s.called = true
	return s.err
}

func TestGetPodConnectionFailsClosedWithoutCommandSender(t *testing.T) {
	srv := newConnectionTestServer(t, 11, nil)

	res, err := srv.GetPodConnection(ctxAsUser(42), connectionRequest())

	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, connect.CodeUnavailable, connectCodeOf(t, err))
}

func TestGetPodConnectionFailsClosedWithoutRunner(t *testing.T) {
	for name, runnerID := range map[string]int64{"zero": 0, "negative": -1} {
		t.Run(name, func(t *testing.T) {
			sender := &connectionCommandSender{}
			srv := newConnectionTestServer(t, runnerID, sender)

			res, err := srv.GetPodConnection(ctxAsUser(42), connectionRequest())

			require.Error(t, err)
			assert.Nil(t, res)
			assert.Equal(t, connect.CodeUnavailable, connectCodeOf(t, err))
			assert.False(t, sender.called)
		})
	}
}

func TestGetPodConnectionFailsClosedWhenSubscriptionFails(t *testing.T) {
	sender := &connectionCommandSender{err: errors.New("runner connection lost")}
	srv := newConnectionTestServer(t, 11, sender)

	res, err := srv.GetPodConnection(ctxAsUser(42), connectionRequest())

	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, connect.CodeUnavailable, connectCodeOf(t, err))
	assert.True(t, sender.called)
}

func TestGetPodConnectionReturnsBrowserConnectionAfterSubscription(t *testing.T) {
	sender := &connectionCommandSender{}
	srv := newConnectionTestServer(t, 11, sender)

	res, err := srv.GetPodConnection(ctxAsUser(42), connectionRequest())

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, sender.called)
	assert.Equal(t, "test-pod", res.Msg.GetPodKey())
	assert.Equal(t, "wss://relay.example", res.Msg.GetRelayUrl())
	assert.NotEmpty(t, res.Msg.GetToken())
}

func newConnectionTestServer(
	t *testing.T,
	runnerID int64,
	sender runner.RunnerCommandSender,
) *Server {
	t.Helper()
	db := testkit.SetupTestDB(t)
	require.NoError(t, db.Create(&agentpod.Pod{
		OrganizationID:  7,
		PodKey:          "test-pod",
		RunnerID:        runnerID,
		CreatedByID:     42,
		Status:          agentpod.StatusRunning,
		AgentStatus:     agentpod.AgentStatusIdle,
		InteractionMode: agentpod.InteractionModePTY,
		AutomationLevel: agentpod.AutomationLevelAutonomous,
	}).Error)

	relayManager := relay.NewManagerWithOptions()
	t.Cleanup(relayManager.Stop)
	require.NoError(t, relayManager.Register(&relay.RelayInfo{
		ID:  "test-relay",
		URL: "wss://relay.example",
	}))

	return NewServer(
		agentpodservice.NewPodService(infra.NewPodRepository(db)),
		&fakeOrgService{role: "member"},
		WithCommandSender(sender),
		WithRelayManager(relayManager),
		WithTokenGenerator(relay.NewTokenGenerator("test-secret", "test-issuer")),
	)
}

func connectionRequest() *connect.Request[podv1.GetPodConnectionRequest] {
	return connect.NewRequest(&podv1.GetPodConnectionRequest{
		OrgSlug: "acme",
		PodKey:  "test-pod",
	})
}
