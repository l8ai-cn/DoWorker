package agentpod

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	agentDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
)

// ==================== AgentFile Resolved Precedence Tests ====================

func TestCreatePod_AgentFilePrompt_ExtractedToDB(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, _ := setupOrchestrator(t, withCoordinator(coord))

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		AgentfileLayer:  ptrStr(`PROMPT "from agentfile"`),
	})

	require.NoError(t, err)
	dbPod, err := podSvc.GetPod(context.Background(), result.Pod.PodKey)
	require.NoError(t, err)
	assert.Equal(t, "from agentfile", dbPod.Prompt, "AgentFile PROMPT should be extracted to DB")
}

func TestCreatePod_AgentFileBranch_OverridesReqBranch(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, _ := setupOrchestrator(t, withCoordinator(coord))

	reqBranch := "req-branch"
	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		BranchName:      &reqBranch,
		AgentfileLayer:  ptrStr(`BRANCH "agentfile-branch"`),
	})

	require.NoError(t, err)
	dbPod, err := podSvc.GetPod(context.Background(), result.Pod.PodKey)
	require.NoError(t, err)
	require.NotNil(t, dbPod.BranchName)
	assert.Equal(t, "agentfile-branch", *dbPod.BranchName, "AgentFile BRANCH should override req.BranchName")
}

func TestCreatePod_AgentFilePermissionMode_ExtractedToDB(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, _ := setupOrchestrator(t, withCoordinator(coord))

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		AgentfileLayer:  ptrStr(`CONFIG permission_mode = "bypassPermissions"`),
	})

	require.NoError(t, err)
	dbPod, err := podSvc.GetPod(context.Background(), result.Pod.PodKey)
	require.NoError(t, err)
	require.NotNil(t, dbPod.PermissionMode)
	assert.Equal(t, "bypassPermissions", *dbPod.PermissionMode, "AgentFile CONFIG permission_mode should be extracted to DB")
}

func TestCreatePod_CodexUsesConfigOverridesNotClaudeLegacyFields(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, _ := setupOrchestrator(t,
		withCoordinator(coord),
		withAgentConfigProvider(newCodexTestProvider()),
	)

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "codex-cli",
		ModelResourceID: testModelResourceID(),
		AgentfileLayer:  ptrStr(`CONFIG approval_mode = "never"`),
	})

	require.NoError(t, err)
	dbPod, err := podSvc.GetPod(context.Background(), result.Pod.PodKey)
	require.NoError(t, err)
	assert.Nil(t, dbPod.Model)
	assert.Nil(t, dbPod.PermissionMode)
	// Default autonomous forces codex onto MODE acp with approval_mode=never;
	// the `--ask-for-approval` arg is guarded by `mode != "acp"` so it drops
	// out and the acp launch arg ("app-server") takes its place.
	assert.Equal(t, "never", dbPod.ResolvedConfig["approval_mode"])
	assert.Equal(t, []string{"app-server"}, coord.lastCmd.LaunchArgs)
}

// Explicit MODE pty must survive the default-autonomous automation adapter,
// otherwise CLI/PTY workers are unreachable. The adapter's CONFIG overrides
// still apply so the PTY worker stays non-interactive via native CLI flags.
func TestCreatePod_ExplicitPTYMode_SurvivesAutonomous_Claude(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, _ := setupOrchestrator(t, withCoordinator(coord))

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		AgentfileLayer:  ptrStr("MODE pty"),
	})

	require.NoError(t, err)
	dbPod, err := podSvc.GetPod(context.Background(), result.Pod.PodKey)
	require.NoError(t, err)
	assert.Equal(t, "pty", dbPod.InteractionMode, "explicit MODE pty must not be overridden by autonomous")
	assert.Equal(t, "pty", coord.lastCmd.InteractionMode)
	require.NotNil(t, dbPod.PermissionMode)
	assert.Equal(t, "bypassPermissions", *dbPod.PermissionMode, "autonomous CONFIG still applied to PTY worker")
}

func TestCreatePod_ExplicitPTYMode_SurvivesAutonomous_Codex(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, _ := setupOrchestrator(t,
		withCoordinator(coord),
		withAgentConfigProvider(newCodexTestProvider()),
	)

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "codex-cli",
		ModelResourceID: testModelResourceID(),
		AgentfileLayer:  ptrStr("MODE pty"),
	})

	require.NoError(t, err)
	dbPod, err := podSvc.GetPod(context.Background(), result.Pod.PodKey)
	require.NoError(t, err)
	assert.Equal(t, "pty", dbPod.InteractionMode, "explicit MODE pty must not be overridden by autonomous")
	assert.Equal(t, "pty", coord.lastCmd.InteractionMode)
	assert.Equal(t, "never", dbPod.ResolvedConfig["approval_mode"], "autonomous CONFIG still applied to PTY worker")
	assert.NotEqual(t, []string{"app-server"}, coord.lastCmd.LaunchArgs, "PTY worker must not use the ACP app-server launch arg")
}

func TestCreatePod_NoLayer_BranchInheritedFromResume(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, _ := setupOrchestrator(t, withCoordinator(coord))

	reqBranch := "my-branch"
	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		BranchName:      &reqBranch,
	})

	require.NoError(t, err)
	dbPod, err := podSvc.GetPod(context.Background(), result.Pod.PodKey)
	require.NoError(t, err)
	require.NotNil(t, dbPod.BranchName)
	assert.Equal(t, "my-branch", *dbPod.BranchName, "Without AgentFile Layer, req.BranchName (resume inheritance) should be used")
}

func TestSystemConfigKeys_SSOT(t *testing.T) {
	for key := range systemConfigKeySet {
		assert.True(t, isSystemConfigKey(key), "expected %q to be a system config key", key)
	}
	for _, key := range []string{"", agentDomain.ConfigKeyModel, agentDomain.ConfigKeyPermissionMode, "approval_mode", "mcp_enabled"} {
		assert.False(t, isSystemConfigKey(key), "expected %q to NOT be a system config key", key)
	}

	for _, tc := range []struct {
		name               string
		isResume           bool
		resumeAgentSession bool
	}{
		{"fresh_create", false, false},
		{"resume_with_session", true, true},
		{"resume_without_session", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			overrides := newSystemOverrides("sid-1", tc.isResume, tc.resumeAgentSession, "")
			for key := range overrides {
				_, ok := systemConfigKeySet[key]
				assert.True(t, ok, "newSystemOverrides emitted non-system key %q", key)
			}
		})
	}
}
