package runner

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/client"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
	"github.com/anthropics/agentsmesh/runner/internal/workspace"
)

// setupGitWorktree creates a git worktree for the pod.
func (b *PodBuilder) setupGitWorktree(ctx context.Context, sandboxRoot string, cfg *runnerv1.SandboxConfig) (string, string, error) {
	// Determine repository URL from HttpCloneUrl or SshCloneUrl
	var repoURL string
	if cfg.HttpCloneUrl != "" {
		repoURL = cfg.HttpCloneUrl
	} else if cfg.SshCloneUrl != "" {
		repoURL = cfg.SshCloneUrl
	} else {
		return "", "", &client.PodError{
			Code:    client.ErrCodeGitClone,
			Message: "http_clone_url or ssh_clone_url is required for worktree creation",
		}
	}

	// Use workspace manager if available
	if b.deps.Workspace == nil {
		return "", "", &client.PodError{
			Code:    client.ErrCodeGitWorktree,
			Message: "workspace manager not available for git operations",
		}
	}
	sourceCommitSHA, err := workspace.RequireCommitSHA("source_commit_sha", cfg.GetSourceCommitSha())
	if err != nil {
		return "", "", &client.PodError{
			Code:    client.ErrCodeGitWorktree,
			Message: err.Error(),
		}
	}

	// Report cloning progress
	b.sendProgress("cloning", 30, "Cloning repository...")

	opts, err := b.gitCredentialOptions(ctx, sandboxRoot, cfg)
	if err != nil {
		return "", "", err
	}

	// Pass new clone URLs for smart probing
	if cfg.HttpCloneUrl != "" {
		opts = append(opts, workspace.WithHttpCloneURL(cfg.HttpCloneUrl))
	}
	if cfg.SshCloneUrl != "" {
		opts = append(opts, workspace.WithSshCloneURL(cfg.SshCloneUrl))
	}
	opts = append(opts, workspace.WithSourceCommitSHA(sourceCommitSHA))

	// Create git worktree inside sandbox directory: sandboxes/{podKey}/workspace
	workspaceTarget := filepath.Join(sandboxRoot, "workspace")
	result, err := b.deps.Workspace.CreateWorktreeWithOptions(
		ctx,
		repoURL,
		cfg.SourceBranch,
		workspaceTarget,
		opts...,
	)
	if err != nil {
		// Determine error type
		errMsg := err.Error()
		errCode := client.ErrCodeGitWorktree
		if strings.Contains(errMsg, "authentication") || strings.Contains(errMsg, "Permission denied") {
			errCode = client.ErrCodeGitAuth
		} else if strings.Contains(errMsg, "clone") {
			errCode = client.ErrCodeGitClone
		}
		return "", "", &client.PodError{
			Code:    errCode,
			Message: fmt.Sprintf("failed to create workspace: %v", err),
			Details: map[string]string{
				"repository":        workspace.RepositoryURLForDisplay(repoURL),
				"branch":            cfg.SourceBranch,
				"source_commit_sha": sourceCommitSHA,
			},
		}
	}

	// Report progress after successful clone
	b.sendProgress("cloning", 60, "Repository cloned successfully")

	// WorktreeResult.Branch already falls back to the requested branch
	// when detached HEAD is detected, so no additional fallback is needed.
	branchName := result.Branch

	logger.Pod().InfoContext(ctx, "Git worktree created",
		"pod_key", b.cmd.PodKey,
		"workspace", result.Path,
		"branch", branchName)

	return result.Path, branchName, nil
}
