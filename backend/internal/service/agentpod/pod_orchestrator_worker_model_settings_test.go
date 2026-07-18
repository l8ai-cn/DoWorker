package agentpod

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoAgentModelSettingsOmitsEmptyBaseURL(t *testing.T) {
	settings, err := doAgentModelSettings(resolvedResource("openai", "", "gpt-5.1"))

	require.NoError(t, err)
	providers := settings["provider"].(map[string]interface{})
	openai := providers["openai"].(map[string]interface{})
	options := openai["options"].(map[string]interface{})
	assert.NotContains(t, options, "baseURL")
	assert.Equal(t, "openai", options["kind"])
	assert.Equal(t, "openai/gpt-5.1", settings["model"])
}

func TestDoAgentModelSettingsUsesAnthropicWireForMiniMax(t *testing.T) {
	settings, err := doAgentModelSettings(
		resolvedResource("minimax", "https://api.minimax.io/v1", "MiniMax-M2.5"),
	)

	require.NoError(t, err)
	providers := settings["provider"].(map[string]interface{})
	minimax := providers["minimax"].(map[string]interface{})
	options := minimax["options"].(map[string]interface{})
	assert.Equal(t, "anthropic", options["kind"])
	assert.Equal(t, "minimax/MiniMax-M2.5", settings["model"])
}

func TestDoAgentModelSettingsRequiresAPIKeyAndModel(t *testing.T) {
	withoutKey := resolvedResource("openai", "", "gpt-5.1")
	withoutKey.Credentials = map[string]string{}
	_, err := doAgentModelSettings(withoutKey)
	require.ErrorIs(t, err, ErrMissingModelResource)

	withoutModel := resolvedResource("openai", "", "")
	_, err = doAgentModelSettings(withoutModel)
	require.ErrorIs(t, err, ErrMissingModelResource)
}

func TestApplyWorkerModelPropagatesResolverError(t *testing.T) {
	resolveErr := errors.New("resolve exact")
	resolver := &recordingModelResourceResolver{err: resolveErr}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{ModelResources: resolver})
	resourceID := int64(9)
	req := &OrchestrateCreatePodRequest{
		AgentSlug: "codex-cli", UserID: 7, OrganizationID: 11, ModelResourceID: &resourceID,
	}

	err := orchestrator.applyWorkerModel(context.Background(), req, nil)

	require.ErrorIs(t, err, resolveErr)
}

func TestApplyModelResourceEnvRejectsConflictWithoutPartialWrites(t *testing.T) {
	existing := map[string]string{
		"OPENAI_API_KEY": "custom-key",
		"FEATURE_FLAG":   "enabled",
	}

	err := applyModelResourceEnv(existing, map[string]string{
		"OPENAI_API_KEY": "resource-key",
		"OPENAI_MODEL":   "gpt-5.1",
	})

	require.ErrorIs(t, err, ErrModelResourceEnvConflict)
	assert.Equal(t, map[string]string{
		"OPENAI_API_KEY": "custom-key",
		"FEATURE_FLAG":   "enabled",
	}, existing)
}

func TestApplyModelResourceEnvMergesMatchingValues(t *testing.T) {
	existing := map[string]string{"OPENAI_API_KEY": "resource-key"}

	require.NoError(t, applyModelResourceEnv(existing, map[string]string{
		"OPENAI_API_KEY": "resource-key",
		"OPENAI_MODEL":   "gpt-5.1",
	}))

	assert.Equal(t, "gpt-5.1", existing["OPENAI_MODEL"])
}
