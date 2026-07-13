package skill

import (
	"context"
	"fmt"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
)

func (s *Service) Delete(ctx context.Context, orgID, id int64) error {
	row, err := s.store.GetByID(ctx, orgID, id)
	if err != nil {
		return err
	}
	if err := s.store.Delete(ctx, orgID, id); err != nil {
		return err
	}
	if row.GitRepoPath != "" {
		repoName := s.gitops.RepoNameFromPath(row.GitRepoPath)
		if delErr := s.gitops.DeleteRepo(ctx, repoName); delErr != nil {
			s.logger.Warn("skill: repo delete failed", "repo", repoName, "error", delErr)
		}
	}
	return nil
}

func (s *Service) Get(ctx context.Context, orgID int64, slug string) (*skilldom.Skill, error) {
	return s.store.GetBySlug(ctx, orgID, slug)
}

func (s *Service) List(ctx context.Context, orgID int64, limit, offset int) ([]skilldom.Skill, int64, error) {
	return s.store.List(ctx, orgID, limit, offset)
}

func (s *Service) packageFromGit(ctx context.Context, repoName, branch string) (*packagedSkill, error) {
	dir, cleanup, err := materializeRepo(ctx, s.gitops, repoName, branch)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	pkg, err := s.packager.PackageFromDir(ctx, dir)
	if err != nil {
		return nil, fmt.Errorf("skill: package: %w", err)
	}
	return &packagedSkill{
		ContentSha:  pkg.ContentSha,
		StorageKey:  pkg.StorageKey,
		PackageSize: pkg.PackageSize,
		Created:     pkg.Created,
	}, nil
}
