package workspace

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneBareRepositoryPostCloneFetchFailsClosedAndRedactsToken(t *testing.T) {
	origin, clone := createPinnedOrigin(t)
	commitPinnedFile(t, clone, "clone")
	pushPinnedBranch(t, clone)
	root := t.TempDir()
	installCloneFetchFailGit(t, origin)
	mgr, err := NewManager(root, "")
	require.NoError(t, err)

	err = mgr.cloneBareRepository(
		testGitContext(t),
		"https://private.test/org/repo.git",
		filepath.Join(root, "repos", "repo"),
		&WorktreeOptions{GitUsername: "oauth2", GitToken: "secret-token"},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch cloned repository refs")
	assert.NotContains(t, err.Error(), "secret-token")
	assert.Contains(t, err.Error(), "[REDACTED]")
	assert.NoDirExists(t, filepath.Join(root, "repos", "repo"))
}

func installCloneFetchFailGit(t *testing.T, origin string) {
	t.Helper()
	realGit, err := exec.LookPath("git")
	require.NoError(t, err)
	dir := t.TempDir()
	name := "git"
	content := cloneFetchFailGitScriptUnix
	if runtime.GOOS == "windows" {
		name = "git.bat"
		content = cloneFetchFailGitScriptWindows
	}
	script := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(script, []byte(content), 0755))
	t.Setenv("REAL_GIT", realGit)
	t.Setenv("GIT_TEST_ORIGIN", origin)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

const cloneFetchFailGitScriptUnix = `#!/bin/sh
set -eu
if [ "${1:-}" = "clone" ]; then
	exec "$REAL_GIT" clone --bare "$GIT_TEST_ORIGIN" "$4"
fi
if [ "${1:-}" = "fetch" ]; then
	printf 'fatal token %s\n' "${GIT_CONFIG_VALUE_2:-}" >&2
	exit 42
fi
exec "$REAL_GIT" "$@"
`

const cloneFetchFailGitScriptWindows = `@echo off
if "%1"=="clone" (
	"%REAL_GIT%" clone --bare "%GIT_TEST_ORIGIN%" "%4"
	exit /b %ERRORLEVEL%
)
if "%1"=="fetch" (
	echo fatal token %GIT_CONFIG_VALUE_2% 1>&2
	exit /b 42
)
"%REAL_GIT%" %*
`
