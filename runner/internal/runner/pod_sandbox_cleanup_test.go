package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanupPodSandboxRemovesOwnedWorktree(t *testing.T) {
	root := t.TempDir()
	podKey := "cleanup-pod"
	sandboxPath := filepath.Join(root, "sandboxes", podKey)
	worktreePath := filepath.Join(sandboxPath, "workspace")
	require.NoError(t, os.MkdirAll(worktreePath, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(worktreePath, ".git"), []byte("gitdir: test"), 0644))
	manager := &mockWorkspace{}

	err := cleanupPodSandbox(context.Background(), manager, root, podKey, sandboxPath)

	require.NoError(t, err)
	assert.Equal(t, []string{worktreePath}, manager.removedWorktreeIDs)
	assert.NoDirExists(t, sandboxPath)
}

func TestCleanupPodSandboxPreservesReusedWorkspace(t *testing.T) {
	root := t.TempDir()
	podKey := "resumed-pod"
	ownedSandboxPath := filepath.Join(root, "sandboxes", podKey)
	reusedSandboxPath := filepath.Join(root, "sandboxes", "source-pod")
	require.NoError(t, os.MkdirAll(ownedSandboxPath, 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(reusedSandboxPath, "workspace"), 0755))
	sentinel := filepath.Join(reusedSandboxPath, "workspace", "sentinel.txt")
	require.NoError(t, os.WriteFile(sentinel, []byte("keep"), 0644))

	err := cleanupPodSandbox(
		context.Background(),
		&mockWorkspace{},
		root,
		podKey,
		reusedSandboxPath,
	)

	require.NoError(t, err)
	assert.NoDirExists(t, ownedSandboxPath)
	assert.FileExists(t, sentinel)
}
