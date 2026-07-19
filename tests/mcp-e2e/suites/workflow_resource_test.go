package suites

import (
	"context"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/tests/mcp-e2e/fixture"
	"github.com/stretchr/testify/require"
)

func TestCreateWorkflow_ResourceManifestPersistsSnapshotBinding(t *testing.T) {
	env := fixture.LoadEnv(t)
	rest := fixture.SharedREST(t, env)
	fixture.NewEchoPod(t, env, rest)
	templateName := fixture.NewEchoWorkerTemplateResource(
		t,
		env,
		rest,
		"workflow-worker",
	)
	promptName := fixture.NewPromptResource(
		t,
		env,
		rest,
		"workflow",
		"Review the current delivery evidence.",
	)
	db := fixture.OpenDB(t, env)
	workflowName := fixture.UniqueResourceName("mcp-workflow")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	applied, err := rest.ApplyOrchestrationResource(ctx, env.DevOrgSlug, "Workflow", map[string]any{
		"apiVersion": "agentsmesh.io/v1alpha1",
		"kind":       "Workflow",
		"metadata": map[string]any{
			"name": workflowName, "namespace": env.DevOrgSlug,
			"displayName": workflowName,
		},
		"spec": map[string]any{
			"workerTemplateRef": map[string]any{
				"kind": "WorkerTemplate", "name": templateName,
			},
			"promptRef": map[string]any{
				"kind": "Prompt", "name": promptName,
			},
			"inputs": map[string]any{}, "executionMode": "direct",
			"cronExpression": "", "sandboxStrategy": "fresh",
			"sessionPersistence": false, "concurrencyPolicy": "skip",
			"maxConcurrentRuns": 1, "maxRetainedRuns": 0,
			"timeoutMinutes": 1, "idleTimeoutSeconds": 30,
		},
	})
	require.NoError(t, err)
	require.Positive(t, applied.WorkflowID)
	require.Positive(t, applied.WorkerSpecSnapshotID)

	binding, err := db.GetWorkflowOrchestrationBinding(
		ctx,
		env.DevOrgSlug,
		workflowName,
	)
	require.NoError(t, err)
	require.Equal(t, "enabled", binding.Status)
	require.Positive(t, binding.ResourceID)
	require.Positive(t, binding.ResourceRevision)
	require.Positive(t, binding.WorkerSpecSnapshotID)
}
