package workerruntime

import (
	"context"
)

type Repository interface {
	ResolveSelection(
		ctx context.Context,
		request Request,
	) (*RepositorySelection, error)
}
