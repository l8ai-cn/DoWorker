package expert

import (
	"context"
	"errors"
	"strings"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
)

// ErrGitBackingDisabled is returned by the repo-content read methods when the
// expert has no backing repo (gitops disabled or a legacy row not yet
// provisioned).
var ErrGitBackingDisabled = errors.New("expert: git backing not available")

// ErrFileNotFound mirrors gitops.ErrNotFound so REST callers can map content
// misses to 404 without importing gitops.
var ErrFileNotFound = gitops.ErrNotFound

// provisionExpertRepo creates the expert repo and seeds it. On success the
// caller stores the returned repo descriptor on the DB row (cache).
func (s *Service) provisionExpertRepo(
	ctx context.Context, e *expertdom.Expert, layer string, avatar *AvatarInput,
) (*gitops.Repo, error) {
	seed, err := s.renderExpertFiles(e, layer, avatar, true)
	if err != nil {
		return nil, err
	}
	return s.gitops.Provision(ctx, gitops.ProvisionParams{
		OrgID:         e.OrganizationID,
		Slug:          e.Slug,
		CommitMessage: "init: expert scaffold (agent.md, expert.json, README.md)",
		Seed:          seed,
	})
}

// applyRepo copies a provisioned repo descriptor onto the DB row cache columns.
func applyRepo(e *expertdom.Expert, repo *gitops.Repo) {
	path := repo.Path
	e.GitRepoPath = &path
	e.DefaultBranch = repo.DefaultBranch
	if e.DefaultBranch == "" {
		e.DefaultBranch = "main"
	}
	url := repo.HTTPCloneURL
	e.HTTPCloneURL = &url
}

// ensureExpertRepo lazily provisions a repo for a legacy row (GitRepoPath nil)
// on its next update/run. No-op when gitops is disabled or the row is already
// backed. Returns whether a repo was newly provisioned.
func (s *Service) ensureExpertRepo(
	ctx context.Context, e *expertdom.Expert, layer string, avatar *AvatarInput,
) (bool, error) {
	if s.gitops == nil || e.GitRepoPath != nil {
		return false, nil
	}
	repo, err := s.provisionExpertRepo(ctx, e, layer, avatar)
	if err != nil {
		return false, err
	}
	applyRepo(e, repo)
	return true, nil
}

// commitExpertChanges re-renders agent.md/expert.json (+ avatar) before the
// database cache update.
func (s *Service) commitExpertChanges(
	ctx context.Context, e *expertdom.Expert, layer string, avatar *AvatarInput,
) error {
	if s.gitops == nil || e.GitRepoPath == nil {
		return nil
	}
	changes, err := s.renderExpertFiles(e, layer, avatar, false)
	if err != nil {
		return err
	}
	repoName := s.gitops.RepoNameFromPath(*e.GitRepoPath)
	return s.gitops.Commit(ctx, repoName, s.branchOf(e), "update: expert configuration", gitops.Author{}, changes)
}

// GitEnabled reports whether git-backing is configured for this service.
func (s *Service) GitEnabled() bool { return s.gitops != nil }

// ReadExpertFile returns the decoded content + entry of a file in the expert's
// repo. relPath must already be sanitized by the caller. Returns
// ErrGitBackingDisabled when the expert has no repo, ErrFileNotFound on miss.
func (s *Service) ReadExpertFile(
	ctx context.Context, orgID int64, slug, relPath string,
) ([]byte, *gitops.Entry, error) {
	e, err := s.store.GetBySlug(ctx, orgID, slug)
	if err != nil {
		return nil, nil, err
	}
	if s.gitops == nil || e.GitRepoPath == nil {
		return nil, nil, ErrGitBackingDisabled
	}
	repoName := s.gitops.RepoNameFromPath(*e.GitRepoPath)
	return s.gitops.ReadFile(ctx, repoName, s.branchOf(e), relPath)
}

// ListExpertTree enumerates the expert repo tree. Returns
// ErrGitBackingDisabled when the expert has no repo.
func (s *Service) ListExpertTree(
	ctx context.Context, orgID int64, slug string,
) ([]gitops.Entry, error) {
	e, err := s.store.GetBySlug(ctx, orgID, slug)
	if err != nil {
		return nil, err
	}
	if s.gitops == nil || e.GitRepoPath == nil {
		return nil, ErrGitBackingDisabled
	}
	repoName := s.gitops.RepoNameFromPath(*e.GitRepoPath)
	return s.gitops.ListTree(ctx, repoName, s.branchOf(e))
}

func (s *Service) branchOf(e *expertdom.Expert) string {
	if strings.TrimSpace(e.DefaultBranch) != "" {
		return e.DefaultBranch
	}
	return "main"
}
