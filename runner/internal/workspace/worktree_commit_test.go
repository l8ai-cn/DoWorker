package workspace

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateWorktreeWithPinnedCommitIgnoresMovedBranch(t *testing.T) {
	origin, clone := createPinnedOrigin(t)
	first := commitPinnedFile(t, clone, "first")
	second := commitPinnedFile(t, clone, "second")
	pushPinnedBranch(t, clone)
	assert.NotEqual(t, first, second)

	mgr, err := NewManager(t.TempDir(), "")
	require.NoError(t, err)
	path := filepath.Join(t.TempDir(), "workspace")
	result, err := mgr.CreateWorktreeWithOptions(
		testGitContext(t),
		origin,
		"main",
		path,
		WithSourceCommitSHA(first),
	)
	require.NoError(t, err)
	assert.Equal(t, first, gitHead(t, result.Path))
	assert.True(t, strings.HasPrefix(result.Branch, "worktree-"))
	assert.Equal(t, "refs/heads/"+result.Branch, gitSymbolicRef(t, result.Path))
	assert.Equal(t, "first", readPinnedFile(t, result.Path))
	require.NoError(t, os.WriteFile(filepath.Join(result.Path, "local.txt"), []byte("local"), 0644))
	runPinnedGit(t, result.Path, "add", ".")
	runPinnedGit(t, result.Path, "commit", "-m", "local")
}

func TestCreateWorktreeWithPinnedCommitAppliesGitConfig(t *testing.T) {
	origin, clone := createPinnedOrigin(t)
	commit := commitPinnedFile(t, clone, "configured")
	pushPinnedBranch(t, clone)
	root := t.TempDir()
	globalConfig := filepath.Join(root, "global.gitconfig")
	require.NoError(t, os.WriteFile(globalConfig, []byte("[credential]\n\thelper = forbidden-global\n[http]\n\textraHeader = Authorization: global\n"), 0644))
	t.Setenv("GIT_CONFIG_GLOBAL", globalConfig)
	configPath := filepath.Join(root, "git.config")
	require.NoError(t, os.WriteFile(configPath, []byte("[user]\n\tname = Pinned User\n"), 0644))
	mgr, err := NewManager(root, configPath)
	require.NoError(t, err)

	result, err := mgr.CreateWorktreeWithOptions(
		testGitContext(t),
		origin,
		"main",
		filepath.Join(root, "workspace"),
		WithSourceCommitSHA(commit),
	)
	require.NoError(t, err)
	assert.Equal(t, "Pinned User", gitWorktreeConfig(t, result.Path, "user.name"))
	assert.Equal(t, "Pinned User", gitEffectiveConfig(t, result.Path, "user.name"))
}

func TestCreateWorktreeAnonymousPinnedCommitIsolatesAfterRunnerLocal(t *testing.T) {
	origin, clone := createPinnedOrigin(t)
	commit := commitPinnedFile(t, clone, "anonymous")
	pushPinnedBranch(t, clone)
	root := t.TempDir()
	configPath := filepath.Join(root, "git.config")
	config := "[user]\n\tname = Pinned User\n[credential]\n\thelper = must-not-leak\n[http]\n\textraHeader = Authorization: bearer secret\n"
	require.NoError(t, os.WriteFile(configPath, []byte(config), 0644))
	mgr, err := NewManager(root, configPath)
	require.NoError(t, err)

	local, err := mgr.CreateWorktreeWithOptions(
		testGitContext(t),
		origin,
		"main",
		filepath.Join(root, "sandboxes", "local", "workspace"),
		WithSourceCommitSHA(commit),
	)
	require.NoError(t, err)
	anonymous, err := mgr.CreateWorktreeWithOptions(
		testGitContext(t),
		origin,
		"main",
		filepath.Join(root, "sandboxes", "anonymous", "workspace"),
		WithSourceCommitSHA(commit),
		WithAnonymousAuth(),
	)
	require.NoError(t, err)
	assert.Equal(t, "Pinned User", gitEffectiveConfig(t, local.Path, "user.name"))
	assert.Equal(t, "must-not-leak", gitEffectiveConfig(t, local.Path, "credential.helper"))
	assert.Equal(t, "Authorization: bearer secret", gitEffectiveConfig(t, local.Path, "http.extraHeader"))
	assert.Equal(t, "Pinned User", gitEffectiveConfig(t, local.Path, "user.name"))

	assert.Empty(t, gitEffectiveConfig(t, anonymous.Path, "credential.helper"))
	assert.Empty(t, gitEffectiveConfig(t, anonymous.Path, "http.extraHeader"))
	assert.Contains(t, gitWorktreeConfig(t, anonymous.Path, "core.sshCommand"), "IdentityAgent=none")
	assert.Equal(t, "must-not-leak", gitEffectiveConfig(t, local.Path, "credential.helper"))
	repo := filepath.Join(root, "repos", extractRepoName(origin))
	_, err = gitBareConfigError(repo, "credential.helper")
	require.Error(t, err)
	_, err = gitBareConfigError(repo, "http.extraHeader")
	require.Error(t, err)
}

