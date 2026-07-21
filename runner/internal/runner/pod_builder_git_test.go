package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/l8ai-cn/agentcloud/runner/internal/client"
	"github.com/l8ai-cn/agentcloud/runner/internal/config"
	"github.com/l8ai-cn/agentcloud/runner/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testCommitSHA = "0123456789abcdef0123456789abcdef01234567"

// mockWorkspace implements workspace.WorkspaceManagerInterface for testing.
type mockWorkspace struct {
	result             *workspace.WorktreeResult
	err                error
	opts               []workspace.WorktreeOption // captured options
	createGitMetadata  bool
	removedWorktreeIDs []string
	removeErr          error
}

func (m *mockWorkspace) CreateWorktree(_ context.Context, _, _, _ string) (*workspace.WorktreeResult, error) {
	return m.result, m.err
}
func (m *mockWorkspace) CreateWorktreeWithOptions(_ context.Context, _, _, worktreePath string, opts ...workspace.WorktreeOption) (*workspace.WorktreeResult, error) {
	m.opts = opts
	if m.createGitMetadata {
		if err := os.MkdirAll(worktreePath, 0755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(worktreePath, ".git"), []byte("gitdir: test"), 0644); err != nil {
			return nil, err
		}
		return &workspace.WorktreeResult{Path: worktreePath, Branch: "main"}, nil
	}
	return m.result, m.err
}
func (m *mockWorkspace) RemoveWorktree(_ context.Context, path string) error {
	m.removedWorktreeIDs = append(m.removedWorktreeIDs, path)
	return m.removeErr
}
func (m *mockWorkspace) CleanupOldWorktrees(_ context.Context) error { return nil }
func (m *mockWorkspace) TempWorkspace(_ string) string               { return "" }
func (m *mockWorkspace) GetWorkspaceRoot() string                    { return "" }
func (m *mockWorkspace) ListWorktrees() ([]string, error)            { return nil, nil }

func gitBuilder(ws workspace.WorkspaceManagerInterface, cfg *runnerv1.SandboxConfig) *PodBuilder {
	r := &Runner{cfg: &config.Config{WorkspaceRoot: os.TempDir()}}
	cmd := &runnerv1.CreatePodCommand{
		PodKey:          "git-test-pod",
		AgentfileSource: "AGENT echo\n",
		SandboxConfig:   cfg,
	}
	return NewPodBuilder(PodBuilderDeps{Config: r.cfg, Workspace: ws}).WithCommand(cmd)
}

func collectWorktreeOptions(opts []workspace.WorktreeOption) workspace.WorktreeOptions {
	var collected workspace.WorktreeOptions
	for _, opt := range opts {
		opt(&collected)
	}
	return collected
}

func TestSetupGitWorktree_Success(t *testing.T) {
	ws := &mockWorkspace{result: &workspace.WorktreeResult{Path: "/tmp/ws", Branch: "main"}}
	b := gitBuilder(ws, &runnerv1.SandboxConfig{
		HttpCloneUrl:    "https://github.com/org/repo.git",
		SourceBranch:    "main",
		SourceCommitSha: testCommitSHA,
		CredentialType:  "runner_local",
	})
	path, branch, err := b.setupGitWorktree(context.Background(), t.TempDir(), b.cmd.SandboxConfig)
	require.NoError(t, err)
	assert.Equal(t, "/tmp/ws", path)
	assert.Equal(t, "main", branch)
}

