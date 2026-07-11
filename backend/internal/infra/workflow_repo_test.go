package infra

import (
	"context"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowRepository_Create(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRepository(db)
	ctx := context.Background()

	l := &workflow.Workflow{
		OrganizationID:    1,
		Name:              "Test Workflow",
		Slug:              "test-workflow",
		PromptTemplate:    "Review code in {{branch}}",
		ExecutionMode:     workflow.ExecutionModeAutopilot,
		Status:            workflow.StatusEnabled,
		SandboxStrategy:   workflow.SandboxStrategyPersistent,
		ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
		MaxConcurrentRuns: 1,
		TimeoutMinutes:    60,
		AutopilotConfig:   []byte("{}"),
		ConfigOverrides:   []byte("{}"),
		CreatedByID:       1,
	}

	err := repo.Create(ctx, l)
	require.NoError(t, err)
	assert.NotZero(t, l.ID)
}

func TestWorkflowRepository_GetByID(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRepository(db)
	ctx := context.Background()

	// Seed
	l := &workflow.Workflow{
		OrganizationID: 1, Name: "Test", Slug: "test",
		PromptTemplate: "prompt",
		ExecutionMode:  workflow.ExecutionModeAutopilot, Status: workflow.StatusEnabled,
		SandboxStrategy: workflow.SandboxStrategyPersistent, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
		MaxConcurrentRuns: 1, TimeoutMinutes: 60,
		AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"),
		CreatedByID: 1,
	}
	require.NoError(t, repo.Create(ctx, l))

	t.Run("should return workflow by ID", func(t *testing.T) {
		got, err := repo.GetByID(ctx, l.ID)
		require.NoError(t, err)
		assert.Equal(t, "test", got.Slug)
		assert.Equal(t, "Test", got.Name)
	})

	t.Run("should return ErrNotFound for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 99999)
		assert.ErrorIs(t, err, workflow.ErrNotFound)
	})
}

func TestWorkflowRepository_GetBySlug(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRepository(db)
	ctx := context.Background()

	l := &workflow.Workflow{
		OrganizationID: 1, Name: "My Workflow", Slug: "my-workflow",
		PromptTemplate: "prompt",
		ExecutionMode:  workflow.ExecutionModeAutopilot, Status: workflow.StatusEnabled,
		SandboxStrategy: workflow.SandboxStrategyPersistent, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
		MaxConcurrentRuns: 1, TimeoutMinutes: 60,
		AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"),
		CreatedByID: 1,
	}
	require.NoError(t, repo.Create(ctx, l))

	t.Run("should return workflow by org_id and slug", func(t *testing.T) {
		got, err := repo.GetBySlug(ctx, 1, "my-workflow")
		require.NoError(t, err)
		assert.Equal(t, "My Workflow", got.Name)
	})

	t.Run("should return ErrNotFound for different org", func(t *testing.T) {
		_, err := repo.GetBySlug(ctx, 999, "my-workflow")
		assert.ErrorIs(t, err, workflow.ErrNotFound)
	})

	t.Run("should return ErrNotFound for non-existent slug", func(t *testing.T) {
		_, err := repo.GetBySlug(ctx, 1, "no-such-workflow")
		assert.ErrorIs(t, err, workflow.ErrNotFound)
	})
}

