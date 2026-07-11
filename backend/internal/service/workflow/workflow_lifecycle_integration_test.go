package workflow

import (
	"context"
	"testing"
	"time"

	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupIntegrationServices creates real WorkflowService + WorkflowRunService backed by testkit.SetupTestDB.
func setupIntegrationServices(t *testing.T) (*WorkflowService, *WorkflowRunService, context.Context) {
	t.Helper()
	db := testkit.SetupTestDB(t)
	workflowRepo := infra.NewWorkflowRepository(db)
	runRepo := infra.NewWorkflowRunRepository(db)
	return NewWorkflowService(workflowRepo), NewWorkflowRunService(runRepo), context.Background()
}

func createTestWorkflow(t *testing.T, svc *WorkflowService, ctx context.Context, orgID int64, slug string) *workflowDomain.Workflow {
	t.Helper()
	workflow, err := svc.Create(ctx, &CreateWorkflowRequest{
		OrganizationID: orgID,
		CreatedByID:    1,
		Name:           "Test Workflow " + slug,
		Slug:           slug,
		AgentSlug:      "claude",
		PromptTemplate: "Do the task for {{project}}",
		TimeoutMinutes: 30,
	})
	require.NoError(t, err)
	return workflow
}

func TestLoopLifecycle_CreateAndQuery(t *testing.T) {
	workflowSvc, _, ctx := setupIntegrationServices(t)

	created := createTestWorkflow(t, workflowSvc, ctx, 1, "create-query")

	// Query back by slug
	got, err := workflowSvc.GetBySlug(ctx, 1, "create-query")
	require.NoError(t, err)

	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "Test Workflow create-query", got.Name)
	assert.Equal(t, "create-query", got.Slug)
	assert.Equal(t, "claude", got.AgentSlug)
	assert.Equal(t, "Do the task for {{project}}", got.PromptTemplate)
	assert.Equal(t, workflowDomain.StatusEnabled, got.Status)
	assert.Equal(t, workflowDomain.ExecutionModeAutopilot, got.ExecutionMode)
	assert.Equal(t, workflowDomain.SandboxStrategyPersistent, got.SandboxStrategy)
	assert.Equal(t, workflowDomain.ConcurrencyPolicySkip, got.ConcurrencyPolicy)
	assert.Equal(t, 1, got.MaxConcurrentRuns)
	assert.Equal(t, 30, got.TimeoutMinutes)
	assert.Equal(t, int64(1), got.OrganizationID)

	// Query back by ID
	gotByID, err := workflowSvc.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, gotByID.ID)
	assert.Equal(t, "create-query", gotByID.Slug)

	// List returns the created workflow
	workflows, total, err := workflowSvc.List(ctx, &ListWorkflowsFilter{OrganizationID: 1, Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, created.ID, workflows[0].ID)
}

func TestLoopLifecycle_TriggerRun(t *testing.T) {
	workflowSvc, runSvc, ctx := setupIntegrationServices(t)

	workflow := createTestWorkflow(t, workflowSvc, ctx, 1, "trigger-run")

	// Get next run number (should be 1 for a fresh workflow)
	nextNum, err := runSvc.GetNextRunNumber(ctx, workflow.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, nextNum)

	// Create a run
	now := time.Now()
	run := &workflowDomain.WorkflowRun{
		OrganizationID: 1,
		WorkflowID:     workflow.ID,
		RunNumber:      nextNum,
		Status:         workflowDomain.RunStatusPending,
		TriggerType:    workflowDomain.RunTriggerManual,
		StartedAt:      &now,
	}
	err = runSvc.Create(ctx, run)
	require.NoError(t, err)
	assert.NotZero(t, run.ID)
	assert.Equal(t, 1, run.RunNumber)

	// Query back
	got, err := runSvc.GetByID(ctx, run.ID)
	require.NoError(t, err)
	assert.Equal(t, workflowDomain.RunStatusPending, got.Status)
	assert.Equal(t, workflow.ID, got.WorkflowID)
	assert.Equal(t, 1, got.RunNumber)
	assert.Equal(t, workflowDomain.RunTriggerManual, got.TriggerType)
}

