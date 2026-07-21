package agentpod

import (
	"context"
	"testing"

	resourceDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyWorkerToolModelsInjectsSeedanceEnvironment(t *testing.T) {
	spec := podServiceWorkerSpec()
	resource := resolvedResource(
		"doubao",
		"https://ark.cn-beijing.volces.com/api/v3",
		"doubao-seedance-2-0-260128",
	)
	resource.Connection.ID = 31
	resource.Connection.Revision = 4
	resource.Resource.ID = 41
	resource.Resource.ProviderConnectionID = 31
	resource.Resource.Revision = 5
	binding := specdomain.ToolModelBinding{
		Role: slugkit.MustNewForTest("seedance-video"),
		ModelBinding: specdomain.ModelBinding{
			ResourceID: 41, ResourceRevision: 5,
			ConnectionID: 31, ConnectionRevision: 4,
			ProviderKey:     slugkit.MustNewForTest("doubao"),
			ProtocolAdapter: slugkit.MustNewForTest("openai-compatible"),
			ModelID:         "doubao-seedance-2-0-260128",
		},
		Modality:   resourceDomain.ModalityVideo,
		Capability: resourceDomain.CapabilityVideoGeneration,
		Environment: specdomain.ToolModelEnvironment{
			APIKey: "SEEDANCE_API_KEY", BaseURL: "SEEDANCE_BASE_URL",
			ModelID: "SEEDANCE_MODEL",
		},
	}
	spec.Runtime.ToolModelBindings = []specdomain.ToolModelBinding{binding}
	resolver := &recordingModelResourceResolver{resource: resource}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{ModelResources: resolver})
	req := &OrchestrateCreatePodRequest{
		UserID: 7, OrganizationID: 11, preparedWorkerSpec: &spec,
	}

	err := orchestrator.applyWorkerToolModels(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, resourceDomain.ModalityVideo, resolver.requirements.Modality)
	assert.Equal(t, resourceDomain.CapabilityVideoGeneration, resolver.requirements.Capability)
	assert.Equal(t, "sk-test", req.ModelResourceEnv["SEEDANCE_API_KEY"])
	assert.Equal(t, resource.Connection.BaseURL, req.ModelResourceEnv["SEEDANCE_BASE_URL"])
	assert.Equal(t, resource.Resource.ModelID, req.ModelResourceEnv["SEEDANCE_MODEL"])
}

func TestApplyWorkerToolModelsRejectsRevisionDrift(t *testing.T) {
	spec := podServiceWorkerSpec()
	binding := specdomain.ToolModelBinding{
		Role: slugkit.MustNewForTest("seedance-video"),
		ModelBinding: specdomain.ModelBinding{
			ResourceID: 41, ResourceRevision: 5,
			ConnectionID: 31, ConnectionRevision: 4,
			ProviderKey:     slugkit.MustNewForTest("doubao"),
			ProtocolAdapter: slugkit.MustNewForTest("openai-compatible"),
			ModelID:         "doubao-seedance-2-0-260128",
		},
		Modality:   resourceDomain.ModalityVideo,
		Capability: resourceDomain.CapabilityVideoGeneration,
		Environment: specdomain.ToolModelEnvironment{
			APIKey: "SEEDANCE_API_KEY", BaseURL: "SEEDANCE_BASE_URL",
			ModelID: "SEEDANCE_MODEL",
		},
	}
	spec.Runtime.ToolModelBindings = []specdomain.ToolModelBinding{binding}
	resource := resolvedResource("doubao", "https://ark.example/api/v3", binding.ModelBinding.ModelID)
	resource.Connection.ID = binding.ModelBinding.ConnectionID
	resource.Connection.Revision = binding.ModelBinding.ConnectionRevision
	resource.Resource.ID = binding.ModelBinding.ResourceID
	resource.Resource.ProviderConnectionID = resource.Connection.ID
	resource.Resource.Revision = binding.ModelBinding.ResourceRevision + 1
	resolver := &recordingModelResourceResolver{resource: resource}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{ModelResources: resolver})
	req := &OrchestrateCreatePodRequest{
		UserID: 7, OrganizationID: 11, preparedWorkerSpec: &spec,
	}

	err := orchestrator.applyWorkerToolModels(context.Background(), req)

	require.ErrorIs(t, err, ErrWorkerSpecModelChanged)
	assert.Empty(t, req.ModelResourceEnv)
}

func TestApplyWorkerToolModelsRejectsDoubaoSeedLanguageModel(t *testing.T) {
	spec := podServiceWorkerSpec()
	resource := resolvedResource(
		"doubao",
		"https://ark.cn-beijing.volces.com/api/v3",
		"doubao-seed-1-8-251228",
	)
	resource.Connection.ID = 31
	resource.Connection.Revision = 4
	resource.Resource.ID = 41
	resource.Resource.ProviderConnectionID = 31
	resource.Resource.Revision = 5
	spec.Runtime.ToolModelBindings = []specdomain.ToolModelBinding{{
		Role: slugkit.MustNewForTest("seedance-video"),
		ModelBinding: specdomain.ModelBinding{
			ResourceID: 41, ResourceRevision: 5,
			ConnectionID: 31, ConnectionRevision: 4,
			ProviderKey:     slugkit.MustNewForTest("doubao"),
			ProtocolAdapter: slugkit.MustNewForTest("openai-compatible"),
			ModelID:         resource.Resource.ModelID,
		},
		Modality:   resourceDomain.ModalityVideo,
		Capability: resourceDomain.CapabilityVideoGeneration,
		Environment: specdomain.ToolModelEnvironment{
			APIKey: "SEEDANCE_API_KEY", BaseURL: "SEEDANCE_BASE_URL",
			ModelID: "SEEDANCE_MODEL",
		},
	}}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{
		ModelResources: &recordingModelResourceResolver{resource: resource},
	})
	req := &OrchestrateCreatePodRequest{
		UserID: 7, OrganizationID: 11, preparedWorkerSpec: &spec,
	}

	err := orchestrator.applyWorkerToolModels(context.Background(), req)

	require.ErrorContains(t, err, "Seedance video model")
	assert.Empty(t, req.ModelResourceEnv)
}
