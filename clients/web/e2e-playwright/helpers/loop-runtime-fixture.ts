import type { DbFixture } from "../fixtures/db.fixture";
import { TEST_ORG_SLUG } from "./env";

export interface LoopRuntimeFixture {
  alias: string;
  snapshotId: string;
}

export function createLoopRuntimeFixture(db: DbFixture): LoopRuntimeFixture {
  const suffix = Array.from(
    { length: 12 },
    () => String.fromCharCode(97 + Math.floor(Math.random() * 26)),
  ).join("");
  const alias = `循环测试环境 ${suffix}`;
  const modelBinding = {
    resource_id: 1,
    resource_revision: 1,
    connection_id: 1,
    connection_revision: 1,
    provider_key: "openai",
    protocol_adapter: "openai-compatible",
    model_id: "loop-e2e",
  };
  const workerType = {
    slug: "loopal",
    definition_hash:
      "29c0b8aa03cda214e72282983db4de6938b54f7c28e943c038d8cfa94966039b",
  };
  const runtimeImage = {
    id: 1,
    digest: `sha256:${"a".repeat(64)}`,
  };
  const placement = {
    policy: "automatic",
    compute_target: { id: 1, kind: "runner-pool" },
    deployment_mode: "pooled",
    resource_profile: {
      id: 1,
      resources: {
        cpu_request_millicpu: 200,
        cpu_limit_millicpu: 1000,
        memory_request_bytes: 268435456,
        memory_limit_bytes: 1073741824,
      },
    },
  };
  const lifecycle = {
    termination_policy: "manual",
    idle_timeout_minutes: 0,
  };
  const spec = {
    version: 1,
    runtime: {
      model_binding: modelBinding,
      worker_type: workerType,
      image: runtimeImage,
    },
    placement,
    type_config: {
      schema_version: 1,
      values: {},
      secret_refs: {},
      interaction_mode: "pty",
      automation_level: "autonomous",
    },
    workspace: {
      branch: "",
      skill_ids: [],
      knowledge_mounts: [],
      env_bundle_ids: [],
      instructions: "",
      initial_task: "",
    },
    lifecycle,
    metadata: { alias },
  };
  const summary = {
    version: 1,
    model_binding: modelBinding,
    worker_type: workerType,
    runtime_image: runtimeImage,
    placement,
    alias,
    branch: "",
    skill_count: 0,
    knowledge_mount_count: 0,
    env_bundle_count: 0,
    lifecycle,
  };

  const snapshotResult = db.queryValue(`
    INSERT INTO worker_spec_snapshots (
      organization_id, version, spec_json, summary_json
    ) VALUES (
      (SELECT id FROM organizations WHERE slug = '${TEST_ORG_SLUG}'),
      1,
      '${JSON.stringify(spec)}'::jsonb,
      '${JSON.stringify(summary)}'::jsonb
    )
    RETURNING id;
  `);
  const snapshotId = snapshotResult?.split(/\s+/)[0];
  if (!snapshotId) throw new Error("failed to create Loop runtime snapshot");
  return { alias, snapshotId };
}

export function cleanupLoopRuntimeFixture(
  db: DbFixture,
  fixture: LoopRuntimeFixture,
) {
  db.cleanup(`
    DELETE FROM goal_loops WHERE worker_spec_snapshot_id = ${fixture.snapshotId};
    DELETE FROM worker_spec_snapshots WHERE id = ${fixture.snapshotId};
  `);
}
