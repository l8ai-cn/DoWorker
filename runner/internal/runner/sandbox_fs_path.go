package runner

import (
	"fmt"
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
	rel = strings.TrimPrefix(strings.TrimSpace(rel), "/")
	rel = filepath.Clean(rel)
	if rel == "." {
		rel = ""
	}
	if strings.HasPrefix(rel, "..") {
		return "", "", fmt.Errorf("path escapes workspace")
	}
	abs := workspaceRoot
	if rel != "" {
		abs = filepath.Join(workspaceRoot, rel)
	}
	abs, err := filepath.Abs(abs)
	if err != nil {
		return "", "", err
	}
	if abs != workspaceRoot && !strings.HasPrefix(abs, workspaceRoot+string(filepath.Separator)) {
		return "", "", fmt.Errorf("path escapes workspace")
	}
	display := rel
	if display == "" {
		display = "."
	}
	return abs, display, nil
}
