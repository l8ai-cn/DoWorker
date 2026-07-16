package runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const workspaceAliasFile = "workspace_alias.json"

var errWorkspaceAliasOutsideRunnerRoot = errors.New("workspace alias escapes runner root")

type workspaceAlias struct {
	WorkspacePath string `json:"workspace_path"`
}

func persistWorkspaceAlias(runnerRoot, podSandboxRoot, workspaceRoot string) error {
	relativeWorkspace, err := runnerRelativePath(runnerRoot, workspaceRoot)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(podSandboxRoot, 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(workspaceAlias{WorkspacePath: relativeWorkspace})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(podSandboxRoot, workspaceAliasFile), data, 0o600)
}

func readWorkspaceAlias(
	runnerRoot *os.Root,
	podKey string,
) (string, bool, error) {
	data, err := runnerRoot.ReadFile(
		filepath.Join("sandboxes", podKey, workspaceAliasFile),
	)
	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	var alias workspaceAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return "", false, fmt.Errorf("invalid workspace alias: %w", err)
	}
	path := filepath.Clean(strings.TrimSpace(alias.WorkspacePath))
	if path == "." || filepath.IsAbs(path) || path == ".." ||
		strings.HasPrefix(path, ".."+string(filepath.Separator)) {
		return "", false, fmt.Errorf("invalid workspace alias path")
	}
	return path, true, nil
}

func runnerRelativePath(runnerRoot, target string) (string, error) {
	root, err := filepath.Abs(runnerRoot)
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return "", err
	}
	relative = filepath.Clean(relative)
	if relative == "." || relative == ".." ||
		strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", errWorkspaceAliasOutsideRunnerRoot
	}
	return relative, nil
}
