package skill

import (
	"context"
	"errors"
	"fmt"

	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
)

func (s *Service) restoreMutation(
	ctx context.Context,
	repoName, branch string,
	snapshot *gitops.TreeSnapshot,
	cause error,
) error {
	restoreCtx := context.WithoutCancel(ctx)
	if err := gitops.RestoreTree(restoreCtx, s.gitops, repoName, branch, snapshot); err != nil {
		restoreErr := fmt.Errorf("skill: restore Git after failed mutation: %w", err)
		if cause == nil {
			return restoreErr
		}
		return errors.Join(cause, restoreErr)
	}
	return cause
}

func (s *Service) compensatePackagedMutation(
	ctx context.Context,
	repoName, branch string,
	snapshot *gitops.TreeSnapshot,
	previousStorageKey string,
	pkg *packagedSkill,
	cause error,
) error {
	compensationErr := s.restoreMutation(
		ctx, repoName, branch, snapshot, cause,
	)
	if pkg == nil || pkg.StorageKey == "" || pkg.StorageKey == previousStorageKey {
		return compensationErr
	}
	cleanupCtx := context.WithoutCancel(ctx)
	if err := s.packager.DeletePackage(cleanupCtx, pkg.StorageKey); err != nil {
		return errors.Join(
			compensationErr,
			fmt.Errorf("skill: delete unreferenced package: %w", err),
		)
	}
	return compensationErr
}
