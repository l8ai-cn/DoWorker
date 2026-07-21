package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/l8ai-cn/agentcloud/runner/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersistWorkspaceAliasRejectsPathOutsideRunnerRoot(t *testing.T) {
	runnerRoot := t.TempDir()
	err := persistWorkspaceAlias(
		runnerRoot,
		filepath.Join(runnerRoot, "sandboxes", "resume-pod"),
		t.TempDir(),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "escapes runner root")
}

func TestDetachedPodWorkspaceRootRejectsEscapingAlias(t *testing.T) {
	runnerRoot := t.TempDir()
	podSandbox := filepath.Join(runnerRoot, "sandboxes", "resume-pod")
	require.NoError(t, os.MkdirAll(podSandbox, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(podSandbox, workspaceAliasFile),
		[]byte(`{"workspace_path":"../../outside"}`),
		0o600,
	))

	_, err := detachedPodWorkspaceRoot(
		&config.Config{WorkspaceRoot: runnerRoot},
		"resume-pod",
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid workspace alias path")
}
