//go:build windows

package runner

import (
	"fmt"
	"os"

	"github.com/anthropics/agentsmesh/runner/internal/poddaemon"
)

type sandboxWorkspace struct {
	path     string
	root     *os.Root
	dir      *os.File
	identity *poddaemon.WorkspaceIdentity
}

func (workspace *sandboxWorkspace) pinForPod(
	expected *poddaemon.WorkspaceIdentity,
) error {
	if expected == nil {
		var err error
		expected, err = poddaemon.CaptureWorkspaceIdentity(workspace.path)
		if err != nil {
			return err
		}
	}
	if err := poddaemon.ValidateWorkspaceFile(workspace.dir, expected); err != nil {
		return err
	}
	if err := poddaemon.ValidateWorkspaceIdentity(workspace.path, expected); err != nil {
		return err
	}
	identity := *expected
	workspace.identity = &identity
	workspace.Close()
	return nil
}

func (workspace *sandboxWorkspace) workspaceForOperation() (
	*sandboxWorkspace,
	bool,
	error,
) {
	if workspace.identity == nil {
		return nil, false, fmt.Errorf("workspace identity is missing")
	}
	active, err := openSandboxWorkspace(workspace.path)
	if err != nil {
		return nil, false, err
	}
	if err := poddaemon.ValidateWorkspaceFile(
		active.dir,
		workspace.identity,
	); err != nil {
		active.Close()
		return nil, false, err
	}
	if err := poddaemon.ValidateWorkspaceIdentity(
		workspace.path,
		workspace.identity,
	); err != nil {
		active.Close()
		return nil, false, err
	}
	return active, true, nil
}
