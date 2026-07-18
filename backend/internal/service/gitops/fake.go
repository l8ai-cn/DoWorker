package gitops

import (
	"context"
	"errors"
	"path"
	"strings"
)

// Fake is an in-memory Service implementation for unit-testing downstream
// consumers (expert/skill services) without a live Gitea or any network.
type Fake struct {
	NS    string
	Repos map[string]*fakeRepo // repoName -> repo

	// Failure injections for exercising error paths.
	FailProvision bool
	FailCommit    bool

	CloneBaseURL string // "" -> "https://gitea.local"
}

type fakeRepo struct {
	Branch string
	Files  map[string][]byte // path -> content
	SHAs   map[string]string // path -> pseudo-SHA
}

// NewFake returns an empty in-memory gitops service bound to ns.
func NewFake(ns string) *Fake {
	return &Fake{NS: ns, Repos: map[string]*fakeRepo{}}
}

var errFakeInjected = errors.New("gitops.Fake: injected failure")

func (f *Fake) Namespace() string { return f.NS }

func (f *Fake) EnsureNamespace(context.Context) error { return nil }

func (f *Fake) Provision(_ context.Context, p ProvisionParams) (*Repo, error) {
	if f.FailProvision {
		return nil, errFakeInjected
	}
	branch := p.DefaultBranch
	if branch == "" {
		branch = "main"
	}
	name := repoName(p.OrgID, p.Slug)
	r := &fakeRepo{Branch: branch, Files: map[string][]byte{}, SHAs: map[string]string{}}
	for _, ch := range p.Seed {
		r.put(ch.Path, ch.Content)
	}
	f.Repos[name] = r
	return &Repo{
		Namespace:     f.NS,
		Name:          name,
		Path:          f.NS + "/" + name,
		DefaultBranch: branch,
		HTTPCloneURL:  f.CloneURL(name),
	}, nil
}

func (f *Fake) Commit(
	_ context.Context, repoName, _ /*branch*/, _ /*message*/ string, _ Author, changes []FileChange,
) error {
	if f.FailCommit {
		return errFakeInjected
	}
	r, ok := f.Repos[repoName]
	if !ok {
		return ErrNotFound
	}
	for _, ch := range changes {
		r.put(ch.Path, ch.Content)
	}
	return nil
}

func (f *Fake) ReadFile(
	_ context.Context, repoName, _ /*branch*/, filePath string,
) ([]byte, *Entry, error) {
	r, ok := f.Repos[repoName]
	if !ok {
		return nil, nil, ErrNotFound
	}
	content, ok := r.Files[filePath]
	if !ok {
		return nil, nil, ErrNotFound
	}
	out := make([]byte, len(content))
	copy(out, content)
	return out, &Entry{
		Name: path.Base(filePath),
		Path: filePath,
		Type: "file",
		Size: int64(len(content)),
		SHA:  r.SHAs[filePath],
	}, nil
}

func (f *Fake) ListDir(
	_ context.Context, repoName, _ /*branch*/, dir string,
) ([]Entry, error) {
	r, ok := f.Repos[repoName]
	if !ok {
		return nil, ErrNotFound
	}
	dir = strings.Trim(dir, "/")
	seen := map[string]Entry{}
	for p, content := range r.Files {
		rest, ok := underDir(p, dir)
		if !ok {
			continue
		}
		if i := strings.IndexByte(rest, '/'); i >= 0 {
			// Nested entry -> surface the immediate subdirectory once.
			name := rest[:i]
			childPath := name
			if dir != "" {
				childPath = dir + "/" + name
			}
			seen[name] = Entry{Name: name, Path: childPath, Type: "dir"}
			continue
		}
		seen[rest] = Entry{
			Name: rest,
			Path: p,
			Type: "file",
			Size: int64(len(content)),
			SHA:  r.SHAs[p],
		}
	}
	return sortedEntries(seen), nil
}

func (f *Fake) ListTree(_ context.Context, repoName, _ /*ref*/ string) ([]Entry, error) {
	r, ok := f.Repos[repoName]
	if !ok {
		return nil, ErrNotFound
	}
	seen := map[string]Entry{}
	for p, content := range r.Files {
		seen[p] = Entry{
			Name: path.Base(p),
			Path: p,
			Type: "file",
			Size: int64(len(content)),
			SHA:  r.SHAs[p],
		}
		// Synthesize the intermediate directory entries.
		for dir := path.Dir(p); dir != "." && dir != "/"; dir = path.Dir(dir) {
			if _, exists := seen[dir]; !exists {
				seen[dir] = Entry{Name: path.Base(dir), Path: dir, Type: "dir"}
			}
		}
	}
	return sortedEntries(seen), nil
}

func (f *Fake) DeleteRepo(_ context.Context, repoName string) error {
	delete(f.Repos, repoName)
	return nil
}

func (f *Fake) RepoName(orgID int64, slug string) string { return repoName(orgID, slug) }
func (f *Fake) RepoPath(orgID int64, slug string) string { return repoPath(f.NS, orgID, slug) }
func (f *Fake) RepoNameFromPath(p string) string         { return repoNameFromPath(p) }

func (f *Fake) CloneURL(repoName string) string {
	base := f.CloneBaseURL
	if base == "" {
		base = "https://gitea.local"
	}
	return strings.TrimRight(base, "/") + "/" + f.NS + "/" + repoName + ".git"
}

// Compile-time assertion that Fake satisfies the Service interface.
var _ Service = (*Fake)(nil)
