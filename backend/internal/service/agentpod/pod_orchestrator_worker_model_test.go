package agentpod

import (
	"context"
	"testing"

	agentDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agent"
	resourceDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	resourcesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/airesource"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
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
	connectionID   int64
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

func (r *recordingModelResourceResolver) ResolvePinnedCredentials(
	_ context.Context,
	actor resourcesvc.Actor,
	orgID int64,
	resourceID int64,
	connectionID int64,
) (map[string]string, error) {
	r.calls++
	r.userID = actor.UserID
	r.organizationID = orgID
	r.resourceID = resourceID
	r.connectionID = connectionID
	if r.err != nil {
		return nil, r.err
	}
	if r.resource == nil ||
		r.resource.Resource.ID != resourceID ||
		r.resource.Connection.ID != connectionID {
		return nil, ErrMissingModelResource
	}
	return r.resource.Credentials, nil
}

func resolvedOpenAIResource() *resourcesvc.ResolvedResource {
	provider, _ := resourceDomain.Provider("openai")
	return &resourcesvc.ResolvedResource{
		Provider: provider,
		Connection: resourceDomain.Connection{
			ID: 1, ProviderKey: slugkit.Slug("openai"), BaseURL: "https://api.openai.com/v1", Revision: 1,
		},
		Resource:    resourceDomain.ModelResource{ID: 9, ProviderConnectionID: 1, ModelID: "gpt-5.1", Revision: 1},
		Credentials: map[string]string{"api_key": "sk-test"},
	}
}

func resolvedResource(providerKey, baseURL, modelID string) *resourcesvc.ResolvedResource {
	provider, _ := resourceDomain.Provider(providerKey)
	return &resourcesvc.ResolvedResource{
		Provider: provider,
		Connection: resourceDomain.Connection{
			ID: 1, ProviderKey: slugkit.Slug(providerKey), BaseURL: baseURL, Revision: 1,
		},
		Resource:    resourceDomain.ModelResource{ID: 9, ProviderConnectionID: 1, ModelID: modelID, Revision: 1},
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
