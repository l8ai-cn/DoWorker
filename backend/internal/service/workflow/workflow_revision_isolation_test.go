package workflow

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowRunCompletionKeepsPinnedFreshSandboxAfterRevisionChange(
	t *testing.T,
) {
	env := setupWorkflowTest(t, func(item *workflowDomain.Workflow) {
		item.SandboxStrategy = workflowDomain.SandboxStrategyFresh
		item.SessionPersistence = false
	})
	result, err := env.orchestrator.TriggerRun(env.ctx, &TriggerRunRequest{
		WorkflowID: env.workflow.ID, TriggerType: workflowDomain.RunTriggerManual,
	})
	require.NoError(t, err)
	podKey := "revision-isolation-pod"
	require.NoError(t, env.orchestrator.SetRunPodKey(
		env.ctx,
		result.Run.ID,
		podKey,
		"",
	))

	nextRevision := *env.workflow.OrchestrationResourceRevision + 1
	require.NoError(t, env.workflowSvc.repo.Update(
		env.ctx,
		env.workflow.ID,
		map[string]any{
			"sandbox_strategy":                workflowDomain.SandboxStrategyPersistent,
			"session_persistence":             true,
			"orchestration_resource_revision": nextRevision,
		},
	))

	run, err := env.runSvc.GetByID(env.ctx, result.Run.ID)
	require.NoError(t, err)
	env.orchestrator.HandleRunCompleted(
		env.ctx,
		run,
		workflowDomain.RunStatusCompleted,
	)

	latest, err := env.workflowSvc.GetByID(env.ctx, env.workflow.ID)
	require.NoError(t, err)
	assert.Nil(t, latest.LastPodKey)
}

func TestWorkflowRunCompletionUsesPinnedCallbackAfterRevisionChange(
	t *testing.T,
) {
	firstCalled := make(chan struct{}, 1)
	first := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		firstCalled <- struct{}{}
		writer.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(first.Close)
	secondCalled := make(chan struct{}, 1)
	second := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		secondCalled <- struct{}{}
		writer.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(second.Close)

	env := setupWorkflowTest(t)
	env.orchestrator.httpClient = &http.Client{Timeout: time.Second}
	require.NoError(t, env.workflowSvc.repo.Update(
		env.ctx,
		env.workflow.ID,
		map[string]any{"callback_url": first.URL},
	))
	env.workflow, _ = env.workflowSvc.GetByID(env.ctx, env.workflow.ID)
	result, err := env.orchestrator.TriggerRun(env.ctx, &TriggerRunRequest{
		WorkflowID: env.workflow.ID, TriggerType: workflowDomain.RunTriggerManual,
	})
	require.NoError(t, err)

	nextRevision := *env.workflow.OrchestrationResourceRevision + 1
	require.NoError(t, env.workflowSvc.repo.Update(
		env.ctx,
		env.workflow.ID,
		map[string]any{
			"callback_url":                    second.URL,
			"orchestration_resource_revision": nextRevision,
		},
	))

	env.orchestrator.HandleRunCompleted(
		env.ctx,
		result.Run,
		workflowDomain.RunStatusCompleted,
	)

	require.Eventually(t, func() bool {
		return len(firstCalled) == 1
	}, time.Second, 10*time.Millisecond)
	assert.Never(t, func() bool {
		return len(secondCalled) > 0
	}, 100*time.Millisecond, 10*time.Millisecond)
}
