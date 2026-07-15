package workercreation

import (
	"testing"

	resourcedomain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolModelRequirementsFromDefinitionPreservesVideoContract(t *testing.T) {
	requirements, err := toolModelRequirementsFromDefinition(workerdefinition.Definition{
		ToolModelRequirements: []workerdefinition.ToolModelRequirement{
			{
				ID:               "video-generator",
				ProviderKeys:     []string{"openai"},
				ProtocolAdapters: []string{"openai-compatible"},
				Modality:         "video",
				Capability:       "video-generation",
				Environment: workerdefinition.ToolModelEnvironment{
					APIKey: "VIDEO_MODEL_API_KEY", BaseURL: "VIDEO_MODEL_BASE_URL",
					ModelID: "VIDEO_MODEL_ID",
				},
			},
		},
	})

	require.NoError(t, err)
	require.Len(t, requirements, 1)
	assert.Equal(t, "video-generator", requirements[0].Role.String())
	assert.Equal(t, resourcedomain.ModalityVideo, requirements[0].Modality)
	assert.Equal(t, resourcedomain.CapabilityVideoGeneration, requirements[0].Capability)
	assert.Equal(t, "VIDEO_MODEL_ID", requirements[0].Environment.ModelID)
}
