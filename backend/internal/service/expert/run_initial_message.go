package expert

import (
	"context"
	"errors"
	"strings"

	sessiondomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentsession"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	itemservice "github.com/l8ai-cn/agentcloud/backend/internal/service/conversationitem"
)

func (s *Service) prepareRunInitialMessage(
	ctx context.Context,
	organizationID int64,
	snapshotID int64,
	promptOverride *string,
) (func(context.Context, *sessiondomain.Session) error, error) {
	text, err := s.resolveRunInitialMessage(
		ctx,
		organizationID,
		snapshotID,
		promptOverride,
	)
	if err != nil {
		return nil, err
	}
	if text == "" {
		return nil, nil
	}
	return func(ctx context.Context, session *sessiondomain.Session) error {
		if session == nil {
			return errors.New("session is required")
		}
		return itemservice.AppendUserText(ctx, s.items, session.ID, text)
	}, nil
}

func (s *Service) resolveRunInitialMessage(
	ctx context.Context,
	organizationID int64,
	snapshotID int64,
	promptOverride *string,
) (string, error) {
	if s.workerSpecs == nil {
		return "", ErrWorkerSpecSnapshotUnavailable
	}
	snapshot, err := s.workerSpecs.GetByID(ctx, organizationID, snapshotID)
	if err != nil {
		if errors.Is(err, specdomain.ErrNotFound) {
			return "", ErrWorkerSpecSnapshotMismatch
		}
		return "", errors.Join(ErrWorkerSpecSnapshotUnavailable, err)
	}
	if snapshot.ID != snapshotID || snapshot.OrganizationID != organizationID {
		return "", ErrWorkerSpecSnapshotMismatch
	}
	spec, err := specdomain.NormalizeAndValidate(snapshot.Spec)
	if err != nil {
		return "", errors.Join(ErrWorkerSpecSnapshotMismatch, err)
	}
	if promptOverride != nil && strings.TrimSpace(*promptOverride) != "" {
		return strings.TrimSpace(*promptOverride), nil
	}
	return spec.Workspace.InitialTask, nil
}
