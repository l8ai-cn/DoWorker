package workerdefinition

import (
	"encoding/json"
	"testing"

	agentfileschema "github.com/anthropics/agentsmesh/agentfile/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeToolModelRequirementsAcceptsSeedanceVideoRole(t *testing.T) {
	requirements, err := decodeToolModelRequirements([]json.RawMessage{
		json.RawMessage(`{
			"id":"seedance-video",
			"provider_keys":["doubao"],
			"protocol_adapters":["openai-compatible"],
			"modality":"video",
			"capability":"video-generation",
			"environment":{
				"api_key":"SEEDANCE_API_KEY",
				"base_url":"SEEDANCE_BASE_URL",
				"model_id":"SEEDANCE_MODEL"
			}
		}`),
	})

	require.NoError(t, err)
	require.Len(t, requirements, 1)
	assert.Equal(t, "seedance-video", requirements[0].ID)
	assert.Equal(t, []string{"doubao"}, requirements[0].ProviderKeys)
	assert.Equal(t, "video", requirements[0].Modality)
	assert.Equal(t, "video-generation", requirements[0].Capability)
	assert.Equal(t, "SEEDANCE_API_KEY", requirements[0].Environment.APIKey)
}

func TestValidateCredentialBindingSchemaAcceptsToolModelEnvironment(t *testing.T) {
	schema, err := agentfileschema.FromSource(
		"ENV SEEDANCE_API_KEY SECRET OPTIONAL\n" +
			"ENV SEEDANCE_BASE_URL TEXT OPTIONAL\n" +
			"ENV SEEDANCE_MODEL TEXT OPTIONAL\n",
	)
	require.NoError(t, err)

	err = validateCredentialBindingSchema(
		ModelRequirement{},
		schema,
		nil,
		[]ToolModelRequirement{{
			ID: "seedance-video",
			Environment: ToolModelEnvironment{
				APIKey: "SEEDANCE_API_KEY", BaseURL: "SEEDANCE_BASE_URL",
				ModelID: "SEEDANCE_MODEL",
			},
		}},
	)

	require.NoError(t, err)
}

func TestDecodeToolModelRequirementsRejectsDuplicateEnvironmentTarget(t *testing.T) {
	_, err := decodeToolModelRequirements([]json.RawMessage{
		json.RawMessage(`{
			"id":"seedance-video",
			"provider_keys":["doubao"],
			"protocol_adapters":["openai-compatible"],
			"modality":"video",
			"capability":"video-generation",
			"environment":{"api_key":"SEEDANCE_API_KEY","base_url":"SEEDANCE_API_KEY","model_id":"SEEDANCE_MODEL"}
		}`),
	})

	require.Error(t, err)
	assert.ErrorContains(t, err, "duplicate environment target")
}

func TestDecodeToolModelRequirementsRejectsDuplicateRole(t *testing.T) {
	raw := json.RawMessage(`{
		"id":"seedance-video",
		"provider_keys":["doubao"],
		"protocol_adapters":["openai-compatible"],
		"modality":"video",
		"capability":"video-generation",
		"environment":{"api_key":"SEEDANCE_API_KEY","base_url":"SEEDANCE_BASE_URL","model_id":"SEEDANCE_MODEL"}
	}`)

	_, err := decodeToolModelRequirements([]json.RawMessage{raw, raw})

	require.Error(t, err)
	assert.ErrorContains(t, err, "duplicate tool model requirement")
}