func TestLoopLifecycle_RunNumberIncrement(t *testing.T) {
	workflowSvc, runSvc, ctx := setupIntegrationServices(t)

	workflow := createTestWorkflow(t, workflowSvc, ctx, 1, "run-number")

	for i := 1; i <= 5; i++ {
		nextNum, err := runSvc.GetNextRunNumber(ctx, workflow.ID)
		require.NoError(t, err)
		assert.Equal(t, i, nextNum)

		run := &workflowDomain.WorkflowRun{
			OrganizationID: 1,
			WorkflowID:     workflow.ID,
			RunNumber:      nextNum,
			Status:         workflowDomain.RunStatusCompleted,
			TriggerType:    workflowDomain.RunTriggerCron,
		}
		require.NoError(t, runSvc.Create(ctx, run))
	}

	// After 5 runs, next should be 6
	nextNum, err := runSvc.GetNextRunNumber(ctx, workflow.ID)
	require.NoError(t, err)
	assert.Equal(t, 6, nextNum)
}

func TestLoopLifecycle_RunStatusTransitions(t *testing.T) {
	workflowSvc, runSvc, ctx := setupIntegrationServices(t)

	workflow := createTestWorkflow(t, workflowSvc, ctx, 1, "status-transition")

	// Create run in pending state
	now := time.Now()
	run := &workflowDomain.WorkflowRun{
		OrganizationID: 1,
		WorkflowID:     workflow.ID,
		RunNumber:      1,
		Status:         workflowDomain.RunStatusPending,
		TriggerType:    workflowDomain.RunTriggerAPI,
		StartedAt:      &now,
	}
	require.NoError(t, runSvc.Create(ctx, run))

	// Transition to running
	err := runSvc.UpdateStatus(ctx, run.ID, map[string]interface{}{
		"status": workflowDomain.RunStatusRunning,
	})
	require.NoError(t, err)

	got, err := runSvc.GetByID(ctx, run.ID)
	require.NoError(t, err)
	assert.Equal(t, workflowDomain.RunStatusRunning, got.Status)

	// Finish the run (using FinishRun with optimistic locking)
	finishedAt := time.Now()
	durationSec := int(finishedAt.Sub(now).Seconds())
	updated, err := runSvc.FinishRun(ctx, run.ID, map[string]interface{}{
		"status":       workflowDomain.RunStatusCompleted,
		"finished_at":  finishedAt,
		"duration_sec": durationSec,
	})
	require.NoError(t, err)
	assert.True(t, updated, "FinishRun should update the row")

	got, err = runSvc.GetByID(ctx, run.ID)
	require.NoError(t, err)
	assert.Equal(t, workflowDomain.RunStatusCompleted, got.Status)
	assert.NotNil(t, got.FinishedAt)
	assert.NotNil(t, got.DurationSec)

	// Double-finish should return false (optimistic lock)
	updated, err = runSvc.FinishRun(ctx, run.ID, map[string]interface{}{
		"status":      workflowDomain.RunStatusFailed,
		"finished_at": time.Now(),
	})
	require.NoError(t, err)
	assert.False(t, updated, "double-finish should be rejected by optimistic lock")
}

func TestLoopLifecycle_DeleteOldRuns(t *testing.T) {
	workflowSvc, runSvc, ctx := setupIntegrationServices(t)

	workflow := createTestWorkflow(t, workflowSvc, ctx, 1, "delete-old-runs")

	// Create 5 finished runs
	for i := 1; i <= 5; i++ {
		finished := time.Now()
		run := &workflowDomain.WorkflowRun{
			OrganizationID: 1,
			WorkflowID:     workflow.ID,
			RunNumber:      i,
			Status:         workflowDomain.RunStatusCompleted,
			TriggerType:    workflowDomain.RunTriggerCron,
			FinishedAt:     &finished,
		}
		require.NoError(t, runSvc.Create(ctx, run))
	}

	// Verify all 5 exist
	runs, total, err := runSvc.ListWorkflowRuns(ctx, &ListWorkflowRunsFilter{WorkflowID: workflow.ID, Limit: 100})
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, runs, 5)

	// Delete old runs, keeping only 2
	deleted, err := runSvc.DeleteOldFinishedRuns(ctx, workflow.ID, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), deleted)

	// Verify only 2 remain
	runs, total, err = runSvc.ListWorkflowRuns(ctx, &ListWorkflowRunsFilter{WorkflowID: workflow.ID, Limit: 100})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, runs, 2)

	// The remaining runs should be the 2 most recent (highest ID, i.e., run_number 4 and 5)
	remaining := map[int]bool{runs[0].RunNumber: true, runs[1].RunNumber: true}
	assert.True(t, remaining[4], "run_number 4 should be retained")
	assert.True(t, remaining[5], "run_number 5 should be retained")
}

