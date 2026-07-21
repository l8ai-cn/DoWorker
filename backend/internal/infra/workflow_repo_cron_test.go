package infra

import (
	"context"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowRepository_GetDueCronWorkflows(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRepository(db)
	ctx := context.Background()

	cron := "0 9 * * *"
	pastTime := time.Now().Add(-1 * time.Hour)
	futureTime := time.Now().Add(1 * time.Hour)

	// Due cron workflow
	due := &workflow.Workflow{
		OrganizationID: 1, Name: "Due", Slug: "due",
		PromptTemplate: "prompt",
		ExecutionMode:  workflow.ExecutionModeAutopilot, Status: workflow.StatusEnabled,
		CronExpression: &cron, NextRunAt: &pastTime,
		SandboxStrategy: workflow.SandboxStrategyPersistent, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
		MaxConcurrentRuns: 1, TimeoutMinutes: 60,
		AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"),
		CreatedByID: 1,
	}
	require.NoError(t, repo.Create(ctx, due))

	// Not yet due
	notDue := &workflow.Workflow{
		OrganizationID: 1, Name: "Not Due", Slug: "not-due",
		PromptTemplate: "prompt",
		ExecutionMode:  workflow.ExecutionModeAutopilot, Status: workflow.StatusEnabled,
		CronExpression: &cron, NextRunAt: &futureTime,
		SandboxStrategy: workflow.SandboxStrategyPersistent, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
		MaxConcurrentRuns: 1, TimeoutMinutes: 60,
		AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"),
		CreatedByID: 1,
	}
	require.NoError(t, repo.Create(ctx, notDue))

	// Disabled workflow
	disabled := &workflow.Workflow{
		OrganizationID: 1, Name: "Disabled", Slug: "disabled",
		PromptTemplate: "prompt",
		ExecutionMode:  workflow.ExecutionModeAutopilot, Status: workflow.StatusDisabled,
		CronExpression: &cron, NextRunAt: &pastTime,
		SandboxStrategy: workflow.SandboxStrategyPersistent, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
		MaxConcurrentRuns: 1, TimeoutMinutes: 60,
		AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"),
		CreatedByID: 1,
	}
	require.NoError(t, repo.Create(ctx, disabled))

	result, err := repo.GetDueCronWorkflows(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "due", result[0].Slug)
}

func TestWorkflowRepository_FindWorkflowsNeedingNextRun(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRepository(db)
	ctx := context.Background()

	cron := "0 9 * * *"
	pastTime := time.Now().Add(-1 * time.Hour)

	// Enabled cron workflow with next_run_at IS NULL -> should be found
	needsInit := &workflow.Workflow{
		OrganizationID: 1, Name: "Needs Init", Slug: "needs-init",
		PromptTemplate: "p",
		ExecutionMode:  workflow.ExecutionModeAutopilot, Status: workflow.StatusEnabled,
		CronExpression:  &cron, // next_run_at is nil
		SandboxStrategy: workflow.SandboxStrategyPersistent, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
		MaxConcurrentRuns: 1, TimeoutMinutes: 60,
		AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"),
		CreatedByID: 1,
	}
	require.NoError(t, repo.Create(ctx, needsInit))

	// Enabled cron workflow with next_run_at set -> should NOT be found
	hasNextRun := &workflow.Workflow{
		OrganizationID: 1, Name: "Has Next", Slug: "has-next",
		PromptTemplate: "p",
		ExecutionMode:  workflow.ExecutionModeAutopilot, Status: workflow.StatusEnabled,
		CronExpression: &cron, NextRunAt: &pastTime,
		SandboxStrategy: workflow.SandboxStrategyPersistent, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
		MaxConcurrentRuns: 1, TimeoutMinutes: 60,
		AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"),
		CreatedByID: 1,
	}
	require.NoError(t, repo.Create(ctx, hasNextRun))

	// Disabled cron workflow with next_run_at IS NULL -> should NOT be found
	disabled := &workflow.Workflow{
		OrganizationID: 1, Name: "Disabled", Slug: "disabled",
		PromptTemplate: "p",
		ExecutionMode:  workflow.ExecutionModeAutopilot, Status: workflow.StatusDisabled,
		CronExpression:  &cron,
		SandboxStrategy: workflow.SandboxStrategyPersistent, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
		MaxConcurrentRuns: 1, TimeoutMinutes: 60,
		AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"),
		CreatedByID: 1,
	}
	require.NoError(t, repo.Create(ctx, disabled))

	// API-only workflow (no cron) -> should NOT be found
	apiOnly := &workflow.Workflow{
		OrganizationID: 1, Name: "API Only", Slug: "api-only",
		PromptTemplate: "p",
		ExecutionMode:  workflow.ExecutionModeAutopilot, Status: workflow.StatusEnabled,
		SandboxStrategy: workflow.SandboxStrategyPersistent, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
		MaxConcurrentRuns: 1, TimeoutMinutes: 60,
		AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"),
		CreatedByID: 1,
	}
	require.NoError(t, repo.Create(ctx, apiOnly))

	result, err := repo.FindWorkflowsNeedingNextRun(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "needs-init", result[0].Slug)
}

func TestWorkflowRepository_IncrementRunStats(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRepository(db)
	ctx := context.Background()

	l := &workflow.Workflow{
		OrganizationID: 1, Name: "Stats Workflow", Slug: "stats-workflow",
		PromptTemplate: "p",
		ExecutionMode:  workflow.ExecutionModeAutopilot, Status: workflow.StatusEnabled,
		SandboxStrategy: workflow.SandboxStrategyPersistent, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
		MaxConcurrentRuns: 1, TimeoutMinutes: 60,
		AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"),
		CreatedByID: 1,
	}
	require.NoError(t, repo.Create(ctx, l))

	now := time.Now()

	t.Run("should increment total and successful for completed", func(t *testing.T) {
		err := repo.IncrementRunStats(ctx, l.ID, workflow.RunStatusCompleted, now)
		require.NoError(t, err)

		got, err := repo.GetByID(ctx, l.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, got.TotalRuns)
		assert.Equal(t, 1, got.SuccessfulRuns)
		assert.Equal(t, 0, got.FailedRuns)
	})

	t.Run("should increment total and failed for failed", func(t *testing.T) {
		err := repo.IncrementRunStats(ctx, l.ID, workflow.RunStatusFailed, now)
		require.NoError(t, err)

		got, err := repo.GetByID(ctx, l.ID)
		require.NoError(t, err)
		assert.Equal(t, 2, got.TotalRuns)
		assert.Equal(t, 1, got.SuccessfulRuns)
		assert.Equal(t, 1, got.FailedRuns)
	})

	t.Run("should increment total and failed for timeout", func(t *testing.T) {
		err := repo.IncrementRunStats(ctx, l.ID, workflow.RunStatusTimeout, now)
		require.NoError(t, err)

		got, err := repo.GetByID(ctx, l.ID)
		require.NoError(t, err)
		assert.Equal(t, 3, got.TotalRuns)
		assert.Equal(t, 1, got.SuccessfulRuns)
		assert.Equal(t, 2, got.FailedRuns)
	})

	t.Run("should only increment total for skipped", func(t *testing.T) {
		err := repo.IncrementRunStats(ctx, l.ID, workflow.RunStatusSkipped, now)
		require.NoError(t, err)

		got, err := repo.GetByID(ctx, l.ID)
		require.NoError(t, err)
		assert.Equal(t, 4, got.TotalRuns)
		assert.Equal(t, 1, got.SuccessfulRuns)
		assert.Equal(t, 2, got.FailedRuns)
	})
}

// ========== Org-scoped filtering tests ==========

func TestWorkflowRepository_GetDueCronWorkflows_WithOrgFilter(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRepository(db)
	ctx := context.Background()

	cron := "0 9 * * *"
	pastTime := time.Now().Add(-1 * time.Hour)

	// Due workflow in org 1
	org1Loop := &workflow.Workflow{
		OrganizationID: 1, Name: "Org1 Due", Slug: "org1-due",
		PromptTemplate: "p",
		ExecutionMode:  workflow.ExecutionModeAutopilot, Status: workflow.StatusEnabled,
		CronExpression: &cron, NextRunAt: &pastTime,
		SandboxStrategy: workflow.SandboxStrategyPersistent, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
		MaxConcurrentRuns: 1, TimeoutMinutes: 60,
		AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"),
		CreatedByID: 1,
	}
	require.NoError(t, repo.Create(ctx, org1Loop))

	// Due workflow in org 2
	org2Loop := &workflow.Workflow{
		OrganizationID: 2, Name: "Org2 Due", Slug: "org2-due",
		PromptTemplate: "p",
		ExecutionMode:  workflow.ExecutionModeAutopilot, Status: workflow.StatusEnabled,
		CronExpression: &cron, NextRunAt: &pastTime,
		SandboxStrategy: workflow.SandboxStrategyPersistent, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
		MaxConcurrentRuns: 1, TimeoutMinutes: 60,
		AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"),
		CreatedByID: 2,
	}
	require.NoError(t, repo.Create(ctx, org2Loop))

	// Due workflow in org 3
	org3Loop := &workflow.Workflow{
		OrganizationID: 3, Name: "Org3 Due", Slug: "org3-due",
		PromptTemplate: "p",
		ExecutionMode:  workflow.ExecutionModeAutopilot, Status: workflow.StatusEnabled,
		CronExpression: &cron, NextRunAt: &pastTime,
		SandboxStrategy: workflow.SandboxStrategyPersistent, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
		MaxConcurrentRuns: 1, TimeoutMinutes: 60,
		AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"),
		CreatedByID: 3,
	}
	require.NoError(t, repo.Create(ctx, org3Loop))

	t.Run("nil orgIDs should return all due workflows", func(t *testing.T) {
		result, err := repo.GetDueCronWorkflows(ctx, nil)
		require.NoError(t, err)
		assert.Len(t, result, 3)
	})

	t.Run("should filter to specific org", func(t *testing.T) {
		result, err := repo.GetDueCronWorkflows(ctx, []int64{1})
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "org1-due", result[0].Slug)
	})

	t.Run("should filter to multiple orgs", func(t *testing.T) {
		result, err := repo.GetDueCronWorkflows(ctx, []int64{1, 3})
		require.NoError(t, err)
		assert.Len(t, result, 2)
		slugs := []string{result[0].Slug, result[1].Slug}
		assert.ElementsMatch(t, []string{"org1-due", "org3-due"}, slugs)
	})

	t.Run("should return empty for non-matching orgs", func(t *testing.T) {
		result, err := repo.GetDueCronWorkflows(ctx, []int64{999})
		require.NoError(t, err)
		assert.Len(t, result, 0)
	})
}

