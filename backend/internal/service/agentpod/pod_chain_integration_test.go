package agentpod

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	agentDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	envbundledomain "github.com/anthropics/agentsmesh/backend/internal/domain/envbundle"
	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
	"github.com/anthropics/agentsmesh/backend/internal/service/agent"
	envbundleservice "github.com/anthropics/agentsmesh/backend/internal/service/envbundle"
)

// fakeEnvBundleLoader is the minimal mock satisfying ConfigBuilder's
// EnvBundleLoader interface. Returns pre-seeded bundles keyed by name,
// matching the eval-ctx shape produced by the real envbundle service.
type fakeEnvBundleLoader struct {
	bundles map[string]map[string]string
}

func (f *fakeEnvBundleLoader) GetEffectiveForUser(_ context.Context, _, _ int64, _ string) ([]*envbundleservice.EffectiveBundle, error) {
	out := make([]*envbundleservice.EffectiveBundle, 0, len(f.bundles))
	for name, data := range f.bundles {
		out = append(out, &envbundleservice.EffectiveBundle{
			Name: name, Kind: "credential", Data: data,
		})
	}
	return out, nil
}

type malformedPodConfigBundleLoader struct{}

func (malformedPodConfigBundleLoader) GetEffectiveForUser(
	context.Context,
	int64,
	int64,
	string,
) ([]*envbundleservice.EffectiveBundle, error) {
	return []*envbundleservice.EffectiveBundle{{
		Name: "settings",
		Kind: envbundledomain.KindConfig,
		Data: map[string]string{envbundledomain.ConfigJSONDataKey: "{"},
	}}, nil
}

// ==================== Helpers ====================

// acpAgentfile returns a base AgentFile that supports both pty and acp modes.
func acpAgentfile() string {
	return "AGENT claude\nEXECUTABLE claude\nMODE pty\nMCP ON\nPROMPT_POSITION prepend\n"
}

// acpProvider creates a mockAgentConfigProvider for an agent supporting pty+acp.
func acpProvider(agentfileSrc string) *mockAgentConfigProvider {
	return &mockAgentConfigProvider{
		agentDef: &agentDomain.Agent{
			Slug: "claude-code", Name: "Claude Code",
			LaunchCommand: "claude", AdapterID: "claude-stream-json", SupportedModes: "pty,acp",
			AgentfileSource: &agentfileSrc, UsesLegacyColumns: true,
		},
		config: agentDomain.ConfigValues{}, creds: agentDomain.EncryptedCredentials{},
		isRunner: true,
	}
}

// acpResolver builds a mockAgentResolver that supports pty+acp.
func acpResolver(agentfileSrc string) *mockAgentResolver {
	return &mockAgentResolver{
		agentDef: &agentDomain.Agent{
			Slug: "claude-code", AdapterID: "claude-stream-json", SupportedModes: "pty,acp",
			AgentfileSource: &agentfileSrc, UsesLegacyColumns: true,
		},
	}
}

// withConfigBuilder injects a custom ConfigBuilder into PodOrchestratorDeps.
func withConfigBuilder(cb *agent.ConfigBuilder) func(*PodOrchestratorDeps) {
	return func(d *PodOrchestratorDeps) { d.ConfigBuilder = cb }
}

// ==================== Test 1: AgentFile Layer -> Command ====================

