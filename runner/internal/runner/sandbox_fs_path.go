package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const maxSandboxFsReadBytes = 1 << 20

func podWorkspaceRoot(pod *Pod) (string, error) {
	if pod == nil {
		return "", fmt.Errorf("pod not found")
	}
	root := strings.TrimSpace(pod.WorkDir)
	if root == "" && pod.SandboxPath != "" {
		root = filepath.Join(pod.SandboxPath, "workspace")
	}
	if root == "" {
		return "", fmt.Errorf("workspace not configured")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	return abs, nil
}

func resolveWorkspacePath(workspaceRoot, rel string) (string, string, error) {
	relative, display, err := resolveSandboxWorkspaceRelativePath(rel)
	if err != nil {
		return "", "", err
	}
	abs := workspaceRoot
	if relative != "." {
		abs = filepath.Join(workspaceRoot, relative)
	}
	abs, err = filepath.Abs(abs)
	if err != nil {
		return "", "", err
	}
	if abs != workspaceRoot && !strings.HasPrefix(abs, workspaceRoot+string(filepath.Separator)) {
		return "", "", fmt.Errorf("path escapes workspace")
	}
	return abs, display, nil
}

func resolveSandboxWorkspaceRelativePath(rel string) (string, string, error) {
	rel = strings.TrimPrefix(strings.TrimSpace(rel), "/")
	rel = filepath.Clean(rel)
	if rel == "." {
		return ".", ".", nil
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("path escapes workspace")
	}
	return rel, rel, nil
}

func openSandboxWorkspaceRoot(workspaceRoot string) (*os.Root, error) {
	abs, err := filepath.Abs(workspaceRoot)
	if err != nil {
		return nil, err
	}
	return os.OpenRoot(abs)
}
