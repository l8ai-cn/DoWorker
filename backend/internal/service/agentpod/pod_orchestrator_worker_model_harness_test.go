package agentpod

import (
	"context"
	"testing"

	agentDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyWorkerModelConfiguresGeminiModelExactly(t *testing.T) {
	resolver := &recordingModelResourceResolver{resource: resolvedResource("gemini", "", "gemini-pro")}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{ModelResources: resolver})
	resourceID := int64(9)
	req := &OrchestrateCreatePodRequest{
		AgentSlug: "gemini-cli", UserID: 7, OrganizationID: 11, ModelResourceID: &resourceID,
	}

	require.NoError(t, orchestrator.applyWorkerModel(context.Background(), req, nil))

	assert.Equal(t, "sk-test", req.ModelResourceEnv["GEMINI_API_KEY"])
	assert.NotContains(t, req.ModelResourceEnv, "GOOGLE_API_KEY")
	assert.Equal(t, []string{"--model", "gemini-pro"}, req.ModelResourceArgs)
	assert.Nil(t, req.AgentfileLayer)
}

func TestApplyWorkerModelConfiguresOpenClawWithFormalOpenAIContract(t *testing.T) {
	resolver := &recordingModelResourceResolver{resource: resolvedResource("xai", "https://api.x.ai/v1", `grok "fast"`)}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{ModelResources: resolver})
	resourceID := int64(9)
	req := &OrchestrateCreatePodRequest{
		AgentSlug: "openclaw", UserID: 7, OrganizationID: 11, ModelResourceID: &resourceID,
	}

	require.NoError(t, orchestrator.applyWorkerModel(context.Background(), req, nil))

	assert.Equal(t, []string{"openai-compatible"}, resolver.requirements.AllowedProtocolAdapters)
	assert.Equal(t, map[string]string{
		"OPENAI_API_KEY":  "sk-test",
		"OPENAI_BASE_URL": "https://api.x.ai/v1",
		"OPENAI_MODEL":    `grok "fast"`,
	}, req.ModelResourceEnv)
	assert.Nil(t, req.AgentfileLayer)
}

func TestApplyWorkerModelConfiguresHermesWithFormalOpenAIContract(t *testing.T) {
	resolver := &recordingModelResourceResolver{resource: resolvedOpenAIResource()}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{ModelResources: resolver})
	resourceID := int64(9)
	req := &OrchestrateCreatePodRequest{
		AgentSlug: "hermes", UserID: 7, OrganizationID: 11, ModelResourceID: &resourceID,
	}

	require.NoError(t, orchestrator.applyWorkerModel(context.Background(), req, nil))

	assert.Equal(t, []string{"openai-compatible"}, resolver.requirements.AllowedProtocolAdapters)
	assert.Equal(t, map[string]string{
		"OPENAI_API_KEY":  "sk-test",
		"OPENAI_BASE_URL": "https://api.openai.com/v1",
		"OPENAI_MODEL":    "gpt-5.1",
	}, req.ModelResourceEnv)
	assert.Nil(t, req.AgentfileLayer)
}

func TestApplyModelResourceArgsRejectsConflictingModel(t *testing.T) {
	_, err := applyModelResourceArgs(
		[]string{"--sandbox", "--model", "gemini-flash"},
		[]string{"--model", "gemini-pro"},
	)

	require.ErrorIs(t, err, ErrModelResourceCommandConflict)
}

func TestApplyModelResourceArgsAddsBaseURL(t *testing.T) {
	args, err := applyModelResourceArgs(
		[]string{"text", "repl"},
		[]string{"--base-url", "https://api.minimaxi.com/v1"},
	)

	require.NoError(t, err)
	assert.Equal(t, []string{"text", "repl", "--base-url", "https://api.minimaxi.com/v1"}, args)
}

func TestApplyWorkerModelConfiguresMiniMaxCLIModel(t *testing.T) {
	resolver := &recordingModelResourceResolver{
		resource: resolvedResource("minimax", "https://api.minimax.io/v1", "MiniMax-M2.5"),
	}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{ModelResources: resolver})
	resourceID := int64(9)
	req := &OrchestrateCreatePodRequest{
		AgentSlug: "minimax-cli", UserID: 7, OrganizationID: 11, ModelResourceID: &resourceID,
	}

	require.NoError(t, orchestrator.applyWorkerModel(
		context.Background(),
		req,
		&agentDomain.Agent{Executable: "mmx"},
	))

	assert.Equal(t, []string{"minimax"}, resolver.requirements.AllowedProtocolAdapters)
	assert.Equal(t, map[string]string{"MINIMAX_API_KEY": "sk-test"}, req.ModelResourceEnv)
	require.NotNil(t, req.AgentfileLayer)
	assert.Contains(t, *req.AgentfileLayer, `CONFIG model = "MiniMax-M2.5"`)
	assert.Equal(t, []string{"--base-url", "https://api.minimax.io"}, req.ModelResourceArgs)
	assert.Contains(t, *req.AgentfileLayer, `CONFIG base_url = "https://api.minimax.io"`)
}

