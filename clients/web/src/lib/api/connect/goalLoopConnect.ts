import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import {
  CreateGoalLoopRequestSchema,
  GetGoalLoopRequestSchema,
  GoalLoopActionRequestSchema,
  GoalLoopSchema,
  ListGoalLoopsRequestSchema,
  ListGoalLoopsResponseSchema,
  type GoalLoop,
} from "@proto/goalloop/v1/goalloop_pb";
import { getGoalLoopService, initWasmCore } from "@/lib/wasm-core";
import type { CreateGoalLoopInput, GoalLoopData } from "@/lib/viewModels/goal-loop";

function toGoalLoopData(loop: GoalLoop): GoalLoopData {
  return {
    id: Number(loop.id),
    slug: loop.slug,
    name: loop.name,
    description: loop.description,
    worker_spec_snapshot_id: Number(loop.workerSpecSnapshotId),
    objective: loop.objective,
    acceptance_criteria: loop.acceptanceCriteria,
    verification_command: loop.verificationCommand,
    status: loop.status as GoalLoopData["status"],
    pod_key: loop.podKey,
    max_iterations: loop.maxIterations,
    token_budget: loop.tokenBudget === undefined ? undefined : Number(loop.tokenBudget),
    timeout_minutes: loop.timeoutMinutes,
    no_progress_limit: loop.noProgressLimit,
    same_error_limit: loop.sameErrorLimit,
    escalation_policy: loop.escalationPolicy as GoalLoopData["escalation_policy"],
    verification_exit_code: loop.verificationExitCode,
    verification_output: loop.verificationOutput,
    verification_output_truncated: loop.verificationOutputTruncated,
    verification_error: loop.verificationError,
    started_at: loop.startedAt,
    verified_at: loop.verifiedAt,
    completed_at: loop.completedAt,
    created_at: loop.createdAt,
    updated_at: loop.updatedAt,
  };
}

async function service() {
  await initWasmCore();
  return getGoalLoopService();
}

export async function listGoalLoops(orgSlug: string): Promise<GoalLoopData[]> {
  const request = create(ListGoalLoopsRequestSchema, { orgSlug, limit: 100 });
  const response = fromBinary(
    ListGoalLoopsResponseSchema,
    new Uint8Array(await (await service()).listGoalLoopsConnect(toBinary(ListGoalLoopsRequestSchema, request))),
  );
  return response.items.map(toGoalLoopData);
}

export async function getGoalLoop(orgSlug: string, loopSlug: string): Promise<GoalLoopData> {
  const request = create(GetGoalLoopRequestSchema, { orgSlug, loopSlug });
  const response = fromBinary(
    GoalLoopSchema,
    new Uint8Array(await (await service()).getGoalLoopConnect(toBinary(GetGoalLoopRequestSchema, request))),
  );
  return toGoalLoopData(response);
}

export async function createGoalLoop(
  orgSlug: string,
  input: CreateGoalLoopInput,
): Promise<GoalLoopData> {
  const request = create(CreateGoalLoopRequestSchema, {
    orgSlug,
    name: input.name,
    description: input.description ?? "",
    workerSpecSnapshotId: BigInt(input.worker_spec_snapshot_id),
    objective: input.objective,
    acceptanceCriteria: input.acceptance_criteria,
    verificationCommand: input.verification_command,
    maxIterations: input.max_iterations,
    tokenBudget: input.token_budget === undefined ? undefined : BigInt(input.token_budget),
    timeoutMinutes: input.timeout_minutes,
    noProgressLimit: input.no_progress_limit,
    sameErrorLimit: input.same_error_limit,
    escalationPolicy: input.escalation_policy,
  });
  const response = fromBinary(
    GoalLoopSchema,
    new Uint8Array(await (await service()).createGoalLoopConnect(toBinary(CreateGoalLoopRequestSchema, request))),
  );
  return toGoalLoopData(response);
}

async function action(
  method: "startGoalLoopConnect" | "verifyGoalLoopConnect" | "cancelGoalLoopConnect",
  orgSlug: string,
  loopSlug: string,
): Promise<GoalLoopData> {
  const request = create(GoalLoopActionRequestSchema, { orgSlug, loopSlug });
  const response = fromBinary(
    GoalLoopSchema,
    new Uint8Array(await (await service())[method](toBinary(GoalLoopActionRequestSchema, request))),
  );
  return toGoalLoopData(response);
}

export function startGoalLoop(orgSlug: string, loopSlug: string) {
  return action("startGoalLoopConnect", orgSlug, loopSlug);
}

export function verifyGoalLoop(orgSlug: string, loopSlug: string) {
  return action("verifyGoalLoopConnect", orgSlug, loopSlug);
}

export function cancelGoalLoop(orgSlug: string, loopSlug: string) {
  return action("cancelGoalLoopConnect", orgSlug, loopSlug);
}
