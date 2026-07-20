package workspace

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProbeRejectsTokenWithOnlySSHURLWithoutLocalFallback(t *testing.T) {
	mgr, err := NewManager(t.TempDir(), "")
	require.NoError(t, err)
	marker := installNoProbeFallbackHarness(t)

	_, err = mgr.probeRepositoryAccess(
		context.Background(),
		"",
		"ssh://git@private.test/org/repo.git",
		&WorktreeOptions{GitToken: "token"},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no clone URL is compatible")
	assert.NoFileExists(t, marker)
}

func TestProbeRejectsSSHKeyWithOnlyHTTPURLWithoutLocalFallback(t *testing.T) {
	mgr, err := NewManager(t.TempDir(), "")
	require.NoError(t, err)
	marker := installNoProbeFallbackHarness(t)

	_, err = mgr.probeRepositoryAccess(
		context.Background(),
		"https://private.test/org/repo.git",
		"",
		&WorktreeOptions{SSHKeyPath: filepath.Join(t.TempDir(), "id_ed25519")},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no clone URL is compatible")
	assert.NoFileExists(t, marker)
}

func TestProbeRejectsTokenOverHTTPBeforeRunningGit(t *testing.T) {
	mgr, err := NewManager(t.TempDir(), "")
	require.NoError(t, err)
	marker := installNoProbeFallbackHarness(t)

	_, err = mgr.probeRepositoryAccess(
		context.Background(),
		"http://private.test/org/repo.git",
		"",
		&WorktreeOptions{GitUsername: "oauth2", GitToken: "token"},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires an HTTPS")
	assert.NoFileExists(t, marker)
}

func TestCloneRejectsTokenOverHTTPBeforeRunningGit(t *testing.T) {
	mgr, err := NewManager(t.TempDir(), "")
	require.NoError(t, err)
	marker := installNoProbeFallbackHarness(t)

	err = mgr.cloneBareRepository(
		context.Background(),
		"http://private.test/org/repo.git",
		filepath.Join(t.TempDir(), "repo.git"),
		&WorktreeOptions{GitUsername: "oauth2", GitToken: "token"},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires an HTTPS")
	assert.NoFileExists(t, marker)
}

func TestProbeRejectsHTTPQueryCredentialsBeforeRunningGit(t *testing.T) {
	mgr, err := NewManager(t.TempDir(), "")
	require.NoError(t, err)
	marker := installNoProbeFallbackHarness(t)

	_, err = mgr.probeRepositoryAccess(
		context.Background(),
		"https://private.test/org/repo.git?access_token=query-secret",
		"",
		&WorktreeOptions{AnonymousAuth: true},
	)

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "query-secret")
	assert.NoFileExists(t, marker)
}

func installNoProbeFallbackHarness(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	marker := filepath.Join(dir, "git-called")
	name := "git"
	content := "#!/bin/sh\n: > \"$GIT_FALLBACK_MARKER\"\nexit 42\n"
	if runtime.GOOS == "windows" {
		name = "git.bat"
		content = "@echo off\r\necho called>\"%GIT_FALLBACK_MARKER%\"\r\nexit /b 42\r\n"
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0755))
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GIT_FALLBACK_MARKER", marker)
	return marker
}
