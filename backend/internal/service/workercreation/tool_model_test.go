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

func TestModelResolverBindsSeedanceVideoResource(t *testing.T) {
	resources := validModelResourceService()
	resources.resolved.Provider.Key = slugkit.MustNewForTest("doubao")
	resources.resolved.Connection.ProviderKey = slugkit.MustNewForTest("doubao")
	resources.resolved.Resource.ModelID = "doubao-seedance-2-0-260128"

	binding, err := newModelResolver(resources).ResolveToolModel(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		seedanceToolRequirement(),
		101,
	)

	require.NoError(t, err)
	assert.Equal(t, slugkit.MustNewForTest("seedance-video"), binding.Role)
	assert.Equal(t, slugkit.MustNewForTest("doubao"), binding.ModelBinding.ProviderKey)
	assert.Equal(t, resourcedomain.ModalityVideo, resources.requirements.Modality)
	assert.Equal(t, resourcedomain.CapabilityVideoGeneration, resources.requirements.Capability)
	assert.Equal(t, "SEEDANCE_API_KEY", binding.Environment.APIKey)
}

func TestModelResolverBindsSub2APISeedanceVideoResource(t *testing.T) {
	resources := validModelResourceService()
	resources.resolved.Provider.Key = slugkit.MustNewForTest("sub2api-seedance")
	resources.resolved.Provider.ProtocolAdapter = "ark-seedance"
	resources.resolved.Connection.ProviderKey = slugkit.MustNewForTest("sub2api-seedance")
	resources.resolved.Resource.ModelID = "doubao-seedance-2-0-260128"

	binding, err := newModelResolver(resources).ResolveToolModel(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		seedanceToolRequirement(),
		101,
	)

	require.NoError(t, err)
	assert.Equal(t, slugkit.MustNewForTest("sub2api-seedance"), binding.ModelBinding.ProviderKey)
	assert.Equal(t, slugkit.MustNewForTest("ark-seedance"), binding.ModelBinding.ProtocolAdapter)
}

func TestModelResolverRejectsToolProviderSubstitution(t *testing.T) {
	resources := validModelResourceService()
	resources.resolved.Provider.Key = slugkit.MustNewForTest("openai")
	resources.resolved.Connection.ProviderKey = slugkit.MustNewForTest("openai")

	_, err := newModelResolver(resources).ResolveToolModel(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		seedanceToolRequirement(),
		101,
	)

	require.Error(t, err)
	assert.ErrorContains(t, err, "provider was substituted")
}

func TestModelResolverRejectsDoubaoSeedLanguageModelForSeedanceRole(t *testing.T) {
	resources := validModelResourceService()
	resources.resolved.Provider.Key = slugkit.MustNewForTest("doubao")
	resources.resolved.Connection.ProviderKey = slugkit.MustNewForTest("doubao")
	resources.resolved.Resource.ModelID = "doubao-seed-1-8-251228"

	_, err := newModelResolver(resources).ResolveToolModel(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		seedanceToolRequirement(),
		101,
	)

	require.Error(t, err)
	assert.ErrorContains(t, err, "Seedance video model")
}

func seedanceToolRequirement() specdomain.ToolModelRequirement {
	return specdomain.ToolModelRequirement{
		Role: slugkit.MustNewForTest("seedance-video"),
		ProviderKeys: []slugkit.Slug{
			slugkit.MustNewForTest("doubao"),
			slugkit.MustNewForTest("sub2api-seedance"),
		},
		ProtocolAdapters: []slugkit.Slug{
			slugkit.MustNewForTest("openai-compatible"),
			slugkit.MustNewForTest("ark-seedance"),
		},
		Modality:   resourcedomain.ModalityVideo,
		Capability: resourcedomain.CapabilityVideoGeneration,
		Environment: specdomain.ToolModelEnvironment{
			APIKey: "SEEDANCE_API_KEY", BaseURL: "SEEDANCE_BASE_URL",
			ModelID: "SEEDANCE_MODEL",
		},
	}
}
