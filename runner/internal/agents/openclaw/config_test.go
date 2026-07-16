package openclaw

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyOpenAIProviderFromEnvWritesRunnableConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "openclaw.json")

	require.NoError(t, ApplyOpenAIProviderFromEnv(
		configPath,
		"/workspace",
		"https://proxy.example.test/v1",
		"gpt-5.6-terra",
	))

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var config map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &config))
	assert.Equal(t, "/workspace", config["agents"].(map[string]interface{})["defaults"].(map[string]interface{})["workspace"])
	assert.Equal(t, "openai/gpt-5.6-terra", config["agents"].(map[string]interface{})["defaults"].(map[string]interface{})["model"])
	assert.NotContains(t, string(data), "sk-")

	provider := config["models"].(map[string]interface{})["providers"].(map[string]interface{})["openai"].(map[string]interface{})
	assert.Equal(t, "https://proxy.example.test/v1", provider["baseUrl"])
	assert.Equal(t, "openai-completions", provider["api"])
	assert.Equal(t, "api-key", provider["auth"])
	assert.Equal(t, map[string]interface{}{
		"source":   "env",
		"provider": "default",
		"id":       "OPENAI_API_KEY",
	}, provider["apiKey"])
}

func TestApplyOpenAIProviderFromEnvPreservesExistingConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "openclaw.json")
	require.NoError(t, os.WriteFile(configPath, []byte(`{"gateway":{"mode":"local"},"models":{"providers":{"custom":{"baseUrl":"https://custom.example.test"}}}}`), 0o600))

	require.NoError(t, ApplyOpenAIProviderFromEnv(
		configPath,
		"/workspace",
		"https://proxy.example.test/v1",
		"openai/gpt-5.6-terra",
	))

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	var config map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &config))
	assert.Equal(t, "local", config["gateway"].(map[string]interface{})["mode"])
	assert.NotNil(t, config["models"].(map[string]interface{})["providers"].(map[string]interface{})["custom"])
	assert.Equal(t, "openai/gpt-5.6-terra", config["agents"].(map[string]interface{})["defaults"].(map[string]interface{})["model"])
}
