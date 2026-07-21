package skill

import (
	"context"

	"github.com/l8ai-cn/agentcloud/backend/internal/service/gitops"
)

// ErrFileNotFound mirrors gitops.ErrNotFound so REST callers can map content
// misses to 404 without importing gitops.
var ErrFileNotFound = gitops.ErrNotFound

// ReadSkillFile returns the decoded content + entry of a file in the authored
// skill's repo. relPath must already be sanitized by the caller. Returns the
// underlying repository error (skilldom.ErrNotFound) when the skill is unknown,
// gitops.ErrNotFound on a content miss.
func (s *Service) ReadSkillFile(
	ctx context.Context, orgID int64, slug, relPath string,
) ([]byte, *gitops.Entry, error) {
	row, err := s.store.GetBySlug(ctx, orgID, slug)
	if err != nil {
		return nil, nil, err
	}
	repoName := s.gitops.RepoNameFromPath(row.GitRepoPath)
	return s.gitops.ReadFile(ctx, repoName, branchOrDefault(row.DefaultBranch), relPath)
}

// ListSkillTree enumerates the authored skill's repo tree.
func (s *Service) ListSkillTree(
	ctx context.Context, orgID int64, slug string,
) ([]gitops.Entry, error) {
	row, err := s.store.GetBySlug(ctx, orgID, slug)
	if err != nil {
		return nil, err
	}
	repoName := s.gitops.RepoNameFromPath(row.GitRepoPath)
	return s.gitops.ListTree(ctx, repoName, branchOrDefault(row.DefaultBranch))
}
