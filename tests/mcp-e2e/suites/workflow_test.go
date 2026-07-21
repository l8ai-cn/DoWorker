package suites

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/tests/mcp-e2e/fixture"
)

func TestWorkflow_ListDecodes(t *testing.T) {
	env := fixture.LoadEnv(t)
	rest := fixture.SharedREST(t, env)
	pod := fixture.NewEchoPod(t, env, rest)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := pod.MCP.CallToolText(ctx, "list_workflows", map[string]any{
		"limit": 10,
	}); err != nil {
		t.Fatalf("list_workflows: %v", err)
	}
}

func TestWorkflow_TriggerUnknownErrors(t *testing.T) {
	env := fixture.LoadEnv(t)
	rest := fixture.SharedREST(t, env)
	pod := fixture.NewEchoPod(t, env, rest)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := pod.MCP.CallToolText(ctx, "trigger_workflow", map[string]any{
		"workflow_slug": "definitely-not-a-real-workflow-e2e",
	})
	if err == nil {
		t.Fatalf("expected error for unknown workflow_slug, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "not") && !strings.Contains(msg, "workflow") {
		t.Errorf("unexpected error shape: %v", err)
	}
}
