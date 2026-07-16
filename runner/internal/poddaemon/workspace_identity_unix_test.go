//go:build !windows

package poddaemon

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceIdentityValidatesOriginalDirectory(t *testing.T) {
	workspace := t.TempDir()
	identity, err := CaptureWorkspaceIdentity(workspace)
	require.NoError(t, err)

	assert.NoError(t, ValidateWorkspaceIdentity(workspace, identity))
}

func TestWorkspaceIdentityRejectsDirectoryReplacement(t *testing.T) {
	parent := t.TempDir()
	workspace := filepath.Join(parent, "workspace")
	require.NoError(t, os.Mkdir(workspace, 0o700))
	identity, err := CaptureWorkspaceIdentity(workspace)
	require.NoError(t, err)
	require.NoError(t, os.Rename(workspace, workspace+"-moved"))
	require.NoError(t, os.Mkdir(workspace, 0o700))

	err = ValidateWorkspaceIdentity(workspace, identity)

	assert.Contains(t, err.Error(), "workspace identity changed")
}

func TestWorkspaceIdentityRejectsSymlink(t *testing.T) {
	parent := t.TempDir()
	workspace := filepath.Join(parent, "workspace")
	require.NoError(t, os.Symlink(t.TempDir(), workspace))

	identity, err := CaptureWorkspaceIdentity(workspace)

	assert.Nil(t, identity)
	assert.Error(t, err)
}

func TestDaemonProcessUsesPinnedWorkspaceAfterPathReplacement(t *testing.T) {
	parent := t.TempDir()
	workspace := filepath.Join(parent, "workspace")
	require.NoError(t, os.Mkdir(workspace, 0o700))
	require.NoError(t, os.WriteFile(
		filepath.Join(workspace, "marker"),
		[]byte("trusted"),
		0o600,
	))
	identity, err := CaptureWorkspaceIdentity(workspace)
	require.NoError(t, err)
	guard, err := OpenWorkspaceLaunchGuard(workspace, identity)
	require.NoError(t, err)
	defer guard.Close()

	require.NoError(t, os.Rename(workspace, workspace+"-moved"))
	require.NoError(t, os.Mkdir(workspace, 0o700))
	require.NoError(t, os.WriteFile(
		filepath.Join(workspace, "marker"),
		[]byte("replacement"),
		0o600,
	))

	proc, err := startDaemonProcessInWorkspace(
		"sh",
		[]string{"-c", "cat marker"},
		workspace,
		guard,
		os.Environ(),
		80,
		24,
	)
	require.NoError(t, err)
	defer proc.Close()
	buffer := make([]byte, 128)
	n, err := proc.Read(buffer)
	require.NoError(t, err)
	assert.Contains(t, string(buffer[:n]), "trusted")
	code, err := proc.Wait()
	require.NoError(t, err)
	assert.Equal(t, 0, code)
}
