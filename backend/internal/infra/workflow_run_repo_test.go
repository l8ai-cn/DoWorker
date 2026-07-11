package infra

import (
	"context"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunRepository_Create(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRunRepository(db)
	ctx := context.Background()

	// Seed a parent workflow
	workflowRepo := NewWorkflowRepository(db)
	l := &workflow.Workflow{
		OrganizationID: 1, Name: "Parent", Slug: "parent",
		PromptTemplate: "p",
		ExecutionMode:  workflow.ExecutionModeAutopilot, Status: workflow.StatusEnabled,
		SandboxStrategy: workflow.SandboxStrategyPersistent, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
		MaxConcurrentRuns: 1, TimeoutMinutes: 60,
		AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"),
		CreatedByID: 1,
	}
	require.NoError(t, workflowRepo.Create(ctx, l))

	run := &workflow.WorkflowRun{
		OrganizationID: 1,
		WorkflowID:     l.ID,
		RunNumber:      1,
		Status:         workflow.RunStatusPending,
		TriggerType:    workflow.RunTriggerManual,
	}
	err := repo.Create(ctx, run)
	require.NoError(t, err)
	assert.NotZero(t, run.ID)
}

func TestRunRepository_GetByID(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRunRepository(db)
	ctx := context.Background()

	run := &workflow.WorkflowRun{
		OrganizationID: 1, WorkflowID: 1, RunNumber: 1,
		Status: workflow.RunStatusPending, TriggerType: workflow.RunTriggerManual,
	}
	require.NoError(t, repo.Create(ctx, run))

	t.Run("should return run by ID", func(t *testing.T) {
		got, err := repo.GetByID(ctx, run.ID)
		require.NoError(t, err)
		assert.Equal(t, workflow.RunStatusPending, got.Status)
		assert.Equal(t, 1, got.RunNumber)
	})

	t.Run("should return ErrNotFound for non-existent", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 99999)
		assert.ErrorIs(t, err, workflow.ErrNotFound)
	})
}

func TestRunRepository_List(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRunRepository(db)
	ctx := context.Background()

	// Seed runs
	for i := 1; i <= 5; i++ {
		run := &workflow.WorkflowRun{
			OrganizationID: 1, WorkflowID: 1, RunNumber: i,
			Status: workflow.RunStatusCompleted, TriggerType: workflow.RunTriggerCron,
		}
		require.NoError(t, repo.Create(ctx, run))
	}
	// Different workflow
	run := &workflow.WorkflowRun{
		OrganizationID: 1, WorkflowID: 2, RunNumber: 1,
		Status: workflow.RunStatusPending, TriggerType: workflow.RunTriggerAPI,
	}
	require.NoError(t, repo.Create(ctx, run))

	t.Run("should list runs for specific workflow", func(t *testing.T) {
		result, total, err := repo.List(ctx, &workflow.WorkflowRunListFilter{WorkflowID: 1})
		require.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, result, 5)
	})

	t.Run("should respect limit", func(t *testing.T) {
		result, total, err := repo.List(ctx, &workflow.WorkflowRunListFilter{WorkflowID: 1, Limit: 2})
		require.NoError(t, err)
		assert.Equal(t, int64(5), total) // total unaffected
		assert.Len(t, result, 2)
	})

	t.Run("should isolate by workflow_id", func(t *testing.T) {
		result, total, err := repo.List(ctx, &workflow.WorkflowRunListFilter{WorkflowID: 2})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, result, 1)
	})
}

func TestRunRepository_Update(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRunRepository(db)
	ctx := context.Background()

	run := &workflow.WorkflowRun{
		OrganizationID: 1, WorkflowID: 1, RunNumber: 1,
		Status: workflow.RunStatusPending, TriggerType: workflow.RunTriggerManual,
	}
	require.NoError(t, repo.Create(ctx, run))

	podKey := "pod-123"
	err := repo.Update(ctx, run.ID, map[string]interface{}{
		"status":  workflow.RunStatusRunning,
		"pod_key": podKey,
	})
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, run.ID)
	require.NoError(t, err)
	assert.Equal(t, workflow.RunStatusRunning, got.Status)
	assert.Equal(t, &podKey, got.PodKey)
}

