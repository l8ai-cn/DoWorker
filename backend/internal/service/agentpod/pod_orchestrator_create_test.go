package agentpod

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	agentDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	runnerDomain "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	"github.com/anthropics/agentsmesh/backend/internal/service/agent"
)

func TestNewPodOrchestrator(t *testing.T) {
	db := setupTestDB(t)
	podSvc := newTestPodService(db)
	coord := &mockPodCoordinator{}

	deps := &PodOrchestratorDeps{
		PodService:     podSvc,
		PodCoordinator: coord,
	}
	orch := NewPodOrchestrator(deps)

	assert.NotNil(t, orch)
	assert.Equal(t, podSvc, orch.podService)
	assert.Equal(t, coord, orch.podCoordinator)
}

func TestCreatePod_NormalMode_Success(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, _, _ := setupOrchestrator(t, withCoordinator(coord))

	result, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		RunnerID:       1,
		AgentSlug:      "claude-code",
		AgentfileLayer: ptrStr("CONFIG mcp_enabled = true"),
		Cols:           120,
		Rows:           40,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotNil(t, result.Pod)
	assert.Empty(t, result.Warning)
	assert.Equal(t, podDomain.StatusInitializing, result.Pod.Status)
	assert.True(t, coord.createPodCalled)
	assert.Equal(t, int64(1), coord.lastRunnerID)
	assert.Equal(t, result.Pod.PodKey, coord.lastCmd.PodKey)
}

func TestCreatePod_NormalMode_MissingRunnerID(t *testing.T) {
	orch, _, _ := setupOrchestrator(t, withRunnerSelector(nil))

	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		RunnerID:       0, // missing
		AgentSlug:      "claude-code",
		AgentfileLayer: ptrStr("CONFIG mcp_enabled = true"),
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrMissingRunnerID))
}

// ==================== Auto-Select Runner Tests ====================

func TestCreatePod_AutoSelectRunner_Success(t *testing.T) {
	coord := &mockPodCoordinator{}
	selector := &mockRunnerSelector{
		runner: &runnerDomain.Runner{ID: 42, NodeID: "auto-runner"},
	}
	resolver := &mockAgentResolver{
		agentDef: &agentDomain.Agent{Slug: "claude-code", SupportedModes: "pty", AgentfileSource: ptrStr("AGENT claude\nPROMPT_POSITION prepend")},
	}

	orch, _, _ := setupOrchestrator(t,
		withCoordinator(coord),
		withRunnerSelector(selector),
		withAgentResolver(resolver),
	)

	result, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		RunnerID:       0, // auto-select
		AgentSlug:      "claude-code",
		AgentfileLayer: ptrStr("CONFIG mcp_enabled = true"),
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotNil(t, result.Pod)
	assert.Equal(t, int64(42), result.Pod.RunnerID) // auto-selected runner
	assert.True(t, selector.selectCalled)
	assert.Nil(t, selector.resolveCall)
	assert.True(t, coord.createPodCalled)
	assert.Equal(t, int64(42), coord.lastRunnerID)
}

func TestCreatePod_AutoSelectRunner_NoAvailableRunner(t *testing.T) {
	selector := &mockRunnerSelector{
		err: errors.New("no available runner supports the requested agent"),
	}
	resolver := &mockAgentResolver{
		agentDef: &agentDomain.Agent{Slug: "claude-code", SupportedModes: "pty", AgentfileSource: ptrStr("AGENT claude\nPROMPT_POSITION prepend")},
	}

	orch, _, _ := setupOrchestrator(t,
		withRunnerSelector(selector),
		withAgentResolver(resolver),
	)

	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		RunnerID:       0,
		AgentSlug:      "claude-code",
		AgentfileLayer: ptrStr("CONFIG mcp_enabled = true"),
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoAvailableRunner))
}

func TestCreatePod_AutoSelectRunner_AgentResolveError(t *testing.T) {
	selector := &mockRunnerSelector{
		runner: &runnerDomain.Runner{ID: 42},
	}
	resolver := &mockAgentResolver{
		err: errors.New("agent not found"),
	}

	orch, _, _ := setupOrchestrator(t,
		withRunnerSelector(selector),
		withAgentResolver(resolver),
	)

	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		RunnerID:       0,
		AgentSlug:      "claude-code",
		AgentfileLayer: ptrStr("CONFIG mcp_enabled = true"),
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrMissingAgentSlug))
}

