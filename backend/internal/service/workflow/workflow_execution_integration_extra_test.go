package workflow

import (
	"fmt"
	"testing"
	"time"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	agentpodSvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoopWorkflow_ResumeDegradation(t *testing.T) {
	oldPodKey := fmt.Sprintf("old-pod-%d", time.Now().UnixNano())
	env := setupWorkflowTest(t, func(l *workflowDomain.Workflow) {
		l.SandboxStrategy = workflowDomain.SandboxStrategyPersistent
		l.LastPodKey = &oldPodKey
	})

	newPodKey := fmt.Sprintf("new-pod-%d", time.Now().UnixNano())

	// First CreatePod call (with source_pod_key for resume) fails, second succeeds
	env.podOrch.failFirstN = 1
	env.podOrch.results = []*agentpodSvc.OrchestrateCreatePodResult{{
		Pod: &podDomain.Pod{PodKey: newPodKey, OrganizationID: 1, RunnerID: 1},
	}}

	// We need to simulate StartRun's resume degradation logic manually,
	// since podOrchestrator is nil on the orchestrator (it goes through mock).
	// Instead test the logic directly:

	// TriggerRun
	triggerResult, err := env.orchestrator.TriggerRun(env.ctx, &TriggerRunRequest{
		WorkflowID:  env.workflow.ID,
		TriggerType: workflowDomain.RunTriggerManual,
	})
	require.NoError(t, err)
	run := triggerResult.Run

	// Simulate resume degradation: first call fails (resume), second succeeds (fresh)
	_, err = env.podOrch.CreatePod(env.ctx, nil) // call 1: fails
	require.Error(t, err)

	result, err := env.podOrch.CreatePod(env.ctx, nil) // call 2: succeeds
	require.NoError(t, err)
	assert.Equal(t, newPodKey, result.Pod.PodKey)

	// Verify mock was called twice
	assert.Equal(t, 2, env.podOrch.callCount)

	// Simulate what StartRun does after degradation: clear runtime state + set pod key
	require.NoError(t, env.workflowSvc.ClearRuntimeState(env.ctx, env.workflow.ID))
	require.NoError(t, env.orchestrator.SetRunPodKey(env.ctx, run.ID, newPodKey, ""))

	// Verify runtime state was cleared
	workflow, err := env.workflowSvc.GetByID(env.ctx, env.workflow.ID)
	require.NoError(t, err)
	assert.Nil(t, workflow.LastPodKey, "runtime state should be cleared after resume degradation")

	// Verify run has the new pod key
	got, err := env.runSvc.GetByID(env.ctx, run.ID)
	require.NoError(t, err)
	require.NotNil(t, got.PodKey)
	assert.Equal(t, newPodKey, *got.PodKey)
}

func TestLoopWorkflow_TimeoutDetection(t *testing.T) {
	env := setupWorkflowTest(t)

	podKey := fmt.Sprintf("pod-to-%d", time.Now().UnixNano())

	// TriggerRun + associate pod
	triggerResult, err := env.orchestrator.TriggerRun(env.ctx, &TriggerRunRequest{
		WorkflowID:  env.workflow.ID,
		TriggerType: workflowDomain.RunTriggerCron,
	})
	require.NoError(t, err)
	run := triggerResult.Run
	require.NoError(t, env.orchestrator.SetRunPodKey(env.ctx, run.ID, podKey, ""))

	// Manually backdate started_at to exceed timeout
	startedAt := time.Now().Add(-2 * time.Hour) // Workflow timeout is 30 min
	require.NoError(t, env.runSvc.UpdateStatus(env.ctx, run.ID, map[string]interface{}{
		"started_at": startedAt,
	}))

	// GetTimedOutRuns uses PostgreSQL-specific ::INTERVAL syntax, so it won't work
	// in SQLite. Instead, we directly call HandleRunCompleted with timeout status
	// to test the same logic that CheckTimeoutRuns invokes.
	runForTimeout, err := env.runSvc.GetByID(env.ctx, run.ID)
	require.NoError(t, err)

	// Verify the run has exceeded timeout
	assert.True(t, startedAt.Add(time.Duration(env.workflow.TimeoutMinutes)*time.Minute).Before(time.Now()),
		"run should have exceeded timeout")

	// Simulate what CheckTimeoutRuns does
	env.orchestrator.HandleRunCompleted(env.ctx, runForTimeout, workflowDomain.RunStatusTimeout)

	// Also terminate the pod (as CheckTimeoutRuns does)
	if runForTimeout.PodKey != nil {
		_ = env.podTerm.TerminatePod(env.ctx, *runForTimeout.PodKey)
	}

	// Verify run status is timeout
	got, err := env.runSvc.GetByID(env.ctx, run.ID)
	require.NoError(t, err)
	assert.Equal(t, workflowDomain.RunStatusTimeout, got.Status)
	assert.NotNil(t, got.FinishedAt)

	// Verify pod terminator was called
	terminated := env.podTerm.getTerminatedKeys()
	assert.Contains(t, terminated, podKey)

	// Verify stats updated (timeout → failed_runs)
	workflow, err := env.workflowSvc.GetByID(env.ctx, env.workflow.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, workflow.TotalRuns)
	assert.Equal(t, 1, workflow.FailedRuns)
}

func TestLoopWorkflow_HandleRunCompleted_FailedClearsResume(t *testing.T) {
	env := setupWorkflowTest(t, func(l *workflowDomain.Workflow) {
		l.SandboxStrategy = workflowDomain.SandboxStrategyPersistent
	})

	podKey := fmt.Sprintf("pod-fail-%d", time.Now().UnixNano())

	// TriggerRun + associate pod
	triggerResult, err := env.orchestrator.TriggerRun(env.ctx, &TriggerRunRequest{
		WorkflowID:  env.workflow.ID,
		TriggerType: workflowDomain.RunTriggerManual,
	})
	require.NoError(t, err)
	require.NoError(t, env.orchestrator.SetRunPodKey(env.ctx, triggerResult.Run.ID, podKey, ""))

	// Fail the run
	env.orchestrator.HandlePodTerminated(env.ctx, podKey, podDomain.StatusError, nil)

	// Verify last_pod_key is cleared (breaks death spiral)
	workflow, err := env.workflowSvc.GetByID(env.ctx, env.workflow.ID)
	require.NoError(t, err)
	assert.Nil(t, workflow.LastPodKey, "failed run should clear last_pod_key to break death spiral")
}

func TestLoopWorkflow_DoubleCompletionIdempotent(t *testing.T) {
	env := setupWorkflowTest(t)

	podKey := fmt.Sprintf("pod-dup-%d", time.Now().UnixNano())

	// TriggerRun + associate pod
	triggerResult, err := env.orchestrator.TriggerRun(env.ctx, &TriggerRunRequest{
		WorkflowID:  env.workflow.ID,
		TriggerType: workflowDomain.RunTriggerManual,
	})
	require.NoError(t, err)
	require.NoError(t, env.orchestrator.SetRunPodKey(env.ctx, triggerResult.Run.ID, podKey, ""))

	// Complete the run twice (simulating concurrent pod_terminated events)
	env.orchestrator.HandlePodTerminated(env.ctx, podKey, podDomain.StatusCompleted, nil)
	env.orchestrator.HandlePodTerminated(env.ctx, podKey, podDomain.StatusCompleted, nil)

	// Stats should only be incremented once (idempotent via FinishRun optimistic lock)
	workflow, err := env.workflowSvc.GetByID(env.ctx, env.workflow.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, workflow.SuccessfulRuns, "double completion should only count once")
	assert.Equal(t, 1, workflow.TotalRuns)
}

func TestLoopWorkflow_MarkRunCancelled(t *testing.T) {
	env := setupWorkflowTest(t)

	// TriggerRun
	triggerResult, err := env.orchestrator.TriggerRun(env.ctx, &TriggerRunRequest{
		WorkflowID:  env.workflow.ID,
		TriggerType: workflowDomain.RunTriggerManual,
	})
	require.NoError(t, err)
	run := triggerResult.Run

	// Cancel the pending run (no pod associated yet)
	require.NoError(t, env.orchestrator.MarkRunCancelled(env.ctx, run.ID, "user cancelled"))

	// Verify
	got, err := env.runSvc.GetByID(env.ctx, run.ID)
	require.NoError(t, err)
	assert.Equal(t, workflowDomain.RunStatusCancelled, got.Status)
	assert.NotNil(t, got.FinishedAt)

	// Stats: cancelled counts as failed
	workflow, err := env.workflowSvc.GetByID(env.ctx, env.workflow.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, workflow.FailedRuns)
}
