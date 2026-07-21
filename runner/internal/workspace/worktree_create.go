package workspace

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/l8ai-cn/agentcloud/runner/internal/logger"
)

// WorktreeResult encapsulates the result of a worktree creation operation.
type WorktreeResult struct {
	// Path is the filesystem path where the worktree was created.
	Path string
	// Branch is the actual git branch name checked out in the worktree.
	Branch string
}

// CreateWorktree creates a git worktree for a repository.
// The worktree is created inside the sandbox directory: sandboxes/{podKey}/workspace
func (m *Manager) CreateWorktree(ctx context.Context, repoURL, branch, podKey string) (*WorktreeResult, error) {
	workspacePath := filepath.Join(m.root, "sandboxes", podKey, "workspace")
	return m.CreateWorktreeWithOptions(ctx, repoURL, branch, workspacePath)
}

// CreateWorktreeWithOptions creates a git worktree with additional options.
// worktreePath is the full path where the worktree should be created.
func (m *Manager) CreateWorktreeWithOptions(ctx context.Context, repoURL, branch, worktreePath string, opts ...WorktreeOption) (*WorktreeResult, error) {
	options := &WorktreeOptions{}
	for _, opt := range opts {
		opt(options)
	}
	for _, candidate := range []string{repoURL, options.HttpCloneURL, options.SshCloneURL} {
		if err := validateRepositoryURL(candidate); err != nil {
			return nil, err
		}
	}
	log := logger.Workspace()
	log.Info("Creating worktree", "repo", RepositoryURLForDisplay(repoURL), "branch", branch, "path", worktreePath)

	// If multiple clone URLs are available, probe to find the accessible one
	if options.HttpCloneURL != "" || options.SshCloneURL != "" {
		httpURL := options.HttpCloneURL
		sshURL := options.SshCloneURL

		probeURL, err := m.probeRepositoryAccess(ctx, httpURL, sshURL, options)
		if err != nil {
			return nil, fmt.Errorf("repository access probe failed: %w", err)
		}
		log.Info("Repository access probe selected URL", "url", RepositoryURLForDisplay(probeURL))
		repoURL = probeURL
	}

	// Parse repo name from URL (needed before locking)
	repoName := extractRepoName(repoURL)
	if repoName == "" {
		return nil, fmt.Errorf("invalid repository URL: %s", RepositoryURLForDisplay(repoURL))
	}
	log.Debug("Parsed repo name", "name", repoName)

	// Per-repo lock: allows concurrent worktree creation for different repositories
	// while serializing git operations (clone/fetch) on the same repository.
	repoLock := m.getRepoLock(repoName)
	repoLock.Lock()
	defer repoLock.Unlock()

	// Main repository path (bare repo cache, shared across pods)
	mainRepoPath := filepath.Join(m.root, "repos", repoName)

	// Clone or fetch the repository with authentication
	if err := m.ensureRepositoryWithAuth(ctx, repoURL, mainRepoPath, options); err != nil {
		return nil, fmt.Errorf("failed to ensure repository: %w", err)
	}

	// Remove existing worktree if it exists
	if _, err := os.Stat(worktreePath); err == nil {
		if err := m.removeWorktreeInternal(ctx, mainRepoPath, worktreePath); err != nil {
			return nil, fmt.Errorf("failed to remove existing worktree: %w", err)
		}
	}

	// Create worktree parent directory
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create worktree parent dir: %w", err)
	}

	if options.SourceCommitSHA != "" {
		commitSHA, err := NormalizeCommitSHA(options.SourceCommitSHA)
		if err != nil {
			return nil, err
		}
		pinnedBranch := worktreeBranchName(worktreePath)
		if err := m.createPinnedWorktree(ctx, repoURL, worktreePath, mainRepoPath, commitSHA, pinnedBranch, options); err != nil {
			return nil, err
		}
		result, err := m.finalizeWorktree(ctx, worktreePath, "", repoURL, options)
		if err != nil {
			m.cleanupCreatedWorktree(ctx, mainRepoPath, worktreePath, pinnedBranch)
			return nil, err
		}
		return result, nil
	}

	// Fetch the branch
	if branch == "" {
		branch = "main"
	}

	fetchBranch := func(fetchRef string) ([]byte, error) {
		remote := "origin"
		refspec := fetchRef
		if options != nil && options.GitToken != "" {
			authURL := m.prepareAuthURL(repoURL, options)
			remote = authURL
			refspec = fmt.Sprintf("refs/heads/%s:refs/remotes/origin/%s", fetchRef, fetchRef)
		}
		fetchCmd := exec.CommandContext(ctx, "git", "fetch", remote, refspec)
		fetchCmd.Dir = mainRepoPath
		m.setGitAuthEnv(fetchCmd, options)
		return fetchCmd.CombinedOutput()
	}
	if output, err := fetchBranch(branch); err != nil {
		// Try 'master' if 'main' fails
		if branch == "main" {
			branch = "master"
			if output, err = fetchBranch(branch); err != nil {
				return nil, fmt.Errorf("failed to fetch branch: %s, output: %s", err, m.redactGitOutput(options, output))
			}
		} else {
			return nil, fmt.Errorf("failed to fetch branch: %s, output: %s", err, m.redactGitOutput(options, output))
		}
	}

	// Create worktree
	// Use a unique branch name based on parent directory name (sandbox podKey)
	// e.g., /path/sandboxes/pod-123/worktree -> worktree-pod-123
	worktreeBranch := worktreeBranchName(worktreePath)

	worktreeCmd := exec.CommandContext(ctx, "git", "worktree", "add", "-b", worktreeBranch, worktreePath, fmt.Sprintf("origin/%s", branch))
	worktreeCmd.Dir = mainRepoPath
	m.setLocalGitEnv(worktreeCmd)
	createdBranch := true
	if _, err := worktreeCmd.CombinedOutput(); err != nil {
		// If branch already exists, try without -b
		worktreeCmd = exec.CommandContext(ctx, "git", "worktree", "add", worktreePath, fmt.Sprintf("origin/%s", branch))
		worktreeCmd.Dir = mainRepoPath
		m.setLocalGitEnv(worktreeCmd)
		if output, retryErr := worktreeCmd.CombinedOutput(); retryErr != nil {
			return nil, fmt.Errorf("failed to create worktree: %s, output: %s", retryErr, output)
		}
		createdBranch = false
	}

	result, err := m.finalizeWorktree(ctx, worktreePath, branch, repoURL, options)
	if err != nil {
		branchToDelete := ""
		if createdBranch {
			branchToDelete = worktreeBranch
		}
		m.cleanupCreatedWorktree(ctx, mainRepoPath, worktreePath, branchToDelete)
		return nil, err
	}
	return result, nil
}

func worktreeBranchName(worktreePath string) string {
	return fmt.Sprintf("worktree-%s", filepath.Base(filepath.Dir(worktreePath)))
}

func (m *Manager) cleanupCreatedWorktree(
	ctx context.Context,
	mainRepoPath, worktreePath, branchName string,
) {
	log := logger.Workspace()
	if err := m.removeWorktreeInternal(ctx, mainRepoPath, worktreePath); err != nil {
		log.Warn("Failed to clean up worktree after setup failure", "path", worktreePath, "error", err)
	}
	if branchName == "" {
		return
	}
	cmd := exec.CommandContext(ctx, "git", "branch", "-D", branchName)
	cmd.Dir = mainRepoPath
	m.setLocalGitEnv(cmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Warn("Failed to clean up worktree branch after setup failure", "branch", branchName, "error", err, "output", string(output))
	}
}
