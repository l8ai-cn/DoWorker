package expert

import (
	"context"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
)

func (s *Service) RebindWorkerSpecSnapshot(
	ctx context.Context,
	organizationID int64,
	expertID int64,
	snapshotID int64,
) (*expertdom.Expert, error) {
	row, err := s.store.GetByID(ctx, organizationID, expertID)
	if err != nil {
		return nil, err
	}
	if row.IsResourceManaged() {
		return nil, ErrExpertManagedByResourceApply
	}
	row.WorkerSpecSnapshotID = &snapshotID
	if err := s.store.Update(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}
