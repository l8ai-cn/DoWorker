package workflow

import (
	"testing"

	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildResourceManagedWorkflowPodRequestUsesRunPinsOnly(t *testing.T) {
	resourceID := int64(90)
	resourceRevision := int64(3)
	snapshotID := int64(42)
	modelResourceID := int64(77)
	workflow := &workflowDomain.Workflow{
		OrganizationID: 1, AgentSlug: "legacy-agent",
		ModelResourceID: &modelResourceID, RunnerID: &modelResourceID,
		OrchestrationResourceID:       &resourceID,
		OrchestrationResourceRevision: &resourceRevision,
		WorkerSpecSnapshotID:          &snapshotID,
	}
	run := &workflowDomain.WorkflowRun{
		OrchestrationResourceID:       &resourceID,
		OrchestrationResourceRevision: &resourceRevision,
		WorkerSpecSnapshotID:          &snapshotID,
	}

	request, err := buildWorkflowRunPodRequest(
		workflow,
		run,
		9,
		"Review authorization",
		"PROMPT \"legacy\"",
		"",
		false,
	)

	require.NoError(t, err)
	require.NotNil(t, request.WorkerSpecSnapshotID)
	assert.Equal(t, snapshotID, *request.WorkerSpecSnapshotID)
	require.NotNil(t, request.WorkerSpecPromptOverride)
	assert.Equal(t, "Review authorization", *request.WorkerSpecPromptOverride)
	assert.Empty(t, request.AgentSlug)
	assert.Zero(t, request.RunnerID)
	assert.Nil(t, request.ModelResourceID)
	assert.Nil(t, request.AgentfileLayer)
}

func TestBuildResourceManagedWorkflowPodRequestRejectsMismatchedRunPin(t *testing.T) {
	resourceID := int64(90)
	resourceRevision := int64(3)
	snapshotID := int64(42)
	otherSnapshotID := int64(43)
	workflow := &workflowDomain.Workflow{
		OrganizationID:                1,
		OrchestrationResourceID:       &resourceID,
		OrchestrationResourceRevision: &resourceRevision,
		WorkerSpecSnapshotID:          &snapshotID,
	}
	run := &workflowDomain.WorkflowRun{
		OrchestrationResourceID:       &resourceID,
		OrchestrationResourceRevision: &resourceRevision,
		WorkerSpecSnapshotID:          &otherSnapshotID,
	}

	_, err := buildWorkflowRunPodRequest(
		workflow,
		run,
		9,
		"Review authorization",
		"",
		"",
		false,
	)

	require.ErrorIs(t, err, ErrWorkflowResourceBindingCorrupt)
}
