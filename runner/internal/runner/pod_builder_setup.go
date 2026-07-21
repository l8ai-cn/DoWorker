package runner

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/l8ai-cn/agentcloud/runner/internal/cache"
	"github.com/l8ai-cn/agentcloud/runner/internal/client"
	"github.com/l8ai-cn/agentcloud/runner/internal/fsutil"
	"github.com/l8ai-cn/agentcloud/runner/internal/logger"
	"github.com/l8ai-cn/agentcloud/runner/internal/workspace"
)

// setup sets up the sandbox and working directory.
// Returns (sandboxRoot, workingDir, branchName, error).
// Uses Strategy Pattern to select the appropriate setup strategy based on SandboxConfig.
func (b *PodBuilder) setup(ctx context.Context) (string, string, string, error) {
	b.sendProgress("preparing", 10, "Creating sandbox directory...")
	podSandboxRoot := filepath.Join(b.deps.Config.WorkspaceRoot, "sandboxes", b.cmd.PodKey)
	sandboxRoot := podSandboxRoot
	if err := os.MkdirAll(sandboxRoot, 0755); err != nil {
		return "", "", "", &client.PodError{
			Code:    client.ErrCodeSandboxCreate,
			Message: fmt.Sprintf("failed to create sandbox directory: %v", err),
		}
	}
	logger.Pod().DebugContext(ctx, "Sandbox root created", "pod_key", b.cmd.PodKey, "path", sandboxRoot)

	cfg := b.cmd.SandboxConfig

	b.sendProgress("preparing", 20, "Setting up working directory...")

	strategy := b.selectSetupStrategy(cfg)
	logger.Pod().InfoContext(ctx, "Setup strategy selected", "pod_key", b.cmd.PodKey, "strategy", strategy.Name())

	result, err := strategy.Setup(ctx, sandboxRoot, cfg)
	if err != nil {
		logger.Pod().ErrorContext(ctx, "Setup strategy failed", "pod_key", b.cmd.PodKey, "strategy", strategy.Name(), "error", err)
		return "", "", "", errors.Join(err, b.cleanupSandbox(ctx, sandboxRoot, "setup strategy error"))
	}

	sandboxOwned := true
	reusesWorkspace := false
	if result.SandboxRoot != "" && result.SandboxRoot != sandboxRoot {
		_ = fsutil.RemoveAll(sandboxRoot)
		sandboxRoot = result.SandboxRoot
		sandboxOwned = false
		reusesWorkspace = true
	}

	if err := b.setupKnowledgeMounts(ctx, sandboxRoot); err != nil {
		if sandboxOwned {
			err = errors.Join(err, b.cleanupSandbox(ctx, sandboxRoot, "knowledge mount error"))
		}
		return "", "", "", err
	}

	if err := b.prepareAgentHome(sandboxRoot, result.WorkingDir); err != nil {
		if sandboxOwned {
			err = errors.Join(err, b.cleanupSandbox(ctx, sandboxRoot, "agent home error"))
		}
		return "", "", "", err
	}

	if len(b.cmd.FilesToCreate) > 0 {
		b.sendProgress("preparing", 70, "Creating files...")
	}
	if err := b.createFiles(sandboxRoot, result.WorkingDir); err != nil {
		if sandboxOwned {
			err = errors.Join(err, b.cleanupSandbox(ctx, sandboxRoot, "file creation error"))
		}
		return "", "", "", err
	}

	// Download skill packages
	if err := b.downloadResources(ctx, sandboxRoot, result.WorkingDir); err != nil {
		err = fmt.Errorf("failed to download resources: %w", err)
		if sandboxOwned {
			err = errors.Join(err, b.cleanupSandbox(ctx, sandboxRoot, "resource download error"))
		}
		return "", "", "", err
	}
	if reusesWorkspace {
		if err := persistWorkspaceAlias(b.deps.Config.WorkspaceRoot, podSandboxRoot, result.WorkingDir); err != nil {
			if errors.Is(err, errWorkspaceAliasOutsideRunnerRoot) {
				logger.Pod().InfoContext(ctx, "Detached workspace alias unavailable for unmanaged local path",
					"pod_key", b.cmd.PodKey, "working_dir", result.WorkingDir)
			} else {
				_ = fsutil.RemoveAll(podSandboxRoot)
				return "", "", "", fmt.Errorf("persist workspace alias: %w", err)
			}
		}
	}

	logger.Pod().InfoContext(ctx, "Sandbox setup completed",
		"pod_key", b.cmd.PodKey,
		"sandbox_root", sandboxRoot,
		"working_dir", result.WorkingDir,
		"branch", result.BranchName)

	return sandboxRoot, result.WorkingDir, result.BranchName, nil
}

