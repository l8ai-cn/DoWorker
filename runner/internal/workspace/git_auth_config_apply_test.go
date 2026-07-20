package workspace

import (
	"context"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyTokenGitConfigStoresCredentialBesideSandboxWorktree(t *testing.T) {
	manager, worktreePath, bareRepoPath := createAuthConfigWorktree(t)
	options := &WorktreeOptions{
		GitUsername: "oauth2",
		GitToken:    "secret-token",
	}
	repoURL := "https://example.test/org/repo.git"

	require.NoError(t, manager.applyTokenGitConfig(context.Background(), worktreePath, repoURL, options))
	require.NoError(t, manager.applyWorktreeRemote(context.Background(), worktreePath, repoURL))

	helper := gitWorktreeConfig(t, worktreePath, "credential.helper")
	assert.Contains(t, helper, gitCredentialFileName)
	credentialPath := strings.TrimPrefix(helper, "store --file='")
	credentialPath = strings.TrimSuffix(credentialPath, "'")
	assert.Equal(t, filepath.Join(filepath.Dir(worktreePath), gitCredentialFileName), credentialPath)
	assert.NotContains(t, credentialPath, bareRepoPath)
	content, err := os.ReadFile(credentialPath)
	require.NoError(t, err)
	parsed, err := url.Parse(strings.TrimSpace(string(content)))
	require.NoError(t, err)
	password, ok := parsed.User.Password()
	require.True(t, ok)
	assert.Equal(t, "oauth2", parsed.User.Username())
	assert.Equal(t, options.GitToken, password)
	assert.Equal(t, repoURL, gitWorktreeConfig(t, worktreePath, "remote.origin.url"))
	assert.Equal(t, "true", gitWorktreeConfig(t, worktreePath, "credential.useHttpPath"))
	assert.Equal(t, "true", gitBareRepositoryState(t, bareRepoPath))
	_, err = gitBareConfigError(bareRepoPath, "credential.helper")
	require.Error(t, err)

	require.NoError(t, manager.RemoveWorktree(context.Background(), worktreePath))
	assert.NoFileExists(t, credentialPath)
	output, err := gitBareConfigCommand(bareRepoPath, "worktree", "list", "--porcelain")
	require.NoError(t, err, "%s", output)
	assert.NotContains(t, string(output), worktreePath)
}

func TestApplySSHGitConfigPersistsExplicitIdentityOnly(t *testing.T) {
	manager, worktreePath, bareRepoPath := createAuthConfigWorktree(t)
	keyPath := filepath.Join(t.TempDir(), "id key")
	options := &WorktreeOptions{SSHKeyPath: keyPath}

	require.NoError(t, manager.applySSHGitConfig(context.Background(), worktreePath, options))

	command := gitWorktreeConfig(t, worktreePath, "core.sshCommand")
	assert.Contains(t, command, "-F ")
	assert.Contains(t, command, "IdentityFile=none")
	assert.Contains(t, command, keyPath)
	assert.Contains(t, command, "IdentityAgent=none")
	assert.Equal(t, "true", gitBareRepositoryState(t, bareRepoPath))
	_, err := gitBareConfigError(bareRepoPath, "core.sshCommand")
	require.Error(t, err)
}

func TestFinalizeExplicitCredentialSkipsRunnerGitConfig(t *testing.T) {
	manager, worktreePath, _ := createAuthConfigWorktree(t)
	configPath := filepath.Join(t.TempDir(), "runner.gitconfig")
	require.NoError(t, os.WriteFile(configPath, []byte(
		"[url \"https://attacker.test/\"]\n\tinsteadOf = https://example.test/\n",
	), 0600))
	manager.gitConfigPath = configPath
	options := &WorktreeOptions{GitUsername: "oauth2", GitToken: "secret-token"}

	_, err := manager.finalizeWorktree(
		context.Background(),
		worktreePath,
		"main",
		"https://example.test/org/repo.git",
		options,
	)

	require.NoError(t, err)
	cmd := exec.Command("git", "config", "--worktree", "--get", "url.https://attacker.test/.insteadof")
	cmd.Dir = worktreePath
	_, err = cmd.CombinedOutput()
	require.Error(t, err)
}

func TestFinalizeExplicitCredentialsEnablesFreshWorktreeConfig(t *testing.T) {
	cases := []struct {
		name    string
		repoURL string
		options *WorktreeOptions
	}{
		{
			name:    "token",
			repoURL: "https://example.test/org/repo.git",
			options: &WorktreeOptions{GitUsername: "oauth2", GitToken: "secret-token"},
		},
		{
			name:    "ssh",
			repoURL: "ssh://git@example.test/org/repo.git",
			options: &WorktreeOptions{SSHKeyPath: filepath.Join(t.TempDir(), "id_ed25519")},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			manager, worktreePath, bareRepoPath := createFreshAuthConfigWorktree(t)

			_, err := manager.finalizeWorktree(
				context.Background(),
				worktreePath,
				"main",
				testCase.repoURL,
				testCase.options,
			)

			require.NoError(t, err)
			output, err := gitBareConfigCommand(
				bareRepoPath,
				"config",
				"--get",
				"extensions.worktreeConfig",
			)
			require.NoError(t, err, "%s", output)
			assert.Equal(t, "true", strings.TrimSpace(string(output)))
		})
	}
}