func TestPodChain_AgentfileLayerToCommand(t *testing.T) {
	coord := &mockPodCoordinator{}
	agentfileSrc := acpAgentfile()
	provider := acpProvider(agentfileSrc)
	resolver := acpResolver(agentfileSrc)

	orch, podSvc, ctx := setupIntegrationOrchestrator(t,
		withCoordinator(coord),
		withAgentResolver(resolver),
		withConfigBuilder(agent.NewConfigBuilder(provider, noopBundleLoader{})),
	)

	layer := "MODE acp\nBRANCH \"feature-x\"\nPROMPT \"do something\"\nCONFIG permission_mode = \"bypassPermissions\"\n"
	req, preparer := workerSpecPlanRequestForTest(
		t,
		ctx,
		"claude-code",
		layer,
		resolvedResource("anthropic", "https://api.anthropic.com", "claude-test"),
	)
	orch.workerCreation = preparer
	req.RunnerID = ctxRunnerID(ctx)
	req.Cols = 120
	req.Rows = 40
	result, err := createPodWithPlanSourceForTest(t, orch, ctx, req)
	require.NoError(t, err)

	// Verify DB record reflects merged values
	dbPod, err := podSvc.GetPod(ctx, result.Pod.PodKey)
	require.NoError(t, err)
	assert.Equal(t, podDomain.InteractionModeACP, dbPod.InteractionMode)
	require.NotNil(t, dbPod.BranchName)
	assert.Equal(t, "feature-x", *dbPod.BranchName)
	require.NotNil(t, dbPod.PermissionMode)
	assert.Equal(t, "bypassPermissions", *dbPod.PermissionMode)
	assert.Equal(t, "do something", dbPod.Prompt)

	// Verify gRPC command content — Backend eval produces execution instructions
	cmd := coord.lastCmd
	require.NotNil(t, cmd)
	assert.Equal(t, result.Pod.PodKey, cmd.PodKey)
	assert.Equal(t, "claude", cmd.LaunchCommand)
	assert.Equal(t, "acp", cmd.InteractionMode, "MODE acp from layer should be reflected")
	assert.Equal(t, "do something", cmd.Prompt)
	assert.Equal(t, int32(120), cmd.Cols)
	assert.Equal(t, int32(40), cmd.Rows)

	// SandboxConfig.SourceBranch should reflect the BRANCH override
	if cmd.SandboxConfig != nil {
		assert.Equal(t, "feature-x", cmd.SandboxConfig.SourceBranch)
	}
}

// ==================== Test 2: Repo Slug Resolution ====================

func TestPodChain_RepoSlugResolution(t *testing.T) {
	coord := &mockPodCoordinator{}
	agentfileSrc := acpAgentfile()
	provider := acpProvider(agentfileSrc)

	// Mock repo service resolves slug -> repository with clone URL
	repoSvc := &mockRepoService{
		repo: &gitprovider.Repository{
			ID:            77,
			HttpCloneURL:  "https://github.com/org/repo-slug.git",
			DefaultBranch: "main",
		},
	}
	resolver := acpResolver(agentfileSrc)

	orch, _, ctx := setupIntegrationOrchestrator(t,
		withCoordinator(coord),
		withAgentResolver(resolver),
		withRepoSvc(repoSvc),
		withConfigBuilder(agent.NewConfigBuilder(provider, noopBundleLoader{})),
	)

	repositoryID := int64(77)
	layer := "REPO \"org/repo-slug\"\n"
	result, err := createPodWithPlanSourceForTest(t, orch, ctx, &OrchestrateCreatePodRequest{
		OrganizationID:  ctxOrgID(ctx),
		UserID:          ctxUserID(ctx),
		RunnerID:        ctxRunnerID(ctx),
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		RepositoryID:    &repositoryID,
		AgentfileLayer:  &layer,
	})
	require.NoError(t, err)

	// Pod should have RepositoryID set from slug resolution
	require.NotNil(t, result.Pod.RepositoryID)
	assert.Equal(t, int64(77), *result.Pod.RepositoryID)

	// Command sandbox config should carry the repo URL
	cmd := coord.lastCmd
	require.NotNil(t, cmd)
	require.NotNil(t, cmd.SandboxConfig)
	assert.Equal(t, "https://github.com/org/repo-slug.git", cmd.SandboxConfig.HttpCloneUrl)
}

// ==================== Test 3: Credential Flow ====================

