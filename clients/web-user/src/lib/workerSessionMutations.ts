import { authenticatedFetch } from "./identity";
import {
  createWorkerSessionBody,
  crossAgentWorkerSessionBody,
  importWorkerSessionBody,
} from "./workerSessionRequestBodies";
import type { SessionEventInput } from "./types";
import type { SessionInteractionMode } from "./workerSessionPlan";
import type { WorkerCreationSelection } from "./workerSessionPlan";

interface WorkerSessionResult {
  id: string;
}

export async function createWorkerSession(input: WorkerCreationSelection & {
  agentId: string;
  initialItems: SessionEventInput[];
  hostId?: string;
  workspace?: string;
  tokenBudget?: number | null;
  parentSessionId?: string;
  subAgentName?: string | null;
  title?: string;
  mode?: SessionInteractionMode;
  modelResourceId?: number;
}): Promise<WorkerSessionResult> {
  return postWorkerSession("/v1/sessions", createWorkerSessionBody(input));
}

export async function importWorkerSession(input: WorkerCreationSelection & {
  agentId: string;
  sourcePath: string;
  title?: string;
  hostId?: string;
  mode?: SessionInteractionMode;
  modelResourceId?: number;
}): Promise<WorkerSessionResult> {
  return postWorkerSession("/v1/sessions/import", importWorkerSessionBody(input));
}

export async function forkWorkerSession(input: WorkerCreationSelection & {
  sourceId: string;
  sourceAgentId: string;
  agentId: string;
  title?: string;
  upToResponseId?: string;
  mode?: SessionInteractionMode;
  modelResourceId?: number;
}): Promise<WorkerSessionResult> {
  return postWorkerSession(
    `/v1/sessions/${encodeURIComponent(input.sourceId)}/fork`,
    crossAgentWorkerSessionBody(input),
  );
}

export async function forkSnapshotSession(input: {
  sourceId: string;
  title?: string;
  upToResponseId?: string;
}): Promise<WorkerSessionResult> {
  return postWorkerSession(`/v1/sessions/${encodeURIComponent(input.sourceId)}/fork`, {
    ...(input.title === undefined ? {} : { title: input.title }),
    ...(input.upToResponseId === undefined ? {} : { up_to_response_id: input.upToResponseId }),
  });
}

export async function switchWorkerSessionAgent(input: WorkerCreationSelection & {
  sessionId: string;
  sourceAgentId: string;
  agentId: string;
  mode?: SessionInteractionMode;
  modelResourceId?: number;
}): Promise<WorkerSessionResult> {
  return postWorkerSession(
    `/v1/sessions/${encodeURIComponent(input.sessionId)}/switch-agent`,
    crossAgentWorkerSessionBody(input),
  );
}

async function postWorkerSession(
  path: string,
  body: Promise<Record<string, unknown>> | Record<string, unknown>,
): Promise<WorkerSessionResult> {
  const response = await authenticatedFetch(path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(await body),
  });
  if (!response.ok) {
    throw new Error(await sessionMutationError(response));
  }
  const result = (await response.json()) as { id?: unknown; session?: { id?: unknown } };
  const id = result.id ?? result.session?.id;
  if (typeof id !== "string" || !id) throw new Error("Worker session response is missing an id");
  return { id };
}

async function sessionMutationError(response: Response): Promise<string> {
  const text = (await response.text()).trim();
  if (!text) return `Worker session request failed (${response.status})`;
  try {
    const body = JSON.parse(text) as Record<string, unknown>;
    for (const key of ["detail", "message", "error"]) {
      if (typeof body[key] === "string" && body[key].trim()) return body[key];
    }
  } catch {}
  return text;
}
