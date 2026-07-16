package runner

import (
	"context"
	"os/exec"
)

func (workspace *sandboxWorkspace) commandContext(
	ctx context.Context,
	name string,
	args ...string,
) (*exec.Cmd, error) {
	if err := workspace.validateCurrentPath(); err != nil {
		return nil, err
	}
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = workspace.path
	return command, nil
}