// TestPodChain_CredentialFlow asserts that bundles named in the AgentFile's
// USE_ENV_BUNDLE declarations land in cmd.EnvVars at Pod creation. With the
// EnvBundle refactor there's no longer a parallel `cmd.Credentials` channel —
// eval merges the bundle's KV directly into the per-pod env map.
func TestPodChain_CredentialFlow(t *testing.T) {
	coord := &mockPodCoordinator{}
	agentfileSrc := acpAgentfile()

	provider := &mockAgentConfigProvider{
		agentDef: &agentDomain.Agent{
			Slug: "claude-code", Name: "Claude Code",
			LaunchCommand: "claude", AdapterID: "claude-stream-json", SupportedModes: "pty,acp",
			AgentfileSource: &agentfileSrc, UsesLegacyColumns: true,
		},
		config: agentDomain.ConfigValues{},
	}
	resolver := acpResolver(agentfileSrc)

	bundleLoader := &fakeEnvBundleLoader{
		bundles: map[string]map[string]string{
			"runtime": {"FEATURE_FLAG": "enabled"},
		},
	}
	cb := agent.NewConfigBuilder(provider, bundleLoader)

	orch, _, ctx := setupIntegrationOrchestrator(t,
		withCoordinator(coord),
		withAgentResolver(resolver),
		withConfigBuilder(cb),
	)

	layer := "USE_ENV_BUNDLE \"runtime\"\n"
	req, preparer := workerSpecPlanRequestForTest(
		t,
		ctx,
		"claude-code",
		layer,
		resolvedResource("anthropic", "https://api.anthropic.com", "claude-test"),
	)
	orch.workerCreation = preparer
	req.RunnerID = ctxRunnerID(ctx)
	result, err := createPodWithPlanSourceForTest(t, orch, ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result.Pod)

	cmd := coord.lastCmd
	require.NotNil(t, cmd)
	assert.Equal(t, "enabled", cmd.EnvVars["FEATURE_FLAG"])
	assert.Equal(t, "sk-test", cmd.EnvVars["ANTHROPIC_API_KEY"])
}

// ==================== Test 4: ConfigBuilder Failure ====================

func TestPodChain_ConfigBuilderFailure(t *testing.T) {
	coord := &mockPodCoordinator{}

	agentfileSrc := acpAgentfile()
	resolver := acpResolver(agentfileSrc)

	orch, podSvc, ctx := setupIntegrationOrchestrator(t,
		withCoordinator(coord),
		withAgentResolver(resolver),
		withConfigBuilder(agent.NewConfigBuilder(
			acpProvider(agentfileSrc),
			malformedPodConfigBundleLoader{},
		)),
	)

	req, preparer := workerSpecPlanRequestForTest(
		t,
		ctx,
		"claude-code",
		`USE_CONFIG_BUNDLE "settings"`,
		resolvedResource("anthropic", "https://api.anthropic.com", "claude-test"),
	)
	orch.workerCreation = preparer
	req.RunnerID = ctxRunnerID(ctx)
	result, err := createPodWithPlanSourceForTest(t, orch, ctx, req)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigBuildFailed)
	assert.Nil(t, result)
	assert.False(t, coord.createPodCalled, "coordinator should not be called when config build fails")

	// Pod was created in DB before config build; it remains in initializing status
	// (MarkInitFailed is only called on dispatch failure, not config build failure)
	_ = podSvc
}

// ==================== Test 6: Dispatch Failure Marks Error ====================

func TestPodChain_DispatchFailureMarksError(t *testing.T) {
	coord := &mockPodCoordinator{err: errors.New("runner connection refused")}

	agentfileSrc := acpAgentfile()
	provider := acpProvider(agentfileSrc)
	resolver := acpResolver(agentfileSrc)

	orch, podSvc, ctx := setupIntegrationOrchestrator(t,
		withCoordinator(coord),
		withAgentResolver(resolver),
		withConfigBuilder(agent.NewConfigBuilder(provider, noopBundleLoader{})),
	)

	layer := "PROMPT \"deploy fix\"\n"
	req, preparer := workerSpecPlanRequestForTest(
		t,
		ctx,
		"claude-code",
		layer,
		resolvedResource("anthropic", "https://api.anthropic.com", "claude-test"),
	)
	orch.workerCreation = preparer
	req.RunnerID = ctxRunnerID(ctx)
	_, err := createPodWithPlanSourceForTest(t, orch, ctx, req)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRunnerDispatchFailed)

	// The command was built and sent to coordinator (which failed)
	require.NotNil(t, coord.lastCmd, "command should have been built before dispatch failure")

	// Pod should exist in DB with error status
	podKey := coord.lastCmd.PodKey
	dbPod, dbErr := podSvc.GetPod(ctx, podKey)
	require.NoError(t, dbErr)
	assert.Equal(t, podDomain.StatusError, dbPod.Status)
	require.NotNil(t, dbPod.ErrorCode)
	assert.Equal(t, errCodeRunnerUnreachable, *dbPod.ErrorCode)
	require.NotNil(t, dbPod.ErrorMessage)
	assert.Contains(t, *dbPod.ErrorMessage, "runner connection refused")
}
