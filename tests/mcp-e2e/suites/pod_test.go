package suites

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/tests/mcp-e2e/fixture"
	"github.com/stretchr/testify/require"
)

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
	planID, err := rest.PlanOrchestrationResource(ctx, env.DevOrgSlug, map[string]any{
		"apiVersion": "agentcloud.io/v1alpha1",
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
	})
	require.NoError(t, err)
	out, err := pod.MCP.CallToolText(
		ctx,
		"create_pod",
		map[string]any{"plan_id": planID},
	)
	require.NoError(t, err)
	match := nestedPodKeyRE.FindStringSubmatch(out)
	require.Len(t, match, 2, out)
	spawnedKey := match[1]
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