func TestSetupGitWorktreePostCreateFailureRemovesRegisteredWorktree(t *testing.T) {
	root := t.TempDir()
	ws := &mockWorkspace{
		createGitMetadata: true,
		removeErr:         errors.New("prune failed"),
	}
	cmd := &runnerv1.CreatePodCommand{
		PodKey:        "git-cleanup-pod",
		LaunchCommand: "echo",
		SandboxConfig: &runnerv1.SandboxConfig{
			HttpCloneUrl:    "https://github.com/org/repo.git",
			SourceCommitSha: testCommitSHA,
			CredentialType:  "none",
		},
		FilesToCreate: []*runnerv1.FileToCreate{{
			Path:    "/outside/workspace.txt",
			Content: "reject",
		}},
	}
	builder := NewPodBuilder(PodBuilderDeps{
		Config:    &config.Config{WorkspaceRoot: root},
		Workspace: ws,
	}).WithCommand(cmd)

	_, _, _, err := builder.setup(context.Background())
	require.Error(t, err)
	assert.ErrorContains(t, err, "prune failed")
	expectedWorktree := filepath.Join(root, "sandboxes", cmd.PodKey, "workspace")
	assert.Equal(t, []string{expectedWorktree}, ws.removedWorktreeIDs)
	assert.NoDirExists(t, filepath.Join(root, "sandboxes", cmd.PodKey))
}

func TestSetupGitWorktree_NoneCredential(t *testing.T) {
	ws := &mockWorkspace{result: &workspace.WorktreeResult{Path: "/tmp/ws", Branch: "main"}}
	b := gitBuilder(ws, &runnerv1.SandboxConfig{
		HttpCloneUrl:    "https://github.com/org/repo.git",
		SourceCommitSha: testCommitSHA,
		CredentialType:  "none",
	})
	_, _, err := b.setupGitWorktree(context.Background(), t.TempDir(), b.cmd.SandboxConfig)
	require.NoError(t, err)
	assert.True(t, collectWorktreeOptions(ws.opts).AnonymousAuth)
}

func TestSetupGitWorktree_RunnerLocalCredentialKeepsLocalAuth(t *testing.T) {
	ws := &mockWorkspace{result: &workspace.WorktreeResult{Path: "/tmp/ws", Branch: "main"}}
	b := gitBuilder(ws, &runnerv1.SandboxConfig{
		HttpCloneUrl:    "https://github.com/org/repo.git",
		SourceCommitSha: testCommitSHA,
		CredentialType:  "runner_local",
	})
	_, _, err := b.setupGitWorktree(context.Background(), t.TempDir(), b.cmd.SandboxConfig)
	require.NoError(t, err)
	assert.False(t, collectWorktreeOptions(ws.opts).AnonymousAuth)
}

func TestSetupGitWorktree_EmptyURL(t *testing.T) {
	ws := &mockWorkspace{}
	b := gitBuilder(ws, &runnerv1.SandboxConfig{
		SourceCommitSha: testCommitSHA,
		CredentialType:  "runner_local",
	})
	_, _, err := b.setupGitWorktree(context.Background(), t.TempDir(), b.cmd.SandboxConfig)
	require.Error(t, err)
	var podErr *client.PodError
	require.ErrorAs(t, err, &podErr)
	assert.Equal(t, client.ErrCodeGitClone, podErr.Code)
}

func TestSetupGitWorktree_MissingCommitRejected(t *testing.T) {
	ws := &mockWorkspace{result: &workspace.WorktreeResult{Path: "/tmp/ws", Branch: "main"}}
	b := gitBuilder(ws, &runnerv1.SandboxConfig{
		HttpCloneUrl:   "https://github.com/org/repo.git",
		CredentialType: "runner_local",
	})
	_, _, err := b.setupGitWorktree(context.Background(), t.TempDir(), b.cmd.SandboxConfig)
	var podErr *client.PodError
	require.ErrorAs(t, err, &podErr)
	assert.Equal(t, client.ErrCodeGitWorktree, podErr.Code)
	assert.Contains(t, podErr.Message, "source_commit_sha is required")
}

func TestSetupGitWorktree_InvalidCommitRejected(t *testing.T) {
	ws := &mockWorkspace{result: &workspace.WorktreeResult{Path: "/tmp/ws", Branch: "main"}}
	b := gitBuilder(ws, &runnerv1.SandboxConfig{
		HttpCloneUrl:    "https://github.com/org/repo.git",
		SourceCommitSha: "ABC",
		CredentialType:  "runner_local",
	})
	_, _, err := b.setupGitWorktree(context.Background(), t.TempDir(), b.cmd.SandboxConfig)
	var podErr *client.PodError
	require.ErrorAs(t, err, &podErr)
	assert.Equal(t, client.ErrCodeGitWorktree, podErr.Code)
	assert.Contains(t, podErr.Message, "lowercase 40 or 64 hex")
}

