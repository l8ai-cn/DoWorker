package goalloop

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/goalloop"
)

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

func (r *goalLoopTestRepo) GetBySlug(
	_ context.Context,
	orgID int64,
	slug string,
) (*domain.GoalLoop, error) {
	for _, loop := range r.loops {
		if loop.OrganizationID == orgID && loop.Slug == slug {
			return loop, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *goalLoopTestRepo) GetByPodKey(
	_ context.Context,
	podKey string,
) (*domain.GoalLoop, error) {
	for _, loop := range r.loops {
		if loop.PodKey != nil && *loop.PodKey == podKey {
			return loop, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *goalLoopTestRepo) GetByAutopilotControllerKey(
	context.Context,
	string,
) (*domain.GoalLoop, error) {
	return nil, domain.ErrNotFound
}

func (r *goalLoopTestRepo) GetByVerificationRequestID(
	_ context.Context,
	requestID string,
) (*domain.GoalLoop, error) {
	for _, loop := range r.loops {
		if loop.VerificationRequestID != nil && *loop.VerificationRequestID == requestID {
			return loop, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *goalLoopTestRepo) ListVerificationPending(
	context.Context,
) ([]*domain.GoalLoop, error) {
	loops := make([]*domain.GoalLoop, 0)
	for _, loop := range r.loops {
		if loop.Status == domain.StatusVerifying &&
			loop.VerificationRequestID != nil &&
			loop.VerificationExitCode == nil &&
			loop.RetryPromptCommandID == nil {
			loops = append(loops, loop)
		}
	}
	return loops, nil
}

func (r *goalLoopTestRepo) ListRetryPromptPending(
	context.Context,
) ([]*domain.GoalLoop, error) {
	loops := make([]*domain.GoalLoop, 0)
	for _, loop := range r.loops {
		if loop.Status == domain.StatusVerifying && loop.RetryPromptCommandID != nil {
			loops = append(loops, loop)
		}
	}
	return loops, nil
}

func (r *goalLoopTestRepo) ListTimedOut(
	context.Context,
	time.Time,
) ([]*domain.GoalLoop, error) {
	return nil, nil
}

func (r *goalLoopTestRepo) List(
	context.Context,
	domain.ListFilter,
) ([]*domain.GoalLoop, int64, error) {
	return nil, 0, nil
}

func (r *goalLoopTestRepo) ExistsSlug(context.Context, int64, string) (bool, error) {
	return false, nil
}

func (r *goalLoopTestRepo) Update(
	_ context.Context,
	id int64,
	updates map[string]any,
) error {
	loop := r.loops[id]
	if loop == nil {
		return domain.ErrNotFound
	}
	for key, value := range updates {
		applyGoalLoopTestUpdate(loop, key, value)
	}
	return nil
}

func applyGoalLoopTestUpdate(loop *domain.GoalLoop, key string, value any) {
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
	case "verification_request_id":
		loop.VerificationRequestID = stringPointerValue(value)
	case "pod_key":
		loop.PodKey = stringPointerValue(value)
	case "autopilot_controller_key":
		loop.AutopilotControllerKey = stringPointerValue(value)
	case "current_iteration":
		loop.CurrentIteration = value.(int)
	case "no_progress_count":
		loop.NoProgressCount = value.(int)
	case "same_error_count":
		loop.SameErrorCount = value.(int)
	case "last_progress_fingerprint":
		loop.LastProgressFingerprint = stringPointerValue(value)
	case "last_error_fingerprint":
		loop.LastErrorFingerprint = stringPointerValue(value)
	case "retry_prompt_command_id":
		loop.RetryPromptCommandID = stringPointerValue(value)
	case "retry_prompt_created_at":
		loop.RetryPromptCreatedAt = timePointer(value)
	case "started_at":
		loop.StartedAt = timePointer(value)
	}
}

func (r *goalLoopTestRepo) TransitionStatus(
	ctx context.Context,
	id int64,
	from []string,
	updates map[string]any,
) (bool, error) {
	loop := r.loops[id]
	if loop == nil {
		return false, domain.ErrNotFound
	}
	for _, status := range from {
		if loop.Status == status {
			return true, r.Update(ctx, id, updates)
		}
	}
	return false, nil
}

func (r *goalLoopTestRepo) ConsumeVerificationResult(
	ctx context.Context,
	id int64,
	requestID string,
	updates map[string]any,
) (bool, error) {
	loop := r.loops[id]
	if loop == nil {
		return false, domain.ErrNotFound
	}
	if loop.Status != domain.StatusVerifying ||
		loop.VerificationRequestID == nil ||
		*loop.VerificationRequestID != requestID {
		return false, nil
	}
	return true, r.Update(ctx, id, updates)
}

func (r *goalLoopTestRepo) TransitionRetryPrompt(
	ctx context.Context,
	id int64,
	commandID string,
	updates map[string]any,
) (bool, error) {
	loop := r.loops[id]
	if loop == nil {
		return false, domain.ErrNotFound
	}
	if loop.Status != domain.StatusVerifying ||
		loop.RetryPromptCommandID == nil ||
		*loop.RetryPromptCommandID != commandID {
		return false, nil
	}
	return true, r.Update(ctx, id, updates)
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

func TestGoalLoopTestRepoRejectsUnknownUpdate(t *testing.T) {
	repo := newGoalLoopTestRepo()
	err := repo.Update(context.Background(), 99, nil)
	require.True(t, errors.Is(err, domain.ErrNotFound))
}

var _ domain.Repository = (*goalLoopTestRepo)(nil)
