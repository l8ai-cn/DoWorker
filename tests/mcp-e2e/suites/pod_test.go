package suites

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/tests/mcp-e2e/fixture"
	"github.com/stretchr/testify/require"
)

// create_pod returns text like "Pod: <key> ..." — extract the key with a
// regex and use it to clean up via REST.
var nestedPodKeyRE = regexp.MustCompile(`(\d+-standalone-[a-f0-9]+)`)

func TestCreatePod_NestedSpawn(t *testing.T) {
	env := fixture.LoadEnv(t)
	rest := fixture.SharedREST(t, env)
	pod := fixture.NewEchoPod(t, env, rest)
	templateName := fixture.NewEchoWorkerTemplateResource(
		t,
		env,
		rest,
		"nested-worker",
	)
	db := fixture.OpenDB(t, env)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	alias := fmt.Sprintf("e2e-nested-%d", time.Now().UnixMilli())
	out, err := pod.MCP.CallToolText(ctx, "create_pod", map[string]any{
		"resource": map[string]any{
			"apiVersion": "agentsmesh.io/v1alpha1",
			"kind":       "Worker",
			"metadata": map[string]any{
				"name": alias, "namespace": env.DevOrgSlug,
			},
			"spec": map[string]any{
				"workerTemplateRef": map[string]any{
					"kind": "WorkerTemplate", "name": templateName,
				},
				"inputs": map[string]any{}, "alias": alias,
			},
		},
	})
	if err != nil {
		t.Fatalf("create_pod: %v", err)
	}
	// The exact label varies (Pod / pod_key / etc.) — we only need the key
	// itself, and pod keys follow a stable pattern <n>-standalone-<hex>.
	m := nestedPodKeyRE.FindStringSubmatch(out)
	if len(m) != 2 {
		t.Fatalf("could not parse spawned pod key from output:\n%s", out)
	}
	spawnedKey := m[1]
	if !strings.Contains(out, spawnedKey) {
		t.Errorf("expected pod key in output:\n%s", out)
	}
	require.Contains(t, out, "Resource: Worker/")
	require.Contains(t, out, "Snapshot:")
	binding, err := db.GetWorkerLaunchOrchestrationBinding(ctx, spawnedKey)
	require.NoError(t, err)
	require.Positive(t, binding.ResourceID)
	require.Positive(t, binding.ResourceRevision)
	require.Positive(t, binding.WorkerSpecSnapshotID)

	t.Cleanup(func() {
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()
		_ = rest.TerminatePod(ctx2, env.DevOrgSlug, spawnedKey)
	})
}
