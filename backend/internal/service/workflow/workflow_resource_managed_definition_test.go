package workflow

import (
	"context"
	"testing"

	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/stretchr/testify/require"
)

func TestWorkflowServiceRejectsResourceManagedDefinitionMutation(t *testing.T) {
	svc, db := newTestWorkflowService(t)
	ctx := context.Background()
	created, err := svc.Create(ctx, &CreateWorkflowRequest{
		OrganizationID: 1,
		CreatedByID:    1,
		Name:           "Managed workflow",
		Slug:           "managed-workflow",
		PromptTemplate: "original prompt",
	})
	require.NoError(t, err)

	resourceID := int64(21)
	revision := int64(3)
	snapshotID := int64(34)
	require.NoError(t, db.Model(&workflowDomain.Workflow{}).
		Where("id = ?", created.ID).
		Updates(map[string]any{
			"orchestration_resource_id":       resourceID,
			"orchestration_resource_revision": revision,
			"worker_spec_snapshot_id":         snapshotID,
		}).Error)

	t.Run("update", func(t *testing.T) {
		updatedPrompt := "bypassed prompt"
		_, err := svc.Update(ctx, 1, created.Slug, &UpdateWorkflowRequest{
			PromptTemplate: &updatedPrompt,
		})
		require.ErrorIs(t, err, ErrWorkflowManagedByResourceApply)

		stored, getErr := svc.GetBySlug(ctx, 1, created.Slug)
		require.NoError(t, getErr)
		require.Equal(t, "original prompt", stored.PromptTemplate)
	})

	t.Run("delete", func(t *testing.T) {
		err := svc.Delete(ctx, 1, created.Slug)
		require.ErrorIs(t, err, ErrWorkflowManagedByResourceApply)

		_, getErr := svc.GetBySlug(ctx, 1, created.Slug)
		require.NoError(t, getErr)
	})
}
