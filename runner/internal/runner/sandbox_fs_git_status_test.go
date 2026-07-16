package runner

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSandboxFsDiffTreatsStagedAdditionAsCreated(t *testing.T) {
	workspace := initGitStatusRepository(t)
	path := filepath.Join(workspace, "new.txt")
	require.NoError(t, os.WriteFile(path, []byte("new content"), 0o600))
	runGitStatusTestCommand(t, workspace, "add", "new.txt")

	result, err := (&RunnerMessageHandler{}).sandboxFsDiff(
		workspace,
		"new.txt",
	)

	require.NoError(t, err)
	assert.Empty(t, result.GetError())
	assert.Empty(t, result.GetBefore())
	assert.Equal(t, "new content", result.GetAfter())
}

func TestSandboxFsChangesTreatsStagedAdditionAsCreated(t *testing.T) {
	workspace := initGitStatusRepository(t)
	path := filepath.Join(workspace, "new.txt")
	require.NoError(t, os.WriteFile(path, []byte("new content"), 0o600))
	runGitStatusTestCommand(t, workspace, "add", "new.txt")

	result, err := (&RunnerMessageHandler{}).sandboxFsChanges(workspace)

	require.NoError(t, err)
	require.Len(t, result.GetChanges(), 1)
	assert.Equal(t, "created", result.GetChanges()[0].GetStatus())
	assert.Equal(t, "new.txt", result.GetChanges()[0].GetPath())
}

func initGitStatusRepository(t *testing.T) string {
	t.Helper()
	workspace := t.TempDir()
	runGitStatusTestCommand(t, workspace, "init", "-q")
	runGitStatusTestCommand(t, workspace, "config", "user.email", "test@example.com")
	runGitStatusTestCommand(t, workspace, "config", "user.name", "Test")
	return workspace
}

func runGitStatusTestCommand(
	t *testing.T,
	workspace string,
	args ...string,
) {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = workspace
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))
}
