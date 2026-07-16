//go:build windows

package runner

import (
	"os"
	"path/filepath"
	"testing"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActivePodSandboxFsRejectsWorkspaceReplacement(t *testing.T) {
	workspace := filepath.Join(t.TempDir(), "workspace")
	require.NoError(t, os.Mkdir(workspace, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "result.txt"), []byte("original"), 0o644))
	pod := &Pod{PodKey: "active-pod", WorkDir: workspace}
	require.NoError(t, pod.pinWorkspace())
	t.Cleanup(pod.closeWorkspace)

	require.NoError(t, os.Rename(workspace, workspace+"-moved"))
	require.NoError(t, os.Mkdir(workspace, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "result.txt"), []byte("outside"), 0o644))

	result, err := pod.withWorkspace(func(root *sandboxWorkspace) (*runnerv1.SandboxFsResultEvent, error) {
		return (&RunnerMessageHandler{}).sandboxFsReadWorkspace(root, "result.txt")
	})

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace identity changed")
}
