package v1

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublishExpertRequestIgnoresRuntimeConfiguration(t *testing.T) {
	var request publishExpertRequest
	require.NoError(t, json.Unmarshal([]byte(`{
		"name":"Reviewer",
		"slug":"reviewer",
		"description":"Reviews changes",
		"agentfile_layer":"PROMPT \"client supplied\"",
		"runner_id":99,
		"repository_id":23,
		"skill_slugs":["unsafe-client-choice"]
	}`), &request))

	encoded, err := json.Marshal(request)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"name":"Reviewer",
		"slug":"reviewer",
		"description":"Reviews changes"
	}`, string(encoded))
}

func TestRunExpertRequestIgnoresRunnerOverride(t *testing.T) {
	var request runExpertRequest
	require.NoError(t, json.Unmarshal([]byte(`{
		"alias":"one-off",
		"prompt_override":"Review this.",
		"runner_id":99
	}`), &request))

	encoded, err := json.Marshal(request)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"alias":"one-off",
		"prompt_override":"Review this.",
		"cols":0,
		"rows":0
	}`, string(encoded))
}
