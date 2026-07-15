package workerdefinition

import (
	"encoding/json"
	"testing"

	agentfileschema "github.com/anthropics/agentsmesh/agentfile/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeToolModelRequirementsAcceptsVideoRole(t *testing.T) {
	requirements, err := decodeToolModelRequirements([]json.RawMessage{
		json.RawMessage(`{
			"id":"video-generator",
			"provider_keys":["doubao"],
			"protocol_adapters":["openai-compatible"],
			"modality":"video",
			"capability":"video-generation",
			"environment":{
				"api_key":"VIDEO_MODEL_API_KEY",
				"base_url":"VIDEO_MODEL_BASE_URL",
				"model_id":"VIDEO_MODEL_ID"
			}
		}`),
	})

	require.NoError(t, err)
	require.Len(t, requirements, 1)
	assert.Equal(t, "video-generator", requirements[0].ID)
	assert.Equal(t, []string{"doubao"}, requirements[0].ProviderKeys)
	assert.Equal(t, "video", requirements[0].Modality)
	assert.Equal(t, "video-generation", requirements[0].Capability)
	assert.Equal(t, "VIDEO_MODEL_API_KEY", requirements[0].Environment.APIKey)
}

func TestValidateCredentialBindingSchemaAcceptsToolModelEnvironment(t *testing.T) {
	schema, err := agentfileschema.FromSource(
		"ENV VIDEO_MODEL_API_KEY SECRET OPTIONAL\n" +
			"ENV VIDEO_MODEL_BASE_URL TEXT OPTIONAL\n" +
			"ENV VIDEO_MODEL_ID TEXT OPTIONAL\n",
	)
	require.NoError(t, err)

	err = validateCredentialBindingSchema(
		ModelRequirement{},
		schema,
		nil,
		[]ToolModelRequirement{{
			ID: "video-generator",
			Environment: ToolModelEnvironment{
				APIKey: "VIDEO_MODEL_API_KEY", BaseURL: "VIDEO_MODEL_BASE_URL",
				ModelID: "VIDEO_MODEL_ID",
			},
		}},
	)

	require.NoError(t, err)
}

func TestDecodeToolModelRequirementsRejectsDuplicateEnvironmentTarget(t *testing.T) {
	_, err := decodeToolModelRequirements([]json.RawMessage{
		json.RawMessage(`{
			"id":"video-generator",
			"provider_keys":["doubao"],
			"protocol_adapters":["openai-compatible"],
			"modality":"video",
			"capability":"video-generation",
			"environment":{"api_key":"VIDEO_MODEL_API_KEY","base_url":"VIDEO_MODEL_API_KEY","model_id":"VIDEO_MODEL_ID"}
		}`),
	})

	require.Error(t, err)
	assert.ErrorContains(t, err, "duplicate environment target")
}

func TestDecodeToolModelRequirementsRejectsDuplicateRole(t *testing.T) {
	raw := json.RawMessage(`{
		"id":"video-generator",
		"provider_keys":["doubao"],
		"protocol_adapters":["openai-compatible"],
		"modality":"video",
		"capability":"video-generation",
		"environment":{"api_key":"VIDEO_MODEL_API_KEY","base_url":"VIDEO_MODEL_BASE_URL","model_id":"VIDEO_MODEL_ID"}
	}`)

	_, err := decodeToolModelRequirements([]json.RawMessage{raw, raw})

	require.Error(t, err)
	assert.ErrorContains(t, err, "duplicate tool model requirement")
}
