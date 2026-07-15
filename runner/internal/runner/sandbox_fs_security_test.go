package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSandboxFsReadRejectsFileSymlinkOutsideWorkspace(t *testing.T) {
	workspace := t.TempDir()
	outside := t.TempDir()
	secret := "outside-secret-content"
	require.NoError(t, os.WriteFile(filepath.Join(outside, "secret.txt"), []byte(secret), 0o600))
	require.NoError(t, os.Symlink(
		filepath.Join(outside, "secret.txt"),
		filepath.Join(workspace, "artifact.txt"),
	))

	result, err := (&RunnerMessageHandler{}).sandboxFsRead(workspace, "artifact.txt")

	require.NoError(t, err)
	assert.NotEmpty(t, result.GetError())
	assert.Empty(t, result.GetContent())
	assert.NotContains(t, result.GetError(), secret)
}

func TestSandboxFsRejectsParentDirectorySymlinkOutsideWorkspace(t *testing.T) {
	workspace := t.TempDir()
	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0o600))
	require.NoError(t, os.Symlink(outside, filepath.Join(workspace, "outside")))
	handler := &RunnerMessageHandler{}

	t.Run("read", func(t *testing.T) {
		result, err := handler.sandboxFsRead(workspace, "outside/secret.txt")
		require.NoError(t, err)
		assert.NotEmpty(t, result.GetError())
		assert.Empty(t, result.GetContent())
	})

	t.Run("list", func(t *testing.T) {
		result, err := handler.sandboxFsList(workspace, "outside")
		require.NoError(t, err)
		assert.NotEmpty(t, result.GetError())
		assert.Empty(t, result.GetEntries())
	})

	t.Run("write", func(t *testing.T) {
		result, err := handler.sandboxFsWrite(workspace, "outside/result.txt", "escaped")
		require.NoError(t, err)
		assert.NotEmpty(t, result.GetError())
		_, statErr := os.Stat(filepath.Join(outside, "result.txt"))
		assert.ErrorIs(t, statErr, os.ErrNotExist)
	})

	t.Run("mkdir", func(t *testing.T) {
		result, err := handler.sandboxFsMkdir(workspace, "outside/generated")
		require.NoError(t, err)
		assert.NotEmpty(t, result.GetError())
		_, statErr := os.Stat(filepath.Join(outside, "generated"))
		assert.ErrorIs(t, statErr, os.ErrNotExist)
	})
}

func TestSandboxFsRootPreservesRegularFileAndDirectoryOperations(t *testing.T) {
	workspace := t.TempDir()
	handler := &RunnerMessageHandler{}

	mkdirResult, err := handler.sandboxFsMkdir(workspace, "output/videos")
	require.NoError(t, err)
	require.Empty(t, mkdirResult.GetError())

	writeResult, err := handler.sandboxFsWrite(
		workspace,
		"output/videos/result.txt",
		"validation",
	)
	require.NoError(t, err)
	require.Empty(t, writeResult.GetError())

	readResult, err := handler.sandboxFsRead(workspace, "output/videos/result.txt")
	require.NoError(t, err)
	require.Empty(t, readResult.GetError())
	assert.Equal(t, "validation", readResult.GetContent())

	listResult, err := handler.sandboxFsList(workspace, "output/videos")
	require.NoError(t, err)
	require.Empty(t, listResult.GetError())
	require.Len(t, listResult.GetEntries(), 1)
	assert.Equal(t, "output/videos/result.txt", listResult.GetEntries()[0].GetPath())
}
