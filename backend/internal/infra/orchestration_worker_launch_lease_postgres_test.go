package infra

import (
	"context"
	"testing"
	"time"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	workerplanner "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationworker"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestWorkerLaunchRejectsConcurrentClaimAndAllowsReclaimAfterRelease(
	t *testing.T,
) {
	db, repo := orchestrationPostgresRepository(t)
	applied, scope := createWorkerLaunchForLeaseTest(t, db, repo)
	first, err := repo.ClaimWorkerLaunch(
		context.Background(),
		scope,
		applied.LaunchID,
		time.Minute,
		uuid.NewString(),
	)
	require.NoError(t, err)

	_, err = repo.ClaimWorkerLaunch(
		context.Background(),
		scope,
		applied.LaunchID,
		time.Minute,
		uuid.NewString(),
	)
	assert.ErrorIs(t, err, workerplanner.ErrWorkerLaunchInProgress)
	require.NoError(t, repo.ReleaseWorkerLaunch(
		context.Background(),
		scope,
		first,
		"materialization failed",
	))
	second, err := repo.ClaimWorkerLaunch(
		context.Background(),
		scope,
		applied.LaunchID,
		time.Minute,
		uuid.NewString(),
	)
	require.NoError(t, err)
	assert.NotEqual(t, first.ClaimToken, second.ClaimToken)

	var attempts int
	var lastError *string
	require.NoError(t, db.Table("orchestration_worker_launches").
		Select("attempt_count, last_error").
		Where("id = ?", applied.LaunchID).
		Row().Scan(&attempts, &lastError))
	assert.Equal(t, 2, attempts)
	assert.Nil(t, lastError)
}

func TestWorkerLaunchExpiredLeaseCannotCompleteOutbox(t *testing.T) {
	db, repo := orchestrationPostgresRepository(t)
	applied, scope := createWorkerLaunchForLeaseTest(t, db, repo)
	claim, err := repo.ClaimWorkerLaunch(
		context.Background(),
		scope,
		applied.LaunchID,
		time.Minute,
		uuid.NewString(),
	)
	require.NoError(t, err)
	require.NoError(t, db.Exec(
		`UPDATE orchestration_worker_launches
SET lease_expires_at = now() - interval '1 second'
WHERE id = ?`,
		applied.LaunchID,
	).Error)

	_, err = repo.CompleteWorkerLaunch(
		context.Background(),
		scope,
		claim,
		workerplanner.WorkerPodLaunch{
			PodID:          1,
			PodKey:         "7-standalone-12345678",
			RunnerID:       11,
			CommandPayload: []byte{1},
		},
		time.Hour,
	)
	assert.ErrorIs(t, err, workerplanner.ErrWorkerLaunchLeaseLost)

	var pendingCount int64
	require.NoError(t, db.Table("pending_runner_commands").
		Count(&pendingCount).Error)
	assert.Zero(t, pendingCount)
}

func createWorkerLaunchForLeaseTest(
	t *testing.T,
	db *gorm.DB,
	repo *orchestrationResourceRepo,
) (workerplanner.AppliedWorker, control.Scope) {
	t.Helper()
	applyOrchestrationDomainLinkFixtures(t, db)
	applyWorkerLaunchFixtures(t, db)
	plan := orchestrationWorkerApplyPlan(t)
	require.NoError(t, repo.CreatePlan(context.Background(), plan))
	applied, err := repo.RunWorkerApplyTransaction(
		context.Background(),
		plan.Scope,
		plan.ID,
		workerLaunchMutationBuilder(t),
	)
	require.NoError(t, err)
	return applied, plan.Scope
}
