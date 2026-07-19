import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import {
  ListWorkerCreateOptionsRequestSchema,
  ListWorkerCreateOptionsResponseSchema,
  type ListWorkerCreateOptionsResponse,
} from "@do-worker/proto/pod/v1/worker_creation_pb";
import { apiFetch } from "./api-fetch";
import { readOrgSlug } from "./auth-store";
import { resolveDefaultModelResourceId } from "./model-resources-api";
import { getMobilePodService } from "./mobile-wasm";
import type {
  SessionCreationAgent,
  SessionInteractionMode,
  SessionWire,
} from "./sessions-api";

const OPTIONS_METHOD = "ListWorkerCreateOptions";

export async function createMobileWorkerSession(
  agent: SessionCreationAgent,
  title: string | undefined,
  initialText: string | undefined,
  mode: SessionInteractionMode,
): Promise<SessionWire> {
  const plan = await mobileWorkerPlan(agent, mode);
  const response = await apiFetch("/v1/sessions", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      agent_id: agent.id,
      initial_items: initialItems(initialText),
      ...(title ? { title } : {}),
      ...plan,
    }),
  });
  if (!response.ok) throw new Error((await response.text()) || `HTTP ${response.status}`);
  return response.json() as Promise<SessionWire>;
}

async function mobileWorkerPlan(
  agent: SessionCreationAgent,
  mode: SessionInteractionMode,
): Promise<Record<string, unknown>> {
  const workerTypeSlug = requiredWorkerTypeSlug(agent);
  const initial = await listOptions({ workerTypeSlug });
  const revision = requireRevision(initial.revision);
  const worker = select(initial.workerTypes, (option) => option.slug === workerTypeSlug, "Worker 不可创建");
  verifyWorkerMetadata(agent, worker);
  if (!worker.supportedInteractionModes.includes(mode)) {
    throw new Error(`Worker ${workerTypeSlug} 不支持 ${mode} 交互`);
  }
  const modelResourceId = worker.requiresModelResource
    ? validResourceID(await resolveDefaultModelResourceId())
    : undefined;
  const compute = select(initial.computeTargets, () => true, "没有可用的计算目标");
  const deploymentOptions = await listOptions({ workerTypeSlug, computeTargetId: compute.id });
  assertRevision(revision, deploymentOptions.revision);
  const deployment = select(deploymentOptions.deploymentModes, () => true, "没有可用的部署模式");
  const resolved = await listOptions({
    workerTypeSlug,
    computeTargetId: compute.id,
    deploymentMode: deployment.value,
  });
  assertRevision(revision, resolved.revision);
  const runtime = select(
    resolved.runtimeImages,
    (option) => option.workerTypeSlugs.includes(workerTypeSlug),
    "没有可用的运行镜像",
  );
  const profile = select(resolved.resourceProfiles, () => true, "没有可用的资源规格");
  return {
    worker_spec: {
      options_revision: revision,
      runtime_image_id: numberID(runtime.id, "runtime image id"),
      placement_policy: "automatic",
      compute_target_id: numberID(compute.id, "compute target id"),
      deployment_mode: deployment.value,
      resource_profile_id: numberID(profile.id, "resource profile id"),
    },
    automation_level: mode === "acp" ? "autonomous" : "interactive",
    ...(modelResourceId === undefined ? {} : { model_resource_id: modelResourceId }),
  };
}

async function listOptions(input: {
  workerTypeSlug: string;
  computeTargetId?: bigint;
  deploymentMode?: string;
}): Promise<ListWorkerCreateOptionsResponse> {
  const orgSlug = readOrgSlug();
  if (!orgSlug) throw new Error("未选择组织");
  const request = create(ListWorkerCreateOptionsRequestSchema, { orgSlug, ...input });
  const service = await getMobilePodService();
  const bytes = await service.list_worker_create_options_connect(
    toBinary(ListWorkerCreateOptionsRequestSchema, request),
  );
  return fromBinary(ListWorkerCreateOptionsResponseSchema, new Uint8Array(bytes));
}

function requiredWorkerTypeSlug(agent: SessionCreationAgent): string {
  if (!agent.workerTypeSlug) throw new Error("Worker 缺少权威创建元数据，无法创建任务");
  return agent.workerTypeSlug;
}

function verifyWorkerMetadata(
  agent: SessionCreationAgent,
  worker: { supportedInteractionModes: string[]; requiresModelResource: boolean },
): void {
  const catalogModes = [...agent.supportedModes].sort();
  const optionModes = [...worker.supportedInteractionModes].sort();
  if (
    agent.requiresModelResource !== worker.requiresModelResource ||
    catalogModes.length !== optionModes.length ||
    catalogModes.some((mode, index) => mode !== optionModes[index])
  ) {
    throw new Error("Worker 目录已变化，请重新选择 Worker");
  }
}

function initialItems(text: string | undefined): unknown[] {
  if (!text?.trim()) return [];
  return [{ type: "message", data: { role: "user", content: [{ type: "input_text", text: text.trim() }] } }];
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

function requireRevision(value: string): string {
  if (!value.trim()) throw new Error(`${OPTIONS_METHOD} 未返回 revision`);
  return value;
}

function assertRevision(expected: string, actual: string): void {
  if (expected !== actual) throw new Error("Worker 创建选项已变化，请重新选择");
}

function validResourceID(value: number): number {
  if (!Number.isSafeInteger(value) || value <= 0) throw new Error("无效的模型资源");
  return value;
}

function numberID(value: bigint, label: string): number {
  const result = Number(value);
  if (!Number.isSafeInteger(result) || result <= 0) throw new Error(`无效的 ${label}`);
  return result;
}
