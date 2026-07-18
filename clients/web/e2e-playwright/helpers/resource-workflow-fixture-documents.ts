import { createHash } from "node:crypto";
import { TEST_ORG_SLUG } from "./env";

export const RESOURCE_WORKFLOW_SLUG = "e2e-resource-workflow-v1";
export const RESOURCE_WORKFLOW_NAME = "E2E Resource Workflow";

const LOOPAL_DEFINITION_HASH =
  "29c0b8aa03cda214e72282983db4de6938b54f7c28e943c038d8cfa94966039b";

export function resourceWorkflowFixtureDocuments(
  clusterId: number,
  resourceUID: string,
) {
  const workerSpec = workerSpecDocument(clusterId);
  const workerSummary = workerSummaryDocument(workerSpec);
  const workflowSpec = workflowSpecDocument();
  const canonicalManifest = stableJson({
    apiVersion: "agentsmesh.io/v1alpha1",
    kind: "Workflow",
    metadata: {
      generation: 1,
      labels: { "test-suite": "workflow-runtime" },
      name: RESOURCE_WORKFLOW_SLUG,
      namespace: TEST_ORG_SLUG,
      resourceVersion: "1",
      uid: resourceUID,
    },
    spec: workflowSpec,
    status: {},
  });
  return {
    canonicalManifest,
    digest: `sha256:${createHash("sha256")
      .update(canonicalManifest)
      .digest("hex")}`,
    workerSpec,
    workerSummary,
    workflowSpec,
  };
}

export function fixtureJson(value: unknown): string {
  const json = typeof value === "string" ? value : stableJson(value);
  return `'${json.replace(/'/g, "''")}'::jsonb`;
}

function workerSpecDocument(clusterId: number) {
  const placement = {
    compute_target: { id: clusterId, kind: "runner-pool" },
    deployment_mode: "pooled",
    policy: "automatic",
    resource_profile: {
      custom: true,
      id: 0,
      resources: {
        cpu_limit_millicpu: 1000,
        cpu_request_millicpu: 200,
        memory_limit_bytes: 1073741824,
        memory_request_bytes: 268435456,
        storage_limit_bytes: 2147483648,
        storage_request_bytes: 1073741824,
      },
    },
  };
  const lifecycle = {
    idle_timeout_minutes: 0,
    termination_policy: "completed",
  };
  return {
    lifecycle,
    metadata: { alias: RESOURCE_WORKFLOW_NAME },
    placement,
    runtime: {
      image: { digest: `sha256:${"a".repeat(64)}`, id: 1 },
      model_binding: {},
      worker_type: {
        definition_hash: LOOPAL_DEFINITION_HASH,
        slug: "loopal",
      },
    },
    type_config: {
      automation_level: "autonomous",
      interaction_mode: "pty",
      schema_version: 1,
      secret_refs: {},
      values: { permission_mode: "supervised" },
    },
    version: 1,
    workspace: {
      branch: "",
      config_document_bindings: [],
      env_bundle_ids: [],
      initial_task: "",
      instructions: "",
      knowledge_mounts: [],
      skill_ids: [],
    },
  };
}

function workerSummaryDocument(
  spec: ReturnType<typeof workerSpecDocument>,
) {
  return {
    alias: RESOURCE_WORKFLOW_NAME,
    branch: "",
    env_bundle_count: 0,
    knowledge_mount_count: 0,
    lifecycle: spec.lifecycle,
    model_binding: {},
    placement: spec.placement,
    runtime_image: spec.runtime.image,
    skill_count: 0,
    version: 1,
    worker_type: spec.runtime.worker_type,
  };
}

function workflowSpecDocument() {
  return {
    concurrencyPolicy: "skip",
    executionMode: "direct",
    idleTimeoutSeconds: 30,
    inputs: {},
    maxConcurrentRuns: 1,
    maxRetainedRuns: 0,
    promptRef: { kind: "Prompt", name: "e2e-runtime-prompt" },
    sandboxStrategy: "fresh",
    sessionPersistence: false,
    timeoutMinutes: 1,
    workerTemplateRef: {
      kind: "WorkerTemplate",
      name: "e2e-loopal-template",
    },
  };
}

function stableJson(value: unknown): string {
  if (Array.isArray(value)) return `[${value.map(stableJson).join(",")}]`;
  if (value && typeof value === "object") {
    const object = value as Record<string, unknown>;
    return `{${Object.keys(object).sort().map((key) =>
      `${JSON.stringify(key)}:${stableJson(object[key])}`
    ).join(",")}}`;
  }
  return JSON.stringify(value);
}
