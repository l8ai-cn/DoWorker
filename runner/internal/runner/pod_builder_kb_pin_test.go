package runner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneKnowledgeMountPinnedCommitIgnoresMovedBranch(t *testing.T) {
	origin := initKBOriginRepo(t)
	first := kbHead(t, origin)
	second := addKBCommit(t, origin, "moved")
	require.NotEqual(t, first, second)
	sandbox := t.TempDir()
	builder := &PodBuilder{cmd: &runnerv1.CreatePodCommand{PodKey: "pod-1"}}
	mount := &runnerv1.KnowledgeMount{
		Slug: "docs", HttpCloneUrl: origin, Branch: "main", CommitSha: first, Mode: "ro",
	}

	require.NoError(t, builder.cloneKnowledgeMount(context.Background(), sandbox, mount))

	dest := filepath.Join(sandbox, "kb", "docs")
	assert.Equal(t, first, kbHead(t, dest))
	assert.Equal(t, "# KB\n", readKBFile(t, filepath.Join(dest, "llms.txt")))
}

func TestKnowledgeMountReadWritePinnedCommitUsesBranchAndPushes(t *testing.T) {
	origin, clone := initKBBareOriginRepo(t)
	first := addKBCommit(t, clone, "first")
	runKBRepoGit(t, clone, "branch", "-M", "main")
	runKBRepoGit(t, clone, "push", "-u", "origin", "main", "--force")
	dest := filepath.Join(t.TempDir(), "docs")
	runKBRepoGit(t, t.TempDir(), "clone", "--no-checkout", origin, dest)
	runKBRepoGit(t, dest, "config", "user.email", "test@test.com")
	runKBRepoGit(t, dest, "config", "user.name", "Test")
	mount := &runnerv1.KnowledgeMount{
		Slug: "docs", Branch: "main", CommitSha: first, Mode: "rw",
	}

	require.NoError(t, checkoutKnowledgeMountCommit(context.Background(), dest, mount, first, ""))

	assert.Equal(t, first, kbHead(t, dest))
	assert.Equal(t, "refs/heads/main", kbSymbolicRef(t, dest))
	require.NoError(t, os.WriteFile(filepath.Join(dest, "rw.txt"), []byte("rw"), 0644))
	runKBRepoGit(t, dest, "add", ".")
	runKBRepoGit(t, dest, "commit", "-m", "rw")
	runKBRepoGit(t, dest, "push", "origin", "main")
	assert.Equal(t, kbHead(t, dest), kbBareRef(t, origin, "refs/heads/main"))
}

func TestSetupKnowledgeMountsReadWriteResumePreservesWorktree(t *testing.T) {
	origin := initKBOriginRepo(t)
	sandbox := t.TempDir()
	realGit, _, _ := installKBGitHarness(t, origin, sandbox)
	mount := &runnerv1.KnowledgeMount{
		Slug: "docs", HttpCloneUrl: "https://gitea.test/am-kb/docs.git",
		SshCloneUrl: "ssh://git@gitea.test/am-kb/docs.git", GitPrivateKey: "rw-key",
		GitKnownHosts: kbKnownHosts, Branch: "main", CommitSha: kbHead(t, origin), Mode: "rw",
	}
	builder := &PodBuilder{cmd: &runnerv1.CreatePodCommand{
		PodKey: "resume-rw-pod",
		SandboxConfig: &runnerv1.SandboxConfig{
			KnowledgeMounts: []*runnerv1.KnowledgeMount{mount},
		},
	}}
	require.NoError(t, builder.setupKnowledgeMounts(context.Background(), sandbox))
	dest := filepath.Join(sandbox, "kb", "docs")
	runKBGit(t, realGit, dest, "config", "user.email", "test@test.com")
	runKBGit(t, realGit, dest, "config", "user.name", "Test")
	require.NoError(t, os.WriteFile(filepath.Join(dest, "local.txt"), []byte("local"), 0644))
	runKBGit(t, realGit, dest, "add", "local.txt")
	runKBGit(t, realGit, dest, "commit", "-m", "local")
	localCommit := kbHead(t, dest)
	require.NoError(t, os.WriteFile(filepath.Join(dest, "draft.txt"), []byte("draft"), 0644))

	require.NoError(t, builder.setupKnowledgeMounts(context.Background(), sandbox))

	assert.Equal(t, localCommit, kbHead(t, dest))
	assert.Equal(t, "refs/heads/main", kbSymbolicRef(t, dest))
	assert.Equal(t, "draft", readKBFile(t, filepath.Join(dest, "draft.txt")))
}

