//go:build darwin

package poddaemon

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/creack/pty"
	"golang.org/x/sys/unix"
)

func startDaemonPTY(
	command *exec.Cmd,
	workDir string,
	workspace *os.File,
	size *pty.Winsize,
) (ptyFile *os.File, startErr error) {
	if workspace == nil {
		command.Dir = workDir
		return pty.StartWithSize(command, size)
	}
	original, err := os.Open(".")
	if err != nil {
		return nil, fmt.Errorf("open current directory: %w", err)
	}
	defer original.Close()

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := unix.PthreadFchdir(int(workspace.Fd())); err != nil {
		return nil, fmt.Errorf("enter pinned workspace: %w", err)
	}
	defer func() {
		if err := unix.PthreadFchdir(int(original.Fd())); err != nil {
			if ptyFile != nil {
				ptyFile.Close()
			}
			ptyFile = nil
			startErr = fmt.Errorf("restore current directory: %w", err)
		}
	}()
	return pty.StartWithSize(command, size)
}
