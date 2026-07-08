package agentpod

import (
	"context"
	"time"
)

type PendingCommandRepository interface {
	Enqueue(ctx context.Context, cmd *PendingCommand) error
	CountByRunner(ctx context.Context, runnerID int64) (int, error)
	ListByRunnerFIFO(ctx context.Context, runnerID int64, limit int) ([]*PendingCommand, error)
	Delete(ctx context.Context, id int64) error
	DeleteByPodKey(ctx context.Context, podKey string) (int64, error)
	ListExpired(ctx context.Context, now time.Time, limit int) ([]*PendingCommand, error)
	ListRunnerIDsWithPending(ctx context.Context, limit int) ([]int64, error)
	PositionByPodKey(ctx context.Context, runnerID int64, podKey string) (int, error)
	GetCreatePodByPodKey(ctx context.Context, podKey string) (*PendingCommand, error)
}
