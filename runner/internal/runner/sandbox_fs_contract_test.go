package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/anthropics/agentsmesh/runner/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSandboxFsListReturnsCanonicalRelativeFilePath(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "deliverables", "showcase")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "preview.png"), []byte("png"), 0o644))

	result, err := (&RunnerMessageHandler{}).sandboxFsList(root, "deliverables/showcase")

	require.NoError(t, err)
	require.Empty(t, result.GetError())
	require.Len(t, result.GetEntries(), 1)
	assert.Equal(t, "deliverables/showcase/preview.png", result.GetEntries()[0].GetPath())
}

func TestSandboxFsChangesListsStandaloneWorkspaceFilesAsCreated(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "deliverables", "showcase"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "README.md"), []byte("readme"), 0o644))
	require.NoError(t, os.WriteFile(
		filepath.Join(root, "deliverables", "showcase", "preview.png"),
		[]byte("png"),
		0o644,
	))

	result, err := (&RunnerMessageHandler{}).sandboxFsChanges(root)

	require.NoError(t, err)
	require.Empty(t, result.GetError())
	require.Len(t, result.GetChanges(), 2)
	assert.Equal(t, "README.md", result.GetChanges()[0].GetPath())
	assert.Equal(t, "created", result.GetChanges()[0].GetStatus())
	assert.Equal(t, "deliverables/showcase/preview.png", result.GetChanges()[1].GetPath())
	assert.Equal(t, "created", result.GetChanges()[1].GetStatus())
}

func TestSandboxFsChangesDoesNotMaskBrokenGitWorkspace(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(root, ".git"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "result.png"), []byte("png"), 0o644))

	result, err := (&RunnerMessageHandler{}).sandboxFsChanges(root)

	require.NoError(t, err)
	require.NotEmpty(t, result.GetError())
	assert.Empty(t, result.GetChanges())
}

func TestSandboxFsReadRejectsSymlinkOutsideWorkspace(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret.txt")
	require.NoError(t, os.WriteFile(outside, []byte("secret"), 0o644))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "secret.txt")))

	result, err := (&RunnerMessageHandler{}).sandboxFsRead(root, "secret.txt")

	require.NoError(t, err)
	assert.NotEmpty(t, result.GetError())
	assert.Empty(t, result.GetContent())
}

func TestSandboxFsWriteRejectsSymlinkDirectoryOutsideWorkspace(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "outside")))

	result, err := (&RunnerMessageHandler{}).sandboxFsWrite(
		root,
		"outside/result.txt",
		"must not escape",
	)

	require.NoError(t, err)
	assert.NotEmpty(t, result.GetError())
	_, statErr := os.Stat(filepath.Join(outside, "result.txt"))
	assert.ErrorIs(t, statErr, os.ErrNotExist)
}

func TestSandboxFsMkdirRejectsSymlinkDirectoryOutsideWorkspace(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "outside")))

	result, err := (&RunnerMessageHandler{}).sandboxFsMkdir(root, "outside/generated")

	require.NoError(t, err)
	assert.NotEmpty(t, result.GetError())
	_, statErr := os.Stat(filepath.Join(outside, "generated"))
	assert.ErrorIs(t, statErr, os.ErrNotExist)
}

func TestDetachedPodWorkspaceRootResolvesCompletedSandbox(t *testing.T) {
	workspaceRoot := t.TempDir()
	expected := filepath.Join(
		workspaceRoot,
		"sandboxes",
		"7-standalone-b9f1b3cc",
		"workspace",
	)
	require.NoError(t, os.MkdirAll(expected, 0o755))

	root, err := detachedPodWorkspaceRoot(
		&config.Config{WorkspaceRoot: workspaceRoot},
		"7-standalone-b9f1b3cc",
	)

	require.NoError(t, err)
	assert.Equal(t, expected, root)
}

func TestDetachedPodWorkspaceRootResolvesResumedPodAlias(t *testing.T) {
	workspaceRoot := t.TempDir()
	sourceSandbox := filepath.Join(
		workspaceRoot,
		"sandboxes",
		"7-standalone-b9f1b3cc",
	)
	expected := filepath.Join(sourceSandbox, "workspace")
	require.NoError(t, os.MkdirAll(expected, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(sourceSandbox, "pod_daemon.json"),
		[]byte(`{"pod_key":"7-standalone-bc2bdd58"}`),
		0o600,
	))

	root, err := detachedPodWorkspaceRoot(
		&config.Config{WorkspaceRoot: workspaceRoot},
		"7-standalone-bc2bdd58",
	)

	require.NoError(t, err)
	assert.Equal(t, expected, root)
}

func TestDetachedPodWorkspaceRootRejectsInvalidPodKey(t *testing.T) {
	_, err := detachedPodWorkspaceRoot(
		&config.Config{WorkspaceRoot: t.TempDir()},
		"../outside",
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid pod key")
}

func TestDetachedPodWorkspaceRootRejectsSymlinkOutsideRunnerWorkspace(t *testing.T) {
	workspaceRoot := t.TempDir()
	outside := t.TempDir()
	sandbox := filepath.Join(workspaceRoot, "sandboxes", "completed-pod")
	require.NoError(t, os.MkdirAll(sandbox, 0o755))
	require.NoError(t, os.Symlink(outside, filepath.Join(sandbox, "workspace")))

	_, err := detachedPodWorkspaceRoot(
		&config.Config{WorkspaceRoot: workspaceRoot},
		"completed-pod",
	)

	require.Error(t, err)
}
