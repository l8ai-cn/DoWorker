package agentpod

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	agentDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agent"
	podDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
)

func TestCreatePod_ResumeMode_SessionReused(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, db := setupOrchestrator(t, withCoordinator(coord))

	agentSlug := "claude-code"
	sourcePod := createImmutableResumeSource(t, orch, podSvc, db, &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		AgentSlug:      agentSlug,
	})
	sourcePod = updateResumeSource(t, podSvc, db, sourcePod.PodKey, map[string]interface{}{
		"session_id": "my-session-id",
		"status":     podDomain.StatusTerminated,
	})

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		SourcePodKey:   sourcePod.PodKey,
	})

	require.NoError(t, err)
	assert.NotNil(t, result.Pod.SessionID)
	assert.Equal(t, "my-session-id", *result.Pod.SessionID)
	assert.Contains(t, coord.lastCmd.LaunchArgs, "--resume")
	assert.Contains(t, coord.lastCmd.LaunchArgs, "my-session-id")
	assert.NotContains(t, coord.lastCmd.LaunchArgs, "--session-id")
}

func TestCreatePod_ResumeMode_CodexUsesCodexResumeLast(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, db := setupOrchestrator(t,
		withCoordinator(coord),
		withAgentConfigProvider(newCodexTestProvider()),
	)

	sourceLayer := "MODE pty"
	sourcePod := createImmutableResumeSource(t, orch, podSvc, db, &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		AgentSlug:       "codex-cli",
		AgentfileLayer:  &sourceLayer,
		AutomationLevel: podDomain.AutomationLevelInteractive,
	})

	sandboxPath := "/home/user/sandbox/codex-source"
	sourcePod = updateResumeSource(t, podSvc, db, sourcePod.PodKey, map[string]interface{}{
		"session_id":   "platform-session-id-not-codex-thread",
		"sandbox_path": sandboxPath,
		"status":       podDomain.StatusTerminated,
	})

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		SourcePodKey:   sourcePod.PodKey,
	})

	require.NoError(t, err)
	require.NotNil(t, result.Pod)
	require.True(t, coord.createPodCalled)
	require.NotNil(t, coord.lastCmd)
	require.NotNil(t, coord.lastCmd.SandboxConfig)

	assert.Equal(t, "codex", coord.lastCmd.LaunchCommand)
	assert.Equal(t, "append", coord.lastCmd.PromptPosition)
	assert.Equal(t, sandboxPath, coord.lastCmd.SandboxConfig.LocalPath)
	assert.Equal(t, []string{"resume", "--last", "--ask-for-approval", "untrusted"}, coord.lastCmd.LaunchArgs)
	assert.NotContains(t, coord.lastCmd.LaunchArgs, "platform-session-id-not-codex-thread")
	assert.NotContains(t, coord.lastCmd.LaunchArgs, "--session-id")
}

func TestCreatePod_ResumeMode_CodexPreservesSourceApprovalMode(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, db := setupOrchestrator(t,
		withCoordinator(coord),
		withAgentConfigProvider(newCodexTestProvider()),
	)

	sourceLayer := `CONFIG approval_mode = "never"`
	source, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "codex-cli",
		ModelResourceID: testModelResourceID(),
		AgentfileLayer:  &sourceLayer,
	})
	require.NoError(t, err)

	sourcePod, err := podSvc.GetPod(context.Background(), source.Pod.PodKey)
	require.NoError(t, err)
	assert.Nil(t, sourcePod.Model)
	assert.Nil(t, sourcePod.PermissionMode)
	assert.Equal(t, "never", sourcePod.ResolvedConfig["approval_mode"])

	sandboxPath := "/home/user/sandbox/codex-never"
	db.Model(&podDomain.Pod{}).Where("pod_key = ?", source.Pod.PodKey).Updates(map[string]interface{}{
		"sandbox_path": sandboxPath,
		"status":       podDomain.StatusTerminated,
	})

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		SourcePodKey:   source.Pod.PodKey,
	})
	require.NoError(t, err)

	assert.Nil(t, result.Pod.Model)
	assert.Nil(t, result.Pod.PermissionMode)
	assert.Equal(t, "never", result.Pod.ResolvedConfig["approval_mode"])
	assert.Equal(t, podDomain.InteractionModeACP, result.Pod.InteractionMode)
	assert.Equal(t, []string{"app-server"}, coord.lastCmd.LaunchArgs)
}

func TestCreatePod_ResumeMode_ClaudePreservesSourcePermissionMode(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, db := setupOrchestrator(t,
		withCoordinator(coord),
		withAgentConfigProvider(newClaudePermissionTestProvider()),
	)

	// Seed the source with a distinct, non-default automation level so we can
	// prove resume replays the source's resolved permission verbatim instead of
	// re-applying the autonomous default (bypassPermissions). auto_edit maps to
	// claude's acceptEdits.
	source, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		AutomationLevel: podDomain.AutomationLevelAutoEdit,
	})
	require.NoError(t, err)

	sourcePod, err := podSvc.GetPod(context.Background(), source.Pod.PodKey)
	require.NoError(t, err)
	assert.Equal(t, "acceptEdits", sourcePod.ResolvedConfig[agentDomain.ConfigKeyPermissionMode])

	db.Model(&podDomain.Pod{}).Where("pod_key = ?", source.Pod.PodKey).Update("status", podDomain.StatusTerminated)

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		SourcePodKey:   source.Pod.PodKey,
	})
	require.NoError(t, err)

	assert.Equal(t, "acceptEdits", result.Pod.ResolvedConfig[agentDomain.ConfigKeyPermissionMode])
	assert.Contains(t, coord.lastCmd.LaunchArgs, "--resume")
	assert.Contains(t, coord.lastCmd.LaunchArgs, "--permission-mode")
	assert.Contains(t, coord.lastCmd.LaunchArgs, "acceptEdits")
	assert.NotContains(t, coord.lastCmd.LaunchArgs, "bypassPermissions")
}

