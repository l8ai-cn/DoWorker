package goalloop

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/goalloop"
	workerspecdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
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

type goalLoopTestRepo struct {
	loops map[int64]*domain.GoalLoop
}

func newGoalLoopTestRepo(loops ...*domain.GoalLoop) *goalLoopTestRepo {
	repo := &goalLoopTestRepo{loops: map[int64]*domain.GoalLoop{}}
	for _, loop := range loops {
		repo.loops[loop.ID] = loop
	}
	return repo
}

func (r *goalLoopTestRepo) Create(_ context.Context, loop *domain.GoalLoop) error {
	loop.ID = int64(len(r.loops) + 1)
	r.loops[loop.ID] = loop
	return nil
}

func (r *goalLoopTestRepo) GetBySlug(_ context.Context, orgID int64, slug string) (*domain.GoalLoop, error) {
	for _, loop := range r.loops {
		if loop.OrganizationID == orgID && loop.Slug == slug {
			return loop, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *goalLoopTestRepo) GetByPodKey(_ context.Context, podKey string) (*domain.GoalLoop, error) {
	for _, loop := range r.loops {
		if loop.PodKey != nil && *loop.PodKey == podKey {
			return loop, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *goalLoopTestRepo) GetByAutopilotControllerKey(_ context.Context, key string) (*domain.GoalLoop, error) {
	return nil, domain.ErrNotFound
}

func (r *goalLoopTestRepo) GetByVerificationRequestID(_ context.Context, requestID string) (*domain.GoalLoop, error) {
	for _, loop := range r.loops {
		if loop.VerificationRequestID != nil && *loop.VerificationRequestID == requestID {
			return loop, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *goalLoopTestRepo) ListTimedOut(_ context.Context, _ time.Time) ([]*domain.GoalLoop, error) {
	return nil, nil
}

func (r *goalLoopTestRepo) List(_ context.Context, _ domain.ListFilter) ([]*domain.GoalLoop, int64, error) {
	return nil, 0, nil
}

func (r *goalLoopTestRepo) ExistsSlug(_ context.Context, _ int64, _ string) (bool, error) {
	return false, nil
}

func (r *goalLoopTestRepo) Update(_ context.Context, id int64, updates map[string]any) error {
	loop := r.loops[id]
	if loop == nil {
		return domain.ErrNotFound
	}
	for key, value := range updates {
		switch key {
		case "status":
			loop.Status = value.(string)
		case "completed_at":
			loop.CompletedAt = timePointer(value)
		case "verified_at":
			loop.VerifiedAt = timePointer(value)
		case "verification_exit_code":
			loop.VerificationExitCode = intPointer(value)
		case "verification_output":
			loop.VerificationOutput = stringPointerValue(value)
		case "verification_output_truncated":
			loop.VerificationOutputTruncated = value.(bool)
		case "verification_error":
			loop.VerificationError = stringPointerValue(value)
		}
	}
	return nil
}

func timePointer(value any) *time.Time {
	timeValue, ok := value.(time.Time)
	if !ok {
		return nil
	}
	return &timeValue
}

func intPointer(value any) *int {
	intValue, ok := value.(int)
	if !ok {
		return nil
	}
	return &intValue
}

func stringPointerValue(value any) *string {
	stringValue, ok := value.(string)
	if !ok {
		return nil
	}
	return &stringValue
}

type goalLoopSnapshotLoader struct {
	snapshot workerspecdomain.Snapshot
	err      error
}

func (l goalLoopSnapshotLoader) GetByID(_ context.Context, _ int64, _ int64) (workerspecdomain.Snapshot, error) {
	return l.snapshot, l.err
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

func TestGoalLoopTestRepoRejectsUnknownUpdate(t *testing.T) {
	repo := newGoalLoopTestRepo()
	err := repo.Update(context.Background(), 99, nil)
	require.True(t, errors.Is(err, domain.ErrNotFound))
}