func TestCreatePod_NormalMode_MissingAgentSlug(t *testing.T) {
	orch, _, _ := setupOrchestrator(t)

	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		RunnerID:       1,
		AgentSlug:      "", // missing
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrMissingAgentSlug))
}

func TestCreatePod_QuotaExceeded(t *testing.T) {
	errQuota := errors.New("quota exceeded")
	billing := &mockBillingService{err: errQuota}
	orch, _, _ := setupOrchestrator(t, withBilling(billing))

	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		RunnerID:       1,
		AgentSlug:      "claude-code",
		AgentfileLayer: ptrStr("CONFIG mcp_enabled = true"),
	})

	require.Error(t, err)
	assert.Equal(t, errQuota, err)
}

func TestCreatePod_NilBilling_SkipsQuotaCheck(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, _, _ := setupOrchestrator(t, withCoordinator(coord))

	result, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		RunnerID:       1,
		AgentSlug:      "claude-code",
		AgentfileLayer: ptrStr("CONFIG mcp_enabled = true"),
	})

	require.NoError(t, err)
	assert.NotNil(t, result.Pod)
}

func TestCreatePod_NilCoordinator(t *testing.T) {
	// No coordinator -> pod is created in DB but no command sent
	orch, _, _ := setupOrchestrator(t)

	result, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		RunnerID:       1,
		AgentSlug:      "claude-code",
		AgentfileLayer: ptrStr("CONFIG mcp_enabled = true"),
	})

	require.NoError(t, err)
	assert.NotNil(t, result.Pod)
	assert.Empty(t, result.Warning)
}

func TestCreatePod_CoordinatorSendFailure_ReturnsError(t *testing.T) {
	coord := &mockPodCoordinator{err: errors.New("runner not connected")}
	orch, _, _ := setupOrchestrator(t, withCoordinator(coord))

	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		RunnerID:       1,
		AgentSlug:      "claude-code",
		AgentfileLayer: ptrStr("CONFIG mcp_enabled = true"),
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRunnerDispatchFailed)
}

func TestCreatePod_ConfigBuildFailure(t *testing.T) {
	// Create an orchestrator with a provider that fails on GetAgent
	db := setupTestDB(t)
	podSvc := newTestPodService(db)

	provider := &mockAgentConfigProvider{
		agentErr: errors.New("agent not found"),
	}
	configBuilder := agent.NewConfigBuilder(provider, noopBundleLoader{})

	orch := NewPodOrchestrator(&PodOrchestratorDeps{
		PodService:     podSvc,
		ConfigBuilder:  configBuilder,
		RunnerSelector: &mockRunnerSelector{resolveRunner: &runnerDomain.Runner{ID: 1}},
	})

	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		RunnerID:       1,
		AgentSlug:      "claude-code",
		AgentfileLayer: ptrStr("CONFIG mcp_enabled = true"),
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrConfigBuildFailed))
}

func TestCreatePod_SessionID_SetForNormalMode(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, _, _ := setupOrchestrator(t, withCoordinator(coord))

	result, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		RunnerID:       1,
		AgentSlug:      "claude-code",
		AgentfileLayer: ptrStr("CONFIG mcp_enabled = true"),
	})

	require.NoError(t, err)
	// Session ID should be set on the pod
	assert.NotNil(t, result.Pod.SessionID)
	assert.NotEmpty(t, *result.Pod.SessionID)
	assert.Contains(t, coord.lastCmd.LaunchArgs, "--session-id")
	assert.Contains(t, coord.lastCmd.LaunchArgs, *result.Pod.SessionID)
	assert.NotContains(t, coord.lastCmd.LaunchArgs, "--resume")
}

// Credential routing tests previously asserted CredentialProfileID storage on
// the Pod row. After the EnvBundle refactor that field is gone — credentials
// flow exclusively through USE_ENV_BUNDLE → ConfigBuilder.envBundleSvc →
// cmd.EnvVars. The equivalent end-to-end check lives in
// TestPodChain_CredentialFlow (pod_chain_integration_test.go).