func TestWorkflowRepository_FindWorkflowsNeedingNextRun_WithOrgFilter(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRepository(db)
	ctx := context.Background()

	cron := "0 9 * * *"

	// Workflow needing init in org 1
	org1 := &workflow.Workflow{
		OrganizationID: 1, Name: "Org1 Init", Slug: "org1-init",
		PromptTemplate: "p",
		ExecutionMode:  workflow.ExecutionModeAutopilot, Status: workflow.StatusEnabled,
		CronExpression:  &cron,
		SandboxStrategy: workflow.SandboxStrategyPersistent, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
		MaxConcurrentRuns: 1, TimeoutMinutes: 60,
		AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"),
		CreatedByID: 1,
	}
	require.NoError(t, repo.Create(ctx, org1))

	// Workflow needing init in org 2
	org2 := &workflow.Workflow{
		OrganizationID: 2, Name: "Org2 Init", Slug: "org2-init",
		PromptTemplate: "p",
		ExecutionMode:  workflow.ExecutionModeAutopilot, Status: workflow.StatusEnabled,
		CronExpression:  &cron,
		SandboxStrategy: workflow.SandboxStrategyPersistent, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
		MaxConcurrentRuns: 1, TimeoutMinutes: 60,
		AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"),
		CreatedByID: 2,
	}
	require.NoError(t, repo.Create(ctx, org2))

	t.Run("nil orgIDs should return all", func(t *testing.T) {
		result, err := repo.FindWorkflowsNeedingNextRun(ctx, nil)
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("should filter to specific org", func(t *testing.T) {
		result, err := repo.FindWorkflowsNeedingNextRun(ctx, []int64{2})
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "org2-init", result[0].Slug)
	})

	t.Run("should return empty for non-matching orgs", func(t *testing.T) {
		result, err := repo.FindWorkflowsNeedingNextRun(ctx, []int64{999})
		require.NoError(t, err)
		assert.Len(t, result, 0)
	})
}
