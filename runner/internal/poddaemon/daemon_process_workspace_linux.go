//go:build linux

package poddaemon

import (
	"os"
	"os/exec"
	"strconv"

	"github.com/creack/pty"
)

func startDaemonPTY(
	command *exec.Cmd,
	workDir string,
	workspace *os.File,
	size *pty.Winsize,
) (*os.File, error) {
	if workspace == nil {
		command.Dir = workDir
	} else {
		command.Dir = "/proc/self/fd/" +
			strconv.FormatUint(uint64(workspace.Fd()), 10)
	}
	return pty.StartWithSize(command, size)
}
