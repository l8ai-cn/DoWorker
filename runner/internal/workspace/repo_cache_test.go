package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExistingRepoFetchFailurePreservesCacheAndActiveWorktree(t *testing.T) {
	origin, clone := createPinnedOrigin(t)
	commit := commitPinnedFile(t, clone, "cached")
	pushPinnedBranch(t, clone)
	root := t.TempDir()
	mgr, err := NewManager(root, "")
	require.NoError(t, err)

	first, err := mgr.CreateWorktreeWithOptions(
		testGitContext(t),
		origin,
		"main",
		filepath.Join(root, "sandboxes", "first", "workspace"),
		WithSourceCommitSHA(commit),
	)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(first.Path, "local.txt"), []byte("local"), 0644))
	runPinnedGit(t, first.Path, "add", ".")
	runPinnedGit(t, first.Path, "commit", "-m", "local")
	localCommit := gitHead(t, first.Path)
	require.NotEqual(t, commit, localCommit)

	require.NoError(t, os.Rename(origin, origin+".offline"))
	_, err = mgr.CreateWorktreeWithOptions(
		testGitContext(t),
		origin,
		"main",
		filepath.Join(root, "sandboxes", "second", "workspace"),
		WithSourceCommitSHA(commit),
		WithAnonymousAuth(),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch existing repository")
	assert.DirExists(t, filepath.Join(root, "repos", extractRepoName(origin)))
	assert.Equal(t, localCommit, gitHead(t, first.Path))
	assert.Equal(t, "local", readPinnedFilePath(t, filepath.Join(first.Path, "local.txt")))
}

func readPinnedFilePath(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(content)
}
