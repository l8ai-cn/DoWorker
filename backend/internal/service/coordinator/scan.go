package coordinator

import (
	"context"
	"errors"
	"fmt"

	coordinatordom "github.com/anthropics/agentsmesh/backend/internal/domain/coordinator"
)

type RunResult struct {
	ProjectID  int64
	Scanned    int
	Candidates int
	Claimed    int
	Dispatched int
	Skipped    int
	Errors     []string
}

var ErrCoordinatorWorkerSpecSnapshotStoreRequired = errors.New(
	"coordinator: worker spec snapshot store is required",
)

// RunProject is one coordinator tick for a project: discover external tasks,
// filter by claim policy, then claim+dispatch candidates up to the concurrency
// budget.
func (s *Service) RunProject(ctx context.Context, project *coordinatordom.Project) (*RunResult, error) {
	result := &RunResult{ProjectID: project.ID}

	if s.runnerEnsurer != nil {
		workerType, err := s.projectSnapshotWorkerType(ctx, project)
		if err != nil {
			return result, err
		}
		if err := s.runnerEnsurer.Ensure(ctx, project.OrganizationID, project.CreatedByID, workerType); err != nil {
			return result, fmt.Errorf("ensure runner: %w", err)
		}
	}

	platform, repo, err := s.platform.For(ctx, project)
	if err != nil {
		return result, fmt.Errorf("resolve platform: %w", err)
	}

	policy := project.DecodeClaimPolicy()
	tasks, err := platform.DiscoverTasks(ctx, repo, policy)
	if err != nil {
		return result, fmt.Errorf("discover tasks: %w", err)
	}
	result.Scanned = len(tasks)

	budget, err := s.dispatchBudget(ctx, project)
	if err != nil {
		return result, err
	}

	for i := range tasks {
		task := tasks[i]
		if matched, _ := policy.Matches(task.Candidate()); !matched {
			result.Skipped++
			continue
		}
		result.Candidates++
		if budget <= 0 {
			result.Skipped++
			continue
		}
		dispatched, err := s.claimAndDispatch(ctx, project, platform, repo, task)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", task.ExternalID, err))
			continue
		}
		if dispatched {
			result.Claimed++
			result.Dispatched++
			budget--
		} else {
			result.Skipped++
		}
	}
	return result, nil
}

func (s *Service) projectSnapshotWorkerType(
	ctx context.Context,
	project *coordinatordom.Project,
) (string, error) {
	snapshotID, err := coordinatorSnapshotID(project)
	if err != nil {
		return "", err
	}
	if s.snapshots == nil {
		return "", ErrCoordinatorWorkerSpecSnapshotStoreRequired
	}
	snapshot, err := s.snapshots.GetByID(ctx, project.OrganizationID, *snapshotID)
	if err != nil {
		return "", fmt.Errorf("load coordinator workerspec snapshot: %w", err)
	}
	return snapshot.Spec.Runtime.WorkerType.Slug.String(), nil
}

// dispatchBudget is the number of new dispatches allowed this tick, bounded by
// the project's max_concurrent minus currently-active executions.
func (s *Service) dispatchBudget(ctx context.Context, project *coordinatordom.Project) (int, error) {
	max := project.MaxConcurrent
	if max <= 0 {
		max = 1
	}
	active, err := s.store.CountActiveExecutions(ctx, project.ID)
	if err != nil {
		return 0, fmt.Errorf("count active executions: %w", err)
	}
	budget := max - int(active)
	if budget < 0 {
		budget = 0
	}
	return budget, nil
}
