package suites

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/tests/mcp-e2e/fixture"
)

func TestWorkflow_ListShowsSeededWorkflow(t *testing.T) {
	env := fixture.LoadEnv(t)
	rest := fixture.SharedREST(t, env)
	pod := fixture.NewEchoPod(t, env, rest)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	workflow := fixture.NewWorkflowResource(
		t,
		env,
		rest,
		"e2e-workflow-list",
		"do something",
	)

	out, err := pod.MCP.CallToolText(ctx, "list_workflows", map[string]any{
		"query": workflow.Name,
		"limit": 10,
	})
	if err != nil {
		t.Fatalf("list_workflows: %v", err)
	}
	if !strings.Contains(out, workflow.Name) {
		t.Errorf("seeded workflow %q not surfaced by list_workflows:\n%s",
			workflow.Name, out)
	}
}

func TestWorkflow_TriggerCreatesRun(t *testing.T) {
	env := fixture.LoadEnv(t)
	rest := fixture.SharedREST(t, env)
	pod := fixture.NewEchoPod(t, env, rest)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	workflow := fixture.NewWorkflowResource(
		t,
		env,
		rest,
		"e2e-workflow-trigger",
		"{{.task}}",
	)

	db := fixture.OpenDB(t, env)
	t.Cleanup(func() { _ = db.Close() })

	var beforeRuns int
	if err := db.QueryRow(ctx,
		`SELECT count(*) FROM workflow_runs WHERE workflow_id = $1`, workflow.ID,
	).Scan(&beforeRuns); err != nil {
		t.Fatalf("count workflow_runs before: %v", err)
	}

	if _, err := pod.MCP.CallToolText(ctx, "trigger_workflow", map[string]any{
		"workflow_slug": workflow.Name,
		"variables":     map[string]any{"task": "ping"},
	}); err != nil {
		t.Fatalf("trigger_workflow %s: %v", workflow.Name, err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var after int
		if err := db.QueryRow(ctx,
			`SELECT count(*) FROM workflow_runs WHERE workflow_id = $1`, workflow.ID,
		).Scan(&after); err != nil {
			t.Fatalf("count workflow_runs after: %v", err)
		}
		if after > beforeRuns {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	var final int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM workflow_runs WHERE workflow_id = $1`, workflow.ID).Scan(&final)
	t.Fatalf("trigger_workflow did not produce a workflow_run row within 5s (before=%d final=%d)", beforeRuns, final)
}