// selectSetupStrategy selects the appropriate setup strategy based on configuration.
// Strategies are tried in order; first matching strategy is used.
func (b *PodBuilder) selectSetupStrategy(cfg *runnerv1.SandboxConfig) SetupStrategy {
	for _, strategy := range b.setupStrategies {
		if strategy.CanHandle(cfg) {
			return strategy
		}
	}
	// Fallback to empty sandbox (should not reach here if strategies are properly configured)
	return NewEmptySandboxStrategy(b)
}

// runPreparationScript executes the preparation script in the workspace.
func (b *PodBuilder) runPreparationScript(ctx context.Context, cfg *runnerv1.SandboxConfig, workspacePath, branchName string) error {
	timeout := int(cfg.PreparationTimeout)
	if timeout <= 0 {
		timeout = 300 // Default 5 minutes
	}

	b.sendProgress("preparing", 65, "Running preparation script...")

	preparer := workspace.NewPreparerFromScript(cfg.PreparationScript, timeout)
	if preparer == nil {
		return nil
	}

	prepCtx := &workspace.PreparationContext{
		PodID:        b.cmd.PodKey,
		TicketSlug:   cfg.GetTicketSlug(),
		BranchName:   branchName,
		WorkspaceDir: workspacePath,
		BaseEnvVars:  gitProcessIsolationEnv(cfg.GetCredentialType()),
		UnsetEnvVars: gitProcessIsolationUnsetEnv(cfg.GetCredentialType()),
	}

	if err := preparer.Prepare(ctx, prepCtx); err != nil {
		return &client.PodError{
			Code:    client.ErrCodePrepareScript,
			Message: fmt.Sprintf("preparation script failed: %v", err),
		}
	}

	b.sendProgress("preparing", 75, "Preparation script completed")
	return nil
}

// downloadResources downloads skill packages and other resources into the sandbox.
func (b *PodBuilder) downloadResources(ctx context.Context, sandboxRoot, workDir string) error {
	if len(b.cmd.ResourcesToDownload) == 0 {
		return nil
	}

	cacheDir := filepath.Join(b.deps.Config.WorkspaceRoot, "cache", "skills")
	cacheManager, err := cache.NewSkillCacheManager(cacheDir)
	if err != nil {
		return fmt.Errorf("failed to create skill cache manager: %w", err)
	}

	hostAliases := make(map[string]string, len(b.deps.Config.ResourceHostAliases))
	for _, alias := range b.deps.Config.ResourceHostAliases {
		hostAliases[alias.Host] = alias.DialHost
	}
	downloader := cache.NewDownloaderWithHostAliases(cacheManager, hostAliases)
	for _, res := range b.cmd.ResourcesToDownload {
		result, err := downloader.DownloadAndExtract(ctx, res, sandboxRoot, workDir)
		if err != nil {
			return fmt.Errorf("failed to download resource %s: %w", res.Sha, err)
		}
		if result.CacheHit {
			slog.InfoContext(ctx, "Resource cache hit", "sha", res.Sha)
		} else {
			slog.InfoContext(ctx, "Resource downloaded", "sha", res.Sha, "bytes", result.BytesRead)
		}
	}
	return nil
}
