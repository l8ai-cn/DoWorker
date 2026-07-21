import { createHash } from "node:crypto";
import { TEST_ORG_SLUG } from "./env";

export const RESOURCE_WORKFLOW_SLUG = "e2e-resource-workflow-v2";
export const RESOURCE_WORKFLOW_NAME = "E2E Resource Workflow";

export function resourceWorkflowManifest(resourceUID: string) {
  const spec = {
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
    workerTemplateRef: { kind: "WorkerTemplate", name: "e2e-echo-runtime" },
  };
  const canonicalManifest = stableJson({
    apiVersion: "agentcloud.io/v1alpha1",
    kind: "Workflow",
    metadata: {
      generation: 1,
      labels: { "test-suite": "workflow-runtime" },
      name: RESOURCE_WORKFLOW_SLUG,
      namespace: TEST_ORG_SLUG,
      resourceVersion: "1",
      uid: resourceUID,
    },
    spec,
    status: {},
  });
  return {
    canonicalManifest,
    digest: `sha256:${createHash("sha256").update(canonicalManifest).digest("hex")}`,
    spec,
  };
}

export function fixtureJSON(value: unknown): string {
  const json = typeof value === "string" ? value : stableJson(value);
  return `'${json.replace(/'/g, "''")}'::jsonb`;
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