func TestSetupKnowledgeMountsReadWriteResumeRejectsHeadOutsidePin(t *testing.T) {
	origin := initKBOriginRepo(t)
	pinned := kbHead(t, origin)
	sandbox := t.TempDir()
	realGit, _, _ := installKBGitHarness(t, origin, sandbox)
	mount := &runnerv1.KnowledgeMount{
		Slug: "docs", HttpCloneUrl: "https://gitea.test/am-kb/docs.git",
		SshCloneUrl: "ssh://git@gitea.test/am-kb/docs.git", GitPrivateKey: "rw-key",
		GitKnownHosts: kbKnownHosts, Branch: "main", CommitSha: pinned, Mode: "rw",
	}
	builder := &PodBuilder{cmd: &runnerv1.CreatePodCommand{
		PodKey: "resume-rw-pin-pod",
		SandboxConfig: &runnerv1.SandboxConfig{
			KnowledgeMounts: []*runnerv1.KnowledgeMount{mount},
		},
	}}
	require.NoError(t, builder.setupKnowledgeMounts(context.Background(), sandbox))
	dest := filepath.Join(sandbox, "kb", "docs")
	runKBGit(t, realGit, dest, "checkout", "--orphan", "unrelated")
	require.NoError(t, os.WriteFile(filepath.Join(dest, "unrelated.txt"), []byte("unrelated"), 0644))
	runKBGit(t, realGit, dest, "add", ".")
	runKBGit(t, realGit, dest, "commit", "-m", "unrelated")

	err := builder.setupKnowledgeMounts(context.Background(), sandbox)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not descend from pinned commit")
}

func TestSetupKnowledgeMountsReadOnlyResumeReturnsToPinnedCommit(t *testing.T) {
	origin := initKBOriginRepo(t)
	pinned := kbHead(t, origin)
	sandbox := t.TempDir()
	mount := &runnerv1.KnowledgeMount{
		Slug: "docs", HttpCloneUrl: origin, Branch: "main", CommitSha: pinned, Mode: "ro",
	}
	builder := &PodBuilder{cmd: &runnerv1.CreatePodCommand{
		PodKey: "resume-ro-pod",
		SandboxConfig: &runnerv1.SandboxConfig{
			KnowledgeMounts: []*runnerv1.KnowledgeMount{mount},
		},
	}}
	require.NoError(t, builder.setupKnowledgeMounts(context.Background(), sandbox))
	moved := addKBCommit(t, origin, "moved")
	dest := filepath.Join(sandbox, "kb", "docs")
	runKBRepoGit(t, dest, "fetch", "--no-tags", "--depth=1", "origin", moved)
	runKBRepoGit(t, dest, "checkout", "--detach", moved)
	require.Equal(t, moved, kbHead(t, dest))

	require.NoError(t, builder.setupKnowledgeMounts(context.Background(), sandbox))

	assert.Equal(t, pinned, kbHead(t, dest))
	_, err := kbSymbolicRefError(dest)
	require.Error(t, err)
}

func TestCloneKnowledgeMountRejectsMissingInvalidAndUnreachableCommit(t *testing.T) {
	origin := initKBOriginRepo(t)
	builder := &PodBuilder{cmd: &runnerv1.CreatePodCommand{PodKey: "pod-1"}}

	missing := &runnerv1.KnowledgeMount{Slug: "docs", HttpCloneUrl: origin, Branch: "main", Mode: "ro"}
	err := builder.cloneKnowledgeMount(context.Background(), t.TempDir(), missing)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "commit_sha is required")

	invalid := &runnerv1.KnowledgeMount{Slug: "docs", HttpCloneUrl: origin, Branch: "main", CommitSha: "ABC", Mode: "ro"}
	err = builder.cloneKnowledgeMount(context.Background(), t.TempDir(), invalid)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "lowercase 40 or 64 hex")

	unreachable := &runnerv1.KnowledgeMount{
		Slug: "docs", HttpCloneUrl: origin, Branch: "main", CommitSha: strings.Repeat("0", 40), Mode: "ro",
	}
	err = builder.cloneKnowledgeMount(context.Background(), t.TempDir(), unreachable)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetch")
}

func kbHead(t *testing.T, repo string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repo
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git rev-parse HEAD: %s", output)
	return strings.TrimSpace(string(output))
}

func kbBareRef(t *testing.T, repo, ref string) string {
	t.Helper()
	cmd := exec.Command("git", "--git-dir", repo, "rev-parse", ref)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git rev-parse %s: %s", ref, output)
	return strings.TrimSpace(string(output))
}

func kbSymbolicRef(t *testing.T, repo string) string {
	t.Helper()
	cmd := exec.Command("git", "symbolic-ref", "HEAD")
	cmd.Dir = repo
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git symbolic-ref HEAD: %s", output)
	return strings.TrimSpace(string(output))
}

func kbSymbolicRefError(repo string) ([]byte, error) {
	cmd := exec.Command("git", "symbolic-ref", "HEAD")
	cmd.Dir = repo
	return cmd.CombinedOutput()
}

func initKBBareOriginRepo(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	origin := filepath.Join(dir, "origin.git")
	clone := filepath.Join(dir, "clone")
	require.NoError(t, exec.Command("git", "init", "--bare", origin).Run())
	require.NoError(t, exec.Command("git", "clone", origin, clone).Run())
	runKBRepoGit(t, clone, "config", "user.email", "test@test.com")
	runKBRepoGit(t, clone, "config", "user.name", "Test")
	return origin, clone
}

func addKBCommit(t *testing.T, repo, content string) string {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(repo, "llms.txt"), []byte(content), 0644))
	runKBRepoGit(t, repo, "add", ".")
	runKBRepoGit(t, repo, "commit", "-m", "update")
	return kbHead(t, repo)
}

func runKBRepoGit(t *testing.T, repo string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, output)
}
