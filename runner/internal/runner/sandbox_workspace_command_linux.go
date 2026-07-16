//go:build linux

package runner

import (
	"context"
	"os/exec"
	"strconv"
)

func runSandboxCommand(
	ctx context.Context,
	workspace *sandboxWorkspace,
	name string,
	args ...string,
) (string, error) {
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = "/proc/self/fd/" +
		strconv.FormatUint(uint64(workspace.dir.Fd()), 10)
	return runBoundedCommand(command, maxSandboxFsReadBytes)
}
