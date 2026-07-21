import type { ConnectClient } from "./connect-client";
import type { E2EWorkerSpecDraft } from "./e2e-worker-spec";
import { TEST_ORG_SLUG } from "./env";
import { validatePlanApplyResource } from "./orchestration-resource";

interface AppliedWorkerTemplate {
  workerSpecSnapshotId?: bigint;
}

export async function applyE2EWorkerTemplate(
  client: ConnectClient,
  name: string,
  worker: E2EWorkerSpecDraft,
): Promise<string> {
  const targetName = `${name}-target`;
  const profileName = `${name}-profile`;
  await applyDocument(client, "ComputeTarget", targetName, {
    computeTargetId: safeNumber(worker.computeTargetId, "compute target"),
  });
  await applyDocument(client, "ResourceProfile", profileName, {
    resourceProfileId: safeNumber(worker.resourceProfileId, "resource profile"),
  });
  const applied = await applyDocument(client, "WorkerTemplate", name, {
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
      environmentBundleRefs: [],
      configDocumentBindings: [],
      instructions: worker.instructions,
    },
    lifecycle: {
      terminationPolicy: worker.terminationPolicy,
      idleTimeoutMinutes: worker.idleTimeoutMinutes,
    },
    metadata: { alias: worker.alias },
  }) as AppliedWorkerTemplate;
  const snapshotID = applied.workerSpecSnapshotId;
  if (!snapshotID || snapshotID <= 0n) {
    throw new Error("WorkerTemplate apply returned no immutable WorkerSpec snapshot");
  }
  return snapshotID.toString();
}

async function applyDocument(
  client: ConnectClient,
  kind: "ComputeTarget" | "ResourceProfile" | "WorkerTemplate",
  name: string,
  spec: Record<string, unknown>,
) {
  const displayName = kind === "WorkerTemplate"
    ? (spec.metadata as { alias?: string }).alias ?? name
    : name;
  return validatePlanApplyResource(
    client,
    TEST_ORG_SLUG,
    kind,
    JSON.stringify({
      apiVersion: "agentcloud.io/v1alpha1",
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
  const result = Number(value);
  if (!Number.isSafeInteger(result) || result <= 0) {
    throw new Error(`${label} ID is outside the safe integer range`);
  }
  return result;
}
