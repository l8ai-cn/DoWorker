import { createHash } from "node:crypto";
import { TEST_ORG_SLUG } from "./env";

const E2E_ECHO_DEFINITION_HASH =
  "5f7bcc931f925692e0a5b60a99002bd148070b37112d653dee851ce90587e945";
const E2E_ECHO_RUNTIME_IMAGE_ID = 9;
const E2E_ECHO_WORKER_TYPE = "e2e-echo";

export function loopRuntimeDocuments(input: {
  alias: string;
  clusterId: number;
  resourceUID: string;
  workerTemplateName: string;
}) {
  const spec = workerSpecDocument(input.alias, input.clusterId);
  const workerTemplateSpec = workerTemplateSpecDocument(input.alias);
  const canonicalManifest = stableJson({
    apiVersion: "agentsmesh.io/v1alpha1",
    kind: "WorkerTemplate",
    metadata: {
      displayName: input.alias,
      generation: 1,
      labels: {},
      name: input.workerTemplateName,
      namespace: TEST_ORG_SLUG,
      resourceVersion: "1",
      uid: input.resourceUID,
    },
    spec: workerTemplateSpec,
    status: {},
  });
  return {
    canonicalManifest,
    digest: `sha256:${createHash("sha256")
      .update(canonicalManifest)
      .digest("hex")}`,
    workerSpec: spec,
    workerSummary: workerSummaryDocument(input.alias, spec),
    workerTemplateSpec,
  };
}

function workerSpecDocument(alias: string, clusterId: number) {
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
    metadata: { alias },
    placement,
    runtime: {
      image: {
        digest: `sha256:${"a".repeat(64)}`,
        id: E2E_ECHO_RUNTIME_IMAGE_ID,
      },
      model_binding: {},
      worker_type: {
        definition_hash: E2E_ECHO_DEFINITION_HASH,
        slug: E2E_ECHO_WORKER_TYPE,
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
  alias: string,
  spec: ReturnType<typeof workerSpecDocument>,
) {
  return {
    alias,
    branch: "",
    env_bundle_count: 0,
    knowledge_mount_count: 0,
    lifecycle: spec.lifecycle,
    model_binding: spec.runtime.model_binding,
    placement: spec.placement,
    runtime_image: spec.runtime.image,
    skill_count: 0,
    version: 1,
    worker_type: spec.runtime.worker_type,
  };
}

function workerTemplateSpecDocument(alias: string) {
  return {
    lifecycle: {
      idleTimeoutMinutes: 0,
      terminationPolicy: "completed",
    },
    metadata: { alias },
    optionsRevision: "loop-e2e-runtime",
    runtime: {
      computeTargetRef: { kind: "ComputeTarget", name: "loop-e2e-target" },
      customResources: {
        cpuLimitMilliCPU: 1000,
        cpuRequestMilliCPU: 200,
        memoryLimitBytes: 1073741824,
        memoryRequestBytes: 268435456,
        storageLimitBytes: 2147483648,
        storageRequestBytes: 1073741824,
      },
      deploymentMode: "pooled",
      placementPolicy: "automatic",
      runtimeImageId: E2E_ECHO_RUNTIME_IMAGE_ID,
    },
    toolRefs: {},
    typeConfig: {
      automationLevel: "autonomous",
      interactionMode: "pty",
      schemaVersion: 1,
      secretRefs: {},
      values: { permission_mode: "supervised" },
    },
    workerType: E2E_ECHO_WORKER_TYPE,
    workspace: {
      branch: "",
      configDocumentBindings: [],
      environmentBundleRefs: [],
      instructions: "",
      knowledgeMounts: [],
      skillRefs: [],
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
