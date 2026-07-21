//go:build !windows

package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/l8ai-cn/agentcloud/runner/internal/poddaemon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenRecoveredWorkspaceRejectsMissingIdentity(t *testing.T) {
	state := &poddaemon.PodDaemonState{WorkDir: t.TempDir()}

	workspace, err := openRecoveredSandboxWorkspace(state)

	assert.Nil(t, workspace)
	assert.Contains(t, err.Error(), "workspace identity is missing")
}

func TestOpenRecoveredWorkspaceRejectsReplacement(t *testing.T) {
	parent := t.TempDir()
	workspacePath := filepath.Join(parent, "workspace")
	require.NoError(t, os.Mkdir(workspacePath, 0o700))
	identity, err := poddaemon.CaptureWorkspaceIdentity(workspacePath)
	require.NoError(t, err)
	require.NoError(t, os.Rename(workspacePath, workspacePath+"-moved"))
	require.NoError(t, os.Mkdir(workspacePath, 0o700))
	state := &poddaemon.PodDaemonState{
		WorkDir: workspacePath, WorkspaceID: identity,
	}

	workspace, err := openRecoveredSandboxWorkspace(state)

	assert.Nil(t, workspace)
	assert.Contains(t, err.Error(), "workspace identity changed")
}

func TestOpenRecoveredWorkspaceRejectsSymlinkReplacement(t *testing.T) {
	parent := t.TempDir()
	workspacePath := filepath.Join(parent, "workspace")
	require.NoError(t, os.Mkdir(workspacePath, 0o700))
	identity, err := poddaemon.CaptureWorkspaceIdentity(workspacePath)
	require.NoError(t, err)
	require.NoError(t, os.Rename(workspacePath, workspacePath+"-moved"))
	require.NoError(t, os.Symlink(t.TempDir(), workspacePath))
	state := &poddaemon.PodDaemonState{
		WorkDir: workspacePath, WorkspaceID: identity,
	}

	workspace, err := openRecoveredSandboxWorkspace(state)

	assert.Nil(t, workspace)
	assert.Error(t, err)
}
