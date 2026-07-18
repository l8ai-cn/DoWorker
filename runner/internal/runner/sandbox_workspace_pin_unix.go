//go:build !windows

package runner

import (
	"os"

	"github.com/anthropics/agentsmesh/runner/internal/poddaemon"
)

type sandboxWorkspace struct {
	path string
	root *os.Root
	dir  *os.File
}

func (workspace *sandboxWorkspace) pinForPod(
	_ *poddaemon.WorkspaceIdentity,
) error {
	return nil
}

func (workspace *sandboxWorkspace) workspaceForOperation() (
	*sandboxWorkspace,
	bool,
	error,
) {
	return workspace, false, nil
}
