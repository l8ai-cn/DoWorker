package expert

import (
	"context"
	"strings"

	expertdom "github.com/l8ai-cn/agentcloud/backend/internal/domain/expert"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

func (s *Service) Delete(ctx context.Context, orgID, id int64) error {
	row, err := s.store.GetByID(ctx, orgID, id)
	if err != nil {
		return err
	}
	if row.IsResourceManaged() {
		return ErrExpertManagedByResourceApply
	}
	if err := s.store.Delete(ctx, orgID, id); err != nil {
		return err
	}
	if s.gitops != nil && row.GitRepoPath != nil {
		repoName := s.gitops.RepoNameFromPath(*row.GitRepoPath)
		if delErr := s.gitops.DeleteRepo(ctx, repoName); delErr != nil {
			s.logger.Warn(
				"expert: repo delete failed",
				"repo",
				repoName,
				"error",
				delErr,
			)
		}
	}
	return nil
}

func (s *Service) GetBySlug(
	ctx context.Context,
	orgID int64,
	slug string,
) (*expertdom.Expert, error) {
	return s.store.GetBySlug(ctx, orgID, slug)
}

func (s *Service) GetByID(
	ctx context.Context,
	orgID, id int64,
) (*expertdom.Expert, error) {
	return s.store.GetByID(ctx, orgID, id)
}

func (s *Service) List(
	ctx context.Context,
	orgID int64,
	limit, offset int,
) ([]expertdom.Expert, int64, error) {
	return s.store.List(ctx, orgID, limit, offset)
}

func (s *Service) resolveSlug(
	ctx context.Context,
	orgID int64,
	explicit, nameSeed string,
	excludeID int64,
) (string, error) {
	seed := strings.TrimSpace(explicit)
	if seed == "" {
		seed = nameSeed
	}
	return slugkit.GenerateUnique(
		ctx,
		seed,
		slugkit.FromExistsCheck(
			func(ctx context.Context, candidate string) (bool, error) {
				return s.store.SlugExists(
					ctx,
					orgID,
					candidate,
					excludeID,
				)
			},
		),
	)
}
