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
	"github.com/anthropics/agentsmesh/backend/pkg/embedtoken"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const relayConnectionTokenSecret = "test-secret"

type relayConnectionSender struct {
	runner.RunnerCommandSender
	calls       int
	runnerToken string
	err         error
}

func (s *relayConnectionSender) SendSubscribePod(
	_ context.Context,
	_ int64,
	_, _, token string,
	_ bool,
	_ int32,
) error {
	s.calls++
	s.runnerToken = token
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

func TestGetSessionACPRelayConnectionAcceptsACP(t *testing.T) {
	deps, sender := relayConnectionTestDepsWithMode(t, nil, podDomain.InteractionModeACP)

	response := getSessionACPRelayConnection(t, deps)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Equal(t, 1, sender.calls)
}

func TestGetSessionACPRelayConnectionRejectsPTY(t *testing.T) {
	deps, sender := relayConnectionTestDeps(t, nil)

	response := getSessionACPRelayConnection(t, deps)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, 0, sender.calls)
}

func TestListSessionTerminalsExcludesACP(t *testing.T) {
	deps, _ := relayConnectionTestDepsWithMode(t, nil, podDomain.InteractionModeACP)

	response := getSessionTerminals(t, deps)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.JSONEq(t, `{"data":[]}`, response.Body.String())
}

func TestEmbedRelayConnectionCapsTokensAtEmbedExpiry(t *testing.T) {
	deps, sender := relayConnectionTestDeps(t, nil)
	expiresAt := time.Now().Add(10 * time.Minute).Truncate(time.Second)

	response := getSessionRelayConnectionWithClaims(t, deps, embedRelayClaims(expiresAt))

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var body sessionRelayConnection
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	for name, token := range map[string]string{
		"browser": body.Token,
		"runner":  sender.runnerToken,
	} {
		t.Run(name, func(t *testing.T) {
			claims := parseRelayConnectionToken(t, token)
			require.NotNil(t, claims.ExpiresAt)
			assert.False(t, claims.ExpiresAt.Time.After(expiresAt))
		})
	}
}

func TestEmbedRelayConnectionRejectsExpiredClaims(t *testing.T) {
	deps, sender := relayConnectionTestDeps(t, nil)

	response := getSessionRelayConnectionWithClaims(
		t,
		deps,
		embedRelayClaims(time.Now().Add(-time.Minute)),
	)

	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.JSONEq(t, `{"error":"embed session expired"}`, response.Body.String())
	assert.Equal(t, 0, sender.calls)
	assert.Empty(t, sender.runnerToken)
}

func TestRegularRelayConnectionTokensRemainOneHour(t *testing.T) {
	deps, sender := relayConnectionTestDeps(t, nil)

	response := getSessionRelayConnection(t, deps)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var body sessionRelayConnection
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	for name, token := range map[string]string{
		"browser": body.Token,
		"runner":  sender.runnerToken,
	} {
		t.Run(name, func(t *testing.T) {
			claims := parseRelayConnectionToken(t, token)
			require.NotNil(t, claims.ExpiresAt)
			require.NotNil(t, claims.IssuedAt)
			assert.Equal(t, time.Hour, claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time))
		})
	}
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
		RelayTokens:   relay.NewTokenGenerator(relayConnectionTokenSecret, "test-issuer"),
	}, sender
}

func getSessionRelayConnection(t *testing.T, deps *Deps) *httptest.ResponseRecorder {
	return getSessionRelayConnectionWithClaims(t, deps, nil)
}

func getSessionACPRelayConnection(t *testing.T, deps *Deps) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v1/sessions/session-mobile/acp-relay-connection", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "session-mobile"}}
	ctx.Set("tenant", &middleware.TenantContext{
		OrganizationID: 21, OrganizationSlug: "dev-org", UserID: 11, UserRole: "member",
	})
	deps.handleGetSessionACPRelayConnection(ctx)
	ctx.Writer.WriteHeaderNow()
	return response
}

func getSessionTerminals(t *testing.T, deps *Deps) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v1/sessions/session-mobile/resources/terminals", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "session-mobile"}}
	ctx.Set("tenant", &middleware.TenantContext{
		OrganizationID: 21, OrganizationSlug: "dev-org", UserID: 11, UserRole: "member",
	})
	deps.handleListTerminals(ctx)
	ctx.Writer.WriteHeaderNow()
	return response
}

func getSessionRelayConnectionWithClaims(
	t *testing.T,
	deps *Deps,
	claims *embedtoken.Claims,
) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v1/sessions/session-mobile/relay-connection", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "session-mobile"}}
	ctx.Set("tenant", &middleware.TenantContext{
		OrganizationID: 21, OrganizationSlug: "dev-org", UserID: 11, UserRole: "member",
	})
	if claims != nil {
		ctx.Set(embedClaimsContextKey, claims)
	}
	deps.handleGetSessionRelayConnection(ctx)
	ctx.Writer.WriteHeaderNow()
	return response
}

func parseRelayConnectionToken(t *testing.T, token string) *relay.TokenClaims {
	t.Helper()
	parsed, err := jwt.ParseWithClaims(token, &relay.TokenClaims{}, func(*jwt.Token) (any, error) {
		return []byte(relayConnectionTokenSecret), nil
	})
	require.NoError(t, err)
	claims, ok := parsed.Claims.(*relay.TokenClaims)
	require.True(t, ok)
	return claims
}

func embedRelayClaims(expiresAt time.Time) *embedtoken.Claims {
	return &embedtoken.Claims{
		SessionID: "session-mobile",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
}
