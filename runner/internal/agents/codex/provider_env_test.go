package codex

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyOpenAIProviderFromEnv_PatchesBaseURL(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	require.NoError(t, os.WriteFile(configPath, []byte("model = \"gpt-4\"\n"), 0o644))

	err := ApplyOpenAIProviderFromEnv(configPath, "https://token.aiedulab.cn", "gpt-5.2")
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "token.aiedulab.cn")
	assert.Contains(t, content, "gpt-5.2")
	assert.Contains(t, content, "OpenAI")
}

func TestApplyOpenAIProviderFromEnv_EmptyBaseURLNoop(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	require.NoError(t, os.WriteFile(configPath, []byte("model = \"gpt-4\"\n"), 0o644))

	require.NoError(t, ApplyOpenAIProviderFromEnv(configPath, "", ""))

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, "model = \"gpt-4\"\n", string(data))
}
