import type { ConnectClient } from "./connect-client";
import {
  buildE2EEchoWorkerSpec,
  type E2EWorkerSpecOptions,
} from "./e2e-worker-spec";
import { TEST_ORG_SLUG } from "./env";
import { validatePlanApplyResource } from "./orchestration-resource";

interface EnvironmentBundleInput {
  id: bigint;
  name: string;
}

export interface ResourceWorkflowOptions {
  slug: string;
  name: string;
  prompt: string;
  cronExpression?: string;
  executionMode?: "direct" | "autopilot";
  sandboxStrategy?: "fresh" | "persistent";
  sessionPersistence?: boolean;
  timeoutMinutes?: number;
  environmentBundles?: EnvironmentBundleInput[];
  worker?: E2EWorkerSpecOptions;
}

export interface CreatedResourceWorkflow {
  slug: string;
  name: string;
  workflowId: bigint;
  workerSpecSnapshotId: bigint;
}

export async function createResourceWorkflow(
  client: ConnectClient,
  options: ResourceWorkflowOptions,
): Promise<CreatedResourceWorkflow> {
  const worker = await buildE2EEchoWorkerSpec(client, {
    mode: "pty",
    automationLevel: "autonomous",
    ...options.worker,
  });
  const targetName = `${options.slug}-target`;
  const profileName = `${options.slug}-profile`;
  const promptName = `${options.slug}-prompt`;
  const workerName = `${options.slug}-worker`;
  await applyDocument(client, "ComputeTarget", targetName, {
    computeTargetId: safeNumber(worker.computeTargetId, "compute target"),
  });
  await applyDocument(client, "ResourceProfile", profileName, {
    resourceProfileId: safeNumber(worker.resourceProfileId, "resource profile"),
  });
  const environmentBundleRefs = [];
  for (const [index, bundle] of (options.environmentBundles ?? []).entries()) {
    const name = `${options.slug}-env-${index + 1}`;
    await applyDocument(client, "EnvironmentBundle", name, {
      environmentBundleId: safeNumber(bundle.id, `environment bundle ${bundle.name}`),
    });
    environmentBundleRefs.push({ kind: "EnvironmentBundle", name });
  }
  await applyDocument(client, "Prompt", promptName, {
    content: options.prompt,
    variables: {},
  });
  await applyDocument(client, "WorkerTemplate", workerName, {
    optionsRevision: worker.optionsRevision,
    workerType: worker.workerTypeSlug,
    toolRefs: {},
    runtime: {
      runtimeImageId: safeNumber(worker.runtimeImageId, "runtime image"),
      placementPolicy: worker.placementPolicy,
      computeTargetRef: { kind: "ComputeTarget", name: targetName },
      deploymentMode: worker.deploymentMode,
      resourceProfileRef: { kind: "ResourceProfile", name: profileName },
    },
    typeConfig: {
      schemaVersion: worker.typeSchemaVersion,
      values: JSON.parse(worker.typeConfigValuesJson),
      secretRefs: {},
      interactionMode: worker.interactionMode,
      automationLevel: worker.automationLevel,
    },
    workspace: {
      branch: worker.branch,
      skillRefs: [],
      knowledgeMounts: [],
      environmentBundleRefs,
      configDocumentBindings: [],
      instructions: worker.instructions,
    },
    lifecycle: {
      terminationPolicy: "completed",
      idleTimeoutMinutes: worker.idleTimeoutMinutes,
    },
    metadata: { alias: worker.alias },
  });
  const applied = await applyDocument(client, "Workflow", options.slug, {
    workerTemplateRef: { kind: "WorkerTemplate", name: workerName },
    promptRef: { kind: "Prompt", name: promptName },
    inputs: {},
    executionMode: options.executionMode ?? "direct",
    cronExpression: options.cronExpression ?? "",
    sandboxStrategy: options.sandboxStrategy ?? "fresh",
    sessionPersistence: options.sessionPersistence ?? false,
    concurrencyPolicy: "skip",
    maxConcurrentRuns: 1,
    maxRetainedRuns: 0,
    timeoutMinutes: options.timeoutMinutes ?? 1,
    idleTimeoutSeconds: 30,
  }, options.name) as { workflowId?: bigint; workerSpecSnapshotId?: bigint };
  if (!applied.workflowId || !applied.workerSpecSnapshotId) {
    throw new Error(`Workflow apply returned incomplete result for ${options.slug}`);
  }
  return {
    slug: options.slug,
    name: options.name,
    workflowId: applied.workflowId,
    workerSpecSnapshotId: applied.workerSpecSnapshotId,
  };
}

async function applyDocument(
  client: ConnectClient,
  kind: Parameters<typeof validatePlanApplyResource>[2],
  name: string,
  spec: Record<string, unknown>,
  displayName = name,
) {
  return validatePlanApplyResource(
    client,
    TEST_ORG_SLUG,
    kind,
    JSON.stringify({
      apiVersion: "agentsmesh.io/v1alpha1",
      kind,
      metadata: {
        name,
        namespace: TEST_ORG_SLUG,
        displayName,
      },
      spec,
    }),
  );
}

function safeNumber(value: bigint, label: string): number {
  const number = Number(value);
  if (!Number.isSafeInteger(number) || number <= 0) {
    throw new Error(`${label} ID is outside the safe integer range`);
  }
  return number;
}