func TestSetupGitWorktree_EmptyCredentialTypeRejected(t *testing.T) {
	ws := &mockWorkspace{result: &workspace.WorktreeResult{Path: "/tmp/ws", Branch: "main"}}
	b := gitBuilder(ws, &runnerv1.SandboxConfig{
		HttpCloneUrl:    "https://github.com/org/repo.git",
		SourceCommitSha: testCommitSHA,
	})
	_, _, err := b.setupGitWorktree(context.Background(), t.TempDir(), b.cmd.SandboxConfig)
	var podErr *client.PodError
	require.ErrorAs(t, err, &podErr)
	assert.Equal(t, client.ErrCodeGitAuth, podErr.Code)
	assert.Contains(t, podErr.Message, "credential_type is required")
}

func TestSetupGitWorktree_NilWorkspace(t *testing.T) {
	b := gitBuilder(nil, &runnerv1.SandboxConfig{
		HttpCloneUrl:    "https://github.com/org/repo.git",
		SourceCommitSha: testCommitSHA,
		CredentialType:  "runner_local",
	})
	// Set Workspace to nil explicitly
	b.deps.Workspace = nil
	_, _, err := b.setupGitWorktree(context.Background(), t.TempDir(), b.cmd.SandboxConfig)
	require.Error(t, err)
	var podErr *client.PodError
	require.ErrorAs(t, err, &podErr)
	assert.Equal(t, client.ErrCodeGitWorktree, podErr.Code)
}

func TestSetupGitWorktreeErrorRedactsRepositoryUserinfo(t *testing.T) {
	ws := &mockWorkspace{err: errors.New("repository rejected")}
	b := gitBuilder(ws, &runnerv1.SandboxConfig{
		HttpCloneUrl:    "https://user:secret@example.test/org/repo.git?token=hidden",
		SourceCommitSha: testCommitSHA,
		CredentialType:  "none",
	})

	_, _, err := b.setupGitWorktree(context.Background(), t.TempDir(), b.cmd.SandboxConfig)

	var podErr *client.PodError
	require.ErrorAs(t, err, &podErr)
	assert.Equal(t, "https://example.test/org/repo.git", podErr.Details["repository"])
	assert.NotContains(t, fmt.Sprintf("%v", podErr.Details), "secret")
	assert.NotContains(t, fmt.Sprintf("%v", podErr.Details), "hidden")
}

func TestSetupGitWorktree_OAuthToken(t *testing.T) {
	ws := &mockWorkspace{result: &workspace.WorktreeResult{Path: "/tmp/ws", Branch: "dev"}}
	b := gitBuilder(ws, &runnerv1.SandboxConfig{
		HttpCloneUrl:    "https://github.com/org/repo.git",
		SourceBranch:    "dev",
		SourceCommitSha: testCommitSHA,
		CredentialType:  "oauth",
		GitToken:        "gho_xxx",
	})
	_, _, err := b.setupGitWorktree(context.Background(), t.TempDir(), b.cmd.SandboxConfig)
	require.NoError(t, err)
	credentials := collectWorktreeOptions(ws.opts)
	assert.Equal(t, "oauth2", credentials.GitUsername)
	assert.Equal(t, "gho_xxx", credentials.GitToken)
}

func TestSetupGitWorktree_OAuthMissingTokenRejected(t *testing.T) {
	ws := &mockWorkspace{result: &workspace.WorktreeResult{Path: "/tmp/ws", Branch: "dev"}}
	b := gitBuilder(ws, &runnerv1.SandboxConfig{
		HttpCloneUrl:    "https://github.com/org/repo.git",
		SourceCommitSha: testCommitSHA,
		CredentialType:  "oauth",
	})
	_, _, err := b.setupGitWorktree(context.Background(), t.TempDir(), b.cmd.SandboxConfig)
	var podErr *client.PodError
	require.ErrorAs(t, err, &podErr)
	assert.Equal(t, client.ErrCodeGitAuth, podErr.Code)
	assert.Contains(t, podErr.Message, "git_token is required")
}

