package workspace

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/anthropics/agentsmesh/runner/internal/fsutil"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
)

// ensureRepository clones or fetches a repository
func (m *Manager) ensureRepository(ctx context.Context, repoURL, path string) error {
	return m.ensureRepositoryWithAuth(ctx, repoURL, path, nil)
}

// ensureRepositoryWithAuth clones or fetches a repository with authentication options
func (m *Manager) ensureRepositoryWithAuth(ctx context.Context, repoURL, path string, opts *WorktreeOptions) error {
	if err := validateRepositoryAuthURL(repoURL, opts); err != nil {
		return err
	}
	log := logger.Workspace()

	// Check if repository exists (bare repo has HEAD file directly in path, not in .git subdirectory)
	if _, err := os.Stat(filepath.Join(path, "HEAD")); err == nil {
		log.Debug("Repository exists, fetching updates", "path", path)
		authURL := m.prepareAuthURL(repoURL, opts)
		fetchCmd := exec.CommandContext(ctx, "git", "fetch", authURL, "+refs/heads/*:refs/remotes/origin/*")
		fetchCmd.Dir = path
		m.setGitAuthEnv(fetchCmd, opts)
		if output, err := fetchCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to fetch existing repository: %w, output: %s", err, m.redactGitOutput(opts, output))
		}
		log.Debug("Repository fetched successfully", "path", path)
		return nil
	}

	// Directory may exist but is not a valid bare repo (e.g., previous clone was interrupted).
	// Clean it up before cloning.
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		log.Warn("Directory exists but is not a valid bare repo, removing before clone", "path", path)
		if err := fsutil.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove invalid repo directory: %w", err)
		}
	}

	return m.cloneBareRepository(ctx, repoURL, path, opts)
}

// cloneBareRepository performs a bare clone and configures the repository for worktree usage.
func (m *Manager) cloneBareRepository(ctx context.Context, repoURL, path string, opts *WorktreeOptions) error {
	if err := validateRepositoryAuthURL(repoURL, opts); err != nil {
		return err
	}
	log := logger.Workspace()

	// Clone the repository (bare clone for worktree support)
	log.Debug("Cloning repository", "url", RepositoryURLForDisplay(repoURL), "path", path)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create repo parent dir: %w", err)
	}

	cloneURL := m.prepareAuthURL(repoURL, opts)

	cloneCmd := exec.CommandContext(ctx, "git", "clone", "--bare", cloneURL, path)
	m.setGitAuthEnv(cloneCmd, opts)
	if output, err := cloneCmd.CombinedOutput(); err != nil {
		// Clean up any partial clone artifacts to avoid blocking future retries
		if removeErr := fsutil.RemoveAll(path); removeErr != nil {
			log.Warn("Failed to clean up partial clone", "path", path, "error", removeErr)
		}
		return fmt.Errorf("failed to clone: %w, output: %s", err, m.redactGitOutput(opts, output))
	}
	log.Debug("Repository cloned successfully", "path", path)

	if err := m.runBareRepoConfig(ctx, path, opts, "remote", "set-url", "origin", repoURL); err != nil {
		_ = fsutil.RemoveAll(path)
		return err
	}

	if err := m.runBareRepoConfig(ctx, path, opts, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*"); err != nil {
		_ = fsutil.RemoveAll(path)
		return err
	}

	authURL := m.prepareAuthURL(repoURL, opts)
	fetchCmd := exec.CommandContext(ctx, "git", "fetch", authURL, "+refs/heads/*:refs/remotes/origin/*")
	fetchCmd.Dir = path
	m.setGitAuthEnv(fetchCmd, opts)
	if output, err := fetchCmd.CombinedOutput(); err != nil {
		_ = fsutil.RemoveAll(path)
		return fmt.Errorf("failed to fetch cloned repository refs: %w, output: %s", err, m.redactGitOutput(opts, output))
	}

	return nil
}

func (m *Manager) runBareRepoConfig(ctx context.Context, repoPath string, opts *WorktreeOptions, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = repoPath
	m.setLocalGitEnv(cmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run git %s: %w, output: %s", strings.Join(args, " "), err, m.redactGitOutput(opts, output))
	}
	return nil
}
