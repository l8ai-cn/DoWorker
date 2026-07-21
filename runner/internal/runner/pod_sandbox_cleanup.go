package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/l8ai-cn/agentcloud/runner/internal/fsutil"
	"github.com/l8ai-cn/agentcloud/runner/internal/logger"
	"github.com/l8ai-cn/agentcloud/runner/internal/workspace"
)

func cleanupPodSandbox(
	ctx context.Context,
	manager workspace.WorkspaceManagerInterface,
	workspaceRoot, podKey, sandboxPath string,
) error {
	ownedSandboxPath := filepath.Join(workspaceRoot, "sandboxes", podKey)
	if filepath.Clean(sandboxPath) != filepath.Clean(ownedSandboxPath) {
		return fsutil.RemoveAll(ownedSandboxPath)
	}
	return cleanupSandboxRoot(ctx, manager, ownedSandboxPath)
}

func cleanupSandboxRoot(
	ctx context.Context,
	manager workspace.WorkspaceManagerInterface,
	sandboxPath string,
) error {
	var worktreeErr error
	worktreePath := filepath.Join(sandboxPath, "workspace")
	if manager != nil {
		if _, err := os.Stat(filepath.Join(worktreePath, ".git")); err == nil {
			if err := manager.RemoveWorktree(ctx, worktreePath); err != nil {
				worktreeErr = fmt.Errorf("remove Git worktree: %w", err)
			}
		} else if !os.IsNotExist(err) {
			worktreeErr = fmt.Errorf("inspect Git worktree: %w", err)
		}
	}
	return errors.Join(worktreeErr, fsutil.RemoveAll(sandboxPath))
}

func (b *PodBuilder) cleanupSandbox(ctx context.Context, sandboxPath, reason string) error {
	err := cleanupPodSandbox(
		ctx,
		b.deps.Workspace,
		b.deps.Config.WorkspaceRoot,
		b.cmd.PodKey,
		sandboxPath,
	)
	if err != nil {
		logger.Pod().WarnContext(ctx, "Failed to clean up pod sandbox",
			"pod_key", b.cmd.PodKey, "path", sandboxPath, "reason", reason, "error", err)
	}
	return err
}