func TestRunRepository_GetMaxRunNumber(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRunRepository(db)
	ctx := context.Background()

	t.Run("should return 0 for no runs", func(t *testing.T) {
		max, err := repo.GetMaxRunNumber(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, 0, max)
	})

	// Seed runs
	for i := 1; i <= 3; i++ {
		run := &workflow.WorkflowRun{
			OrganizationID: 1, WorkflowID: 1, RunNumber: i,
			Status: workflow.RunStatusCompleted, TriggerType: workflow.RunTriggerCron,
		}
		require.NoError(t, repo.Create(ctx, run))
	}

	t.Run("should return max run number", func(t *testing.T) {
		max, err := repo.GetMaxRunNumber(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, 3, max)
	})

	t.Run("should be scoped to workflow_id", func(t *testing.T) {
		max, err := repo.GetMaxRunNumber(ctx, 999)
		require.NoError(t, err)
		assert.Equal(t, 0, max)
	})
}

func TestRunRepository_GetByAutopilotKey(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRunRepository(db)
	ctx := context.Background()

	apKey := "ap-ctrl-123"
	run := &workflow.WorkflowRun{
		OrganizationID: 1, WorkflowID: 1, RunNumber: 1,
		Status: workflow.RunStatusRunning, TriggerType: workflow.RunTriggerManual,
		AutopilotControllerKey: &apKey,
	}
	require.NoError(t, repo.Create(ctx, run))

	t.Run("should find run by autopilot key", func(t *testing.T) {
		got, err := repo.GetByAutopilotKey(ctx, "ap-ctrl-123")
		require.NoError(t, err)
		assert.Equal(t, run.ID, got.ID)
	})

	t.Run("should return ErrNotFound for unknown key", func(t *testing.T) {
		_, err := repo.GetByAutopilotKey(ctx, "unknown-key")
		assert.ErrorIs(t, err, workflow.ErrNotFound)
	})
}

