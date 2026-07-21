package knowledgebase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	kbdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/knowledgebase"
)

func (s *Service) failCreateAndCleanupRepo(ctx context.Context, repoName string, cause error) error {
	if cleanupErr := s.git.DeleteRepo(ctx, repoName); cleanupErr != nil {
		return errors.Join(cause, fmt.Errorf("knowledgebase: cleanup repository: %w", cleanupErr))
	}
	return cause
}

func (s *Service) Delete(ctx context.Context, orgID, id int64) error {
	kb, err := s.repo.Get(ctx, orgID, id)
	if err != nil {
		return err
	}
	if repoName := repoNameFromPath(kb.GitRepoPath); repoName != "" {
		keys, err := s.git.ListDeployKeys(ctx, repoName)
		if err != nil {
			return fmt.Errorf("knowledgebase: list deploy keys: %w", err)
		}
		for _, key := range keys {
			if err := s.git.DeleteDeployKey(ctx, repoName, key.ID); err != nil {
				return fmt.Errorf("knowledgebase: revoke deploy key %d: %w", key.ID, err)
			}
		}
		if err := s.git.DeleteRepo(ctx, repoName); err != nil {
			return fmt.Errorf("knowledgebase: delete repository: %w", err)
		}
	}
	if err := s.repo.Delete(ctx, orgID, id); err != nil {
		syncError := "repository deleted; retry knowledge base deletion"
		disableErr := s.repo.Update(ctx, orgID, id, map[string]any{
			"git_repo_path":  "",
			"http_clone_url": "",
			"sync_status":    kbdomain.SyncStatusFailed,
			"sync_error":     syncError,
		})
		if disableErr != nil {
			return errors.Join(
				fmt.Errorf("knowledgebase: delete database record: %w", err),
				fmt.Errorf("knowledgebase: disable deleted repository record: %w", disableErr),
			)
		}
		return fmt.Errorf("knowledgebase: delete database record: %w", err)
	}
	return nil
}

func repoNameFromPath(path string) string {
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[i+1:]
	}
	return path
}