func TestCreateWorktreeGitConfigFailureFailsClosed(t *testing.T) {
	origin, clone := createPinnedOrigin(t)
	commit := commitPinnedFile(t, clone, "bad-config")
	pushPinnedBranch(t, clone)
	root := t.TempDir()
	configPath := filepath.Join(root, "git.config")
	require.NoError(t, os.WriteFile(configPath, []byte("[broken\n"), 0644))
	mgr, err := NewManager(root, configPath)
	require.NoError(t, err)

	_, err = mgr.CreateWorktreeWithOptions(
		testGitContext(t),
		origin,
		"main",
		filepath.Join(root, "workspace"),
		WithSourceCommitSHA(commit),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list git config entries")
	assert.NoDirExists(t, filepath.Join(root, "workspace"))
	assert.Empty(t, gitBareBranch(t, filepath.Join(root, "repos", extractRepoName(origin)), "worktree-"+filepath.Base(root)))
}

func TestCreateWorktreeWithPinnedCommitRejectsInvalidAndMissing(t *testing.T) {
	origin, clone := createPinnedOrigin(t)
	commitPinnedFile(t, clone, "first")
	pushPinnedBranch(t, clone)
	mgr, err := NewManager(t.TempDir(), "")
	require.NoError(t, err)
	_, err = mgr.CreateWorktreeWithOptions(testGitContext(t), origin, "main", filepath.Join(t.TempDir(), "bad"), WithSourceCommitSHA("ABC"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "lowercase 40 or 64 hex")

	missing := strings.Repeat("0", 40)
	_, err = mgr.CreateWorktreeWithOptions(testGitContext(t), origin, "main", filepath.Join(t.TempDir(), "missing"), WithSourceCommitSHA(missing))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch pinned commit")
}

func createPinnedOrigin(t *testing.T) (string, string) {
	t.Helper()
	requireGit(t)
	dir := t.TempDir()
	origin := filepath.Join(dir, "origin.git")
	clone := filepath.Join(dir, "clone")
	require.NoError(t, exec.Command("git", "init", "--bare", origin).Run())
	require.NoError(t, exec.Command("git", "clone", origin, clone).Run())
	runPinnedGit(t, clone, "config", "user.email", "test@test.com")
	runPinnedGit(t, clone, "config", "user.name", "Test")
	return origin, clone
}

func commitPinnedFile(t *testing.T, clone, content string) string {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(clone, "file.txt"), []byte(content), 0644))
	runPinnedGit(t, clone, "add", ".")
	runPinnedGit(t, clone, "commit", "-m", content)
	runPinnedGit(t, clone, "branch", "-M", "main")
	return gitHead(t, clone)
}

func pushPinnedBranch(t *testing.T, clone string) {
	t.Helper()
	runPinnedGit(t, clone, "push", "-u", "origin", "main", "--force")
}

func runPinnedGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, output)
}

func gitHead(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git rev-parse HEAD: %s", output)
	return strings.TrimSpace(string(output))
}

func gitWorktreeConfig(t *testing.T, dir, key string) string {
	t.Helper()
	cmd := exec.Command("git", "config", "--worktree", "--get", key)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git config %s: %s", key, output)
	return strings.TrimSpace(string(output))
}

func gitEffectiveConfig(t *testing.T, dir, key string) string {
	t.Helper()
	output, err := gitConfigError(dir, key)
	require.NoError(t, err, "git config %s: %s", key, output)
	return strings.TrimSpace(string(output))
}

func gitConfigError(dir, key string) ([]byte, error) {
	cmd := exec.Command("git", "config", "--get", key)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

func gitBareConfigError(repo, key string) ([]byte, error) {
	cmd := exec.Command("git", "--git-dir", repo, "config", "--get", key)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")
	return cmd.CombinedOutput()
}

func gitBareBranch(t *testing.T, repo, branch string) string {
	t.Helper()
	cmd := exec.Command("git", "--git-dir", repo, "branch", "--list", branch)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git branch --list %s: %s", branch, output)
	return strings.TrimSpace(string(output))
}

func gitSymbolicRef(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "symbolic-ref", "HEAD")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git symbolic-ref HEAD: %s", output)
	return strings.TrimSpace(string(output))
}

func readPinnedFile(t *testing.T, dir string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(dir, "file.txt"))
	require.NoError(t, err)
	return string(content)
}

func testGitContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}
