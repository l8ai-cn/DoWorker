//go:build !windows

package runner

import (
	"os"
	"path/filepath"
	"testing"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActivePodSandboxFsUsesPinnedWorkspaceAfterPathReplacement(t *testing.T) {
	workspace := filepath.Join(t.TempDir(), "workspace")
	require.NoError(t, os.Mkdir(workspace, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "result.txt"), []byte("original"), 0o644))
	pod := &Pod{PodKey: "active-pod", WorkDir: workspace}
	require.NoError(t, pod.pinWorkspace())
	t.Cleanup(pod.closeWorkspace)

	require.NoError(t, os.Rename(workspace, workspace+"-moved"))
	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outside, "result.txt"), []byte("outside"), 0o644))
	require.NoError(t, os.Symlink(outside, workspace))

	result, err := pod.withWorkspace(func(root *sandboxWorkspace) (*runnerv1.SandboxFsResultEvent, error) {
		return (&RunnerMessageHandler{}).sandboxFsReadWorkspace(root, "result.txt")
	})

	require.NoError(t, err)
	require.Empty(t, result.GetError())
	assert.Equal(t, "original", result.GetContent())
}
