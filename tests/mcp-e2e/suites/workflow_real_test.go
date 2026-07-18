package suites

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/tests/mcp-e2e/client"
	"github.com/anthropics/agentsmesh/tests/mcp-e2e/fixture"
)

// These specs create workflows through Connect, then exercise their MCP surface.

func TestWorkflow_ListShowsSeededWorkflow(t *testing.T) {
	env := fixture.LoadEnv(t)
	rest := fixture.SharedREST(t, env)
	runner := fixture.DiscoverRunner(t, env, rest)
	pod := fixture.NewEchoPod(t, env, rest, runner.ID)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	workflowName := fmt.Sprintf("e2e-workflow-list-%d", time.Now().UnixMilli())
	workflow, err := rest.CreateWorkflow(ctx, env.DevOrgSlug, client.CreateWorkflowRequest{
		Name:           workflowName,
		AgentSlug:      "e2e-echo",
		PromptTemplate: "do something",
		RunnerID:       &runner.ID,
	})
	if err != nil {
		t.Fatalf("create workflow via Connect: %v", err)
	}

	out, err := pod.MCP.CallToolText(ctx, "list_workflows", map[string]any{
		"query": workflowName,
		"limit": 10,
	})
	if err != nil {
		t.Fatalf("list_workflows: %v", err)
	}
	if !strings.Contains(out, workflowName) && !strings.Contains(out, workflow.Slug) {
		t.Errorf("seeded workflow %q (slug=%s) not surfaced by list_workflows:\n%s",
			workflowName, workflow.Slug, out)
	}
}

func TestWorkflow_TriggerCreatesRun(t *testing.T) {
	env := fixture.LoadEnv(t)
	rest := fixture.SharedREST(t, env)
	runner := fixture.DiscoverRunner(t, env, rest)
	pod := fixture.NewEchoPod(t, env, rest, runner.ID)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	workflowName := fmt.Sprintf("e2e-workflow-trigger-%d", time.Now().UnixMilli())
	workflow, err := rest.CreateWorkflow(ctx, env.DevOrgSlug, client.CreateWorkflowRequest{
		Name:           workflowName,
		AgentSlug:      "e2e-echo",
		PromptTemplate: "{{.task}}",
		RunnerID:       &runner.ID,
	})
	if err != nil {
		t.Fatalf("create workflow: %v", err)
	}
	if err := rest.EnableWorkflow(ctx, env.DevOrgSlug, workflow.Slug); err != nil {
		t.Fatalf("enable workflow %s: %v", workflow.Slug, err)
	}

	db, err := client.OpenDB(env.PostgresDSN)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	var beforeRuns int
	if err := db.QueryRow(ctx,
		`SELECT count(*) FROM workflow_runs WHERE workflow_id = $1`, workflow.ID,
	).Scan(&beforeRuns); err != nil {
		t.Fatalf("count workflow_runs before: %v", err)
	}

	if _, err := pod.MCP.CallToolText(ctx, "trigger_workflow", map[string]any{
		"workflow_slug": workflow.Slug,
		"variables":     map[string]any{"task": "ping"},
	}); err != nil {
		t.Fatalf("trigger_workflow %s: %v", workflow.Slug, err)
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
