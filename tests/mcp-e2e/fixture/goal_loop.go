package fixture

import (
	"context"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/tests/mcp-e2e/client"
)

const goalLoopSpecJSON = `{
  "version": 1,
  "runtime": {
    "model_binding": {
      "resource_id": 1001,
      "resource_revision": 1,
      "connection_id": 2001,
      "connection_revision": 1,
      "provider_key": "openai",
      "protocol_adapter": "openai",
      "model_id": "gpt-5"
    },
    "worker_type": {
      "slug": "codex-cli",
      "definition_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
    },
    "image": {
      "id": 1,
      "digest": "sha256:963c99fb047c0a4fed518eb9949e805fd31329a8395526fbb1fe34d8254ebea1"
    }
  },
  "placement": {
    "policy": "automatic",
    "compute_target": {"id": 1, "kind": "runner-pool"},
    "deployment_mode": "pooled",
    "resource_profile": {
      "id": 1,
      "resources": {
        "cpu_request_millicpu": 200,
        "cpu_limit_millicpu": 1000,
        "memory_request_bytes": 268435456,
        "memory_limit_bytes": 1073741824
      }
    }
  },
  "type_config": {
    "schema_version": 1,
    "values": {"approval_mode": "never"},
    "secret_refs": {},
    "interaction_mode": "acp",
    "automation_level": "autonomous"
  },
  "workspace": {
    "skill_ids": [],
    "knowledge_mounts": [],
    "env_bundle_ids": [],
    "instructions": "",
    "initial_task": ""
  },
  "lifecycle": {"termination_policy": "manual", "idle_timeout_minutes": 0},
  "metadata": {"alias": "e2e-goal-loop"}
}`

const goalLoopSummaryJSON = `{
  "version": 1,
  "model_binding": {
    "resource_id": 1001,
    "resource_revision": 1,
    "connection_id": 2001,
    "connection_revision": 1,
    "provider_key": "openai",
    "protocol_adapter": "openai",
    "model_id": "gpt-5"
  },
  "worker_type": {
    "slug": "codex-cli",
    "definition_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
  },
  "runtime_image": {
    "id": 1,
    "digest": "sha256:963c99fb047c0a4fed518eb9949e805fd31329a8395526fbb1fe34d8254ebea1"
  },
  "placement": {
    "policy": "automatic",
    "compute_target": {"id": 1, "kind": "runner-pool"},
    "deployment_mode": "pooled",
    "resource_profile": {
      "id": 1,
      "resources": {
        "cpu_request_millicpu": 200,
        "cpu_limit_millicpu": 1000,
        "memory_request_bytes": 268435456,
        "memory_limit_bytes": 1073741824
      }
    }
  },
  "alias": "e2e-goal-loop",
  "branch": "",
  "skill_count": 0,
  "knowledge_mount_count": 0,
  "env_bundle_count": 0,
  "lifecycle": {"termination_policy": "manual", "idle_timeout_minutes": 0}
}`

func NewGoalLoopWorkerTemplate(t *testing.T, env *Env) string {
	t.Helper()
	db, err := client.OpenDB(env.PostgresDSN)
	if err != nil {
		t.Fatalf("open database for goal loop worker template: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	name, err := seedGoalLoopWorkerTemplate(ctx, db, env)
	if err != nil {
		t.Fatalf("create goal loop worker template: %v", err)
	}
	return name
}