func TestApplyWorkerModelDoAgentUsesConfigBundle(t *testing.T) {
	resolver := &recordingModelResourceResolver{resource: resolvedOpenAIResource()}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{ModelResources: resolver})
	resourceID := int64(9)
	req := &OrchestrateCreatePodRequest{
		AgentSlug: "do-agent", UserID: 7, OrganizationID: 11, ModelResourceID: &resourceID,
	}

	require.NoError(t, orchestrator.applyWorkerModel(
		context.Background(), req, &agentDomain.Agent{Executable: "do-agent"},
	))

	require.NotNil(t, req.AgentfileLayer)
	assert.Contains(t, *req.AgentfileLayer, `USE_CONFIG_BUNDLE "worker-model"`)
	assert.Contains(t, req.SessionConfigBundles, "worker-model")
	assert.Nil(t, req.ModelResourceEnv)
}

func TestApplyWorkerModelConfiguresMiniMaxCLI(t *testing.T) {
	resolver := &recordingModelResourceResolver{
		resource: resolvedResource("minimax", "https://api.minimax.io/v1", "MiniMax-M2.5"),
	}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{ModelResources: resolver})
	resourceID := int64(9)
	req := &OrchestrateCreatePodRequest{
		AgentSlug: "minimax-cli", UserID: 7, OrganizationID: 11, ModelResourceID: &resourceID,
	}

	require.NoError(t, orchestrator.applyWorkerModel(
		context.Background(), req, &agentDomain.Agent{Executable: "mmx"},
	))

	assert.Equal(t, []string{"minimax"}, resolver.requirements.AllowedProtocolAdapters)
	assert.Equal(t, map[string]string{"MINIMAX_API_KEY": "sk-test"}, req.ModelResourceEnv)
	require.NotNil(t, req.AgentfileLayer)
	assert.Contains(t, *req.AgentfileLayer, `CONFIG model = "MiniMax-M2.5"`)
}

func TestApplyWorkerModelConfiguresGrokBuild(t *testing.T) {
	resolver := &recordingModelResourceResolver{
		resource: resolvedResource("xai", "https://api.x.ai/v1", "grok-4"),
	}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{ModelResources: resolver})
	resourceID := int64(9)
	req := &OrchestrateCreatePodRequest{
		AgentSlug: "grok-build", UserID: 7, OrganizationID: 11, ModelResourceID: &resourceID,
	}

	require.NoError(t, orchestrator.applyWorkerModel(
		context.Background(), req, &agentDomain.Agent{Executable: "grok"},
	))

	assert.Equal(t, []string{"openai-compatible"}, resolver.requirements.AllowedProtocolAdapters)
	assert.Equal(t, map[string]string{"XAI_API_KEY": "sk-test"}, req.ModelResourceEnv)
	require.NotNil(t, req.AgentfileLayer)
	assert.Contains(t, *req.AgentfileLayer, `CONFIG model = "grok-4"`)
}

func TestApplyWorkerModelRejectsNonXaiGrokBuildResource(t *testing.T) {
	resolver := &recordingModelResourceResolver{resource: resolvedOpenAIResource()}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{ModelResources: resolver})
	resourceID := int64(9)
	req := &OrchestrateCreatePodRequest{
		AgentSlug: "grok-build", UserID: 7, OrganizationID: 11, ModelResourceID: &resourceID,
	}

	err := orchestrator.applyWorkerModel(
		context.Background(), req, &agentDomain.Agent{Executable: "grok"},
	)

	require.ErrorIs(t, err, ErrModelResourceProviderUnsupported)
}

func TestApplyWorkerModelConfiguresHermesProviderAndModel(t *testing.T) {
	resolver := &recordingModelResourceResolver{resource: resolvedOpenAIResource()}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{ModelResources: resolver})
	resourceID := int64(9)
	req := &OrchestrateCreatePodRequest{
		AgentSlug: "hermes", UserID: 7, OrganizationID: 11, ModelResourceID: &resourceID,
	}

	require.NoError(t, orchestrator.applyWorkerModel(
		context.Background(), req, &agentDomain.Agent{Executable: "hermes"},
	))

	require.NotNil(t, req.AgentfileLayer)
	assert.Contains(t, *req.AgentfileLayer, `CONFIG provider = "openai"`)
	assert.Contains(t, *req.AgentfileLayer, `CONFIG model = "gpt-5.1"`)
}
