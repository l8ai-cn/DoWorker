package workspace

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/l8ai-cn/agentcloud/runner/internal/logger"
)

func (m *Manager) CleanupOldWorktrees(ctx context.Context) error {
	log := logger.Workspace()
	log.Info("Starting worktree cleanup")

	m.mu.Lock()
	defer m.mu.Unlock()

	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()
	cleanedCount, err := m.removeOrphanedSandboxAuth()
	if err != nil {
		return err
	}
	if err := m.pruneCachedRepositoryWorktrees(cleanupCtx); err != nil {
		return err
	}
	log.Info("Worktree cleanup completed", "cleaned_count", cleanedCount)
	return nil
}

func (m *Manager) removeOrphanedSandboxAuth() (int, error) {
	sandboxesDir := filepath.Join(m.root, "sandboxes")
	entries, err := os.ReadDir(sandboxesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	cleanedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		workspacePath := filepath.Join(sandboxesDir, entry.Name(), "workspace")
		if _, err := os.Stat(workspacePath); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return cleanedCount, err
		}
		if err := removeWorktreeCredential(workspacePath); err != nil {
			return cleanedCount, err
		}
		cleanedCount++
	}
	return cleanedCount, nil
}

func (m *Manager) pruneCachedRepositoryWorktrees(ctx context.Context) error {
	reposDir := filepath.Join(m.root, "repos")
	entries, err := os.ReadDir(reposDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		repoPath := filepath.Join(reposDir, entry.Name())
		if _, err := os.Stat(filepath.Join(repoPath, "HEAD")); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return err
		}
		cmd := exec.CommandContext(ctx, "git", "worktree", "prune", "--expire", "now")
		cmd.Dir = repoPath
		m.setLocalGitEnv(cmd)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to prune cached repository %s: %w, output: %s",
				repoPath, err, strings.TrimSpace(string(output)))
		}
	}
	return nil
}