func TestSetupGitWorktree_PATToken(t *testing.T) {
	ws := &mockWorkspace{result: &workspace.WorktreeResult{Path: "/tmp/ws", Branch: "main"}}
	b := gitBuilder(ws, &runnerv1.SandboxConfig{
		HttpCloneUrl:    "https://github.com/org/repo.git",
		SourceCommitSha: testCommitSHA,
		CredentialType:  "pat",
		GitToken:        "ghp_xxx",
	})
	_, _, err := b.setupGitWorktree(context.Background(), t.TempDir(), b.cmd.SandboxConfig)
	require.NoError(t, err)
	credentials := collectWorktreeOptions(ws.opts)
	assert.Equal(t, "x-access-token", credentials.GitUsername)
	assert.Equal(t, "ghp_xxx", credentials.GitToken)
}

func TestSetupGitWorktree_PATMissingTokenRejected(t *testing.T) {
	ws := &mockWorkspace{result: &workspace.WorktreeResult{Path: "/tmp/ws", Branch: "main"}}
	b := gitBuilder(ws, &runnerv1.SandboxConfig{
		HttpCloneUrl:    "https://github.com/org/repo.git",
		SourceCommitSha: testCommitSHA,
		CredentialType:  "pat",
	})
	_, _, err := b.setupGitWorktree(context.Background(), t.TempDir(), b.cmd.SandboxConfig)
	var podErr *client.PodError
	require.ErrorAs(t, err, &podErr)
	assert.Equal(t, client.ErrCodeGitAuth, podErr.Code)
	assert.Contains(t, podErr.Message, "git_token is required")
}

func TestSetupGitWorktree_SSHKey(t *testing.T) {
	sandbox := t.TempDir()
	ws := &mockWorkspace{result: &workspace.WorktreeResult{Path: sandbox + "/workspace", Branch: "main"}}
	b := gitBuilder(ws, &runnerv1.SandboxConfig{
		SshCloneUrl:     "git@github.com:org/repo.git",
		SourceCommitSha: testCommitSHA,
		CredentialType:  "ssh_key",
		SshPrivateKey:   "-----BEGIN OPENSSH PRIVATE KEY-----\ntest\n-----END OPENSSH PRIVATE KEY-----",
	})
	_, _, err := b.setupGitWorktree(context.Background(), sandbox, b.cmd.SandboxConfig)
	require.NoError(t, err)
	// Verify SSH key was written
	keyFile := filepath.Join(sandbox, ".ssh_key")
	data, readErr := os.ReadFile(keyFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(data), "OPENSSH PRIVATE KEY")
	// Verify permissions (Unix only — Windows uses ACLs)
	if runtime.GOOS != "windows" {
		info, _ := os.Stat(keyFile)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	}
}

func TestSetupGitWorktree_SSHKeyMissingPrivateKeyRejected(t *testing.T) {
	ws := &mockWorkspace{result: &workspace.WorktreeResult{Path: "/tmp/ws", Branch: "main"}}
	b := gitBuilder(ws, &runnerv1.SandboxConfig{
		SshCloneUrl:     "git@github.com:org/repo.git",
		SourceCommitSha: testCommitSHA,
		CredentialType:  "ssh_key",
	})
	_, _, err := b.setupGitWorktree(context.Background(), t.TempDir(), b.cmd.SandboxConfig)
	var podErr *client.PodError
	require.ErrorAs(t, err, &podErr)
	assert.Equal(t, client.ErrCodeGitAuth, podErr.Code)
	assert.Contains(t, podErr.Message, "ssh_private_key is required")
}

