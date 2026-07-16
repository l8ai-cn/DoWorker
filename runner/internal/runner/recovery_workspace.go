package runner

import (
	"fmt"

	"github.com/anthropics/agentsmesh/runner/internal/poddaemon"
)

func openRecoveredSandboxWorkspace(
	state *poddaemon.PodDaemonState,
) (*sandboxWorkspace, error) {
	if err := poddaemon.ValidateWorkspaceIdentity(
		state.WorkDir,
		state.WorkspaceID,
	); err != nil {
		return nil, err
	}
	workspace, err := openSandboxWorkspace(state.WorkDir)
	if err != nil {
		return nil, err
	}
	if err := poddaemon.ValidateWorkspaceFile(
		workspace.dir,
		state.WorkspaceID,
	); err != nil {
		workspace.Close()
		return nil, fmt.Errorf("validate pinned workspace: %w", err)
	}
	if err := poddaemon.ValidateWorkspaceIdentity(
		state.WorkDir,
		state.WorkspaceID,
	); err != nil {
		workspace.Close()
		return nil, fmt.Errorf("revalidate workspace path: %w", err)
	}
	return workspace, nil
}
