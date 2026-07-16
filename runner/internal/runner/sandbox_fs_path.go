package runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/anthropics/agentsmesh/runner/internal/config"
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

func detachedPodWorkspaceRoot(cfg *config.Config, podKey string) (string, error) {
	workspace, err := openDetachedPodWorkspace(cfg, podKey)
	if err != nil {
		return "", err
	}
	defer workspace.Close()
	return workspace.displayPath(), nil
}

func openDetachedPodWorkspace(
	cfg *config.Config,
	podKey string,
) (*sandboxWorkspace, error) {
	if cfg == nil || strings.TrimSpace(cfg.WorkspaceRoot) == "" {
		return nil, fmt.Errorf("workspace not configured")
	}
	if err := slugkit.Validate(podKey); err != nil {
		return nil, fmt.Errorf("invalid pod key: %w", err)
	}
	runnerRoot, err := os.OpenRoot(cfg.WorkspaceRoot)
	if err != nil {
		return nil, err
	}
	defer runnerRoot.Close()
	workspacePath := filepath.Join("sandboxes", podKey, "workspace")
	workspace, directErr := runnerRoot.OpenRoot(workspacePath)
	if directErr == nil {
		return bindSandboxWorkspace(workspace)
	}
	aliasPath, found, err := readWorkspaceAlias(runnerRoot, podKey)
	if err != nil {
		return nil, err
	}
	if found {
		workspace, err = runnerRoot.OpenRoot(aliasPath)
		if err != nil {
			return nil, fmt.Errorf("pod workspace not found: %w", err)
		}
		return bindSandboxWorkspace(workspace)
	}
	aliasPath, found, err = resumedPodWorkspacePath(runnerRoot, podKey)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("pod workspace not found: %w", directErr)
	}
	workspace, err = runnerRoot.OpenRoot(aliasPath)
	if err != nil {
		return nil, fmt.Errorf("pod workspace not found: %w", err)
	}
	return bindSandboxWorkspace(workspace)
}

func resumedPodWorkspacePath(runnerRoot *os.Root, podKey string) (string, bool, error) {
	entries, err := fs.ReadDir(runnerRoot.FS(), "sandboxes")
	if err != nil {
		return "", false, err
	}
	matches := make([]string, 0, 1)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		statePath := filepath.Join("sandboxes", entry.Name(), "pod_daemon.json")
		data, readErr := runnerRoot.ReadFile(statePath)
		if errors.Is(readErr, os.ErrNotExist) {
			continue
		}
		if readErr != nil {
			return "", false, readErr
		}
		var state struct {
			PodKey string `json:"pod_key"`
		}
		if json.Unmarshal(data, &state) != nil || state.PodKey != podKey {
			continue
		}
		matches = append(matches, filepath.Join("sandboxes", entry.Name(), "workspace"))
	}
	if len(matches) == 0 {
		return "", false, nil
	}
	if len(matches) > 1 {
		return "", false, fmt.Errorf("multiple workspaces found for pod %s", podKey)
	}
	return matches[0], true, nil
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
