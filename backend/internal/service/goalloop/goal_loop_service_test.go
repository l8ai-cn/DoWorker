package goalloop

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/goalloop"
	workerspecdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	workerspecsvc "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func TestCreateRejectsUnavailableWorkerSnapshot(t *testing.T) {
	repo := newGoalLoopTestRepo()
	service := NewService(repo)
	service.SetWorkerSpecSnapshotLoader(goalLoopSnapshotLoader{err: workerspecdomain.ErrNotFound})

	_, err := service.Create(context.Background(), validCreateRequest())

	require.ErrorIs(t, err, ErrInvalidInput)
	require.Empty(t, repo.loops)
}

func TestCreateRejectsLegacyProtocolAdapterSnapshot(t *testing.T) {
	repo := newGoalLoopTestRepo()
	service := NewService(repo)
	service.SetWorkerSpecSnapshotLoader(goalLoopSnapshotLoader{
		snapshot: workerspecdomain.Snapshot{
			ID:             3,
			OrganizationID: 1,
			Spec: workerspecdomain.Spec{
				Runtime: workerspecdomain.Runtime{
					ModelBinding: workerspecdomain.ModelBinding{ResourceID: 1},
				},
			},
		},
	})

	_, err := service.Create(context.Background(), validCreateRequest())

	require.ErrorIs(t, err, ErrInvalidInput)
	require.Empty(t, repo.loops)
}

func TestListWorkerSnapshotsUsesOrganizationScopedSnapshotStore(t *testing.T) {
	service := NewService(newGoalLoopTestRepo())
	service.SetWorkerSpecSnapshotLoader(goalLoopSnapshotLoader{
		snapshots: []workerspecdomain.Snapshot{
			{ID: 4, OrganizationID: 1},
			{ID: 3, OrganizationID: 1},
		},
	})
	service.SetWorkerTypeSnapshotValidator(&goalLoopWorkerTypeValidator{
		errs: []error{workercreation.ErrWorkerTypeDefinitionChanged, nil},
	})

	snapshots, err := service.ListWorkerSnapshots(context.Background(), 1, 2)

	require.NoError(t, err)
	require.Len(t, snapshots, 1)
	require.Equal(t, int64(3), snapshots[0].ID)
}

func TestListWorkerSnapshotsReturnsValidatorFailure(t *testing.T) {
	service := NewService(newGoalLoopTestRepo())
	service.SetWorkerSpecSnapshotLoader(goalLoopSnapshotLoader{
		snapshots: []workerspecdomain.Snapshot{{ID: 3, OrganizationID: 1}},
	})
	expected := errors.New("worker type catalog unavailable")
	service.SetWorkerTypeSnapshotValidator(&goalLoopWorkerTypeValidator{
		errs: []error{expected},
	})

	_, err := service.ListWorkerSnapshots(context.Background(), 1, 2)

	require.ErrorIs(t, err, expected)
}

func TestListWorkerSnapshotsExcludesLegacyProtocolAdapterSnapshot(t *testing.T) {
	service := NewService(newGoalLoopTestRepo())
	legacy := workerspecdomain.Snapshot{
		ID: 4,
		Spec: workerspecdomain.Spec{
			Runtime: workerspecdomain.Runtime{
				ModelBinding: workerspecdomain.ModelBinding{ResourceID: 1},
			},
		},
	}
	available := workerspecdomain.Snapshot{ID: 3}
	service.SetWorkerSpecSnapshotLoader(goalLoopSnapshotLoader{
		snapshots: []workerspecdomain.Snapshot{legacy, available},
	})
	service.SetWorkerTypeSnapshotValidator(&goalLoopWorkerTypeValidator{})

	snapshots, err := service.ListWorkerSnapshots(context.Background(), 1, 2)

	require.NoError(t, err)
	require.Equal(t, []workerspecdomain.Snapshot{available}, snapshots)
}

