package workspace

import (
	"context"
	"os/exec"
	"strings"

	"github.com/l8ai-cn/agentcloud/runner/internal/logger"
)

func (m *Manager) finalizeWorktree(
	ctx context.Context,
	worktreePath, fallbackBranch, repoURL string,
	options *WorktreeOptions,
) (*WorktreeResult, error) {
	log := logger.Workspace()
	if options == nil {
		options = &WorktreeOptions{}
	}
	if usesRunnerLocalGitConfig(options) && m.gitConfigPath != "" {
		if err := m.applyGitConfig(ctx, worktreePath); err != nil {
			return nil, err
		}
	}
	switch {
	case options.AnonymousAuth:
		if err := m.applyAnonymousGitConfig(ctx, worktreePath); err != nil {
			return nil, err
		}
	case options.GitToken != "":
		if err := m.applyTokenGitConfig(ctx, worktreePath, repoURL, options); err != nil {
			return nil, err
		}
	case options.SSHKeyPath != "":
		if err := m.applySSHGitConfig(ctx, worktreePath, options); err != nil {
			return nil, err
		}
	}
	if err := m.applyWorktreeRemote(ctx, worktreePath, repoURL); err != nil {
		return nil, err
	}
	actualBranch := fallbackBranch
	branchCmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	branchCmd.Dir = worktreePath
	m.setLocalGitEnv(branchCmd)
	if branchOutput, err := branchCmd.Output(); err == nil {
		detected := strings.TrimSpace(string(branchOutput))
		if detected != "" && detected != "HEAD" {
			actualBranch = detected
		}
	} else {
		log.Warn("Failed to detect actual branch name", "error", err, "fallback", fallbackBranch)
	}
	log.Info("Worktree created successfully", "path", worktreePath, "branch", actualBranch)
	return &WorktreeResult{Path: worktreePath, Branch: actualBranch}, nil
}

func usesRunnerLocalGitConfig(options *WorktreeOptions) bool {
	return !options.AnonymousAuth && options.GitToken == "" && options.SSHKeyPath == ""
}
