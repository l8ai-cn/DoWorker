package agentpod

import (
	"context"
	"testing"

	resourceDomain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyWorkerToolModelsInjectsVideoModelEnvironment(t *testing.T) {
	spec := podServiceWorkerSpec()
	resource := resolvedResource(
		"openai",
		"https://video.example/api/v1",
		"video-model-v1",
	)
	resource.Connection.ID = 31
	resource.Connection.Revision = 4
	resource.Resource.ID = 41
	resource.Resource.ProviderConnectionID = 31
	resource.Resource.Revision = 5
	binding := specdomain.ToolModelBinding{
		Role: slugkit.MustNewForTest("video-generator"),
		ModelBinding: specdomain.ModelBinding{
			ResourceID: 41, ResourceRevision: 5,
			ConnectionID: 31, ConnectionRevision: 4,
			ProviderKey:     slugkit.MustNewForTest("openai"),
			ProtocolAdapter: slugkit.MustNewForTest("openai-compatible"),
			ModelID:         "video-model-v1",
		},
		Modality:   resourceDomain.ModalityVideo,
		Capability: resourceDomain.CapabilityVideoGeneration,
		Environment: specdomain.ToolModelEnvironment{
			APIKey: "VIDEO_MODEL_API_KEY", BaseURL: "VIDEO_MODEL_BASE_URL",
			ModelID: "VIDEO_MODEL_ID",
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
	assert.Equal(t, "sk-test", req.ModelResourceEnv["VIDEO_MODEL_API_KEY"])
	assert.Equal(t, resource.Connection.BaseURL, req.ModelResourceEnv["VIDEO_MODEL_BASE_URL"])
	assert.Equal(t, resource.Resource.ModelID, req.ModelResourceEnv["VIDEO_MODEL_ID"])
}

func TestApplyWorkerToolModelsRejectsRevisionDrift(t *testing.T) {
	spec := podServiceWorkerSpec()
	binding := specdomain.ToolModelBinding{
		Role: slugkit.MustNewForTest("video-generator"),
		ModelBinding: specdomain.ModelBinding{
			ResourceID: 41, ResourceRevision: 5,
			ConnectionID: 31, ConnectionRevision: 4,
			ProviderKey:     slugkit.MustNewForTest("openai"),
			ProtocolAdapter: slugkit.MustNewForTest("openai-compatible"),
			ModelID:         "video-model-v1",
		},
		Modality:   resourceDomain.ModalityVideo,
		Capability: resourceDomain.CapabilityVideoGeneration,
		Environment: specdomain.ToolModelEnvironment{
			APIKey: "VIDEO_MODEL_API_KEY", BaseURL: "VIDEO_MODEL_BASE_URL",
			ModelID: "VIDEO_MODEL_ID",
		},
	}
	spec.Runtime.ToolModelBindings = []specdomain.ToolModelBinding{binding}
	resource := resolvedResource(
		"openai",
		"https://video.example/api/v1",
		binding.ModelBinding.ModelID,
	)
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
