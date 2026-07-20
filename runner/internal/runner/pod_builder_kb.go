package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/client"
	"github.com/anthropics/agentsmesh/runner/internal/fsutil"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
	"github.com/anthropics/agentsmesh/runner/internal/workspace"
)

const (
	knowledgeMountDeployKey       = "agentsmesh-kb-deploy-key"
	knowledgeMountKnownHosts      = "agentsmesh-kb-known-hosts"
	temporaryKnownHostsFileSuffix = ".known-hosts"
)

func (b *PodBuilder) setupKnowledgeMounts(ctx context.Context, sandboxRoot string) error {
	if err := clearManagedKnowledgeCredentials(ctx, sandboxRoot); err != nil {
		return &client.PodError{Code: client.ErrCodeGitClone, Message: err.Error()}
	}
	mounts := b.cmd.GetSandboxConfig().GetKnowledgeMounts()
	if len(mounts) == 0 {
		return nil
	}
	b.sendProgress("preparing", 68, "Mounting knowledge bases...")
	for _, m := range mounts {
		if err := b.cloneKnowledgeMount(ctx, sandboxRoot, m); err != nil {
			return err
		}
	}
	return nil
}

func (b *PodBuilder) cloneKnowledgeMount(ctx context.Context, sandboxRoot string, m *runnerv1.KnowledgeMount) error {
	commitSHA, err := knowledgeMountCommitSHA(m)
	if err != nil {
		return &client.PodError{Code: client.ErrCodeGitClone, Message: err.Error(),
			Details: map[string]string{"kb_slug": m.GetSlug(), "mode": m.GetMode()}}
	}
	for _, cloneURL := range []string{m.GetHttpCloneUrl(), m.GetSshCloneUrl()} {
		if err := workspace.ValidateRepositoryURL(cloneURL); err != nil {
			return &client.PodError{Code: client.ErrCodeGitClone, Message: err.Error(),
				Details: map[string]string{"kb_slug": m.GetSlug(), "mode": m.GetMode()}}
		}
	}
	mountPath := m.GetMountPath()
	if mountPath == "" {
		mountPath = filepath.Join("kb", m.GetSlug())
	}
	dest := filepath.Join(sandboxRoot, filepath.FromSlash(mountPath))
	if err := validateKnowledgeMountDestinationPath(sandboxRoot, dest); err != nil {
		return &client.PodError{Code: client.ErrCodeGitClone, Message: err.Error(),
			Details: map[string]string{"kb_slug": m.GetSlug(), "mode": m.GetMode()}}
	}
	sshURL, privateKey := m.GetSshCloneUrl(), m.GetGitPrivateKey()
	knownHosts := m.GetGitKnownHosts()
	cloneURL := m.GetHttpCloneUrl()
	var temporaryKey string
	if privateKey != "" {
		if sshURL == "" || knownHosts == "" {
			return &client.PodError{Code: client.ErrCodeGitClone, Message: fmt.Sprintf(
				"ssh_clone_url and git_known_hosts are required for knowledge base %q when git_private_key is set", m.GetSlug())}
		}
		cloneURL = sshURL
	}
	if m.GetMode() == "rw" && privateKey == "" {
		return &client.PodError{Code: client.ErrCodeGitClone, Message: fmt.Sprintf(
			"git_private_key is required for read-write knowledge base %q", m.GetSlug())}
	}
	if _, err := os.Stat(filepath.Join(dest, ".git")); err == nil {
		if err := reconcileExistingKnowledgeMount(ctx, sandboxRoot, dest, m); err != nil {
			return &client.PodError{Code: client.ErrCodeGitClone, Message: err.Error(),
				Details: map[string]string{"kb_slug": m.GetSlug(), "mode": m.GetMode()}}
		}
		if m.GetMode() == "rw" {
			if knowledgeMountIsDetached(ctx, dest) {
				if err := checkoutExistingKnowledgeMountCommit(ctx, sandboxRoot, dest, m, commitSHA); err != nil {
					return &client.PodError{Code: client.ErrCodeGitClone, Message: err.Error(),
						Details: map[string]string{"kb_slug": m.GetSlug(), "mode": m.GetMode()}}
				}
			} else if err := verifyReadWriteKnowledgeMountPin(ctx, dest, m, commitSHA); err != nil {
				return &client.PodError{Code: client.ErrCodeGitClone, Message: err.Error(),
					Details: map[string]string{"kb_slug": m.GetSlug(), "mode": m.GetMode()}}
			}
			return nil
		}
		if err := checkoutExistingKnowledgeMountCommit(ctx, sandboxRoot, dest, m, commitSHA); err != nil {
			return &client.PodError{Code: client.ErrCodeGitClone, Message: err.Error(),
				Details: map[string]string{"kb_slug": m.GetSlug(), "mode": m.GetMode()}}
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return &client.PodError{Code: client.ErrCodeSandboxCreate, Message: fmt.Sprintf("failed to create knowledge mount dir: %v", err)}
	}
	if privateKey != "" {
		var err error
		temporaryKey, err = writeKnowledgeMountKey(sandboxRoot, privateKey, knownHosts)
		if err != nil {
			return &client.PodError{Code: client.ErrCodeFileCreate, Message: err.Error()}
		}
	}

	args := []string{"clone", "--no-checkout", cloneURL, dest}
	cloneCmd := exec.CommandContext(ctx, "git", args...)
	if temporaryKey != "" {
		cloneCmd.Env = knowledgeMountSSHEnv(temporaryKey, temporaryKey+temporaryKnownHostsFileSuffix)
	} else {
		cloneCmd.Env = knowledgeMountAnonymousGitEnv()
	}
	if output, err := cloneCmd.CombinedOutput(); err != nil {
		cleanupKnowledgeMount(ctx, dest)
		cloneErr := &client.PodError{Code: client.ErrCodeGitClone,
			Message: knowledgeMountCommandError("clone", m.GetSlug(), err, output, privateKey),
			Details: map[string]string{"kb_slug": m.GetSlug(), "mode": m.GetMode()}}
		return joinKnowledgeMountCredentialCleanup(temporaryKey, cloneErr)
	}
	if err := checkoutKnowledgeMountCommit(ctx, dest, m, commitSHA, temporaryKey); err != nil {
		cleanupKnowledgeMount(ctx, dest)
		checkoutErr := &client.PodError{Code: client.ErrCodeGitClone, Message: err.Error(),
			Details: map[string]string{"kb_slug": m.GetSlug(), "mode": m.GetMode()}}
		return joinKnowledgeMountCredentialCleanup(temporaryKey, checkoutErr)
	}
	if temporaryKey != "" {
		if m.GetMode() == "rw" {
			if err := persistKnowledgeMountKey(ctx, sandboxRoot, dest, temporaryKey, privateKey); err != nil {
				cleanupKnowledgeMount(ctx, dest)
				persistErr := &client.PodError{Code: client.ErrCodeGitClone, Message: err.Error(),
					Details: map[string]string{"kb_slug": m.GetSlug(), "mode": m.GetMode()}}
				return joinKnowledgeMountCredentialCleanup(temporaryKey, persistErr)
			}
		} else {
			if err := removeKnowledgeMountCredential(temporaryKey); err != nil {
				cleanupKnowledgeMount(ctx, dest)
				return &client.PodError{Code: client.ErrCodeGitClone, Message: fmt.Sprintf(
					"failed to remove temporary SSH credentials for knowledge base %q: %v", m.GetSlug(), err)}
			}
		}
	}
	remoteURL := knowledgeMountRemoteURL(m)
	if err := setKnowledgeMountRemote(ctx, dest, remoteURL, privateKey); err != nil {
		cleanupKnowledgeMount(ctx, dest)
		return &client.PodError{Code: client.ErrCodeGitClone, Message: err.Error(),
			Details: map[string]string{"kb_slug": m.GetSlug(), "mode": m.GetMode()}}
	}

	logger.Pod().InfoContext(ctx, "Knowledge base mounted",
		"pod_key", b.cmd.PodKey, "kb_slug", m.GetSlug(), "mode", m.GetMode(), "path", dest)
	return nil
}

func knowledgeMountCommandError(action, slug string, err error, output []byte, privateKey string) string {
	safeOutput := trimKnowledgeMountOutput(redactKnowledgeMountSecret(string(output), privateKey))
	message := fmt.Sprintf("failed to %s knowledge base %q: %v", action, slug, err)
	if safeOutput != "" {
		message += ", output: " + safeOutput
	}
	return message
}

func redactKnowledgeMountSecret(value, privateKey string) string {
	return redactKnowledgeMountPrivateKey(value, privateKey)
}

func cleanupKnowledgeMount(ctx context.Context, dest string) {
	if err := fsutil.RemoveAll(dest); err != nil {
		logger.Pod().WarnContext(ctx, "Failed to clean up partial knowledge clone", "path", dest, "error", err)
	}
}
