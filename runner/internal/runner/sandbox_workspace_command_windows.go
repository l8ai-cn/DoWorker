//go:build windows

package runner

import (
	"context"
	"os/exec"

	"github.com/anthropics/agentsmesh/runner/internal/poddaemon"
)

func runSandboxCommand(
	ctx context.Context,
	workspace *sandboxWorkspace,
	name string,
	args ...string,
) (string, error) {
	identity, err := poddaemon.CaptureWorkspaceIdentity(
		workspace.displayPath(),
	)
	if err != nil {
		return "", err
	}
	if err := poddaemon.ValidateWorkspaceFile(
		workspace.dir,
		identity,
	); err != nil {
		return "", err
	}
	guard, err := poddaemon.OpenWorkspaceLaunchGuard(
		workspace.displayPath(),
		identity,
	)
	if err != nil {
		return "", err
	}
	defer guard.Close()

	command := exec.CommandContext(ctx, name, args...)
	command.Dir = workspace.displayPath()
	return runBoundedCommand(command, maxSandboxFsReadBytes)
}
