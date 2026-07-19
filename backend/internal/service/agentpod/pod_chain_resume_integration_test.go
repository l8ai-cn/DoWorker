package agentpod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	agentDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/service/agent"
)

func TestResumeIntegration_CodexFullChain(t *testing.T) {
	coord := &mockPodCoordinator{}
	codexAgentfile := "AGENT codex\nEXECUTABLE codex\nMODE pty\n" +
		"CONFIG approval_mode SELECT(\"untrusted\", \"on-request\", \"never\") = \"untrusted\"\n" +
		"PROMPT_POSITION append\n" +
		"arg \"resume\" \"--last\" when config.resume_enabled and mode != \"acp\"\n" +
		"arg \"--ask-for-approval\" config.approval_mode when config.approval_mode != \"\" and mode != \"acp\"\n"

	codexProvider := &mockAgentConfigProvider{
		agentDef: &agentDomain.Agent{
			Slug:              "codex-cli",
			Name:              "Codex CLI",
			LaunchCommand:     "codex",
			AdapterID:         "codex-app-server",
			SupportedModes:    "pty",
			AgentfileSource:   &codexAgentfile,
			UsesLegacyColumns: false,
		},
		config:   agentDomain.ConfigValues{},
		creds:    agentDomain.EncryptedCredentials{},
		isRunner: true,
	}
	codexResolver := &mockAgentResolver{agentDef: codexProvider.agentDef}

	orch, podSvc, ctx := setupIntegrationOrchestrator(t,
		withCoordinator(coord),
		withAgentResolver(codexResolver),
		withConfigBuilder(agent.NewConfigBuilder(codexProvider, noopBundleLoader{})),
		withModelResources(&recordingModelResourceResolver{resource: resolvedOpenAIResource()}),
	)

	sourceLayer := `CONFIG approval_mode = "never"`
	req, preparer := workerSpecPlanRequestForTest(
		t,
		ctx,
		"codex-cli",
		sourceLayer,
		resolvedOpenAIResource(),
	)
	orch.workerCreation = preparer
	req.RunnerID = ctxRunnerID(ctx)
	source, err := createPodWithPlanSourceForTest(t, orch, ctx, req)
	require.NoError(t, err)
	require.NotNil(t, source.Pod)

	sourceDB, err := podSvc.GetPod(ctx, source.Pod.PodKey)
	require.NoError(t, err)
	assert.Equal(t, "never", sourceDB.ResolvedConfig["approval_mode"])
	assert.Nil(t, sourceDB.Model)
	assert.Nil(t, sourceDB.PermissionMode)

	sandboxPath := "/home/user/sandbox/codex-source"
	_, err = podSvc.UpdateByKey(ctx, source.Pod.PodKey, map[string]interface{}{
		"sandbox_path": sandboxPath,
		"status":       podDomain.StatusTerminated,
	})
	require.NoError(t, err)

	resumed, err := createPodWithPlanSourceForTest(t, orch, ctx, &OrchestrateCreatePodRequest{
		OrganizationID: ctxOrgID(ctx),
		UserID:         ctxUserID(ctx),
		SourcePodKey:   source.Pod.PodKey,
	})
	require.NoError(t, err)
	require.NotNil(t, resumed.Pod)

	resumedDB, err := podSvc.GetPod(ctx, resumed.Pod.PodKey)
	require.NoError(t, err)
	assert.Equal(t, "never", resumedDB.ResolvedConfig["approval_mode"])
	assert.Nil(t, resumedDB.Model)
	assert.Nil(t, resumedDB.PermissionMode)
	assert.Equal(t, "codex-cli", resumedDB.AgentSlug)

	require.NotNil(t, coord.lastCmd)
	assert.Equal(t, "codex", coord.lastCmd.LaunchCommand)
	assert.Equal(t, "append", coord.lastCmd.PromptPosition)
	assert.Equal(t, []string{"app-server"}, coord.lastCmd.LaunchArgs)
	require.NotNil(t, coord.lastCmd.SandboxConfig)
	assert.Equal(t, sandboxPath, coord.lastCmd.SandboxConfig.LocalPath)
}
