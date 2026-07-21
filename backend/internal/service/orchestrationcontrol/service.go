package orchestrationcontrol

import (
	"fmt"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
)

type Service struct {
	registry          *orchestrationresource.Registry
	repository        Repository
	authorizer        Authorizer
	references        ReferenceResolver
	workerDefinitions WorkerDefinitionPolicyResolver
	planners          map[orchestrationresource.TypeMeta]TargetPlanner
	clock             func() time.Time
	idGenerator       func() string
	planTTL           time.Duration
}

func NewService(deps ServiceDeps) (*Service, error) {
	if deps.Registry == nil || deps.Repository == nil || deps.Authorizer == nil ||
		deps.References == nil || deps.WorkerDefinitions == nil ||
		deps.Clock == nil || deps.IDGenerator == nil ||
		deps.PlanTTL <= 0 || len(deps.RequiredTypes) == 0 {
		return nil, fmt.Errorf("%w: incomplete dependencies", ErrUnavailable)
	}
	planners := make(map[orchestrationresource.TypeMeta]TargetPlanner, len(deps.Planners))
	for _, planner := range deps.Planners {
		if planner == nil {
			return nil, fmt.Errorf("%w: nil target planner", ErrUnavailable)
		}
		meta := planner.TypeMeta()
		if err := meta.Validate(); err != nil {
			return nil, fmt.Errorf("%w: invalid target planner", ErrUnavailable)
		}
		if _, exists := planners[meta]; exists {
			return nil, fmt.Errorf("%w: duplicate target planner", ErrUnavailable)
		}
		planners[meta] = planner
	}
	for _, meta := range deps.RequiredTypes {
		if err := meta.Validate(); err != nil {
			return nil, fmt.Errorf("%w: invalid required target", ErrUnavailable)
		}
		if planners[meta] == nil {
			return nil, fmt.Errorf("%w: missing target planner", ErrUnavailable)
		}
		if !deps.Registry.Has(meta) {
			return nil, fmt.Errorf("%w: missing target schema", ErrUnavailable)
		}
	}
	return &Service{
		registry: deps.Registry, repository: deps.Repository,
		authorizer: deps.Authorizer, references: deps.References,
		workerDefinitions: deps.WorkerDefinitions,
		planners:          planners, clock: deps.Clock,
		idGenerator: deps.IDGenerator, planTTL: deps.PlanTTL,
	}, nil
}
