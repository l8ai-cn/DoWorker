package sessionapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	agentpodservice "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	"github.com/anthropics/agentsmesh/backend/internal/service/relay"
	"github.com/anthropics/agentsmesh/backend/internal/service/runner"
	sessionusagesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionusage"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type relayConnectionSender struct {
	runner.RunnerCommandSender
	calls int
	err   error
}

func (s *relayConnectionSender) SendSubscribePod(
	_ context.Context,
	_ int64,
	_, _, _ string,
	_ bool,
	_ int32,
) error {
	s.calls++
	return s.err
}

func TestGetSessionRelayConnectionReturnsBrowserConnectionAfterSubscription(t *testing.T) {
	deps, sender := relayConnectionTestDeps(t, nil)

	response := getSessionRelayConnection(t, deps)

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, 1, sender.calls)
	var body struct {
		RelayURL string `json:"relay_url"`
		Token    string `json:"token"`
		PodKey   string `json:"pod_key"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, "wss://relay.example", body.RelayURL)
	assert.NotEmpty(t, body.Token)
	assert.Equal(t, "mobile-pod", body.PodKey)
}

func TestGetSessionRelayConnectionFailsClosedWhenSubscriptionFails(t *testing.T) {
	deps, sender := relayConnectionTestDeps(t, errors.New("runner connection lost"))

	response := getSessionRelayConnection(t, deps)

	assert.Equal(t, http.StatusServiceUnavailable, response.Code)
	assert.Equal(t, 1, sender.calls)
}

func TestGetSessionRelayConnectionRejectsACP(t *testing.T) {
	deps, sender := relayConnectionTestDepsWithMode(t, nil, podDomain.InteractionModeACP)

	response := getSessionRelayConnection(t, deps)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, 0, sender.calls)
}

func TestSessionWireIncludesPodInteractionMode(t *testing.T) {
	wire := sessionWireFrom(
		&domain.Session{ID: "session-mode", AgentSlug: "codex-cli", CreatedAt: time.Now()},
		&podDomain.Pod{InteractionMode: podDomain.InteractionModeACP},
		nil,
		nil,
		sessionusagesvc.Aggregate{},
	)

	assert.Equal(t, podDomain.InteractionModeACP, wire.InteractionMode)
}

func relayConnectionTestDeps(t *testing.T, sendErr error) (*Deps, *relayConnectionSender) {
	return relayConnectionTestDepsWithMode(t, sendErr, podDomain.InteractionModePTY)
}

func relayConnectionTestDepsWithMode(
	t *testing.T,
	sendErr error,
	interactionMode string,
) (*Deps, *relayConnectionSender) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := testkit.SetupTestDB(t)
	require.NoError(t, db.Exec(`
		CREATE TABLE agent_sessions (
			id TEXT PRIMARY KEY,
			organization_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			pod_key TEXT NOT NULL UNIQUE,
			agent_slug TEXT NOT NULL,
			runner_node_id TEXT,
			title TEXT,
			status TEXT NOT NULL DEFAULT 'idle',
			parent_session_id TEXT,
			project TEXT,
			archived BOOLEAN NOT NULL DEFAULT FALSE,
			deleted_at DATETIME,
			mcp_servers TEXT NOT NULL DEFAULT '[]',
			codex_goal TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`).Error)
	pod := &podDomain.Pod{
		OrganizationID:  21,
		PodKey:          "mobile-pod",
		RunnerID:        31,
		CreatedByID:     11,
		Status:          podDomain.StatusRunning,
		AgentStatus:     podDomain.AgentStatusIdle,
		AgentSlug:       "codex-cli",
		InteractionMode: interactionMode,
	}
	require.NoError(t, db.Create(pod).Error)
	require.NoError(t, db.Create(&domain.Session{
		ID:             "session-mobile",
		OrganizationID: 21,
		UserID:         11,
		PodKey:         pod.PodKey,
		AgentSlug:      "codex-cli",
		Status:         "idle",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}).Error)

	manager := relay.NewManagerWithOptions()
	t.Cleanup(manager.Stop)
	require.NoError(t, manager.Register(&relay.RelayInfo{ID: "mobile-relay", URL: "wss://relay.example"}))
	sender := &relayConnectionSender{err: sendErr}
	return &Deps{
		Sessions:      sessionsvc.NewService(db),
		Pod:           agentpodservice.NewPodService(infra.NewPodRepository(db)),
		CommandSender: sender,
		RelayManager:  manager,
		RelayTokens:   relay.NewTokenGenerator("test-secret", "test-issuer"),
	}, sender
}

func getSessionRelayConnection(t *testing.T, deps *Deps) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v1/sessions/session-mobile/relay-connection", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "session-mobile"}}
	ctx.Set("tenant", &middleware.TenantContext{
		OrganizationID: 21, OrganizationSlug: "dev-org", UserID: 11, UserRole: "member",
	})
	deps.handleGetSessionRelayConnection(ctx)
	ctx.Writer.WriteHeaderNow()
	return response
}
