package goalloop

import (
	"context"
	"time"
)

type ListFilter struct {
	OrganizationID int64
	Status         string
	Query          string
	Limit          int
	Offset         int
}

type Repository interface {
	Create(ctx context.Context, loop *GoalLoop) error
	GetBySlug(ctx context.Context, organizationID int64, slug string) (*GoalLoop, error)
	GetByPodKey(ctx context.Context, podKey string) (*GoalLoop, error)
	GetByAutopilotControllerKey(ctx context.Context, autopilotKey string) (*GoalLoop, error)
	GetByVerificationRequestID(ctx context.Context, requestID string) (*GoalLoop, error)
	ListTimedOut(ctx context.Context, now time.Time) ([]*GoalLoop, error)
	List(ctx context.Context, filter ListFilter) ([]*GoalLoop, int64, error)
	ExistsSlug(ctx context.Context, organizationID int64, slug string) (bool, error)
	Update(ctx context.Context, id int64, updates map[string]any) error
}
