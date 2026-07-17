package runner

import (
	"fmt"

	"github.com/anthropics/agentsmesh/runner/internal/poddaemon"
)

func (r *Runner) restartDeadPerpetualDaemon(
	state *poddaemon.PodDaemonState,
) (*Pod, error) {
	if err := poddaemon.ValidateWorkspaceIdentity(
		state.WorkDir,
		state.WorkspaceID,
	); err != nil {
		return nil, err
	}
	_, updatedState, err := r.podDaemonManager.CreateSession(
		poddaemon.CreateOpts{
			PodKey: state.PodKey, Agent: state.Agent,
			Command: state.Command, Args: state.Args,
			WorkDir: state.WorkDir, WorkspaceID: state.WorkspaceID,
			Env: state.Env, Cols: state.Cols, Rows: state.Rows,
			SandboxPath:   state.SandboxPath,
			RepositoryURL: state.RepositoryURL,
			Branch:        state.Branch, TicketSlug: state.TicketSlug,
			VTHistoryLimit: state.VTHistoryLimit, Perpetual: true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create daemon session: %w", err)
	}
	return r.recoverSingleSession(updatedState)
}
