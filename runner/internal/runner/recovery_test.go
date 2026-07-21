package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/l8ai-cn/agentcloud/runner/internal/config"
	"github.com/l8ai-cn/agentcloud/runner/internal/poddaemon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- recoverDaemonSessions guard tests ---

func TestRecoverDaemonSessions_NilManager(t *testing.T) {
	r, _ := NewTestRunner(t)

	// Ensure podDaemonManager is nil (default from NewTestRunner).
	r.podDaemonManager = nil

	// Must return immediately without panic.
	r.recoverDaemonSessions()

	assert.Equal(t, 0, r.podStore.Count(), "no pods should be added when manager is nil")
}

func TestRecoverDaemonSessionsRejectsMissingWorkspaceIdentity(t *testing.T) {
	runnerRoot := t.TempDir()
	sandbox := filepath.Join(runnerRoot, "sandbox")
	require.NoError(t, os.Mkdir(sandbox, 0o700))
	require.NoError(t, poddaemon.SaveState(&poddaemon.PodDaemonState{
		PodKey: "missing-id", SandboxPath: sandbox,
		WorkDir: t.TempDir(), Perpetual: true,
	}))
	manager, err := poddaemon.NewPodDaemonManager(runnerRoot)
	require.NoError(t, err)
	r := &Runner{
		cfg:              &config.Config{WorkspaceRoot: runnerRoot},
		podDaemonManager: manager,
		podStore:         NewInMemoryPodStore(),
	}

	r.recoverDaemonSessions()

	assert.Equal(t, 0, r.podStore.Count())
	_, err = os.Stat(poddaemon.StatePath(sandbox))
	assert.True(t, os.IsNotExist(err))
}

func TestRecoverDaemonSessionsRejectsWorkspaceReplacement(t *testing.T) {
	runnerRoot := t.TempDir()
	sandbox := filepath.Join(runnerRoot, "sandbox")
	workspace := filepath.Join(sandbox, "workspace")
	require.NoError(t, os.MkdirAll(workspace, 0o700))
	identity, err := poddaemon.CaptureWorkspaceIdentity(workspace)
	require.NoError(t, err)
	require.NoError(t, poddaemon.SaveState(&poddaemon.PodDaemonState{
		PodKey: "replaced", SandboxPath: sandbox,
		WorkDir: workspace, WorkspaceID: identity, Perpetual: true,
	}))
	require.NoError(t, os.Rename(workspace, workspace+"-moved"))
	require.NoError(t, os.Mkdir(workspace, 0o700))
	manager, err := poddaemon.NewPodDaemonManager(runnerRoot)
	require.NoError(t, err)
	r := &Runner{
		cfg:              &config.Config{WorkspaceRoot: runnerRoot},
		podDaemonManager: manager,
		podStore:         NewInMemoryPodStore(),
	}

	r.recoverDaemonSessions()

	assert.Equal(t, 0, r.podStore.Count())
	_, err = os.Stat(poddaemon.StatePath(sandbox))
	assert.True(t, os.IsNotExist(err))
}

// --- IsDraining / SetDraining with nil upgradeCoord ---

func TestIsDraining_NilUpgradeCoord(t *testing.T) {
	store := NewInMemoryPodStore()
	r := &Runner{
		cfg:      &config.Config{WorkspaceRoot: t.TempDir()},
		podStore: store,
		// upgradeCoord intentionally nil
	}

	assert.False(t, r.IsDraining(), "IsDraining should return false when upgradeCoord is nil")
}

func TestSetDraining_NilUpgradeCoord(t *testing.T) {
	store := NewInMemoryPodStore()
	r := &Runner{
		cfg:      &config.Config{WorkspaceRoot: t.TempDir()},
		podStore: store,
		// upgradeCoord intentionally nil
	}

	// Must not panic.
	r.SetDraining(true)
	assert.False(t, r.IsDraining(), "IsDraining should still return false after SetDraining on nil coord")
}

// --- CanAcceptPod with nil upgradeCoord ---

func TestCanAcceptPod_NilUpgradeCoord(t *testing.T) {
	store := NewInMemoryPodStore()
	r := &Runner{
		cfg: &config.Config{
			WorkspaceRoot:     t.TempDir(),
			MaxConcurrentPods: 5,
		},
		podStore: store,
		// upgradeCoord intentionally nil — IsDraining() returns false
	}

	assert.True(t, r.CanAcceptPod(), "should accept pod when upgradeCoord is nil and below limit")
}

// --- Delegation methods with nil upgradeCoord ---

func TestTryStartUpgrade_NilUpgradeCoord(t *testing.T) {
	r := &Runner{
		cfg:      &config.Config{WorkspaceRoot: t.TempDir()},
		podStore: NewInMemoryPodStore(),
	}

	assert.False(t, r.TryStartUpgrade(), "TryStartUpgrade should return false with nil upgradeCoord")
}

func TestFinishUpgrade_NilUpgradeCoord(t *testing.T) {
	r := &Runner{
		cfg:      &config.Config{WorkspaceRoot: t.TempDir()},
		podStore: NewInMemoryPodStore(),
	}

	// Must not panic.
	r.FinishUpgrade()
}

func TestGetUpdater_NilUpgradeCoord(t *testing.T) {
	r := &Runner{
		cfg:      &config.Config{WorkspaceRoot: t.TempDir()},
		podStore: NewInMemoryPodStore(),
	}

	assert.Nil(t, r.GetUpdater(), "GetUpdater should return nil with nil upgradeCoord")
}

func TestGetRestartFunc_NilUpgradeCoord(t *testing.T) {
	r := &Runner{
		cfg:      &config.Config{WorkspaceRoot: t.TempDir()},
		podStore: NewInMemoryPodStore(),
	}

	assert.Nil(t, r.GetRestartFunc(), "GetRestartFunc should return nil with nil upgradeCoord")
}
