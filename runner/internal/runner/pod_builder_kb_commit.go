package runner

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/l8ai-cn/agentcloud/runner/internal/workspace"
)

func knowledgeMountCommitSHA(m *runnerv1.KnowledgeMount) (string, error) {
	return workspace.RequireCommitSHA(fmt.Sprintf("knowledge_mounts[%s].commit_sha", m.GetSlug()), m.GetCommitSha())
}

func checkoutKnowledgeMountCommit(
	ctx context.Context,
	dest string,
	m *runnerv1.KnowledgeMount,
	commitSHA, temporaryKey string,
) error {
	privateKey := m.GetGitPrivateKey()
	if err := runKnowledgeMountGit(ctx, dest, temporaryKey, privateKey, "fetch", "--no-tags", "--depth=1", "origin", commitSHA); err != nil {
		return err
	}
	if err := runKnowledgeMountGit(ctx, dest, temporaryKey, privateKey, "cat-file", "-e", commitSHA+"^{commit}"); err != nil {
		return err
	}
	if m.GetMode() == "rw" {
		branch := m.GetBranch()
		if branch == "" {
			branch = "main"
		}
		return runKnowledgeMountGit(ctx, dest, temporaryKey, privateKey, "checkout", "-B", branch, commitSHA)
	}
	return runKnowledgeMountGit(ctx, dest, temporaryKey, privateKey, "checkout", "--detach", commitSHA)
}

func checkoutExistingKnowledgeMountCommit(
	ctx context.Context,
	sandboxRoot, dest string,
	m *runnerv1.KnowledgeMount,
	commitSHA string,
) error {
	if m.GetGitPrivateKey() == "" || m.GetMode() == "rw" {
		return checkoutKnowledgeMountCommit(ctx, dest, m, commitSHA, "")
	}
	temporaryKey, err := writeKnowledgeMountKey(sandboxRoot, m.GetGitPrivateKey(), m.GetGitKnownHosts())
	if err != nil {
		return err
	}
	return joinKnowledgeMountCredentialCleanup(
		temporaryKey,
		checkoutKnowledgeMountCommit(ctx, dest, m, commitSHA, temporaryKey),
	)
}

func runKnowledgeMountGit(ctx context.Context, dest, temporaryKey, privateKey string, args ...string) error {
	command := exec.CommandContext(ctx, "git", args...)
	command.Dir = dest
	if temporaryKey != "" {
		command.Env = knowledgeMountSSHEnv(temporaryKey, temporaryKey+temporaryKnownHostsFileSuffix)
	} else {
		command.Env = knowledgeMountGitConfigEnv()
	}
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("%s", knowledgeMountCommandError(args[0], dest, err, output, privateKey))
	}
	return nil
}

func knowledgeMountIsDetached(ctx context.Context, dest string) bool {
	command := exec.CommandContext(ctx, "git", "symbolic-ref", "-q", "HEAD")
	command.Dir = dest
	command.Env = knowledgeMountGitConfigEnv()
	output, err := command.CombinedOutput()
	return err != nil || strings.TrimSpace(string(output)) == ""
}

func verifyReadWriteKnowledgeMountPin(
	ctx context.Context,
	dest string,
	m *runnerv1.KnowledgeMount,
	commitSHA string,
) error {
	privateKey := m.GetGitPrivateKey()
	if err := runKnowledgeMountGit(
		ctx,
		dest,
		"",
		privateKey,
		"fetch",
		"--no-tags",
		"--depth=1",
		"origin",
		commitSHA,
	); err != nil {
		return err
	}
	if err := runKnowledgeMountGit(ctx, dest, "", privateKey, "cat-file", "-e", commitSHA+"^{commit}"); err != nil {
		return err
	}
	if err := runKnowledgeMountGit(ctx, dest, "", privateKey, "merge-base", "--is-ancestor", commitSHA, "HEAD"); err != nil {
		return fmt.Errorf("read-write knowledge mount HEAD does not descend from pinned commit %s: %w", commitSHA, err)
	}
	return nil
}
