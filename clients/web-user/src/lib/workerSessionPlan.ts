import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import {
  ListWorkerCreateOptionsRequestSchema,
  ListWorkerCreateOptionsResponseSchema,
  type ListWorkerCreateOptionsResponse,
} from "@do-worker/proto/pod/v1/worker_creation_pb";

import { authenticatedFetch } from "./identity";
import { readDoWorkerOrgSlug } from "./do-worker";

const OPTIONS_PROCEDURE = "/proto.pod.v1.PodService/ListWorkerCreateOptions";

export type SessionInteractionMode = "acp" | "pty";

export interface WorkerCreationSelection {
  workerTypeSlug: string;
  supportedModes: readonly SessionInteractionMode[];
  requiresModelResource: boolean;
}

export interface SessionWorkerPlan {
  worker_spec: {
    options_revision: string;
    runtime_image_id: number;
    placement_policy: "automatic";
    compute_target_id: number;
    deployment_mode: string;
    resource_profile_id: number;
  };
  automation_level: "autonomous" | "interactive";
  model_resource_id?: number;
}

export async function workerRequiresModelResource(selection: WorkerCreationSelection): Promise<boolean> {
  const options = await listWorkerCreateOptions({ workerTypeSlug: selection.workerTypeSlug });
  const worker = select(
    options.workerTypes,
    (option) => option.slug === selection.workerTypeSlug,
    `Worker ${selection.workerTypeSlug} is not selectable`,
  );
  assertCatalogMetadata(selection, worker);
  return worker.requiresModelResource;
}

export async function buildSessionWorkerPlan(input: {
  selection: WorkerCreationSelection;
  mode: SessionInteractionMode;
  modelResourceId?: number;
  resolveModelResourceId?: () => Promise<number>;
}): Promise<SessionWorkerPlan> {
  const { selection } = input;
  const initial = await listWorkerCreateOptions({ workerTypeSlug: selection.workerTypeSlug });
  const revision = requireRevision(initial.revision);
  const workerType = select(
    initial.workerTypes,
    (option) => option.slug === selection.workerTypeSlug,
    `Worker ${selection.workerTypeSlug} is not selectable`,
  );
  assertCatalogMetadata(selection, workerType);
  if (!workerType.supportedInteractionModes.includes(input.mode)) {
    throw new Error(`Worker ${selection.workerTypeSlug} does not support ${input.mode} sessions`);
  }
  const modelResourceId = modelResourceFor(
    workerType.requiresModelResource,
    input.modelResourceId ?? (workerType.requiresModelResource
      ? await input.resolveModelResourceId?.()
      : undefined),
  );
  const computeTarget = select(
    initial.computeTargets,
    () => true,
    `Worker ${selection.workerTypeSlug} has no selectable compute target`,
  );
  const deploymentOptions = await listWorkerCreateOptions({
    workerTypeSlug: selection.workerTypeSlug,
    computeTargetId: computeTarget.id,
  });
  assertRevision(revision, deploymentOptions.revision);
  const deployment = select(
    deploymentOptions.deploymentModes,
    () => true,
    `Worker ${selection.workerTypeSlug} has no selectable deployment mode`,
  );
  const resolved = await listWorkerCreateOptions({
    workerTypeSlug: selection.workerTypeSlug,
    computeTargetId: computeTarget.id,
    deploymentMode: deployment.value,
  });
  assertRevision(revision, resolved.revision);
  const runtime = select(
    resolved.runtimeImages,
    (option) => option.workerTypeSlugs.includes(selection.workerTypeSlug),
    `Worker ${selection.workerTypeSlug} has no selectable runtime image`,
  );
  const profile = select(
    resolved.resourceProfiles,
    () => true,
    `Worker ${selection.workerTypeSlug} has no selectable resource profile`,
  );
  return {
    worker_spec: {
      options_revision: revision,
      runtime_image_id: safeNumber(runtime.id, "runtime image id"),
      placement_policy: "automatic",
      compute_target_id: safeNumber(computeTarget.id, "compute target id"),
      deployment_mode: deployment.value,
      resource_profile_id: safeNumber(profile.id, "resource profile id"),
    },
    automation_level: input.mode === "acp" ? "autonomous" : "interactive",
    ...(modelResourceId === undefined ? {} : { model_resource_id: modelResourceId }),
  };
}

async function listWorkerCreateOptions(filter: {
  workerTypeSlug?: string;
  computeTargetId?: bigint;
  deploymentMode?: string;
}): Promise<ListWorkerCreateOptionsResponse> {
  const orgSlug = readDoWorkerOrgSlug();
  if (!orgSlug) throw new Error("Select an organization before creating a Worker");
  const request = create(ListWorkerCreateOptionsRequestSchema, {
    orgSlug,
    workerTypeSlug: filter.workerTypeSlug,
    computeTargetId: filter.computeTargetId,
    deploymentMode: filter.deploymentMode,
  });
  const response = await authenticatedFetch(OPTIONS_PROCEDURE, {
    method: "POST",
    headers: {
      Accept: "application/proto",
      "Content-Type": "application/proto",
    },
    body: toBinary(ListWorkerCreateOptionsRequestSchema, request),
  });
  if (!response.ok) {
    throw new Error((await response.text()).trim() || `Worker options failed (${response.status})`);
  }
  return fromBinary(
    ListWorkerCreateOptionsResponseSchema,
    new Uint8Array(await response.arrayBuffer()),
  );
}

function select<T extends { selectable: boolean }>(
  options: T[],
  matches: (option: T) => boolean,
  message: string,
): T {
  const option = options.find((item) => item.selectable && matches(item));
  if (!option) throw new Error(message);
  return option;
}

function safeNumber(value: bigint, label: string): number {
  const result = Number(value);
  if (!Number.isSafeInteger(result) || result <= 0) {
    throw new Error(`Invalid ${label}`);
  }
  return result;
}

function requireRevision(value: string): string {
  if (!value.trim()) throw new Error("Worker options revision is missing");
  return value;
}

function assertRevision(expected: string, actual: string): void {
  if (actual !== expected) {
    throw new Error("Worker options changed while creating the session");
  }
}

function modelResourceFor(required: boolean, value: number | undefined): number | undefined {
  if (value !== undefined && (!Number.isSafeInteger(value) || value <= 0)) {
    throw new Error("Invalid model resource id");
  }
  if (required && value === undefined) {
    throw new Error("Worker requires a model resource");
  }
  if (!required && value !== undefined) {
    throw new Error("Worker does not accept a model resource");
  }
  return value;
}

function assertCatalogMetadata(
  selection: WorkerCreationSelection,
  option: { supportedInteractionModes: string[]; requiresModelResource: boolean },
): void {
  const modes = [...selection.supportedModes].sort();
  const optionModes = [...option.supportedInteractionModes].sort();
  if (
    selection.requiresModelResource !== option.requiresModelResource ||
    modes.length !== optionModes.length ||
    modes.some((mode, index) => mode !== optionModes[index])
  ) {
    throw new Error("Worker catalog metadata changed; reload and choose the Worker again");
  }
}
