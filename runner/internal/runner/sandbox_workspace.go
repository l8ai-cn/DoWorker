package runner

import (
	"fmt"
	"os"
	"path/filepath"
)

type sandboxWorkspace struct {
	path string
	root *os.Root
}

func openSandboxWorkspace(path string) (*sandboxWorkspace, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(abs)
	if err != nil {
		return nil, err
	}
	return bindSandboxWorkspace(root)
}

func bindSandboxWorkspace(root *os.Root) (*sandboxWorkspace, error) {
	dir, err := os.Open(root.Name())
	if err != nil {
		_ = root.Close()
		return nil, err
	}
	dirInfo, dirErr := dir.Stat()
	rootInfo, rootErr := root.Stat(".")
	if dirErr != nil || rootErr != nil || !os.SameFile(dirInfo, rootInfo) {
		_ = root.Close()
		_ = dir.Close()
		return nil, fmt.Errorf("workspace changed while opening")
	}
	_ = dir.Close()
	return &sandboxWorkspace{path: root.Name(), root: root}, nil
}

func (workspace *sandboxWorkspace) Close() {
	if workspace == nil {
		return
	}
	if workspace.root != nil {
		_ = workspace.root.Close()
	}
}

func (workspace *sandboxWorkspace) displayPath() string {
	if workspace == nil {
		return ""
	}
	return workspace.path
}

func (workspace *sandboxWorkspace) validateCurrentPath() error {
	current, err := os.OpenRoot(workspace.path)
	if err != nil {
		return fmt.Errorf("workspace path changed: %w", err)
	}
	defer current.Close()
	pinnedInfo, pinnedErr := workspace.root.Stat(".")
	currentInfo, currentErr := current.Stat(".")
	if pinnedErr != nil || currentErr != nil || !os.SameFile(pinnedInfo, currentInfo) {
		return fmt.Errorf("workspace path changed")
	}
	return nil
}
