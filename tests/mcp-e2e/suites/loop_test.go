package suites

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/tests/mcp-e2e/fixture"
)

func TestGoalLoop_ListDecodes(t *testing.T) {
	env := fixture.LoadEnv(t)
	rest := fixture.SharedREST(t, env)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := rest.ListGoalLoops(ctx, env.DevOrgSlug, "", 10, 0); err != nil {
		t.Fatalf("list goal loops: %v", err)
	}
}

func TestGoalLoop_StartUnknownErrors(t *testing.T) {
	env := fixture.LoadEnv(t)
	rest := fixture.SharedREST(t, env)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := rest.StartGoalLoop(ctx, env.DevOrgSlug, "definitely-not-a-real-goal-loop-e2e")
	if err == nil {
		t.Fatal("expected error for unknown goal loop, got nil")
	}
	if !strings.Contains(err.Error(), "404") && !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error shape: %v", err)
	}
}
