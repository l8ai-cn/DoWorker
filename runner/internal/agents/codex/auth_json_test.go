package codex

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteAuthJSONFromEnv_WritesKey(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, WriteAuthJSONFromEnv(dir, "  sk-test-123  "))

	data, err := os.ReadFile(filepath.Join(dir, "auth.json"))
	require.NoError(t, err)

	var got map[string]string
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, "sk-test-123", got["OPENAI_API_KEY"])
	assert.Equal(t, "apikey", got["auth_mode"])
}

func TestWriteAuthJSONFromEnv_EmptyKeyNoop(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, WriteAuthJSONFromEnv(dir, "   "))

	_, err := os.Stat(filepath.Join(dir, "auth.json"))
	assert.True(t, os.IsNotExist(err))
}

func TestWriteAuthJSONFromEnv_OverridesStaleCopy(t *testing.T) {
	dir := t.TempDir()
	authPath := filepath.Join(dir, "auth.json")
	require.NoError(t, os.WriteFile(authPath, []byte(`{"OPENAI_API_KEY":"stale"}`), 0o600))

	require.NoError(t, WriteAuthJSONFromEnv(dir, "sk-fresh"))

	data, err := os.ReadFile(authPath)
	require.NoError(t, err)
	var got map[string]string
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, "sk-fresh", got["OPENAI_API_KEY"])
}