func TestCreatePod_ResumeMode_ClaudePreservesSourcePermissionConfig(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, db := setupOrchestrator(t,
		withCoordinator(coord),
		withAgentConfigProvider(newClaudePermissionTestProvider()),
	)

	sourceLayer := `CONFIG permission_mode = "dontAsk"`
	sourcePod := createImmutableResumeSource(t, orch, podSvc, db, &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		AgentSlug:      "claude-code",
		AgentfileLayer: &sourceLayer,
	})
	sourcePod = updateResumeSource(t, podSvc, db, sourcePod.PodKey, map[string]interface{}{
		"session_id": "legacy-session",
		"status":     podDomain.StatusTerminated,
	})

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		SourcePodKey:   sourcePod.PodKey,
	})
	require.NoError(t, err)

	assert.Equal(t, "dontAsk", result.Pod.ResolvedConfig[agentDomain.ConfigKeyPermissionMode])
	assert.Contains(t, coord.lastCmd.LaunchArgs, "--permission-mode")
	assert.Contains(t, coord.lastCmd.LaunchArgs, "dontAsk")
	assert.NotContains(t, coord.lastCmd.LaunchArgs, "bypassPermissions")
}

func TestCreatePod_ResumeMode_NoSessionID_GeneratesNew(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, db := setupOrchestrator(t, withCoordinator(coord))

	agentSlug := "claude-code"
	sourcePod := createImmutableResumeSource(t, orch, podSvc, db, &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		AgentSlug:      agentSlug,
	})
	sourcePod = updateResumeSource(t, podSvc, db, sourcePod.PodKey, map[string]interface{}{
		"session_id": nil,
		"status":     podDomain.StatusTerminated,
	})

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		SourcePodKey:   sourcePod.PodKey,
	})

	require.NoError(t, err)
	assert.NotNil(t, result.Pod.SessionID)
	assert.NotEmpty(t, *result.Pod.SessionID)
}

func TestCreatePod_ResumeMode_DisableResumeAgentSession(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, db := setupOrchestrator(t, withCoordinator(coord))

	agentSlug := "claude-code"
	sourcePod := createImmutableResumeSource(t, orch, podSvc, db, &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		AgentSlug:      agentSlug,
	})
	sourcePod = updateResumeSource(t, podSvc, db, sourcePod.PodKey, map[string]interface{}{
		"session_id": "session-1",
		"status":     podDomain.StatusTerminated,
	})

	resumeOff := false
	_, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:     1,
		UserID:             1,
		SourcePodKey:       sourcePod.PodKey,
		ResumeAgentSession: &resumeOff,
	})

	require.NoError(t, err)
	// When ResumeAgentSession is false, resume_enabled/resume_session should NOT be set
	assert.NotContains(t, coord.lastCmd.LaunchArgs, "--resume")
	// Resume-mode create never injects config.session_id, so --session-id must stay off
	assert.NotContains(t, coord.lastCmd.LaunchArgs, "--session-id")
}

func TestCreatePod_ResumeMode_CompletedPod(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, db := setupOrchestrator(t, withCoordinator(coord))

	agentSlug := "claude-code"
	sourcePod := createImmutableResumeSource(t, orch, podSvc, db, &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		AgentSlug:      agentSlug,
	})
	sourcePod = updateResumeSource(t, podSvc, db, sourcePod.PodKey, map[string]interface{}{
		"session_id": "session-1",
		"status":     podDomain.StatusCompleted,
	})

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		SourcePodKey:   sourcePod.PodKey,
	})

	require.NoError(t, err)
	assert.NotNil(t, result.Pod)
}

func TestCreatePod_ResumeMode_OrphanedPod(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, db := setupOrchestrator(t, withCoordinator(coord))

	agentSlug := "claude-code"
	sourcePod := createImmutableResumeSource(t, orch, podSvc, db, &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		AgentSlug:      agentSlug,
	})
	sourcePod = updateResumeSource(t, podSvc, db, sourcePod.PodKey, map[string]interface{}{
		"session_id": "session-1",
		"status":     podDomain.StatusOrphaned,
	})

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		SourcePodKey:   sourcePod.PodKey,
	})

	require.NoError(t, err)
	assert.NotNil(t, result.Pod)
}

func TestCreatePod_ResumeMode_SandboxPath(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, db := setupOrchestrator(t, withCoordinator(coord))

	agentSlug := "claude-code"
	sourcePod := createImmutableResumeSource(t, orch, podSvc, db, &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		AgentSlug:      agentSlug,
	})

	sandboxPath := "/home/user/sandbox/pod-123"
	sourcePod = updateResumeSource(t, podSvc, db, sourcePod.PodKey, map[string]interface{}{
		"session_id":   "session-1",
		"sandbox_path": sandboxPath,
		"status":       podDomain.StatusTerminated,
	})

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		SourcePodKey:   sourcePod.PodKey,
	})

	require.NoError(t, err)
	assert.NotNil(t, result.Pod)
	assert.True(t, coord.createPodCalled)
	// SandboxConfig.LocalPath should be set when sandbox_path exists
	if coord.lastCmd.SandboxConfig != nil {
		assert.Equal(t, sandboxPath, coord.lastCmd.SandboxConfig.LocalPath)
	}
}
