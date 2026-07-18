package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPinWorkflowRunExecutionManifestCapturesExecutionSemantics(
	t *testing.T,
) {
	callbackURL := "https://callbacks.example.com/workflow"
	ticketID := int64(71)
	lastPodKey := "pod-previous"
	item := &Workflow{
		OrganizationID: 9,
		Name:           "Nightly review",
		Slug:           "nightly-review",
		CreatedByID:    17,
		ExecutionMode:  ExecutionModeAutopilot,
		AutopilotConfig: []byte(`{
			"max_iterations": 8,
			"same_error_threshold": 3
		}`),
		SandboxStrategy:    SandboxStrategyPersistent,
		SessionPersistence: true,
		LastPodKey:         &lastPodKey,
		CallbackURL:        &callbackURL,
		TicketID:           &ticketID,
		MaxRetainedRuns:    25,
		TimeoutMinutes:     90,
		IdleTimeoutSec:     45,
	}

	content, err := PinWorkflowRunExecutionManifest(item)
	require.NoError(t, err)
	run := &WorkflowRun{ExecutionManifest: content}
	pinned, err := run.PinnedExecution()
	require.NoError(t, err)

	assert.Equal(t, 1, pinned.Version)
	assert.Equal(t, int64(9), pinned.OrganizationID)
	assert.Equal(t, "Nightly review", pinned.WorkflowName)
	assert.Equal(t, "nightly-review", pinned.WorkflowSlug)
	assert.Equal(t, int64(17), pinned.CreatedByID)
	assert.Equal(t, ExecutionModeAutopilot, pinned.ExecutionMode)
	assert.Equal(t, int32(8), pinned.Autopilot.MaxIterations)
	assert.Equal(t, int32(3), pinned.Autopilot.SameErrorThreshold)
	assert.Equal(t, SandboxStrategyPersistent, pinned.SandboxStrategy)
	assert.True(t, pinned.SessionPersistence)
	assert.Equal(t, "pod-previous", pinned.SourcePodKey)
	assert.Equal(t, callbackURL, pinned.CallbackURL)
	assert.Equal(t, &ticketID, pinned.TicketID)
	assert.Equal(t, 25, pinned.MaxRetainedRuns)
	assert.Equal(t, 90, pinned.TimeoutMinutes)
	assert.Equal(t, 45, pinned.IdleTimeoutSeconds)
}

func TestPinnedWorkflowRunExecutionManifestRejectsMissingOrCorruptData(
	t *testing.T,
) {
	_, err := (&WorkflowRun{}).PinnedExecution()
	require.ErrorIs(t, err, ErrWorkflowRunExecutionManifestRequired)

	run := &WorkflowRun{ExecutionManifest: []byte(`{"version":1}`)}
	_, err = run.PinnedExecution()
	require.ErrorIs(t, err, ErrWorkflowRunExecutionManifestCorrupt)
}
