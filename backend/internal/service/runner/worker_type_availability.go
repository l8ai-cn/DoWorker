package runner

import (
	"context"
	"errors"
)

func (s *Service) HasAvailableRunnerForAgent(
	ctx context.Context,
	orgID int64,
	userID int64,
	agentSlug string,
) (bool, error) {
	_, err := s.SelectAvailableRunnerForAgent(ctx, orgID, userID, agentSlug)
	if errors.Is(err, ErrNoRunnerForAgent) {
		return false, nil
	}
	return err == nil, err
}
