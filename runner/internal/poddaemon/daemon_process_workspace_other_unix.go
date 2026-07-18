//go:build !windows && !linux && !darwin

package poddaemon

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/creack/pty"
)

func startDaemonPTY(
	command *exec.Cmd,
	workDir string,
	workspace *os.File,
	size *pty.Winsize,
) (*os.File, error) {
	if workspace != nil {
		return nil, fmt.Errorf(
			"secure pinned workspace launch is unavailable on this platform",
		)
	}
	command.Dir = workDir
	return pty.StartWithSize(command, size)
}
