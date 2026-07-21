package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/l8ai-cn/agentcloud/runner/internal/client"
	"github.com/l8ai-cn/agentcloud/runner/internal/logger"
	"github.com/l8ai-cn/agentcloud/runner/internal/workspace"
)

func (b *PodBuilder) gitCredentialOptions(
	ctx context.Context,
	sandboxRoot string,
	cfg *runnerv1.SandboxConfig,
) ([]workspace.WorktreeOption, error) {
	logger.Pod().DebugContext(ctx, "Setting up git credentials", "pod_key", b.cmd.PodKey, "credential_type", cfg.CredentialType)
	switch cfg.GetCredentialType() {
	case "none":
		return []workspace.WorktreeOption{workspace.WithAnonymousAuth()}, nil
	case "runner_local":
		return nil, nil
	case "oauth", "pat":
		if cfg.GetGitToken() == "" {
			return nil, gitAuthError(fmt.Sprintf("git_token is required for credential_type %q", cfg.GetCredentialType()))
		}
		username := "x-access-token"
		if cfg.GetCredentialType() == "oauth" {
			username = "oauth2"
		}
		return []workspace.WorktreeOption{
			workspace.WithGitTokenCredentials(username, cfg.GetGitToken()),
		}, nil
	case "ssh_key":
		if cfg.GetSshPrivateKey() == "" {
			return nil, gitAuthError("ssh_private_key is required for credential_type \"ssh_key\"")
		}
		keyPath, err := b.writeSandboxSSHKey(ctx, sandboxRoot, cfg.GetSshPrivateKey())
		if err != nil {
			return nil, err
		}
		return []workspace.WorktreeOption{workspace.WithSSHKeyPath(keyPath)}, nil
	case "":
		return nil, gitAuthError("credential_type is required")
	default:
		return nil, gitAuthError(fmt.Sprintf("unsupported git credential_type %q", cfg.GetCredentialType()))
	}
}

func (b *PodBuilder) writeSandboxSSHKey(ctx context.Context, sandboxRoot, privateKey string) (string, error) {
	keyFile := filepath.Join(sandboxRoot, ".ssh_key")
	if err := os.WriteFile(keyFile, []byte(privateKey), 0600); err != nil {
		return "", &client.PodError{Code: client.ErrCodeFileCreate, Message: fmt.Sprintf("failed to write SSH key: %v", err)}
	}
	if err := secureWindowsPrivateKey(keyFile); err != nil {
		_ = os.Remove(keyFile)
		return "", &client.PodError{Code: client.ErrCodeFileCreate, Message: err.Error()}
	}
	logger.Pod().DebugContext(ctx, "SSH key written to sandbox", "pod_key", b.cmd.PodKey, "key_file", keyFile)
	return keyFile, nil
}

func gitAuthError(message string) error {
	return &client.PodError{Code: client.ErrCodeGitAuth, Message: message}
}
