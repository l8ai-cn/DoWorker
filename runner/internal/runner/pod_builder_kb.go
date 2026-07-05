package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/client"
	"github.com/anthropics/agentsmesh/runner/internal/fsutil"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
)

// setupKnowledgeMounts clones each declared knowledge base into the sandbox.
// Resume pods reuse the source sandbox where clones already exist, so
// existing mount directories are left untouched.
func (b *PodBuilder) setupKnowledgeMounts(ctx context.Context, sandboxRoot string) error {
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
	mountPath := m.GetMountPath()
	if mountPath == "" {
		mountPath = filepath.Join("kb", m.GetSlug())
	}
	dest := filepath.Join(sandboxRoot, filepath.FromSlash(mountPath))

	if _, err := os.Stat(filepath.Join(dest, ".git")); err == nil {
		logger.Pod().DebugContext(ctx, "Knowledge mount already present, skipping clone",
			"pod_key", b.cmd.PodKey, "kb_slug", m.GetSlug(), "path", dest)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return &client.PodError{
			Code:    client.ErrCodeSandboxCreate,
			Message: fmt.Sprintf("failed to create knowledge mount dir: %v", err),
		}
	}

	args := []string{"clone", "--single-branch"}
	if m.GetBranch() != "" {
		args = append(args, "--branch", m.GetBranch())
	}
	args = append(args, kbAuthURL(m.GetHttpCloneUrl(), m.GetGitToken()), dest)

	cloneCmd := exec.CommandContext(ctx, "git", args...)
	if output, err := cloneCmd.CombinedOutput(); err != nil {
		if rmErr := fsutil.RemoveAll(dest); rmErr != nil {
			logger.Pod().WarnContext(ctx, "Failed to clean up partial knowledge clone", "path", dest, "error", rmErr)
		}
		return &client.PodError{
			Code:    client.ErrCodeGitClone,
			Message: fmt.Sprintf("failed to clone knowledge base %q: %v, output: %s", m.GetSlug(), err, output),
			Details: map[string]string{"kb_slug": m.GetSlug(), "mode": m.GetMode()},
		}
	}

	// ro mounts must not retain push credentials: strip the token from the
	// stored remote URL. rw mounts keep it so the agent can commit + push.
	if m.GetMode() != "rw" {
		cleanCmd := exec.CommandContext(ctx, "git", "remote", "set-url", "origin", m.GetHttpCloneUrl())
		cleanCmd.Dir = dest
		if output, err := cleanCmd.CombinedOutput(); err != nil {
			logger.Pod().WarnContext(ctx, "Failed to strip token from ro knowledge mount",
				"kb_slug", m.GetSlug(), "error", err, "output", string(output))
		}
	}

	logger.Pod().InfoContext(ctx, "Knowledge base mounted",
		"pod_key", b.cmd.PodKey, "kb_slug", m.GetSlug(), "mode", m.GetMode(), "path", dest)
	return nil
}

// kbAuthURL embeds the clone token as basic-auth password. Gitea accepts any
// username with an access token; x-access-token matches repo_auth.go usage.
func kbAuthURL(cloneURL, token string) string {
	if token == "" {
		return cloneURL
	}
	for _, scheme := range []string{"https://", "http://"} {
		if strings.HasPrefix(cloneURL, scheme) {
			return scheme + "x-access-token:" + token + "@" + strings.TrimPrefix(cloneURL, scheme)
		}
	}
	return cloneURL
}
