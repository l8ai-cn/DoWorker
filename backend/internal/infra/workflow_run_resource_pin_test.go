package infra

import (
	"context"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTriggerRunAtomicPinsWorkflowResourceRevisionAndSnapshot(t *testing.T) {
	db := setupLoopTestDB(t)
	repo := NewWorkflowRunRepository(db)
	workflowRepo := NewWorkflowRepository(db)
	resourceID := int64(90)
	resourceRevision := int64(3)
	snapshotID := int64(42)
	row := &workflow.Workflow{
		OrganizationID: 1, Name: "Nightly", Slug: "nightly",
		PromptTemplate:    "Review authorization",
		ExecutionMode:     workflow.ExecutionModeDirect,
		Status:            workflow.StatusEnabled,
		SandboxStrategy:   workflow.SandboxStrategyFresh,
		ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
		MaxConcurrentRuns: 1, TimeoutMinutes: 60,
		AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"),
		CreatedByID:                   1,
		OrchestrationResourceID:       &resourceID,
		OrchestrationResourceRevision: &resourceRevision,
		WorkerSpecSnapshotID:          &snapshotID,
	}
	require.NoError(t, workflowRepo.Create(context.Background(), row))

	result, err := repo.TriggerRunAtomic(
		context.Background(),
		&workflow.TriggerRunAtomicParams{
			WorkflowID: row.ID, TriggerType: workflow.RunTriggerManual,
			TriggerSource: "test",
		},
	)

	require.NoError(t, err)
	require.NotNil(t, result.Run)
	assert.Equal(t, &resourceID, result.Run.OrchestrationResourceID)
	assert.Equal(t, &resourceRevision, result.Run.OrchestrationResourceRevision)
	assert.Equal(t, &snapshotID, result.Run.WorkerSpecSnapshotID)
}

func TestOlderWorkflowRunKeepsPinsAfterWorkflowRevisionUpdate(t *testing.T) {
	db := setupLoopTestDB(t)
	runRepo := NewWorkflowRunRepository(db)
	workflowRepo := NewWorkflowRepository(db)
	resourceID := int64(90)
	firstRevision := int64(3)
	firstSnapshot := int64(42)
	row := &workflow.Workflow{
		OrganizationID: 1, Name: "Nightly", Slug: "nightly-pins",
		PromptTemplate:    "Review authorization",
		ExecutionMode:     workflow.ExecutionModeDirect,
		Status:            workflow.StatusEnabled,
		SandboxStrategy:   workflow.SandboxStrategyFresh,
		ConcurrencyPolicy: workflow.ConcurrencyPolicySkip,
		MaxConcurrentRuns: 2, TimeoutMinutes: 60,
		AutopilotConfig: []byte("{}"), ConfigOverrides: []byte("{}"),
		CreatedByID:                   1,
		OrchestrationResourceID:       &resourceID,
		OrchestrationResourceRevision: &firstRevision,
		WorkerSpecSnapshotID:          &firstSnapshot,
	}
	require.NoError(t, workflowRepo.Create(context.Background(), row))
	first, err := runRepo.TriggerRunAtomic(
		context.Background(),
		&workflow.TriggerRunAtomicParams{
			WorkflowID: row.ID, TriggerType: workflow.RunTriggerManual,
			TriggerSource: "first",
		},
	)
	require.NoError(t, err)

	secondRevision := int64(4)
	secondSnapshot := int64(43)
	require.NoError(t, workflowRepo.Update(context.Background(), row.ID, map[string]any{
		"orchestration_resource_revision": secondRevision,
		"worker_spec_snapshot_id":         secondSnapshot,
	}))
	second, err := runRepo.TriggerRunAtomic(
		context.Background(),
		&workflow.TriggerRunAtomicParams{
			WorkflowID: row.ID, TriggerType: workflow.RunTriggerManual,
			TriggerSource: "second",
		},
	)
	require.NoError(t, err)
	persistedFirst, err := runRepo.GetByID(context.Background(), first.Run.ID)
	require.NoError(t, err)

	assert.Equal(t, &firstRevision, persistedFirst.OrchestrationResourceRevision)
	assert.Equal(t, &firstSnapshot, persistedFirst.WorkerSpecSnapshotID)
	assert.Equal(t, &secondRevision, second.Run.OrchestrationResourceRevision)
	assert.Equal(t, &secondSnapshot, second.Run.WorkerSpecSnapshotID)
}
