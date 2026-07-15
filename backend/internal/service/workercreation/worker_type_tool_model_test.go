package workercreation

import (
	"testing"

	resourcedomain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolModelRequirementsFromDefinitionPreservesSeedanceContract(t *testing.T) {
	requirements, err := toolModelRequirementsFromDefinition(workerdefinition.Definition{
		ToolModelRequirements: []workerdefinition.ToolModelRequirement{
			{
				ID:               "seedance-video",
				ProviderKeys:     []string{"doubao"},
				ProtocolAdapters: []string{"openai-compatible"},
				Modality:         "video",
				Capability:       "video-generation",
				Environment: workerdefinition.ToolModelEnvironment{
					APIKey: "SEEDANCE_API_KEY", BaseURL: "SEEDANCE_BASE_URL",
					ModelID: "SEEDANCE_MODEL",
				},
			},
		},
	})

	require.NoError(t, err)
	require.Len(t, requirements, 1)
	assert.Equal(t, "seedance-video", requirements[0].Role.String())
	assert.Equal(t, resourcedomain.ModalityVideo, requirements[0].Modality)
	assert.Equal(t, resourcedomain.CapabilityVideoGeneration, requirements[0].Capability)
	assert.Equal(t, "SEEDANCE_MODEL", requirements[0].Environment.ModelID)
}
