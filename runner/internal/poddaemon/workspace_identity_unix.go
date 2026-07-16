//go:build linux || darwin

package poddaemon

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

const workspaceIdentityKind = "unix-dev-inode-v1"

func openWorkspaceIdentityFile(path string) (*os.File, error) {
	fd, err := unix.Open(
		path,
		unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY|unix.O_NOFOLLOW,
		0,
	)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), path), nil
}

func openWorkspaceLaunchFile(path string) (*os.File, error) {
	return openWorkspaceIdentityFile(path)
}

func workspaceFileIdentity(file *os.File) (string, uint64, uint64, error) {
	info, err := file.Stat()
	if err != nil {
		return "", 0, 0, fmt.Errorf("stat workspace identity: %w", err)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "", 0, 0, fmt.Errorf("workspace identity unavailable")
	}
	return workspaceIdentityKind, uint64(stat.Dev), uint64(stat.Ino), nil
}
