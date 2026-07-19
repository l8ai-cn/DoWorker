package operatorcatalog

import (
	"context"

	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
)

type DependencyArtifactStore interface {
	Create(context.Context, int64, int64, []byte, string) error
	Delete(context.Context, int64, int64) error
	GetBySnapshotID(context.Context, int64, int64) (workerdependency.Document, error)
}
