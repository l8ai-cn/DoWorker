package sessionapi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSessionBodyAcceptsModelResourceID(t *testing.T) {
	var body createSessionBody

	err := json.Unmarshal([]byte(`{"agent_id":"do-agent","model_resource_id":42}`), &body)

	require.NoError(t, err)
	require.NotNil(t, body.ModelResourceID)
	assert.Equal(t, int64(42), *body.ModelResourceID)
}

func TestLegacySessionCreateModelFieldsAreRejected(t *testing.T) {
	for _, field := range []string{
		"credential" + "_profile_id",
		"model",
		"model" + "_config_id",
		"virtual_api" + "_key_id",
	} {
		t.Run(field, func(t *testing.T) {
			got, ok := legacySessionCreateModelField([]byte(`{"agent_id":"do-agent","` + field + `":99}`))

			require.True(t, ok)
			assert.Equal(t, field, got)
		})
	}
}

func TestSessionCreatePodRequestCarriesModelResourceIDOnly(t *testing.T) {
	resourceID := int64(42)
	layer := "MODE acp"

	req := sessionCreatePodRequest(11, 21, createSessionBody{
		AgentID:         "do-agent",
		ModelResourceID: &resourceID,
	}, &layer, "/tmp/workspace")

	assert.Equal(t, int64(11), req.UserID)
	assert.Equal(t, int64(21), req.OrganizationID)
	assert.Equal(t, "do-agent", req.AgentSlug)
	require.NotNil(t, req.ModelResourceID)
	assert.Equal(t, resourceID, *req.ModelResourceID)
	assert.Nil(t, req.SessionConfigBundles)
	assert.Empty(t, req.ModelResourceEnv)
}
