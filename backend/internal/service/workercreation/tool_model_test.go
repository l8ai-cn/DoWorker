package workercreation

import (
	"context"
	"testing"

	resourcedomain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelResolverBindsVideoGenerationResource(t *testing.T) {
	resources := validModelResourceService()
	resources.resolved.Provider.Key = slugkit.MustNewForTest("openai")
	resources.resolved.Connection.ProviderKey = slugkit.MustNewForTest("openai")
	resources.resolved.Resource.ModelID = "video-model-v1"

	binding, err := newModelResolver(resources).ResolveToolModel(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		videoToolRequirement(),
		101,
	)

	require.NoError(t, err)
	assert.Equal(t, slugkit.MustNewForTest("video-generator"), binding.Role)
	assert.Equal(t, slugkit.MustNewForTest("openai"), binding.ModelBinding.ProviderKey)
	assert.Equal(t, resourcedomain.ModalityVideo, resources.requirements.Modality)
	assert.Equal(t, resourcedomain.CapabilityVideoGeneration, resources.requirements.Capability)
	assert.Equal(t, "VIDEO_MODEL_API_KEY", binding.Environment.APIKey)
}

func TestModelResolverRejectsToolProviderSubstitution(t *testing.T) {
	resources := validModelResourceService()
	resources.resolved.Provider.Key = slugkit.MustNewForTest("anthropic")
	resources.resolved.Connection.ProviderKey = slugkit.MustNewForTest("anthropic")

	_, err := newModelResolver(resources).ResolveToolModel(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		videoToolRequirement(),
		101,
	)

	require.Error(t, err)
	assert.ErrorContains(t, err, "provider was substituted")
}

func videoToolRequirement() specdomain.ToolModelRequirement {
	return specdomain.ToolModelRequirement{
		Role:             slugkit.MustNewForTest("video-generator"),
		ProviderKeys:     []slugkit.Slug{slugkit.MustNewForTest("openai")},
		ProtocolAdapters: []slugkit.Slug{slugkit.MustNewForTest("openai-compatible")},
		Modality:         resourcedomain.ModalityVideo,
		Capability:       resourcedomain.CapabilityVideoGeneration,
		Environment: specdomain.ToolModelEnvironment{
			APIKey: "VIDEO_MODEL_API_KEY", BaseURL: "VIDEO_MODEL_BASE_URL",
			ModelID: "VIDEO_MODEL_ID",
		},
	}
}
