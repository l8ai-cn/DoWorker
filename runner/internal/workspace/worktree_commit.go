package workspace

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func (m *Manager) createPinnedWorktree(
	ctx context.Context,
	repoURL, worktreePath, mainRepoPath, commitSHA, branchName string,
	options *WorktreeOptions,
) error {
	if err := m.fetchPinnedCommit(ctx, repoURL, mainRepoPath, commitSHA, options); err != nil {
		return err
	}
	if err := m.verifyPinnedCommit(ctx, mainRepoPath, commitSHA, options); err != nil {
		return err
	}

	branchCmd := exec.CommandContext(ctx, "git", "branch", "-f", branchName, commitSHA)
	branchCmd.Dir = mainRepoPath
	m.setLocalGitEnv(branchCmd)
	if output, err := branchCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to prepare pinned branch %s at commit %s: %w, output: %s",
			branchName, commitSHA, err, strings.TrimSpace(string(output)))
	}
	worktreeCmd := exec.CommandContext(ctx, "git", "worktree", "add", worktreePath, branchName)
	worktreeCmd.Dir = mainRepoPath
	m.setLocalGitEnv(worktreeCmd)
	if output, err := worktreeCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create pinned worktree at commit %s: %w, output: %s",
			commitSHA, err, strings.TrimSpace(string(output)))
	}
	return nil
}

func (m *Manager) fetchPinnedCommit(
	ctx context.Context,
	repoURL, mainRepoPath, commitSHA string,
	options *WorktreeOptions,
) error {
	remote := m.prepareAuthURL(repoURL, options)
	fetchCmd := exec.CommandContext(ctx, "git", "fetch", "--no-tags", "--depth=1", remote, commitSHA)
	fetchCmd.Dir = mainRepoPath
	m.setGitAuthEnv(fetchCmd, options)
	if output, err := fetchCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to fetch pinned commit %s: %w, output: %s",
			commitSHA, err, m.redactGitOutput(options, output))
	}
	return nil
}

func (m *Manager) verifyPinnedCommit(
	ctx context.Context,
	mainRepoPath, commitSHA string,
	options *WorktreeOptions,
) error {
	verifyCmd := exec.CommandContext(ctx, "git", "cat-file", "-e", commitSHA+"^{commit}")
	verifyCmd.Dir = mainRepoPath
	m.setLocalGitEnv(verifyCmd)
	if output, err := verifyCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pinned commit %s is not available as a commit: %w, output: %s",
			commitSHA, err, strings.TrimSpace(string(output)))
	}
	return nil
}
