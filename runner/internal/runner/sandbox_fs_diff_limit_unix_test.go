//go:build !windows

package runner

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSandboxFsDiffRejectsOversizedWorkspaceFile(t *testing.T) {
	workspace := initSandboxGitRepository(t)
	large := bytes.Repeat([]byte("a"), maxSandboxFsReadBytes+1)
	require.NoError(t, os.WriteFile(
		filepath.Join(workspace, "large.txt"),
		large,
		0o600,
	))

	result, err := (&RunnerMessageHandler{}).sandboxFsDiff(
		workspace,
		"large.txt",
	)

	require.NoError(t, err)
	assert.Contains(
		t,
		result.GetError(),
		fmt.Sprintf(
			"sandbox filesystem content exceeds %d byte limit",
			maxSandboxFsReadBytes,
		),
	)
}

func TestSandboxFsDiffRejectsOversizedGitShow(t *testing.T) {
	workspace := initSandboxGitRepository(t)
	path := filepath.Join(workspace, "large.txt")
	large := bytes.Repeat([]byte("a"), maxSandboxFsReadBytes+1)
	require.NoError(t, os.WriteFile(path, large, 0o600))
	runSandboxGit(t, workspace, "add", "large.txt")
	runSandboxGit(t, workspace, "commit", "-qm", "large")
	require.NoError(t, os.WriteFile(path, []byte("small"), 0o600))

	result, err := (&RunnerMessageHandler{}).sandboxFsDiff(
		workspace,
		"large.txt",
	)

	require.NoError(t, err)
	assert.Contains(
		t,
		result.GetError(),
		fmt.Sprintf(
			"git output exceeds %d byte limit",
			maxSandboxFsReadBytes,
		),
	)
}

func initSandboxGitRepository(t *testing.T) string {
	t.Helper()
	workspace := t.TempDir()
	runSandboxGit(t, workspace, "init", "-q")
	runSandboxGit(t, workspace, "config", "user.email", "test@example.com")
	runSandboxGit(t, workspace, "config", "user.name", "Test")
	return workspace
}

func runSandboxGit(t *testing.T, workspace string, args ...string) {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = workspace
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))
}
