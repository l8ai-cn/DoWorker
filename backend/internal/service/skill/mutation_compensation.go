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