func createAuthConfigWorktree(t *testing.T) (*Manager, string, string) {
	t.Helper()
	origin, clone := createPinnedOrigin(t)
	commit := commitPinnedFile(t, clone, "auth")
	pushPinnedBranch(t, clone)
	root := t.TempDir()
	manager, err := NewManager(root, "")
	require.NoError(t, err)
	result, err := manager.CreateWorktreeWithOptions(
		testGitContext(t),
		origin,
		"main",
		filepath.Join(root, "sandboxes", "auth", "workspace"),
		WithSourceCommitSHA(commit),
	)
	require.NoError(t, err)
	commonDir := exec.Command("git", "rev-parse", "--git-common-dir")
	commonDir.Dir = result.Path
	output, err := commonDir.CombinedOutput()
	require.NoError(t, err, "%s", output)
	return manager, result.Path, strings.TrimSpace(string(output))
}

func createFreshAuthConfigWorktree(t *testing.T) (*Manager, string, string) {
	t.Helper()
	origin, clone := createPinnedOrigin(t)
	commit := commitPinnedFile(t, clone, "fresh-auth")
	pushPinnedBranch(t, clone)
	root := t.TempDir()
	bareRepoPath := filepath.Join(root, "repos", "repo.git")
	require.NoError(t, os.MkdirAll(filepath.Dir(bareRepoPath), 0755))
	output, err := exec.Command("git", "clone", "--bare", origin, bareRepoPath).CombinedOutput()
	require.NoError(t, err, "%s", output)
	worktreePath := filepath.Join(root, "sandboxes", "fresh-auth", "workspace")
	require.NoError(t, os.MkdirAll(filepath.Dir(worktreePath), 0755))
	command := exec.Command("git", "--git-dir", bareRepoPath, "worktree", "add", "--detach", worktreePath, commit)
	output, err = command.CombinedOutput()
	require.NoError(t, err, "%s", output)
	manager, err := NewManager(root, "")
	require.NoError(t, err)
	return manager, worktreePath, bareRepoPath
}

func gitBareRepositoryState(t *testing.T, repoPath string) string {
	t.Helper()
	output, err := gitBareConfigCommand(repoPath, "rev-parse", "--is-bare-repository")
	require.NoError(t, err, "%s", output)
	return strings.TrimSpace(string(output))
}

func gitBareConfigCommand(repoPath string, args ...string) ([]byte, error) {
	command := append([]string{"--git-dir", repoPath}, args...)
	return exec.Command("git", command...).CombinedOutput()
}
