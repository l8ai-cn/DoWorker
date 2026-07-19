import { listWorkerCreateOptions } from "@/lib/api/facade/podConnect";

export interface SessionImportWorkerPlan {
  worker_spec: {
    options_revision: string;
    runtime_image_id: number;
    placement_policy: "automatic";
    compute_target_id: number;
    deployment_mode: string;
    resource_profile_id: number;
  };
  automation_level: "autonomous";
  model_resource_id?: number;
}

export interface SessionImportWorkerRequirement {
  requiresModelResource: boolean;
  modelProtocolAdapters: string[];
}

export async function getSessionImportWorkerRequirement(
  orgSlug: string,
  workerTypeSlug: string,
): Promise<SessionImportWorkerRequirement> {
  const options = await listWorkerCreateOptions(orgSlug, { worker_type_slug: workerTypeSlug });
  const worker = select(
    options.worker_types,
    (option) => option.slug === workerTypeSlug,
    "所选 Worker 当前不可创建",
  );
  if (!worker.supported_interaction_modes.includes("acp")) {
    throw new Error("所选 Worker 不支持 ACP 会话导入");
  }
  return {
    requiresModelResource: worker.requires_model_resource,
    modelProtocolAdapters: worker.model_protocol_adapters,
  };
}

export async function buildSessionImportWorkerPlan(input: {
  orgSlug: string;
  workerTypeSlug: string;
  modelResourceId?: number;
}): Promise<SessionImportWorkerPlan> {
  const initial = await listWorkerCreateOptions(input.orgSlug, { worker_type_slug: input.workerTypeSlug });
  const revision = requiredRevision(initial.revision);
  const worker = select(
    initial.worker_types,
    (option) => option.slug === input.workerTypeSlug,
    "所选 Worker 当前不可创建",
  );
  if (!worker.supported_interaction_modes.includes("acp")) {
    throw new Error("所选 Worker 不支持 ACP 会话导入");
  }
  const modelResourceId = validModelResource(worker.requires_model_resource, input.modelResourceId);
  const compute = select(initial.compute_targets, () => true, "没有可用的计算目标");
  const deploymentOptions = await listWorkerCreateOptions(input.orgSlug, {
    worker_type_slug: input.workerTypeSlug,
    compute_target_id: compute.id,
  });
  assertRevision(revision, deploymentOptions.revision);
  const deployment = select(
    deploymentOptions.deployment_modes,
    () => true,
    "没有可用的部署模式",
  );
  const resolved = await listWorkerCreateOptions(input.orgSlug, {
    worker_type_slug: input.workerTypeSlug,
    compute_target_id: compute.id,
    deployment_mode: deployment.value,
  });
  assertRevision(revision, resolved.revision);
  const runtime = select(
    resolved.runtime_images,
    (option) => option.worker_type_slugs.includes(input.workerTypeSlug),
    "没有可用的运行镜像",
  );
  const profile = select(resolved.resource_profiles, () => true, "没有可用的资源规格");
  return {
    worker_spec: {
      options_revision: revision,
      runtime_image_id: runtime.id,
      placement_policy: "automatic",
      compute_target_id: compute.id,
      deployment_mode: deployment.value,
      resource_profile_id: profile.id,
    },
    automation_level: "autonomous",
    ...(modelResourceId === undefined ? {} : { model_resource_id: modelResourceId }),
  };
}

function select<T extends { selectable: boolean }>(
  options: T[],
  matches: (option: T) => boolean,
  message: string,
): T {
  const value = options.find((option) => option.selectable && matches(option));
  if (!value) throw new Error(message);
  return value;
}

function requiredRevision(value: string): string {
  if (!value.trim()) throw new Error("Worker 创建选项缺少 revision");
  return value;
}

function assertRevision(expected: string, actual: string): void {
  if (actual !== expected) throw new Error("Worker 创建选项已变化，请重新选择");
}

function validModelResource(required: boolean, value: number | undefined): number | undefined {
  if (value !== undefined && (!Number.isSafeInteger(value) || value <= 0)) {
    throw new Error("模型资源无效");
  }
  if (required && value === undefined) {
    throw new Error("所选 Worker 需要明确选择模型资源");
  }
  if (!required && value !== undefined) {
    throw new Error("所选 Worker 不接受模型资源");
  }
  return value;
}
