//go:build darwin

package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"golang.org/x/sys/unix"
)

func runSandboxCommand(
	ctx context.Context,
	workspace *sandboxWorkspace,
	name string,
	args ...string,
) (output string, commandErr error) {
	original, err := os.Open(".")
	if err != nil {
		return "", fmt.Errorf("open current directory: %w", err)
	}
	defer original.Close()

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := unix.PthreadFchdir(int(workspace.dir.Fd())); err != nil {
		return "", fmt.Errorf("enter pinned workspace: %w", err)
	}
	defer func() {
		if err := unix.PthreadFchdir(int(original.Fd())); err != nil {
			output = ""
			commandErr = fmt.Errorf("restore current directory: %w", err)
		}
	}()

	command := exec.CommandContext(ctx, name, args...)
	return runBoundedCommand(command, maxSandboxFsReadBytes)
}