func TestSetupGitWorktree_HttpCloneURL_Preferred(t *testing.T) {
	ws := &mockWorkspace{result: &workspace.WorktreeResult{Path: "/tmp/ws", Branch: "main"}}
	b := gitBuilder(ws, &runnerv1.SandboxConfig{
		HttpCloneUrl:    "https://new-url.com/repo.git",
		SshCloneUrl:     "git@github.com:org/repo.git",
		SourceCommitSha: testCommitSHA,
		CredentialType:  "runner_local",
	})
	_, _, err := b.setupGitWorktree(context.Background(), t.TempDir(), b.cmd.SandboxConfig)
	require.NoError(t, err)
	// HTTP clone URL should be used when both available
	assert.True(t, len(ws.opts) >= 2, "should pass both HttpCloneURL and SshCloneURL options")
}

func TestSetupGitWorktree_AuthError(t *testing.T) {
	ws := &mockWorkspace{err: fmt.Errorf("authentication failed: Permission denied")}
	b := gitBuilder(ws, &runnerv1.SandboxConfig{
		HttpCloneUrl:    "https://github.com/org/repo.git",
		SourceCommitSha: testCommitSHA,
		CredentialType:  "runner_local",
	})
	_, _, err := b.setupGitWorktree(context.Background(), t.TempDir(), b.cmd.SandboxConfig)
	require.Error(t, err)
	var podErr *client.PodError
	require.ErrorAs(t, err, &podErr)
	assert.Equal(t, client.ErrCodeGitAuth, podErr.Code)
}

func TestSetupGitWorktree_CloneError(t *testing.T) {
	ws := &mockWorkspace{err: fmt.Errorf("failed to clone repository")}
	b := gitBuilder(ws, &runnerv1.SandboxConfig{
		HttpCloneUrl:    "https://github.com/org/repo.git",
		SourceCommitSha: testCommitSHA,
		CredentialType:  "runner_local",
	})
	_, _, err := b.setupGitWorktree(context.Background(), t.TempDir(), b.cmd.SandboxConfig)
	require.Error(t, err)
	var podErr *client.PodError
	require.ErrorAs(t, err, &podErr)
	assert.Equal(t, client.ErrCodeGitClone, podErr.Code)
}

// TestSetupGitWorktree_DeprecatedRepositoryUrl_Ignored is a regression test
// ensuring the deprecated RepositoryUrl proto field is never used as a fallback
// when HttpCloneUrl and SshCloneUrl are both empty. The builder must return
// ErrCodeGitClone instead of silently cloning from the legacy field.
func TestSetupGitWorktree_DeprecatedRepositoryUrl_Ignored(t *testing.T) {
	ws := &mockWorkspace{result: &workspace.WorktreeResult{Path: "/tmp/ws", Branch: "main"}}
	b := gitBuilder(ws, &runnerv1.SandboxConfig{
		RepositoryUrl:   "https://github.com/org/repo.git", //nolint:staticcheck // testing deprecated field
		SourceCommitSha: testCommitSHA,
		CredentialType:  "runner_local",
		// HttpCloneUrl and SshCloneUrl intentionally left empty
	})
	_, _, err := b.setupGitWorktree(context.Background(), t.TempDir(), b.cmd.SandboxConfig)
	require.Error(t, err, "should not fall back to deprecated RepositoryUrl")
	var podErr *client.PodError
	require.ErrorAs(t, err, &podErr)
	assert.Equal(t, client.ErrCodeGitClone, podErr.Code)
	assert.Contains(t, podErr.Message, "http_clone_url or ssh_clone_url is required")
}

func TestSetupGitWorktree_UnknownCredentialType(t *testing.T) {
	ws := &mockWorkspace{result: &workspace.WorktreeResult{Path: "/tmp/ws", Branch: "main"}}
	b := gitBuilder(ws, &runnerv1.SandboxConfig{
		HttpCloneUrl:    "https://github.com/org/repo.git",
		SourceCommitSha: testCommitSHA,
		CredentialType:  "unknown_type",
	})
	_, _, err := b.setupGitWorktree(context.Background(), t.TempDir(), b.cmd.SandboxConfig)
	var podErr *client.PodError
	require.ErrorAs(t, err, &podErr)
	assert.Equal(t, client.ErrCodeGitAuth, podErr.Code)
}
