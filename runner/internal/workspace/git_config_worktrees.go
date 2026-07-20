package workspace

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type listedWorktree struct {
	path     string
	bare     bool
	prunable bool
}

func (m *Manager) configurableWorktrees(
	ctx context.Context,
	worktreePath string,
) ([]string, error) {
	commonDir, err := m.gitCommonDir(ctx, worktreePath)
	if err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, "git", "worktree", "list", "--porcelain", "-z")
	cmd.Dir = worktreePath
	m.setLocalGitEnv(cmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees for Git config migration: %w, output: %s", err, output)
	}
	var records []listedWorktree
	var current *listedWorktree
	for _, field := range strings.Split(string(output), "\x00") {
		switch {
		case strings.HasPrefix(field, "worktree "):
			if current != nil {
				records = append(records, *current)
			}
			current = &listedWorktree{path: strings.TrimPrefix(field, "worktree ")}
		case field == "bare" && current != nil:
			current.bare = true
		case strings.HasPrefix(field, "prunable") && current != nil:
			current.prunable = true
		}
	}
	if current != nil {
		records = append(records, *current)
	}
	var paths []string
	for _, record := range records {
		if !record.bare && !record.prunable && record.path != "" &&
			canonicalGitPath(record.path) != commonDir {
			paths = append(paths, record.path)
		}
	}
	return paths, nil
}

func (m *Manager) gitCommonDir(
	ctx context.Context,
	worktreePath string,
) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-common-dir")
	cmd.Dir = worktreePath
	m.setLocalGitEnv(cmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to locate common Git directory: %w, output: %s", err, output)
	}
	path := strings.TrimSpace(string(output))
	if !filepath.IsAbs(path) {
		path = filepath.Join(worktreePath, path)
	}
	return canonicalGitPath(path), nil
}

func canonicalGitPath(path string) string {
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		path = resolved
	}
	absolute, err := filepath.Abs(path)
	if err == nil {
		path = absolute
	}
	return filepath.Clean(path)
}

func (m *Manager) prepareNonBareWorktreeConfig(
	ctx context.Context,
	worktreePath string,
) error {
	gitDir, err := m.worktreeGitDir(ctx, worktreePath)
	if err != nil {
		return err
	}
	configPath := filepath.Join(gitDir, "config.worktree")
	cmd := exec.CommandContext(ctx, "git", "config", "--file", configPath, "--replace-all", "core.bare", "false")
	m.setLocalGitEnv(cmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to prepare non-bare Git config for %s: %w, output: %s", worktreePath, err, output)
	}
	return nil
}
