package podconnect

import (
	"context"
	"errors"
	"testing"
	"time"

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
	events      *connectionEventLog
	err         error
	called      bool
	runnerToken string
}

func (s *connectionCommandSender) SendSubscribePod(
	_ context.Context,
	_ int64,
	_, _ string,
	runnerToken string,
	_ bool,
	_ int32,
) error {
	s.called = true
	s.runnerToken = runnerToken
	s.events.record("dispatch")
	return s.err
}

type recordingConnectionTokenGenerator struct {
	events *connectionEventLog
}

func (g *recordingConnectionTokenGenerator) GenerateToken(
	_ string,
	_, userID, _ int64,
	_ time.Duration,
) (string, error) {
	if userID == 0 {
		g.events.record("runner-token")
		return "runner-token", nil
	}
	g.events.record("browser-token")
	return "browser-token", nil
}

type connectionEventLog struct {
	entries []string
}

func (l *connectionEventLog) record(event string) {
	if l != nil {
		l.entries = append(l.entries, event)
	}
}

func TestGetPodConnectionFailsClosedWithoutCommandSender(t *testing.T) {
	srv := newConnectionTestServer(t, 11, nil)

	res, err := srv.GetPodConnection(ctxAsUser(42), connectionRequest())

	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, connect.CodeUnavailable, connectCodeOf(t, err))
}

func TestGetPodConnectionRejectsInitializingPodBeforeSubscription(t *testing.T) {
	sender := &connectionCommandSender{}
	srv := newConnectionTestServerWithStatus(t, 11, sender, agentpod.StatusInitializing)

	res, err := srv.GetPodConnection(ctxAsUser(42), connectionRequest())

	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, connect.CodeFailedPrecondition, connectCodeOf(t, err))
	assert.False(t, sender.called)
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
	events := &connectionEventLog{}
	sender := &connectionCommandSender{
		events: events,
		err:    errors.New("runner connection lost"),
	}
	srv := newConnectionTestServerWithTokenGenerator(
		t,
		11,
		sender,
		&recordingConnectionTokenGenerator{events: events},
		agentpod.StatusRunning,
	)

	res, err := srv.GetPodConnection(ctxAsUser(42), connectionRequest())

	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, connect.CodeUnavailable, connectCodeOf(t, err))
	assert.True(t, sender.called)
	assert.Equal(t, "runner-token", sender.runnerToken)
	assert.Equal(t, []string{"runner-token", "dispatch"}, events.entries)
}

func TestGetPodConnectionReturnsBrowserConnectionAfterSubscription(t *testing.T) {
	events := &connectionEventLog{}
	sender := &connectionCommandSender{events: events}
	srv := newConnectionTestServerWithTokenGenerator(
		t,
		11,
		sender,
		&recordingConnectionTokenGenerator{events: events},
		agentpod.StatusRunning,
	)

	res, err := srv.GetPodConnection(ctxAsUser(42), connectionRequest())

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, sender.called)
	assert.Equal(t, "runner-token", sender.runnerToken)
	assert.Equal(t, "test-pod", res.Msg.GetPodKey())
	assert.Equal(t, "wss://relay.example", res.Msg.GetRelayUrl())
	assert.Equal(t, "browser-token", res.Msg.GetToken())
	assert.Equal(t, []string{"runner-token", "dispatch", "browser-token"}, events.entries)
	assert.True(t, hasSafeConnectionTokenOrder(events.entries))
}

func TestConnectionTokenOrderRejectsBrowserBeforeDispatch(t *testing.T) {
	deliberatelyWrong := []string{"runner-token", "browser-token", "dispatch"}

	assert.False(t, hasSafeConnectionTokenOrder(deliberatelyWrong))
}

func hasSafeConnectionTokenOrder(events []string) bool {
	return len(events) == 3 &&
		events[0] == "runner-token" &&
		events[1] == "dispatch" &&
		events[2] == "browser-token"
}

func newConnectionTestServer(
	t *testing.T,
	runnerID int64,
	sender runner.RunnerCommandSender,
) *Server {
	return newConnectionTestServerWithStatus(t, runnerID, sender, agentpod.StatusRunning)
}

func newConnectionTestServerWithStatus(
	t *testing.T,
	runnerID int64,
	sender runner.RunnerCommandSender,
	status string,
) *Server {
	return newConnectionTestServerWithTokenGenerator(
		t,
		runnerID,
		sender,
		relay.NewTokenGenerator("test-secret", "test-issuer"),
		status,
	)
}

func newConnectionTestServerWithTokenGenerator(
	t *testing.T,
	runnerID int64,
	sender runner.RunnerCommandSender,
	generator relayTokenGenerator,
	podStatus string,
) *Server {
	t.Helper()
	db := testkit.SetupTestDB(t)
	require.NoError(t, db.Create(&agentpod.Pod{
		OrganizationID:  7,
		PodKey:          "test-pod",
		RunnerID:        runnerID,
		CreatedByID:     42,
		Status:          podStatus,
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
		withConnectionTokenGenerator(generator),
	)
}

func withConnectionTokenGenerator(generator relayTokenGenerator) Option {
	return func(server *Server) {
		server.tokenGenerator = generator
	}
}

func connectionRequest() *connect.Request[podv1.GetPodConnectionRequest] {
	return connect.NewRequest(&podv1.GetPodConnectionRequest{
		OrgSlug: "acme",
		PodKey:  "test-pod",
	})
}
