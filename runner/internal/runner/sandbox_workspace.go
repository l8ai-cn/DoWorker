package runner

import (
	"fmt"
	"os"
	"path/filepath"
)

func openSandboxWorkspace(path string) (*sandboxWorkspace, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(abs)
	if err != nil {
		return nil, err
	}
	return bindSandboxWorkspace(root, abs)
}

func bindSandboxWorkspace(root *os.Root, path string) (*sandboxWorkspace, error) {
	dir, err := root.Open(".")
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
	return &sandboxWorkspace{path: filepath.Clean(path), root: root, dir: dir}, nil
}

func (workspace *sandboxWorkspace) Close() {
	if workspace == nil {
		return
	}
	if workspace.root != nil {
		_ = workspace.root.Close()
		workspace.root = nil
	}
	if workspace.dir != nil {
		_ = workspace.dir.Close()
		workspace.dir = nil
	}
}

func (workspace *sandboxWorkspace) displayPath() string {
	if workspace == nil {
		return ""
	}
	return workspace.path
}