// TestRunRepository_CountActiveRuns tests the SSOT-based active run counting.
func TestRunRepository_CountActiveRuns(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRunRepository(db)
	ctx := context.Background()

	db.Exec(`INSERT INTO pods (pod_key, status) VALUES ('pod-running', 'running')`)
	db.Exec(`INSERT INTO pods (pod_key, status) VALUES ('pod-init', 'initializing')`)
	db.Exec(`INSERT INTO pods (pod_key, status) VALUES ('pod-done', 'completed')`)
	db.Exec(`INSERT INTO pods (pod_key, status) VALUES ('pod-err', 'error')`)
	db.Exec(`INSERT INTO pods (pod_key, status) VALUES ('pod-term', 'terminated')`)

	runs := []workflow.WorkflowRun{
		{OrganizationID: 1, WorkflowID: 1, RunNumber: 1, Status: workflow.RunStatusRunning,
			TriggerType: workflow.RunTriggerManual, PodKey: workflowStrPtr("pod-running")},
		{OrganizationID: 1, WorkflowID: 1, RunNumber: 2, Status: workflow.RunStatusRunning,
			TriggerType: workflow.RunTriggerManual, PodKey: workflowStrPtr("pod-init")},
		{OrganizationID: 1, WorkflowID: 1, RunNumber: 3, Status: workflow.RunStatusRunning,
			TriggerType: workflow.RunTriggerManual, PodKey: workflowStrPtr("pod-done")},
		{OrganizationID: 1, WorkflowID: 1, RunNumber: 4, Status: workflow.RunStatusRunning,
			TriggerType: workflow.RunTriggerManual, PodKey: workflowStrPtr("pod-err")},
		{OrganizationID: 1, WorkflowID: 1, RunNumber: 5, Status: workflow.RunStatusRunning,
			TriggerType: workflow.RunTriggerManual, PodKey: workflowStrPtr("pod-term")},
		{OrganizationID: 1, WorkflowID: 1, RunNumber: 6, Status: workflow.RunStatusPending,
			TriggerType: workflow.RunTriggerManual},
		{OrganizationID: 1, WorkflowID: 1, RunNumber: 7, Status: workflow.RunStatusSkipped,
			TriggerType: workflow.RunTriggerManual},
	}
	for i := range runs {
		require.NoError(t, repo.Create(ctx, &runs[i]))
	}

	count, err := repo.CountActiveRuns(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

// TestRunRepository_GetActiveRunByPodKey tests finding active runs by pod key.
func TestRunRepository_GetActiveRunByPodKey(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRunRepository(db)
	ctx := context.Background()

	db.Exec(`INSERT INTO pods (pod_key, status) VALUES ('active-pod', 'running')`)
	db.Exec(`INSERT INTO pods (pod_key, status) VALUES ('done-pod', 'completed')`)

	run1 := &workflow.WorkflowRun{
		OrganizationID: 1, WorkflowID: 1, RunNumber: 1,
		Status: workflow.RunStatusRunning, TriggerType: workflow.RunTriggerManual,
		PodKey: workflowStrPtr("active-pod"),
	}
	finishedAt := time.Now()
	run2 := &workflow.WorkflowRun{
		OrganizationID: 1, WorkflowID: 1, RunNumber: 2,
		Status: workflow.RunStatusCompleted, TriggerType: workflow.RunTriggerManual,
		PodKey:     workflowStrPtr("done-pod"),
		FinishedAt: &finishedAt,
	}
	require.NoError(t, repo.Create(ctx, run1))
	require.NoError(t, repo.Create(ctx, run2))

	t.Run("should find run with active pod", func(t *testing.T) {
		got, err := repo.GetActiveRunByPodKey(ctx, "active-pod")
		require.NoError(t, err)
		assert.Equal(t, run1.ID, got.ID)
	})

	t.Run("should not find run with completed pod", func(t *testing.T) {
		_, err := repo.GetActiveRunByPodKey(ctx, "done-pod")
		assert.Error(t, err)
	})
}

// TestRunRepository_ComputeLoopStats tests SSOT statistics computation.
func TestRunRepository_ComputeLoopStats(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRunRepository(db)
	ctx := context.Background()

	db.Exec(`INSERT INTO pods (pod_key, status) VALUES ('stat-completed', 'completed')`)
	db.Exec(`INSERT INTO pods (pod_key, status) VALUES ('stat-terminated', 'terminated')`)
	db.Exec(`INSERT INTO pods (pod_key, status) VALUES ('stat-error', 'error')`)
	db.Exec(`INSERT INTO pods (pod_key, status) VALUES ('stat-running', 'running')`)

	db.Exec(`INSERT INTO autopilot_controllers (autopilot_controller_key, phase) VALUES ('ap-completed', 'completed')`)
	db.Exec(`INSERT INTO autopilot_controllers (autopilot_controller_key, phase) VALUES ('ap-failed', 'failed')`)
	db.Exec(`INSERT INTO autopilot_controllers (autopilot_controller_key, phase) VALUES ('ap-stopped', 'stopped')`)

	runs := []workflow.WorkflowRun{
		{OrganizationID: 1, WorkflowID: 1, RunNumber: 1, Status: workflow.RunStatusRunning,
			TriggerType: workflow.RunTriggerCron, PodKey: workflowStrPtr("stat-completed")},
		{OrganizationID: 1, WorkflowID: 1, RunNumber: 2, Status: workflow.RunStatusRunning,
			TriggerType: workflow.RunTriggerCron, PodKey: workflowStrPtr("stat-terminated")},
		{OrganizationID: 1, WorkflowID: 1, RunNumber: 3, Status: workflow.RunStatusRunning,
			TriggerType: workflow.RunTriggerCron, PodKey: workflowStrPtr("stat-error")},
		{OrganizationID: 1, WorkflowID: 1, RunNumber: 4, Status: workflow.RunStatusRunning,
			TriggerType: workflow.RunTriggerCron, PodKey: workflowStrPtr("stat-running")},
		{OrganizationID: 1, WorkflowID: 1, RunNumber: 5, Status: workflow.RunStatusSkipped,
			TriggerType: workflow.RunTriggerCron},
		{OrganizationID: 1, WorkflowID: 1, RunNumber: 6, Status: workflow.RunStatusRunning,
			TriggerType: workflow.RunTriggerCron, PodKey: workflowStrPtr("stat-running"),
			AutopilotControllerKey: workflowStrPtr("ap-completed")},
		{OrganizationID: 1, WorkflowID: 1, RunNumber: 7, Status: workflow.RunStatusRunning,
			TriggerType: workflow.RunTriggerCron, PodKey: workflowStrPtr("stat-running"),
			AutopilotControllerKey: workflowStrPtr("ap-failed")},
		{OrganizationID: 1, WorkflowID: 1, RunNumber: 8, Status: workflow.RunStatusRunning,
			TriggerType: workflow.RunTriggerCron, PodKey: workflowStrPtr("stat-running"),
			AutopilotControllerKey: workflowStrPtr("ap-stopped")},
	}
	for i := range runs {
		require.NoError(t, repo.Create(ctx, &runs[i]))
	}

	total, successful, failed, err := repo.ComputeLoopStats(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, 8, total)
	assert.Equal(t, 2, successful)
	assert.Equal(t, 4, failed)
}

// TestRunRepository_ComputeLoopStats_PodWinsOverAutopilot tests Pod priority.
func TestRunRepository_ComputeLoopStats_PodWinsOverAutopilot(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRunRepository(db)
	ctx := context.Background()

	db.Exec(`INSERT INTO pods (pod_key, status) VALUES ('pod-wins', 'completed')`)
	db.Exec(`INSERT INTO autopilot_controllers (autopilot_controller_key, phase) VALUES ('ap-stale', 'running')`)

	run := &workflow.WorkflowRun{
		OrganizationID: 1, WorkflowID: 1, RunNumber: 1,
		Status: workflow.RunStatusRunning, TriggerType: workflow.RunTriggerManual,
		PodKey:                 workflowStrPtr("pod-wins"),
		AutopilotControllerKey: workflowStrPtr("ap-stale"),
	}
	require.NoError(t, repo.Create(ctx, run))

	total, successful, failed, err := repo.ComputeLoopStats(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, 1, successful, "Pod terminal (completed) should win over autopilot active (running)")
	assert.Equal(t, 0, failed)
}

func TestRunRepository_GetLatestPodKey(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRunRepository(db)
	ctx := context.Background()

	t.Run("should return nil when no runs exist", func(t *testing.T) {
		result := repo.GetLatestPodKey(ctx, 1)
		assert.Nil(t, result)
	})

	t.Run("should return nil when runs have no pod_key", func(t *testing.T) {
		run := &workflow.WorkflowRun{
			OrganizationID: 1, WorkflowID: 1, RunNumber: 1,
			Status: workflow.RunStatusSkipped, TriggerType: workflow.RunTriggerCron,
		}
		require.NoError(t, repo.Create(ctx, run))

		result := repo.GetLatestPodKey(ctx, 1)
		assert.Nil(t, result)
	})

	t.Run("should return latest pod_key", func(t *testing.T) {
		run1 := &workflow.WorkflowRun{
			OrganizationID: 1, WorkflowID: 2, RunNumber: 1,
			Status: workflow.RunStatusCompleted, TriggerType: workflow.RunTriggerManual,
			PodKey: workflowStrPtr("old-pod"),
		}
		run2 := &workflow.WorkflowRun{
			OrganizationID: 1, WorkflowID: 2, RunNumber: 2,
			Status: workflow.RunStatusCompleted, TriggerType: workflow.RunTriggerManual,
			PodKey: workflowStrPtr("latest-pod"),
		}
		require.NoError(t, repo.Create(ctx, run1))
		require.NoError(t, repo.Create(ctx, run2))

		result := repo.GetLatestPodKey(ctx, 2)
		require.NotNil(t, result)
		assert.Equal(t, "latest-pod", *result)
	})
}
