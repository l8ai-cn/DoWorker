package workflow

import (
	"context"
	"testing"

	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupWorkflowServiceTestDB(t *testing.T) *gorm.DB {
	return testkit.SetupTestDB(t)
}

func newTestWorkflowService(t *testing.T) (*WorkflowService, *gorm.DB) {
	db := setupWorkflowServiceTestDB(t)
	repo := infra.NewWorkflowRepository(db)
	svc := NewWorkflowService(repo)
	return svc, db
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

func TestWorkflowService_Create(t *testing.T) {
	svc, _ := newTestWorkflowService(t)
	ctx := context.Background()

	t.Run("should create workflow with defaults", func(t *testing.T) {
		workflow, err := svc.Create(ctx, &CreateWorkflowRequest{
			OrganizationID: 1,
			CreatedByID:    1,
			Name:           "Daily Review",
			Slug:           "daily-review",
			PromptTemplate: "Review code",
		})
		require.NoError(t, err)
		assert.NotZero(t, workflow.ID)
		assert.Equal(t, "daily-review", workflow.Slug)
		assert.Equal(t, workflowDomain.StatusEnabled, workflow.Status)
		assert.Equal(t, workflowDomain.ExecutionModeAutopilot, workflow.ExecutionMode)
		assert.Equal(t, workflowDomain.SandboxStrategyPersistent, workflow.SandboxStrategy)
		assert.Equal(t, workflowDomain.ConcurrencyPolicySkip, workflow.ConcurrencyPolicy)
		assert.Equal(t, 1, workflow.MaxConcurrentRuns)
		assert.Equal(t, 60, workflow.TimeoutMinutes)
	})

	t.Run("should auto-generate slug from name", func(t *testing.T) {
		workflow, err := svc.Create(ctx, &CreateWorkflowRequest{
			OrganizationID: 1,
			CreatedByID:    1,
			Name:           "My Cool Workflow",
			PromptTemplate: "Do something",
		})
		require.NoError(t, err)
		assert.Equal(t, "my-cool-workflow", workflow.Slug)
	})

	t.Run("should reject invalid slug", func(t *testing.T) {
		_, err := svc.Create(ctx, &CreateWorkflowRequest{
			OrganizationID: 1,
			CreatedByID:    1,
			Name:           "Test",
			Slug:           "AB", // too short
			PromptTemplate: "prompt",
		})
		assert.ErrorIs(t, err, ErrInvalidSlug)
	})
}

func TestWorkflowService_GetBySlug(t *testing.T) {
	svc, _ := newTestWorkflowService(t)
	ctx := context.Background()

	_, err := svc.Create(ctx, &CreateWorkflowRequest{
		OrganizationID: 1, CreatedByID: 1,
		Name: "Test", Slug: "test-get", PromptTemplate: "p",
	})
	require.NoError(t, err)

	t.Run("should return existing workflow", func(t *testing.T) {
		workflow, err := svc.GetBySlug(ctx, 1, "test-get")
		require.NoError(t, err)
		assert.Equal(t, "Test", workflow.Name)
	})

	t.Run("should return ErrWorkflowNotFound for non-existent", func(t *testing.T) {
		_, err := svc.GetBySlug(ctx, 1, "no-such")
		assert.ErrorIs(t, err, ErrWorkflowNotFound)
	})

	t.Run("should return ErrWorkflowNotFound for wrong org", func(t *testing.T) {
		_, err := svc.GetBySlug(ctx, 999, "test-get")
		assert.ErrorIs(t, err, ErrWorkflowNotFound)
	})
}

func TestWorkflowService_Update(t *testing.T) {
	svc, _ := newTestWorkflowService(t)
	ctx := context.Background()

	_, err := svc.Create(ctx, &CreateWorkflowRequest{
		OrganizationID: 1, CreatedByID: 1,
		Name: "Before", Slug: "update-test", PromptTemplate: "original",
	})
	require.NoError(t, err)

	t.Run("should update specified fields", func(t *testing.T) {
		updated, err := svc.Update(ctx, 1, "update-test", &UpdateWorkflowRequest{
			Name:           strPtr("After"),
			PromptTemplate: strPtr("updated prompt"),
			TimeoutMinutes: intPtr(120),
		})
		require.NoError(t, err)
		assert.Equal(t, "After", updated.Name)
		assert.Equal(t, "updated prompt", updated.PromptTemplate)
		assert.Equal(t, 120, updated.TimeoutMinutes)
	})

	t.Run("should return error for non-existent workflow", func(t *testing.T) {
		_, err := svc.Update(ctx, 1, "no-such", &UpdateWorkflowRequest{
			Name: strPtr("X"),
		})
		assert.ErrorIs(t, err, ErrWorkflowNotFound)
	})
}

func TestWorkflowService_Delete(t *testing.T) {
	svc, _ := newTestWorkflowService(t)
	ctx := context.Background()

	_, err := svc.Create(ctx, &CreateWorkflowRequest{
		OrganizationID: 1, CreatedByID: 1,
		Name: "Deletable", Slug: "deletable", PromptTemplate: "p",
	})
	require.NoError(t, err)

	t.Run("should delete existing workflow", func(t *testing.T) {
		err := svc.Delete(ctx, 1, "deletable")
		require.NoError(t, err)

		_, err = svc.GetBySlug(ctx, 1, "deletable")
		assert.ErrorIs(t, err, ErrWorkflowNotFound)
	})

	t.Run("should return error for non-existent", func(t *testing.T) {
		err := svc.Delete(ctx, 1, "no-such")
		assert.ErrorIs(t, err, ErrWorkflowNotFound)
	})
}

func TestWorkflowService_SetStatus(t *testing.T) {
	svc, _ := newTestWorkflowService(t)
	ctx := context.Background()

	_, err := svc.Create(ctx, &CreateWorkflowRequest{
		OrganizationID: 1, CreatedByID: 1,
		Name: "Status Test", Slug: "status-test", PromptTemplate: "p",
	})
	require.NoError(t, err)

	t.Run("should change status to disabled", func(t *testing.T) {
		workflow, err := svc.SetStatus(ctx, 1, "status-test", workflowDomain.StatusDisabled)
		require.NoError(t, err)
		assert.Equal(t, workflowDomain.StatusDisabled, workflow.Status)
	})

	t.Run("should change status back to enabled", func(t *testing.T) {
		workflow, err := svc.SetStatus(ctx, 1, "status-test", workflowDomain.StatusEnabled)
		require.NoError(t, err)
		assert.Equal(t, workflowDomain.StatusEnabled, workflow.Status)
	})
}

func TestWorkflowService_UpdateStats(t *testing.T) {
	svc, _ := newTestWorkflowService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, &CreateWorkflowRequest{
		OrganizationID: 1, CreatedByID: 1,
		Name: "Stats", Slug: "stats-test", PromptTemplate: "p",
	})
	require.NoError(t, err)

	err = svc.UpdateStats(ctx, created.ID, 10, 8, 2)
	require.NoError(t, err)

	got, err := svc.GetBySlug(ctx, 1, "stats-test")
	require.NoError(t, err)
	assert.Equal(t, 10, got.TotalRuns)
	assert.Equal(t, 8, got.SuccessfulRuns)
	assert.Equal(t, 2, got.FailedRuns)
}

func TestWorkflowService_UpdateRuntimeState(t *testing.T) {
	svc, _ := newTestWorkflowService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, &CreateWorkflowRequest{
		OrganizationID: 1, CreatedByID: 1,
		Name: "Runtime", Slug: "runtime-test", PromptTemplate: "p",
	})
	require.NoError(t, err)

	err = svc.UpdateRuntimeState(ctx, created.ID, strPtr("/sandbox/path"), strPtr("pod-abc"))
	require.NoError(t, err)

	got, err := svc.GetBySlug(ctx, 1, "runtime-test")
	require.NoError(t, err)
	assert.NotNil(t, got.SandboxPath)
	assert.Equal(t, "/sandbox/path", *got.SandboxPath)
	assert.NotNil(t, got.LastPodKey)
	assert.Equal(t, "pod-abc", *got.LastPodKey)
}

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase name", "daily review", "daily-review"},
		{"mixed case", "My Cool Workflow", "my-cool-workflow"},
		{"special chars", "PR Review (v2)", "pr-review-v2"},
		{"short name padded", "a", "a-workflow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, generateSlug(tt.input))
		})
	}

	t.Run("chinese name generates valid slug", func(t *testing.T) {
		slug := generateSlug("每日代码审查")
		assert.True(t, isValidSlug(slug), "slug %q should match regex", slug)
		assert.True(t, len(slug) >= 3, "slug should be at least 3 chars")
	})

	t.Run("mixed chinese and ascii", func(t *testing.T) {
		slug := generateSlug("每日review任务")
		assert.True(t, isValidSlug(slug), "slug %q should match regex", slug)
		assert.Contains(t, slug, "review")
	})

	t.Run("emoji name generates valid slug", func(t *testing.T) {
		slug := generateSlug("🚀 deploy bot")
		assert.True(t, isValidSlug(slug), "slug %q should match regex", slug)
	})

	t.Run("pure unicode generates timestamp-based slug", func(t *testing.T) {
		slug := generateSlug("日本語テスト")
		assert.True(t, isValidSlug(slug), "slug %q should match regex", slug)
		assert.Contains(t, slug, "workflow-")
	})

	t.Run("reserved word falls back to timestamp", func(t *testing.T) {
		slug := generateSlug("admin")
		assert.True(t, isValidSlug(slug), "slug %q should be valid", slug)
		assert.NotEqual(t, "admin", slug, "reserved word must not be returned verbatim")
	})

	t.Run("reserved word with prefix collisions falls back", func(t *testing.T) {
		slug := generateSlug("api")
		assert.True(t, isValidSlug(slug), "slug %q should be valid", slug)
		assert.NotEqual(t, "api", slug)
	})
}

func TestWorkflowService_Create_ChineseName(t *testing.T) {
	svc, _ := newTestWorkflowService(t)
	ctx := context.Background()

	workflow, err := svc.Create(ctx, &CreateWorkflowRequest{
		OrganizationID: 1,
		CreatedByID:    1,
		Name:           "每日代码审查",
		PromptTemplate: "Review code daily",
	})
	require.NoError(t, err)
	assert.NotZero(t, workflow.ID)
	assert.True(t, isValidSlug(workflow.Slug), "auto-generated slug %q should be valid", workflow.Slug)
}