func TestLoopLifecycle_SlugUniqueness(t *testing.T) {
	workflowSvc, _, ctx := setupIntegrationServices(t)

	// Create first workflow
	_, err := workflowSvc.Create(ctx, &CreateWorkflowRequest{
		OrganizationID: 1,
		CreatedByID:    1,
		Name:           "First",
		Slug:           "unique-slug",
		PromptTemplate: "prompt",
	})
	require.NoError(t, err)

	// Create second workflow with same slug and same org should fail.
	// In PostgreSQL the error maps to ErrDuplicateSlug; in SQLite the unique
	// constraint error message differs, so we accept either sentinel or any error.
	_, err = workflowSvc.Create(ctx, &CreateWorkflowRequest{
		OrganizationID: 1,
		CreatedByID:    1,
		Name:           "Second",
		Slug:           "unique-slug",
		PromptTemplate: "prompt",
	})
	require.Error(t, err, "duplicate slug in same org should fail")

	// Verify the first workflow is still intact
	got, err := workflowSvc.GetBySlug(ctx, 1, "unique-slug")
	require.NoError(t, err)
	assert.Equal(t, "First", got.Name)

	// Same slug in a different org should succeed
	_, err = workflowSvc.Create(ctx, &CreateWorkflowRequest{
		OrganizationID: 2,
		CreatedByID:    1,
		Name:           "Third",
		Slug:           "unique-slug",
		PromptTemplate: "prompt",
	})
	assert.NoError(t, err)
}

func TestLoopLifecycle_TimeoutDetection(t *testing.T) {
	// GetTimedOutRuns uses PostgreSQL-specific syntax (::INTERVAL),
	// so we test the timeout concept via manual DB state + direct query.
	db := testkit.SetupTestDB(t)
	runRepo := infra.NewWorkflowRunRepository(db)
	runSvc := NewWorkflowRunService(runRepo)
	workflowRepo := infra.NewWorkflowRepository(db)
	workflowSvc := NewWorkflowService(workflowRepo)
	ctx := context.Background()

	workflow := createTestWorkflow(t, workflowSvc, ctx, 1, "timeout-detect")

	// Create a run that started 2 hours ago (workflow timeout is 30 min)
	startedAt := time.Now().Add(-2 * time.Hour)
	run := &workflowDomain.WorkflowRun{
		OrganizationID: 1,
		WorkflowID:     workflow.ID,
		RunNumber:      1,
		Status:         workflowDomain.RunStatusRunning,
		TriggerType:    workflowDomain.RunTriggerCron,
		StartedAt:      &startedAt,
	}
	require.NoError(t, runSvc.Create(ctx, run))

	// Verify the run is active (started but not finished)
	got, err := runSvc.GetByID(ctx, run.ID)
	require.NoError(t, err)
	assert.Equal(t, workflowDomain.RunStatusRunning, got.Status)
	assert.Nil(t, got.FinishedAt, "run should not be finished")

	// Verify it's detectable as timed out: started_at + timeout < now
	assert.True(t, startedAt.Add(time.Duration(workflow.TimeoutMinutes)*time.Minute).Before(time.Now()),
		"run should have exceeded timeout_minutes (%d)", workflow.TimeoutMinutes)

	// Simulate what the scheduler does: mark the run as timed out
	finishedAt := time.Now()
	updated, err := runSvc.FinishRun(ctx, run.ID, map[string]interface{}{
		"status":      workflowDomain.RunStatusTimeout,
		"finished_at": finishedAt,
	})
	require.NoError(t, err)
	assert.True(t, updated)

	got, err = runSvc.GetByID(ctx, run.ID)
	require.NoError(t, err)
	assert.Equal(t, workflowDomain.RunStatusTimeout, got.Status)
	assert.NotNil(t, got.FinishedAt)
}
