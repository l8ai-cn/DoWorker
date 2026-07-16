package workerspec

import (
	"context"
	"testing"

	resourcedomain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolverSnapshotsRequiredToolModelResource(t *testing.T) {
	ports := newResolverPortsForTest()
	requirement := seedanceRequirementForTest()
	ports.workerType.ToolModelRequirements = []domain.ToolModelRequirement{requirement}
	ports.toolModelBinding = domain.ToolModelBinding{
		Role: requirement.Role,
		ModelBinding: domain.ModelBinding{
			ResourceID: 3001, ResourceRevision: 4,
			ConnectionID: 4001, ConnectionRevision: 5,
			ProviderKey:     mustSlugForTest("doubao"),
			ProtocolAdapter: mustSlugForTest("openai-compatible"),
			ModelID:         "doubao-seedance-2-0-260128",
		},
		Modality: requirement.Modality, Capability: requirement.Capability,
		Environment: requirement.Environment,
	}
	draft := validDraftForTest()
	draft.ToolModelResourceIDs = map[string]int64{"seedance-video": 3001}

	resolved, err := NewResolver(ports.deps()).Resolve(context.Background(), validScopeForTest(), draft)

	require.NoError(t, err)
	spec, err := domain.DecodeSpec(resolved.SpecJSON())
	require.NoError(t, err)
	assert.Equal(t, []domain.ToolModelBinding{ports.toolModelBinding}, spec.Runtime.ToolModelBindings)
	assert.Equal(t, int64(3001), ports.toolModelResourceID)
}

func TestResolverRejectsMissingRequiredToolModelResource(t *testing.T) {
	ports := newResolverPortsForTest()
	ports.workerType.ToolModelRequirements = []domain.ToolModelRequirement{
		seedanceRequirementForTest(),
	}

	_, err := NewResolver(ports.deps()).Resolve(
		context.Background(),
		validScopeForTest(),
		validDraftForTest(),
	)

	require.Error(t, err)
	assert.ErrorContains(t, err, "tool_model_resource_ids.seedance-video")
}

func seedanceRequirementForTest() domain.ToolModelRequirement {
	return domain.ToolModelRequirement{
		Role:             mustSlugForTest("seedance-video"),
		ProviderKeys:     []slugkit.Slug{mustSlugForTest("doubao")},
		ProtocolAdapters: []slugkit.Slug{mustSlugForTest("openai-compatible")},
		Modality:         resourcedomain.ModalityVideo,
		Capability:       resourcedomain.CapabilityVideoGeneration,
		Environment: domain.ToolModelEnvironment{
			APIKey: "SEEDANCE_API_KEY", BaseURL: "SEEDANCE_BASE_URL",
			ModelID: "SEEDANCE_MODEL",
		},
	}
}