func TestValidateWorkerSnapshotForExecutionRejectsDefinitionDrift(t *testing.T) {
	service := NewService(newGoalLoopTestRepo())
	service.SetWorkerSpecSnapshotLoader(goalLoopSnapshotLoader{
		snapshot: workerspecdomain.Snapshot{ID: 3, OrganizationID: 1},
	})
	service.SetWorkerTypeSnapshotValidator(&goalLoopWorkerTypeValidator{
		errs: []error{workercreation.ErrWorkerTypeDefinitionChanged},
	})

	err := service.ValidateWorkerSnapshotForExecution(context.Background(), 1, 2, 3)

	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestValidateWorkerSnapshotForExecutionRejectsLegacyProtocolAdapter(t *testing.T) {
	service := NewService(newGoalLoopTestRepo())
	service.SetWorkerSpecSnapshotLoader(goalLoopSnapshotLoader{
		snapshot: workerspecdomain.Snapshot{
			ID:             3,
			OrganizationID: 1,
			Spec: workerspecdomain.Spec{
				Runtime: workerspecdomain.Runtime{
					ModelBinding: workerspecdomain.ModelBinding{ResourceID: 1},
				},
			},
		},
	})
	service.SetWorkerTypeSnapshotValidator(&goalLoopWorkerTypeValidator{})

	err := service.ValidateWorkerSnapshotForExecution(context.Background(), 1, 2, 3)

	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestVerificationSuccessCompletesAndTerminatesPod(t *testing.T) {
	loop := verifyingLoop(domain.EscalationPause)
	repo := newGoalLoopTestRepo(loop)
	pods := &goalLoopPodStore{pod: runningPod(*loop.PodKey)}
	terminator := &goalLoopTerminator{}
	service := NewService(repo)
	service.podLookup = pods
	service.podTerminator = terminator

	err := service.HandleVerificationResult(context.Background(), 7, &runnerv1.VerificationResultEvent{
		RequestId: "verify-1",
		PodKey:    *loop.PodKey,
		ExitCode:  0,
		Output:    "all checks passed",
	})

	require.NoError(t, err)
	require.Equal(t, domain.StatusCompleted, loop.Status)
	require.Equal(t, []string{*loop.PodKey}, terminator.keys)
	require.Equal(t, "all checks passed", stringValue(loop.VerificationOutput))
	require.NotNil(t, loop.CompletedAt)
}

func TestVerificationFailurePausesAndTerminatesPod(t *testing.T) {
	loop := verifyingLoop(domain.EscalationPause)
	repo := newGoalLoopTestRepo(loop)
	pods := &goalLoopPodStore{pod: runningPod(*loop.PodKey)}
	terminator := &goalLoopTerminator{}
	service := NewService(repo)
	service.podLookup = pods
	service.podTerminator = terminator

	err := service.HandleVerificationResult(context.Background(), 7, &runnerv1.VerificationResultEvent{
		RequestId: "verify-1",
		PodKey:    *loop.PodKey,
		ExitCode:  1,
		Output:    "test failed",
	})

	require.NoError(t, err)
	require.Equal(t, domain.StatusPaused, loop.Status)
	require.Equal(t, []string{*loop.PodKey}, terminator.keys)
	require.Equal(t, "verification exited with code 1", stringValue(loop.VerificationError))
}

func validCreateRequest() CreateRequest {
	return CreateRequest{
		OrganizationID:       1,
		CreatedByID:          2,
		Name:                 "Repair checkout total",
		WorkerSpecSnapshotID: 3,
		Objective:            "Repair checkout total",
		AcceptanceCriteria:   []string{"tests pass"},
		VerificationCommand:  "go test ./...",
		EscalationPolicy:     domain.EscalationPause,
	}
}

func verifyingLoop(policy string) *domain.GoalLoop {
	podKey := "goal-loop-pod"
	requestID := "verify-1"
	return &domain.GoalLoop{
		ID:                    1,
		OrganizationID:        1,
		Slug:                  "repair-checkout-total",
		Status:                domain.StatusVerifying,
		PodKey:                &podKey,
		VerificationRequestID: &requestID,
		EscalationPolicy:      policy,
	}
}

func runningPod(key string) *agentpod.Pod {
	return &agentpod.Pod{
		OrganizationID: 1,
		PodKey:         key,
		RunnerID:       7,
		Status:         agentpod.StatusRunning,
	}
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

type goalLoopWorkerTypeValidator struct {
	errs []error
}

func (v *goalLoopWorkerTypeValidator) ValidateWorkerTypeSnapshot(
	_ context.Context,
	_ workerspecsvc.Scope,
	_ workerspecdomain.WorkerType,
) error {
	if len(v.errs) == 0 {
		return nil
	}
	err := v.errs[0]
	v.errs = v.errs[1:]
	return err
}

type goalLoopSnapshotLoader struct {
	snapshot  workerspecdomain.Snapshot
	snapshots []workerspecdomain.Snapshot
	err       error
}

func (l goalLoopSnapshotLoader) GetByID(_ context.Context, _ int64, _ int64) (workerspecdomain.Snapshot, error) {
	return l.snapshot, l.err
}

func (l goalLoopSnapshotLoader) ListByOrganization(_ context.Context, _ int64) ([]workerspecdomain.Snapshot, error) {
	return l.snapshots, l.err
}

type goalLoopPodStore struct {
	pod *agentpod.Pod
	err error
}

func (s *goalLoopPodStore) GetPod(_ context.Context, _ string) (*agentpod.Pod, error) {
	return s.pod, s.err
}

type goalLoopTerminator struct {
	keys []string
	err  error
}

func (t *goalLoopTerminator) TerminatePod(_ context.Context, podKey string) error {
	if t.err != nil {
		return t.err
	}
	t.keys = append(t.keys, podKey)
	return nil
}

var _ domain.Repository = (*goalLoopTestRepo)(nil)
var _ WorkerSpecSnapshotLoader = goalLoopSnapshotLoader{}
var _ PodLookup = (*goalLoopPodStore)(nil)
var _ PodTerminator = (*goalLoopTerminator)(nil)
