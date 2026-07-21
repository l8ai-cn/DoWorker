package workspace

import (
	"fmt"
	"os"
	"sync"

	"github.com/l8ai-cn/agentcloud/runner/internal/logger"
)

// Manager manages workspace directories and git worktrees
type Manager struct {
	root          string
	gitConfigPath string
	mu            sync.Mutex // Global lock for cleanup/list operations
	repoLocks     sync.Map   // repoName -> *sync.Mutex (per-repo locking)
}

// WorktreeOptions contains options for creating a worktree
type WorktreeOptions struct {
	GitToken        string // Git token for HTTPS authentication
	GitUsername     string // Git HTTPS username paired with GitToken
	SSHKeyPath      string // Path to SSH key for SSH authentication
	HttpCloneURL    string // HTTPS clone URL
	SshCloneURL     string // SSH clone URL
	SourceCommitSHA string // Immutable source commit to checkout
	AnonymousAuth   bool   // Disable runner-local credentials
}

// WorktreeOption is a function that modifies WorktreeOptions
type WorktreeOption func(*WorktreeOptions)

// WithGitToken sets the git token for HTTPS authentication
func WithGitToken(token string) WorktreeOption {
	return WithGitTokenCredentials("x-access-token", token)
}

func WithGitTokenCredentials(username, token string) WorktreeOption {
	return func(opts *WorktreeOptions) {
		opts.GitUsername = username
		opts.GitToken = token
	}
}

// WithSSHKeyPath sets the SSH key path for SSH authentication
func WithSSHKeyPath(path string) WorktreeOption {
	return func(opts *WorktreeOptions) {
		opts.SSHKeyPath = path
	}
}

// WithHttpCloneURL sets the HTTPS clone URL
func WithHttpCloneURL(url string) WorktreeOption {
	return func(opts *WorktreeOptions) {
		opts.HttpCloneURL = url
	}
}

// WithSshCloneURL sets the SSH clone URL
func WithSshCloneURL(url string) WorktreeOption {
	return func(opts *WorktreeOptions) {
		opts.SshCloneURL = url
	}
}

func WithSourceCommitSHA(sha string) WorktreeOption {
	return func(opts *WorktreeOptions) {
		opts.SourceCommitSHA = sha
	}
}

func WithAnonymousAuth() WorktreeOption {
	return func(opts *WorktreeOptions) {
		opts.AnonymousAuth = true
	}
}

// NewManager creates a new workspace manager
func NewManager(root, gitConfigPath string) (*Manager, error) {
	// Ensure root directory exists
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workspace root: %w", err)
	}

	logger.Runner().Info("Workspace manager created", "root", root)

	return &Manager{
		root:          root,
		gitConfigPath: gitConfigPath,
	}, nil
}

// getRepoLock returns a per-repository mutex, creating one if needed.
// This allows concurrent worktree creation for different repositories
// while serializing operations on the same repository.
func (m *Manager) getRepoLock(repoName string) *sync.Mutex {
	actual, _ := m.repoLocks.LoadOrStore(repoName, &sync.Mutex{})
	return actual.(*sync.Mutex)
}
