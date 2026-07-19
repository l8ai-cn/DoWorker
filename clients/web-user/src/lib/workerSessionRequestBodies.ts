import type { SessionEventInput } from "./types";
import { listModelResources } from "./modelConfigsApi";
import {
  buildSessionWorkerPlan,
  type WorkerCreationSelection,
  type SessionInteractionMode,
  type SessionWorkerPlan,
} from "./workerSessionPlan";

interface ModelPlanInput extends WorkerCreationSelection {
  agentId: string;
  mode?: SessionInteractionMode;
  modelResourceId?: number;
}

interface CrossAgentPlanInput extends ModelPlanInput {
  sourceAgentId: string;
}

type FreshWorkerPlanBody = SessionWorkerPlan;

type SameAgentSnapshotBody = {
  agent_id: string;
  title?: string;
  up_to_response_id?: string;
};

export async function createWorkerSessionBody(input: ModelPlanInput & {
  initialItems: SessionEventInput[];
  hostId?: string;
  workspace?: string;
  tokenBudget?: number | null;
  parentSessionId?: string;
  subAgentName?: string | null;
  title?: string;
}): Promise<{
  agent_id: string;
  initial_items: SessionEventInput[];
  host_id?: string;
  workspace?: string;
  token_budget?: number;
  parent_session_id?: string;
  sub_agent_name?: string | null;
  title?: string;
} & FreshWorkerPlanBody> {
  return {
    agent_id: input.agentId,
    initial_items: input.initialItems,
    ...(input.hostId?.trim() ? { host_id: input.hostId } : {}),
    ...(input.workspace?.trim() ? { workspace: input.workspace } : {}),
    ...(validTokenBudget(input.tokenBudget) === undefined
      ? {}
      : { token_budget: validTokenBudget(input.tokenBudget) }),
    ...(await workerPlan(input)),
    ...(input.parentSessionId === undefined ? {} : { parent_session_id: input.parentSessionId }),
    ...(input.subAgentName === undefined ? {} : { sub_agent_name: input.subAgentName }),
    ...(input.title === undefined ? {} : { title: input.title }),
  };
}

export async function importWorkerSessionBody(input: ModelPlanInput & {
  sourcePath: string;
  title?: string;
  hostId?: string;
}): Promise<{
  source_path: string;
  agent_id: string;
  title?: string;
  host_id?: string;
} & FreshWorkerPlanBody> {
  return {
    source_path: input.sourcePath,
    agent_id: input.agentId,
    ...(await workerPlan(input)),
    ...(input.title?.trim() ? { title: input.title } : {}),
    ...(input.hostId?.trim() ? { host_id: input.hostId } : {}),
  };
}

export async function crossAgentWorkerSessionBody(input: CrossAgentPlanInput & {
  title?: string;
  upToResponseId?: string;
}): Promise<SameAgentSnapshotBody | (SameAgentSnapshotBody & FreshWorkerPlanBody)> {
  if (input.agentId === input.sourceAgentId) {
    return {
      agent_id: input.agentId,
      ...(input.title === undefined ? {} : { title: input.title }),
      ...(input.upToResponseId === undefined ? {} : { up_to_response_id: input.upToResponseId }),
    };
  }
  return {
    agent_id: input.agentId,
    ...(await workerPlan(input)),
    ...(input.title === undefined ? {} : { title: input.title }),
    ...(input.upToResponseId === undefined ? {} : { up_to_response_id: input.upToResponseId }),
  };
}

async function workerPlan(input: ModelPlanInput): Promise<SessionWorkerPlan> {
  return buildSessionWorkerPlan({
    selection: {
      workerTypeSlug: input.workerTypeSlug,
      supportedModes: input.supportedModes,
      requiresModelResource: input.requiresModelResource,
    },
    mode: input.mode ?? "acp",
    modelResourceId: input.modelResourceId,
    resolveModelResourceId: resolveDefaultModelResourceId,
  });
}

async function resolveDefaultModelResourceId(): Promise<number> {
  const defaults = (await listModelResources()).filter((resource) => resource.is_default);
  if (defaults.length !== 1) throw new Error("No default model resource is configured");
  return validModelResourceId(defaults[0].id);
}

function validModelResourceId(value: number): number {
  if (!Number.isSafeInteger(value) || value <= 0) throw new Error("Invalid model resource id");
  return value;
}

function validTokenBudget(value: number | null | undefined): number | undefined {
  if (value === null || value === undefined) return undefined;
  if (!Number.isSafeInteger(value) || value <= 0) throw new Error("Invalid token budget");
  return value;
}
