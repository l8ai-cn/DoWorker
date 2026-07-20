package workspace

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type gitConfigEntry struct {
	key   string
	value string
}

func (m *Manager) applyGitConfig(ctx context.Context, worktreePath string) error {
	if m.gitConfigPath == "" {
		return nil
	}
	entries, err := m.gitConfigEntries(ctx)
	if err != nil {
		return err
	}
	if err := m.enableWorktreeConfig(ctx, worktreePath); err != nil {
		return err
	}
	for _, entry := range entries {
		if err := m.runWorktreeGitConfig(ctx, worktreePath, "--add", entry.key, entry.value); err != nil {
			return fmt.Errorf("failed to apply git config %s: %w", entry.key, err)
		}
	}
	return nil
}

func (m *Manager) gitConfigEntries(ctx context.Context) ([]gitConfigEntry, error) {
	cmd := exec.CommandContext(ctx, "git", "config", "--file", m.gitConfigPath, "--list")
	m.setLocalGitEnv(cmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list git config entries: %w, output: %s", err, output)
	}
	var entries []gitConfigEntry
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("invalid git config entry: %s", line)
		}
		entries = append(entries, gitConfigEntry{key: key, value: value})
	}
	return entries, nil
}

func (m *Manager) applyAnonymousGitConfig(ctx context.Context, worktreePath string) error {
	if err := m.enableWorktreeConfig(ctx, worktreePath); err != nil {
		return err
	}
	for _, args := range [][]string{
		{"--add", "credential.helper", ""},
		{"--add", "http.extraHeader", ""},
		{"--replace-all", "core.sshCommand", anonymousSSHCommand(false)},
	} {
		if err := m.runWorktreeGitConfig(ctx, worktreePath, args...); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) applyWorktreeRemote(ctx context.Context, worktreePath, repoURL string) error {
	if repoURL == "" {
		return fmt.Errorf("repository URL is required for worktree remote")
	}
	if err := m.enableWorktreeConfig(ctx, worktreePath); err != nil {
		return err
	}
	return m.runWorktreeGitConfig(ctx, worktreePath, "--replace-all", "remote.origin.url", repoURL)
}

func (m *Manager) enableWorktreeConfig(ctx context.Context, worktreePath string) error {
	worktrees, err := m.configurableWorktrees(ctx, worktreePath)
	if err != nil {
		return err
	}
	for _, siblingPath := range worktrees {
		if err := m.prepareNonBareWorktreeConfig(ctx, siblingPath); err != nil {
			return err
		}
	}
	cmd := exec.CommandContext(ctx, "git", "config", "extensions.worktreeConfig", "true")
	cmd.Dir = worktreePath
	m.setLocalGitEnv(cmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to enable worktree git config: %w, output: %s", err, output)
	}
	return nil
}

func (m *Manager) runWorktreeGitConfig(ctx context.Context, worktreePath string, args ...string) error {
	cmdArgs := append([]string{"config", "--worktree"}, args...)
	cmd := exec.CommandContext(ctx, "git", cmdArgs...)
	cmd.Dir = worktreePath
	m.setLocalGitEnv(cmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to write worktree git config: %w, output: %s", err, output)
	}
	return nil
}
