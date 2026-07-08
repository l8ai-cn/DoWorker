package codex

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppendCodexProjectTrust(t *testing.T) {
	dir := t.TempDir()
	workDir := filepath.Join(dir, "workspace")
	configPath := filepath.Join(dir, "config.toml")
	require.NoError(t, os.WriteFile(configPath, []byte("model = \"gpt-4\"\n"), 0644))

	err := AppendCodexProjectTrust(configPath, workDir)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	text := string(data)
	assert.Contains(t, text, workDir)
	assert.Contains(t, text, filepath.Join(workDir, ".codex"))
	assert.Contains(t, text, "trust_level")
	assert.Contains(t, text, "danger-full-access")
}
