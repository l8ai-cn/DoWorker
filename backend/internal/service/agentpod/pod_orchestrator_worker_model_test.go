package agentpod

import (
	"context"
	"testing"

	agentDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	resourceDomain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	resourcesvc "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingModelResourceResolver struct {
	resource       *resourcesvc.ResolvedResource
	err            error
	calls          int
	userID         int64
	organizationID int64
	resourceID     int64
	requirements   resourcesvc.ResolutionRequirements
}

func (r *recordingModelResourceResolver) ResolveExact(
	_ context.Context,
	actor resourcesvc.Actor,
	orgID, resourceID int64,
	requirements resourcesvc.ResolutionRequirements,
) (*resourcesvc.ResolvedResource, error) {
	r.calls++
	r.userID = actor.UserID
	r.organizationID = orgID
	r.resourceID = resourceID
	r.requirements = requirements
	if r.err != nil {
		return nil, r.err
	}
	return r.resource, nil
}

func resolvedOpenAIResource() *resourcesvc.ResolvedResource {
	provider, _ := resourceDomain.Provider("openai")
	return &resourcesvc.ResolvedResource{
		Provider: provider,
		Connection: resourceDomain.Connection{
			ID: 1, ProviderKey: slugkit.Slug("openai"), BaseURL: "https://api.openai.com/v1",
		},
		Resource:    resourceDomain.ModelResource{ID: 9, ModelID: "gpt-5.1"},
		Credentials: map[string]string{"api_key": "sk-test"},
	}
}

func resolvedResource(providerKey, baseURL, modelID string) *resourcesvc.ResolvedResource {
	provider, _ := resourceDomain.Provider(providerKey)
	return &resourcesvc.ResolvedResource{
		Provider: provider,
		Connection: resourceDomain.Connection{
			ID: 1, ProviderKey: slugkit.Slug(providerKey), BaseURL: baseURL,
		},
		Resource:    resourceDomain.ModelResource{ID: 9, ModelID: modelID},
		Credentials: map[string]string{"api_key": "sk-test"},
	}
}

func TestApplyWorkerModelRequiresExactResource(t *testing.T) {
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{ModelResources: &recordingModelResourceResolver{}})
	req := &OrchestrateCreatePodRequest{AgentSlug: "codex-cli", UserID: 7, OrganizationID: 11}

	err := orchestrator.applyWorkerModel(context.Background(), req, nil)

	require.ErrorIs(t, err, ErrMissingModelResource)
}

func TestApplyWorkerModelRequiresResolverWiring(t *testing.T) {
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{})
	resourceID := int64(9)
	req := &OrchestrateCreatePodRequest{
		AgentSlug: "codex-cli", UserID: 7, OrganizationID: 11, ModelResourceID: &resourceID,
	}

	err := orchestrator.applyWorkerModel(context.Background(), req, nil)

	require.ErrorIs(t, err, ErrModelResourceResolverUnavailable)
}

func TestApplyWorkerModelResolvesExactResourceForCodex(t *testing.T) {
	resolver := &recordingModelResourceResolver{resource: resolvedOpenAIResource()}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{ModelResources: resolver})
	resourceID := int64(9)
	req := &OrchestrateCreatePodRequest{
		AgentSlug: "codex-cli", UserID: 7, OrganizationID: 11, ModelResourceID: &resourceID,
	}

	require.NoError(t, orchestrator.applyWorkerModel(context.Background(), req, nil))

	assert.Equal(t, 1, resolver.calls)
	assert.Equal(t, int64(7), resolver.userID)
	assert.Equal(t, int64(11), resolver.organizationID)
	assert.Equal(t, resourceID, resolver.resourceID)
	assert.Equal(t, resourceDomain.ModalityChat, resolver.requirements.Modality)
	assert.Equal(t, resourceDomain.CapabilityTextGeneration, resolver.requirements.Capability)
	assert.Equal(t, []string{"openai-compatible"}, resolver.requirements.AllowedProtocolAdapters)
	assert.Empty(t, req.SessionConfigBundles)
	assert.Nil(t, req.AgentfileLayer)
	assert.Equal(t, "sk-test", req.ModelResourceEnv["OPENAI_API_KEY"])
	assert.Equal(t, "gpt-5.1", req.ModelResourceEnv["OPENAI_MODEL"])
}

func TestApplyWorkerModelRejectsWorkerSpecModelDrift(t *testing.T) {
	spec := podServiceWorkerSpec()
	resource := resolvedOpenAIResource()
	resource.Connection.ID = spec.Runtime.ModelBinding.ConnectionID
	resource.Connection.Revision = spec.Runtime.ModelBinding.ConnectionRevision
	resource.Resource.ID = spec.Runtime.ModelBinding.ResourceID
	resource.Resource.ProviderConnectionID = resource.Connection.ID
	resource.Resource.Revision = spec.Runtime.ModelBinding.ResourceRevision + 1
	resource.Resource.ModelID = spec.Runtime.ModelBinding.ModelID
	resolver := &recordingModelResourceResolver{resource: resource}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{ModelResources: resolver})
	resourceID := spec.Runtime.ModelBinding.ResourceID
	req := &OrchestrateCreatePodRequest{
		AgentSlug:          "codex-cli",
		UserID:             7,
		OrganizationID:     11,
		ModelResourceID:    &resourceID,
		preparedWorkerSpec: &spec,
	}

	err := orchestrator.applyWorkerModel(context.Background(), req, nil)

	require.ErrorIs(t, err, ErrWorkerSpecModelChanged)
	assert.Empty(t, req.ModelResourceEnv)
}

