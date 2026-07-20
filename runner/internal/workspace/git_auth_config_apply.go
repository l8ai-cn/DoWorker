package workspace

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const gitCredentialFileName = "agentsmesh-credentials"

func (m *Manager) applyTokenGitConfig(
	ctx context.Context,
	worktreePath, repoURL string,
	options *WorktreeOptions,
) error {
	if err := m.enableWorktreeConfig(ctx, worktreePath); err != nil {
		return err
	}
	credentialURL, err := tokenCredentialURL(repoURL, options)
	if err != nil {
		return err
	}
	credentialPath := filepath.Join(filepath.Dir(worktreePath), gitCredentialFileName)
	if err := os.WriteFile(credentialPath, []byte(credentialURL+"\n"), 0600); err != nil {
		return fmt.Errorf("failed to write worktree Git credential: %w", err)
	}
	if err := secureGitCredentialFile(credentialPath); err != nil {
		return err
	}
	for _, args := range [][]string{
		{"--replace-all", "credential.helper", ""},
		{"--add", "credential.helper", credentialStoreHelper(credentialPath)},
		{"--replace-all", "credential.useHttpPath", "true"},
		{"--replace-all", "http.extraHeader", ""},
		{"--replace-all", "core.sshCommand", anonymousSSHCommand(false)},
	} {
		if err := m.runWorktreeGitConfig(ctx, worktreePath, args...); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) applySSHGitConfig(
	ctx context.Context,
	worktreePath string,
	options *WorktreeOptions,
) error {
	if err := m.enableWorktreeConfig(ctx, worktreePath); err != nil {
		return err
	}
	for _, args := range [][]string{
		{"--replace-all", "credential.helper", ""},
		{"--replace-all", "http.extraHeader", ""},
		{"--replace-all", "core.sshCommand", sshKeyCommand(options.SSHKeyPath, false)},
	} {
		if err := m.runWorktreeGitConfig(ctx, worktreePath, args...); err != nil {
			return err
		}
	}
	return nil
}

func tokenCredentialURL(repoURL string, options *WorktreeOptions) (string, error) {
	if options == nil || options.GitToken == "" || options.GitUsername == "" {
		return "", fmt.Errorf("Git username and token are required")
	}
	if err := validateRepositoryURL(repoURL); err != nil {
		return "", err
	}
	parsed, err := url.Parse(repoURL)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return "", fmt.Errorf("token authentication requires an HTTPS repository URL")
	}
	parsed.User = url.UserPassword(options.GitUsername, options.GitToken)
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String(), nil
}

func (m *Manager) worktreeGitDir(ctx context.Context, worktreePath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-dir")
	cmd.Dir = worktreePath
	m.setLocalGitEnv(cmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to locate worktree Git directory: %w, output: %s", err, output)
	}
	gitDir := strings.TrimSpace(string(output))
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(worktreePath, gitDir)
	}
	return filepath.Clean(gitDir), nil
}

func credentialStoreHelper(path string) string {
	return "store --file=" + shellQuoteGitConfigPath(path)
}

func shellQuoteGitConfigPath(path string) string {
	return "'" + strings.ReplaceAll(path, "'", "'\"'\"'") + "'"
}
