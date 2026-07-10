package workerspec

import "context"

type Repository interface {
	Create(ctx context.Context, snapshot *Snapshot) error
	GetByID(ctx context.Context, organizationID, id int64) (*Snapshot, error)
}