func TestApplyWorkerModelUsesPreparedWorkerSpecProtocolAdapter(t *testing.T) {
	spec := podServiceWorkerSpec()
	spec.Runtime.WorkerType.Slug = slugkit.MustNewForTest("openclaw")
	resource := resolvedOpenAIResource()
	resource.Connection.ID = spec.Runtime.ModelBinding.ConnectionID
	resource.Connection.Revision = spec.Runtime.ModelBinding.ConnectionRevision
	resource.Connection.ProviderKey = spec.Runtime.ModelBinding.ProviderKey
	resource.Resource.ID = spec.Runtime.ModelBinding.ResourceID
	resource.Resource.ProviderConnectionID = resource.Connection.ID
	resource.Resource.Revision = spec.Runtime.ModelBinding.ResourceRevision
	resource.Resource.ModelID = spec.Runtime.ModelBinding.ModelID
	resolver := &recordingModelResourceResolver{resource: resource}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{ModelResources: resolver})
	resourceID := spec.Runtime.ModelBinding.ResourceID
	req := &OrchestrateCreatePodRequest{
		AgentSlug:          "openclaw",
		UserID:             7,
		OrganizationID:     11,
		ModelResourceID:    &resourceID,
		preparedWorkerSpec: &spec,
	}

	require.NoError(t, orchestrator.applyWorkerModel(context.Background(), req, nil))

	assert.Equal(t, []string{"openai-compatible"}, resolver.requirements.AllowedProtocolAdapters)
}

func TestValidatePreparedModelBindingRejectsProtocolAdapterDrift(t *testing.T) {
	spec := podServiceWorkerSpec()
	spec.Runtime.ModelBinding.ProtocolAdapter = slugkit.MustNewForTest("anthropic")
	resource := resolvedOpenAIResource()
	resource.Connection.ID = spec.Runtime.ModelBinding.ConnectionID
	resource.Connection.Revision = spec.Runtime.ModelBinding.ConnectionRevision
	resource.Connection.ProviderKey = spec.Runtime.ModelBinding.ProviderKey
	resource.Resource.ID = spec.Runtime.ModelBinding.ResourceID
	resource.Resource.ProviderConnectionID = resource.Connection.ID
	resource.Resource.Revision = spec.Runtime.ModelBinding.ResourceRevision
	resource.Resource.ModelID = spec.Runtime.ModelBinding.ModelID

	err := validatePreparedModelBinding(spec.Runtime.ModelBinding, resource)

	require.ErrorIs(t, err, ErrWorkerSpecModelChanged)
}

func TestApplyWorkerModelUsesCustomAgentExecutable(t *testing.T) {
	resolver := &recordingModelResourceResolver{resource: resolvedOpenAIResource()}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{ModelResources: resolver})
	resourceID := int64(9)
	req := &OrchestrateCreatePodRequest{
		AgentSlug: "custom-codex", UserID: 7, OrganizationID: 11, ModelResourceID: &resourceID,
	}

	require.NoError(t, orchestrator.applyWorkerModel(
		context.Background(),
		req,
		&agentDomain.Agent{Executable: "codex-cli"},
	))

	assert.Equal(t, 1, resolver.calls)
	assert.Equal(t, []string{"openai-compatible"}, resolver.requirements.AllowedProtocolAdapters)
	assert.Equal(t, "sk-test", req.ModelResourceEnv["OPENAI_API_KEY"])
	assert.Equal(t, "gpt-5.1", req.ModelResourceEnv["OPENAI_MODEL"])
}

func TestApplyWorkerModelCodexOmitsEmptyBaseURLAndModel(t *testing.T) {
	resolver := &recordingModelResourceResolver{resource: resolvedResource("openai", "", "")}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{ModelResources: resolver})
	resourceID := int64(9)
	req := &OrchestrateCreatePodRequest{
		AgentSlug: "codex-cli", UserID: 7, OrganizationID: 11, ModelResourceID: &resourceID,
	}

	require.NoError(t, orchestrator.applyWorkerModel(context.Background(), req, nil))

	assert.Equal(t, map[string]string{"OPENAI_API_KEY": "sk-test"}, req.ModelResourceEnv)
}

func TestApplyWorkerModelClaudeUsesConfigModelNotModelEnv(t *testing.T) {
	resolver := &recordingModelResourceResolver{resource: resolvedResource("anthropic", "https://api.anthropic.com", `claude "quoted"`)}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{ModelResources: resolver})
	resourceID := int64(9)
	req := &OrchestrateCreatePodRequest{
		AgentSlug: "claude-code", UserID: 7, OrganizationID: 11, ModelResourceID: &resourceID,
	}

	require.NoError(t, orchestrator.applyWorkerModel(context.Background(), req, nil))

	assert.Equal(t, "sk-test", req.ModelResourceEnv["ANTHROPIC_API_KEY"])
	assert.Equal(t, "https://api.anthropic.com", req.ModelResourceEnv["ANTHROPIC_BASE_URL"])
	assert.NotContains(t, req.ModelResourceEnv, "ANTHROPIC_MODEL")
	require.NotNil(t, req.AgentfileLayer)
	assert.Contains(t, *req.AgentfileLayer, `CONFIG model = "claude \"quoted\""`)
}

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
	args := []string{"--sandbox", "--model", "gemini-flash"}

	_, err := applyModelResourceArgs(args, []string{"--model", "gemini-pro"})

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