func TestWorkflowRepository_List(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRepository(db)
	ctx := context.Background()

	// Seed multiple workflows
	cron := "0 9 * * *"
	workflows := []*workflow.Workflow{
		{OrganizationID: 1, Name: "Workflow A", Slug: "workflow-a", Status: workflow.StatusEnabled,
			ExecutionMode: workflow.ExecutionModeAutopilot, CronExpression: &cron,
			PromptTemplate:  "p",
			SandboxStrategy: workflow.SandboxStrategyPersistent, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
			MaxConcurrentRuns: 1, TimeoutMinutes: 60,
			AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"), CreatedByID: 1},
		{OrganizationID: 1, Name: "Workflow B", Slug: "workflow-b", Status: workflow.StatusEnabled,
			ExecutionMode:   workflow.ExecutionModeDirect,
			PromptTemplate:  "p",
			SandboxStrategy: workflow.SandboxStrategyFresh, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
			MaxConcurrentRuns: 1, TimeoutMinutes: 60,
			AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"), CreatedByID: 1},
		{OrganizationID: 1, Name: "Workflow C", Slug: "workflow-c", Status: workflow.StatusDisabled,
			ExecutionMode:   workflow.ExecutionModeAutopilot,
			PromptTemplate:  "p",
			SandboxStrategy: workflow.SandboxStrategyPersistent, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
			MaxConcurrentRuns: 1, TimeoutMinutes: 60,
			AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"), CreatedByID: 1},
		{OrganizationID: 1, Name: "Workflow D", Slug: "workflow-d", Status: workflow.StatusArchived,
			ExecutionMode:   workflow.ExecutionModeDirect,
			PromptTemplate:  "p",
			SandboxStrategy: workflow.SandboxStrategyFresh, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
			MaxConcurrentRuns: 1, TimeoutMinutes: 60,
			AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"), CreatedByID: 1},
		{OrganizationID: 2, Name: "Other Org Workflow", Slug: "other", Status: workflow.StatusEnabled,
			ExecutionMode:   workflow.ExecutionModeAutopilot,
			PromptTemplate:  "p",
			SandboxStrategy: workflow.SandboxStrategyPersistent, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
			MaxConcurrentRuns: 1, TimeoutMinutes: 60,
			AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"), CreatedByID: 2},
	}
	for _, l := range workflows {
		require.NoError(t, repo.Create(ctx, l))
	}

	t.Run("should list non-archived workflows by default", func(t *testing.T) {
		result, total, err := repo.List(ctx, &workflow.ListWorkflowsFilter{OrganizationID: 1})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total) // A, B, C (not D=archived)
		assert.Len(t, result, 3)
	})

	t.Run("should filter by status", func(t *testing.T) {
		result, total, err := repo.List(ctx, &workflow.ListWorkflowsFilter{
			OrganizationID: 1,
			Status:         workflow.StatusEnabled,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total) // A, B
		assert.Len(t, result, 2)
	})

	t.Run("should filter by execution mode", func(t *testing.T) {
		result, total, err := repo.List(ctx, &workflow.ListWorkflowsFilter{
			OrganizationID: 1,
			ExecutionMode:  workflow.ExecutionModeDirect,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total) // B (not D=archived)
		assert.Len(t, result, 1)
		assert.Equal(t, "workflow-b", result[0].Slug)
	})

	t.Run("should filter by cron enabled", func(t *testing.T) {
		enabled := true
		result, _, err := repo.List(ctx, &workflow.ListWorkflowsFilter{
			OrganizationID: 1,
			CronEnabled:    &enabled,
		})
		require.NoError(t, err)
		assert.Len(t, result, 1) // Only Workflow A has cron
		assert.Equal(t, "workflow-a", result[0].Slug)
	})

	t.Run("should respect limit and offset", func(t *testing.T) {
		result, total, err := repo.List(ctx, &workflow.ListWorkflowsFilter{
			OrganizationID: 1,
			Limit:          2,
			Offset:         0,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total) // total count is unaffected
		assert.Len(t, result, 2)
	})

	t.Run("should isolate by organization", func(t *testing.T) {
		result, total, err := repo.List(ctx, &workflow.ListWorkflowsFilter{OrganizationID: 2})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, result, 1)
		assert.Equal(t, "other", result[0].Slug)
	})
}

func TestWorkflowRepository_Update(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRepository(db)
	ctx := context.Background()

	l := &workflow.Workflow{
		OrganizationID: 1, Name: "Original", Slug: "original",
		PromptTemplate: "prompt",
		ExecutionMode:  workflow.ExecutionModeAutopilot, Status: workflow.StatusEnabled,
		SandboxStrategy: workflow.SandboxStrategyPersistent, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
		MaxConcurrentRuns: 1, TimeoutMinutes: 60,
		AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"),
		CreatedByID: 1,
	}
	require.NoError(t, repo.Create(ctx, l))

	err := repo.Update(ctx, l.ID, map[string]interface{}{
		"name":            "Updated",
		"status":          workflow.StatusDisabled,
		"total_runs":      5,
		"successful_runs": 3,
	})
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, l.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated", got.Name)
	assert.Equal(t, workflow.StatusDisabled, got.Status)
	assert.Equal(t, 5, got.TotalRuns)
	assert.Equal(t, 3, got.SuccessfulRuns)
}

func TestWorkflowRepository_Delete(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRepository(db)
	ctx := context.Background()

	l := &workflow.Workflow{
		OrganizationID: 1, Name: "To Delete", Slug: "to-delete",
		PromptTemplate: "prompt",
		ExecutionMode:  workflow.ExecutionModeAutopilot, Status: workflow.StatusEnabled,
		SandboxStrategy: workflow.SandboxStrategyPersistent, ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
		MaxConcurrentRuns: 1, TimeoutMinutes: 60,
		AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"),
		CreatedByID: 1,
	}
	require.NoError(t, repo.Create(ctx, l))

	t.Run("should delete existing workflow", func(t *testing.T) {
		affected, err := repo.Delete(ctx, 1, "to-delete")
		require.NoError(t, err)
		assert.Equal(t, int64(1), affected)

		_, err = repo.GetBySlug(ctx, 1, "to-delete")
		assert.ErrorIs(t, err, workflow.ErrNotFound)
	})

	t.Run("should return 0 affected for non-existent", func(t *testing.T) {
		affected, err := repo.Delete(ctx, 1, "no-such")
		require.NoError(t, err)
		assert.Equal(t, int64(0), affected)
	})
}
