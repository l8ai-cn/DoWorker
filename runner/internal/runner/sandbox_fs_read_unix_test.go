//go:build !windows

package runner

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSandboxFsReadRejectsFIFOWithoutBlocking(t *testing.T) {
	workspace := t.TempDir()
	fifo := filepath.Join(workspace, "stream")
	require.NoError(t, syscall.Mkfifo(fifo, 0o600))
	resultCh := make(chan *runnerv1.SandboxFsResultEvent, 1)

	go func() {
		result, _ := (&RunnerMessageHandler{}).sandboxFsRead(workspace, "stream")
		resultCh <- result
	}()

	select {
	case result := <-resultCh:
		assert.Contains(t, result.GetError(), "not a regular file")
	case <-time.After(200 * time.Millisecond):
		writer, err := os.OpenFile(fifo, os.O_WRONLY|syscall.O_NONBLOCK, 0)
		if err == nil {
			_, _ = writer.Write([]byte("release"))
			_ = writer.Close()
		}
		t.Fatal("sandbox filesystem read blocked on FIFO")
	}
}

func TestSandboxFsGitUsesPinnedWorkspaceAfterPathReplacement(t *testing.T) {
	workspacePath := filepath.Join(t.TempDir(), "workspace")
	require.NoError(t, os.Mkdir(workspacePath, 0o755))
	require.NoError(t, exec.Command("git", "init", "-q", workspacePath).Run())
	require.NoError(t, os.WriteFile(
		filepath.Join(workspacePath, "original.txt"),
		[]byte("original"),
		0o644,
	))
	workspace, err := openSandboxWorkspace(workspacePath)
	require.NoError(t, err)
	t.Cleanup(workspace.Close)

	moved := workspacePath + "-moved"
	require.NoError(t, os.Rename(workspacePath, moved))
	outside := t.TempDir()
	require.NoError(t, exec.Command("git", "init", "-q", outside).Run())
	require.NoError(t, os.WriteFile(
		filepath.Join(outside, "outside.txt"),
		[]byte("outside"),
		0o644,
	))
	require.NoError(t, os.Symlink(outside, workspacePath))

	result, err := (&RunnerMessageHandler{}).sandboxFsChangesWorkspace(workspace)

	require.NoError(t, err)
	require.Empty(t, result.GetError())
	require.Len(t, result.GetChanges(), 1)
	assert.Equal(t, "original.txt", result.GetChanges()[0].GetPath())
}
