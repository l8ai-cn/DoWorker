// Package gitops is the shared choke point for platform-owned Gitea repository
// lifecycle and content operations. It composes over internal/infra/gitea (the
// thin HTTP transport) and adds domain-agnostic policy: repo naming,
// per-domain namespaces, seed-commit + compensating cleanup, create-vs-update
// SHA probing, and error normalization (Gitea 404 -> ErrNotFound).
//
// Higher-level services (expert, skill, and eventually knowledgebase) import
// gitops and must not import internal/infra/gitea directly. One Service
// instance is bound to one namespace (per domain).
package gitops

import "context"

// Author identifies the committer for seed/edit commits.
type Author struct {
	Name  string
	Email string
}

// FileChange is a create-or-update of one path in a commit. Content is bytes
// so binary assets (e.g. avatars) are first-class; the gitea client already
// base64-encodes on commit and decodes on read.
type FileChange struct {
	Path    string
	Content []byte
}

// Entry is a directory/tree listing item.
type Entry struct {
	Name string
	Path string
	Type string // "file" | "dir"
	Size int64
	SHA  string
}

// Repo is the provisioned repository descriptor.
type Repo struct {
	Namespace     string
	Name          string // org<ID>-<slug>
	Path          string // namespace/name -> stored as git_repo_path
	DefaultBranch string
	HTTPCloneURL  string
}

// ProvisionParams drives create-repo + seed-initial-commit atomically.
type ProvisionParams struct {
	OrgID         int64
	Slug          string
	DefaultBranch string       // "" -> "main"
	CommitMessage string       // seed commit message
	Author        Author       // zero -> platform default
	Seed          []FileChange // initial files; empty repo if nil
}

// Service is the single choke point for platform-owned repo operations.
type Service interface {
	// Namespace returns the domain namespace this instance manages.
	Namespace() string

	// EnsureNamespace makes the namespace org exist (idempotent).
	EnsureNamespace(ctx context.Context) error

	// Provision creates the repo and seeds the initial commit in one shot.
	// On seed-commit failure it deletes the repo and returns the error, so
	// callers never observe a half-created repo.
	Provision(ctx context.Context, p ProvisionParams) (*Repo, error)

	// Commit creates/updates files in one commit. Create-vs-update is decided
	// per path by probing the current SHA.
	Commit(ctx context.Context, repoName, branch, message string, a Author, changes []FileChange) error

	// ReadFile returns decoded file content. Maps 404 -> ErrNotFound.
	ReadFile(ctx context.Context, repoName, branch, path string) ([]byte, *Entry, error)

	// ListDir lists one directory level. Maps 404 -> ErrNotFound.
	ListDir(ctx context.Context, repoName, branch, path string) ([]Entry, error)

	// ListTree enumerates the whole tree recursively (one call). Maps 404 ->
	// ErrNotFound.
	ListTree(ctx context.Context, repoName, ref string) ([]Entry, error)

	// DeleteRepo removes the repo (best-effort cleanup helper for callers).
	DeleteRepo(ctx context.Context, repoName string) error

	// RepoName maps (orgID, slug) -> "org<ID>-<slug>".
	RepoName(orgID int64, slug string) string
	// RepoPath maps (orgID, slug) -> "<namespace>/org<ID>-<slug>".
	RepoPath(orgID int64, slug string) string
	// RepoNameFromPath strips the namespace prefix off a stored git_repo_path.
	RepoNameFromPath(path string) string

	// CloneURL returns the runner-facing HTTPS clone URL (no credentials).
	CloneURL(repoName string) string
}

// defaultAuthor is used for seed/edit commits when the caller passes a zero
// Author.
var defaultAuthor = Author{Name: "Agent Cloud", Email: "platform@agentcloud.local"}
