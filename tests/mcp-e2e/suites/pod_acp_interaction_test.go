package suites

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/tests/mcp-e2e/fixture"
)

func TestPodACPInteraction_RoundTrip(t *testing.T) {
	env := fixture.LoadEnv(t)
	rest := fixture.SharedREST(t, env)
	pod := fixture.NewACPEchoPod(t, env, rest)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := rest.SendPodPrompt(ctx, env.DevOrgSlug, pod.Pod.PodKey, "hello ACP"); err != nil {
		t.Fatalf("send ACP prompt: %v", err)
	}

	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		snapshot, err := pod.MCP.CallToolText(ctx, "get_pod_snapshot", map[string]any{
			"pod_key": pod.Pod.PodKey,
			"lines":   200,
		})
		if err == nil && strings.Contains(snapshot, "[assistant] echo: hello ACP") {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatal("expected ACP echo response in pod snapshot within 15s")
}
