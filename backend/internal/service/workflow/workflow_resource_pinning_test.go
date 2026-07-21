package workflow

import (
	"testing"

	workflowDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildResourceManagedWorkflowPodRequestUsesRunPinsOnly(t *testing.T) {
	resourceID := int64(90)
	resourceRevision := int64(3)
	snapshotID := int64(42)
	modelResourceID := int64(77)
	workflow := &workflowDomain.Workflow{
		OrganizationID: 1, Name: "Nightly", Slug: "nightly",
		CreatedByID: 9, AgentSlug: "legacy-agent",
		ExecutionMode:   workflowDomain.ExecutionModeDirect,
		SandboxStrategy: workflowDomain.SandboxStrategyFresh,
		TimeoutMinutes:  60, IdleTimeoutSec: 30,
		ModelResourceID: &modelResourceID, RunnerID: &modelResourceID,
		OrchestrationResourceID:       &resourceID,
		OrchestrationResourceRevision: &resourceRevision,
		WorkerSpecSnapshotID:          &snapshotID,
	}
	manifest, err := workflowDomain.PinWorkflowRunExecutionManifest(workflow)
	require.NoError(t, err)
	run := &workflowDomain.WorkflowRun{
		OrganizationID:                1,
		OrchestrationResourceID:       &resourceID,
		OrchestrationResourceRevision: &resourceRevision,
		WorkerSpecSnapshotID:          &snapshotID,
		ResolvedPrompt:                strPtr("Pinned authorization review"),
		ExecutionManifest:             manifest,
	}

	request, err := buildWorkflowRunPodRequest(
		run,
		9,
	)

	require.NoError(t, err)
	require.NotNil(t, request.WorkerSpecSnapshotID)
	assert.Equal(t, snapshotID, *request.WorkerSpecSnapshotID)
	require.NotNil(t, request.WorkerSpecPromptOverride)
	assert.Equal(t, "Pinned authorization review", *request.WorkerSpecPromptOverride)
	assert.Empty(t, request.AgentSlug)
	assert.Zero(t, request.RunnerID)
	assert.Nil(t, request.ModelResourceID)
	assert.Nil(t, request.AgentfileLayer)
}

func TestBuildResourceManagedWorkflowPodRequestIgnoresNewerWorkflowRevision(
	t *testing.T,
) {
	resourceID := int64(90)
	resourceRevision := int64(3)
	snapshotID := int64(42)
	pinnedWorkflow := &workflowDomain.Workflow{
		OrganizationID:                1,
		Name:                          "Nightly",
		Slug:                          "nightly",
		CreatedByID:                   9,
		ExecutionMode:                 workflowDomain.ExecutionModeDirect,
		SandboxStrategy:               workflowDomain.SandboxStrategyFresh,
		TimeoutMinutes:                60,
		IdleTimeoutSec:                30,
		OrchestrationResourceID:       &resourceID,
		OrchestrationResourceRevision: &resourceRevision,
		WorkerSpecSnapshotID:          &snapshotID,
	}
	manifest, err := workflowDomain.PinWorkflowRunExecutionManifest(
		pinnedWorkflow,
	)
	require.NoError(t, err)
	run := &workflowDomain.WorkflowRun{
		OrganizationID:                1,
		OrchestrationResourceID:       &resourceID,
		OrchestrationResourceRevision: &resourceRevision,
		WorkerSpecSnapshotID:          &snapshotID,
		ResolvedPrompt:                strPtr("Review authorization"),
		ExecutionManifest:             manifest,
	}
	request, err := buildWorkflowRunPodRequest(
		run,
		9,
	)

	require.NoError(t, err)
	require.NotNil(t, request.WorkerSpecSnapshotID)
	assert.Equal(t, snapshotID, *request.WorkerSpecSnapshotID)
}

func TestBuildPersistentWorkflowPodRequestUsesLineageOnly(t *testing.T) {
	resourceID := int64(90)
	resourceRevision := int64(3)
	snapshotID := int64(42)
	sourcePodKey := "previous-workflow-pod"
	workflow := &workflowDomain.Workflow{
		OrganizationID:                1,
		Name:                          "Nightly",
		Slug:                          "nightly",
		CreatedByID:                   9,
		ExecutionMode:                 workflowDomain.ExecutionModeDirect,
		SandboxStrategy:               workflowDomain.SandboxStrategyPersistent,
		SessionPersistence:            true,
		LastPodKey:                    &sourcePodKey,
		TimeoutMinutes:                60,
		IdleTimeoutSec:                30,
		OrchestrationResourceID:       &resourceID,
		OrchestrationResourceRevision: &resourceRevision,
		WorkerSpecSnapshotID:          &snapshotID,
	}
	manifest, err := workflowDomain.PinWorkflowRunExecutionManifest(workflow)
	require.NoError(t, err)
	run := &workflowDomain.WorkflowRun{
		OrganizationID:                1,
		OrchestrationResourceID:       &resourceID,
		OrchestrationResourceRevision: &resourceRevision,
		WorkerSpecSnapshotID:          &snapshotID,
		ResolvedPrompt:                strPtr("Continue the nightly review"),
		ExecutionManifest:             manifest,
	}

	request, err := buildWorkflowRunPodRequest(run, 9)

	require.NoError(t, err)
	assert.Equal(t, sourcePodKey, request.SourcePodKey)
	assert.Nil(t, request.WorkerSpecSnapshotID)
	require.NotNil(t, request.WorkerSpecPromptOverride)
	assert.Equal(t, "Continue the nightly review", *request.WorkerSpecPromptOverride)
}

func TestBuildWorkflowPodRequestRejectsLegacyWorkflow(t *testing.T) {
	prompt := "legacy prompt"
	run := &workflowDomain.WorkflowRun{ResolvedPrompt: &prompt}

	_, err := buildWorkflowRunPodRequest(
		run,
		9,
	)

	require.ErrorIs(t, err, ErrWorkflowResourceBindingCorrupt)
}
