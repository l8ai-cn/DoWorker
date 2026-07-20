package workspace

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/anthropics/agentsmesh/runner/internal/fsutil"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
)

// RemoveWorktree removes a worktree
func (m *Manager) RemoveWorktree(ctx context.Context, worktreePath string) error {
	log := logger.Workspace()
	log.Info("Removing worktree", "path", worktreePath)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Find the main repository
	repoPath, err := m.findMainRepo(worktreePath)
	if err != nil {
		// If we can't find the main repo, just remove the directory
		log.Debug("Main repo not found, removing directory directly", "path", worktreePath)
		if err := fsutil.RemoveAll(worktreePath); err != nil {
			return err
		}
		return removeWorktreeCredential(worktreePath)
	}

	return m.removeWorktreeInternal(ctx, repoPath, worktreePath)
}

// removeWorktreeInternal removes a worktree (internal, no lock)
func (m *Manager) removeWorktreeInternal(ctx context.Context, repoPath, worktreePath string) error {
	// Remove worktree using git
	removeCtx, cancelRemove := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	removeCmd := exec.CommandContext(removeCtx, "git", "worktree", "remove", "--force", worktreePath)
	removeCmd.Dir = repoPath
	m.setLocalGitEnv(removeCmd)
	output, err := removeCmd.CombinedOutput()
	cancelRemove()
	if err != nil {
		// If git worktree remove fails, try manual removal
		logger.Workspace().Warn("Git worktree remove failed, trying manual removal",
			"error", err, "output", string(output))
		if removeErr := fsutil.RemoveAll(worktreePath); removeErr != nil {
			return removeErr
		}
	}

	// Prune worktrees
	pruneCtx, cancelPrune := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancelPrune()
	pruneCmd := exec.CommandContext(pruneCtx, "git", "worktree", "prune", "--expire", "now")
	pruneCmd.Dir = repoPath
	m.setLocalGitEnv(pruneCmd)
	var pruneErr error
	if output, err := pruneCmd.CombinedOutput(); err != nil {
		pruneErr = fmt.Errorf("failed to prune worktrees: %w, output: %s", err, strings.TrimSpace(string(output)))
	}

	return errors.Join(pruneErr, removeWorktreeCredential(worktreePath))
}

func removeWorktreeCredential(worktreePath string) error {
	sandboxPath := filepath.Dir(worktreePath)
	var cleanupErr error
	for _, name := range []string{gitCredentialFileName, ".ssh_key"} {
		if err := os.Remove(filepath.Join(sandboxPath, name)); err != nil && !os.IsNotExist(err) {
			cleanupErr = errors.Join(cleanupErr, err)
		}
	}
	return cleanupErr
}

// findMainRepo finds the main repository for a worktree
func (m *Manager) findMainRepo(worktreePath string) (string, error) {
	// The .git file in a worktree contains the path to the main repo
	gitPath := filepath.Join(worktreePath, ".git")

	data, err := os.ReadFile(gitPath)
	if err != nil {
		return "", fmt.Errorf("failed to read .git file: %w", err)
	}

	// Format: gitdir: /path/to/main/repo/.git/worktrees/name
	content := strings.TrimSpace(string(data))
	if !strings.HasPrefix(content, "gitdir: ") {
		return "", fmt.Errorf("invalid .git file format")
	}

	gitDir := strings.TrimPrefix(content, "gitdir: ")

	// Navigate up from .git/worktrees/name to .git
	mainGitDir := filepath.Dir(filepath.Dir(gitDir))
	mainRepoDir := filepath.Dir(mainGitDir)

	// For bare repos, the path is different
	if filepath.Base(mainGitDir) == ".git" {
		return mainRepoDir, nil
	}

	// For bare repos, mainGitDir is the repo itself
	return mainGitDir, nil
}
