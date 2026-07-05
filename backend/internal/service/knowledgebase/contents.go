package knowledgebase

import (
	"context"
	"strings"

	"github.com/anthropics/agentsmesh/backend/internal/infra/gitea"
)

type File struct {
	Path    string
	Content string
	Size    int64
}

type DirEntry struct {
	Name string
	Path string
	Type string // file | dir
	Size int64
}

func (s *Service) ReadFile(ctx context.Context, orgID int64, slug, path string) (*File, error) {
	kb, err := s.repo.GetBySlug(ctx, orgID, slug)
	if err != nil {
		return nil, err
	}
	entry, err := s.git.GetFile(ctx, repoNameFromPath(kb.GitRepoPath), kb.DefaultBranch, path)
	if err != nil {
		return nil, err
	}
	content, err := entry.DecodedContent()
	if err != nil {
		return nil, err
	}
	return &File{Path: entry.Path, Content: content, Size: entry.Size}, nil
}

func (s *Service) ListDir(ctx context.Context, orgID int64, slug, path string) ([]*DirEntry, error) {
	kb, err := s.repo.GetBySlug(ctx, orgID, slug)
	if err != nil {
		return nil, err
	}
	entries, err := s.git.ListDir(ctx, repoNameFromPath(kb.GitRepoPath), kb.DefaultBranch, strings.TrimPrefix(path, "/"))
	if err != nil {
		return nil, err
	}
	out := make([]*DirEntry, 0, len(entries))
	for _, e := range entries {
		out = append(out, &DirEntry{Name: e.Name, Path: e.Path, Type: e.Type, Size: e.Size})
	}
	return out, nil
}

// CommitFile writes a single file through the Gitea contents API — used by
// the MCP kb_write path when a pod has no rw mount. Detects create vs update
// by probing the current file SHA.
func (s *Service) CommitFile(
	ctx context.Context, orgID int64, slug, path, content, message, authorName string,
) error {
	kb, err := s.repo.GetBySlug(ctx, orgID, slug)
	if err != nil {
		return err
	}
	repoName := repoNameFromPath(kb.GitRepoPath)
	isUpdate := map[string]string{}
	if existing, err := s.git.GetFile(ctx, repoName, kb.DefaultBranch, path); err == nil {
		isUpdate[path] = existing.SHA
	}
	if authorName == "" {
		authorName = "Do Worker"
	}
	return s.git.CommitFiles(ctx, repoName, kb.DefaultBranch, message,
		gitea.CommitAuthor{Name: authorName, Email: "kb@agentsmesh.local"},
		[]gitea.FileChange{{Path: path, Content: content}}, isUpdate)
}
