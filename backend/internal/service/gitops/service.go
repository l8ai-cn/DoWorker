package gitops

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/anthropics/agentsmesh/backend/internal/infra/gitea"
)

type service struct {
	git *gitea.Client
	ns  string
	log *slog.Logger
}

// Compile-time assertion that the gitea-backed service satisfies Service.
var _ Service = (*service)(nil)

// NewService wraps a gitea client bound to its configured namespace. It returns
// nil when git is nil, matching the "feature disabled" convention used by the
// knowledgebase service (callers treat a nil Service as disabled).
func NewService(git *gitea.Client, log *slog.Logger) Service {
	if git == nil {
		return nil
	}
	if log == nil {
		log = slog.Default()
	}
	ns := git.Namespace()
	return &service{
		git: git,
		ns:  ns,
		log: log.With("component", "gitops", "namespace", ns),
	}
}

func (s *service) Namespace() string { return s.ns }

func (s *service) EnsureNamespace(ctx context.Context) error {
	if err := s.git.EnsureNamespace(ctx); err != nil {
		return fmt.Errorf("gitops: ensure namespace: %w", err)
	}
	return nil
}

func (s *service) Provision(ctx context.Context, p ProvisionParams) (*Repo, error) {
	branch := p.DefaultBranch
	if branch == "" {
		branch = "main"
	}
	if err := s.EnsureNamespace(ctx); err != nil {
		return nil, err
	}
	name := repoName(p.OrgID, p.Slug)
	repo, err := s.git.CreateRepo(ctx, name, branch)
	if err != nil {
		return nil, fmt.Errorf("gitops: create repo: %w", err)
	}

	if len(p.Seed) > 0 {
		msg := p.CommitMessage
		if msg == "" {
			msg = "init: repository scaffold"
		}
		author := p.Author
		if author == (Author{}) {
			author = defaultAuthor
		}
		if err := s.git.CommitFiles(ctx, name, branch, msg,
			giteaAuthor(author), toGiteaChanges(p.Seed), nil); err != nil {
			// Compensating cleanup: never leave a half-created repo behind.
			if delErr := s.git.DeleteRepo(ctx, name); delErr != nil {
				s.log.Warn("gitops: seed cleanup delete failed", "repo", name, "error", delErr)
			}
			return nil, fmt.Errorf("gitops: seed commit: %w", err)
		}
	}

	defaultBranch := repo.DefaultBranch
	if defaultBranch == "" {
		defaultBranch = branch
	}
	return &Repo{
		Namespace:     s.ns,
		Name:          name,
		Path:          s.ns + "/" + name,
		DefaultBranch: defaultBranch,
		HTTPCloneURL:  s.git.CloneURL(name),
	}, nil
}

func (s *service) Commit(
	ctx context.Context, repoName, branch, message string, a Author, changes []FileChange,
) error {
	if a == (Author{}) {
		a = defaultAuthor
	}
	// Probe existing SHAs to switch create -> update per path.
	isUpdate := map[string]string{}
	for _, ch := range changes {
		existing, err := s.git.GetFile(ctx, repoName, branch, ch.Path)
		if err != nil {
			if isNotFound(err) {
				continue
			}
			return fmt.Errorf("gitops: probe %s: %w", ch.Path, err)
		}
		isUpdate[ch.Path] = existing.SHA
	}
	if err := s.git.CommitFiles(ctx, repoName, branch, message,
		giteaAuthor(a), toGiteaChanges(changes), isUpdate); err != nil {
		return fmt.Errorf("gitops: commit: %w", err)
	}
	return nil
}

func (s *service) ReadFile(
	ctx context.Context, repoName, branch, path string,
) ([]byte, *Entry, error) {
	entry, err := s.git.GetFile(ctx, repoName, branch, path)
	if err != nil {
		return nil, nil, mapNotFound(err)
	}
	content, err := entry.DecodedContent()
	if err != nil {
		return nil, nil, fmt.Errorf("gitops: decode %s: %w", path, err)
	}
	return []byte(content), &Entry{
		Name: entry.Name,
		Path: entry.Path,
		Type: entry.Type,
		Size: entry.Size,
		SHA:  entry.SHA,
	}, nil
}

func (s *service) ListDir(
	ctx context.Context, repoName, branch, path string,
) ([]Entry, error) {
	entries, err := s.git.ListDir(ctx, repoName, branch, path)
	if err != nil {
		return nil, mapNotFound(err)
	}
	out := make([]Entry, 0, len(entries))
	for _, e := range entries {
		out = append(out, Entry{
			Name: e.Name,
			Path: e.Path,
			Type: e.Type, // gitea contents API already returns "file"/"dir"
			Size: e.Size,
			SHA:  e.SHA,
		})
	}
	return out, nil
}

func (s *service) ListTree(ctx context.Context, repoName, ref string) ([]Entry, error) {
	entries, err := s.git.ListTree(ctx, repoName, ref)
	if err != nil {
		return nil, mapNotFound(err)
	}
	out := make([]Entry, 0, len(entries))
	for _, e := range entries {
		out = append(out, Entry{
			Name: baseName(e.Path),
			Path: e.Path,
			Type: treeType(e.Type),
			Size: e.Size,
			SHA:  e.SHA,
		})
	}
	return out, nil
}

func (s *service) DeleteRepo(ctx context.Context, repoName string) error {
	if err := s.git.DeleteRepo(ctx, repoName); err != nil {
		return fmt.Errorf("gitops: delete repo: %w", err)
	}
	return nil
}

func (s *service) RepoName(orgID int64, slug string) string { return repoName(orgID, slug) }
func (s *service) RepoPath(orgID int64, slug string) string { return repoPath(s.ns, orgID, slug) }
func (s *service) RepoNameFromPath(path string) string      { return repoNameFromPath(path) }

func (s *service) CloneURL(repoName string) string { return s.git.CloneURL(repoName) }
func (s *service) CloneToken() string              { return s.git.CloneToken() }

func giteaAuthor(a Author) gitea.CommitAuthor {
	return gitea.CommitAuthor{Name: a.Name, Email: a.Email}
}

func toGiteaChanges(changes []FileChange) []gitea.FileChange {
	out := make([]gitea.FileChange, 0, len(changes))
	for _, ch := range changes {
		out = append(out, gitea.FileChange{Path: ch.Path, Content: string(ch.Content)})
	}
	return out
}

// treeType normalizes git tree object types ("blob"/"tree") to the
// file/dir vocabulary the contents API and gitops.Entry use.
func treeType(t string) string {
	switch t {
	case "blob":
		return "file"
	case "tree":
		return "dir"
	default:
		return t
	}
}

func baseName(path string) string {
	return repoNameFromPath(path)
}

// isNotFound reports whether err is a gitea 404.
func isNotFound(err error) bool {
	var he *gitea.HTTPError
	return errors.As(err, &he) && he.StatusCode == http.StatusNotFound
}

// mapNotFound translates a gitea 404 into ErrNotFound; other errors are
// wrapped with a domain prefix.
func mapNotFound(err error) error {
	if isNotFound(err) {
		return ErrNotFound
	}
	return fmt.Errorf("gitops: %w", err)
}
