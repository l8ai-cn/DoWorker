import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import {
  CompileLoopProgramResponseSchema,
  GenerateLoopProgramRequestSchema,
  RepairLoopProgramRequestSchema,
  RepairLoopProgramResponseSchema,
} from "@proto/goalloop/v1/goalloop_pb";
import {
  getGoalLoopService,
  getLoopBuilderState,
  initWasmCore,
} from "@/lib/wasm-core";
import type { LoopWorkbenchSnapshot } from "@/lib/viewModels/loop-program";
import { readLoopSnapshot } from "./loopProgramConnect";

interface LoopAIRequestSnapshot {
  orgSlug: string;
  currentSource: string;
  modelResourceId: number;
  locale: string;
  revision: number;
}

interface LoopAIDraftRequest extends LoopAIRequestSnapshot {
  prompt: string;
}

export interface LoopAIRepairRequest extends Omit<LoopAIRequestSnapshot, "currentSource"> {
  source: string;
  diagnosticCode: string;
  nodeId: string;
  fieldPath: string;
  prompt: string;
}

export interface LoopAIRepairExpectation {
  revision: number;
  nodeId: string;
  fieldPath: string;
}

function assertModelResourceId(modelResourceId: number) {
  if (!Number.isSafeInteger(modelResourceId) || modelResourceId <= 0) {
    throw new Error("invalid model resource id");
  }
}

export async function requestLoopAIDraft(
  input: LoopAIDraftRequest,
): Promise<Uint8Array> {
  await initWasmCore();
  assertModelResourceId(input.modelResourceId);
  const request = create(GenerateLoopProgramRequestSchema, {
    orgSlug: input.orgSlug,
    prompt: input.prompt,
    currentSource: input.currentSource,
    modelResourceId: BigInt(input.modelResourceId),
    locale: input.locale,
    revision: BigInt(input.revision),
  });
  return new Uint8Array(
    await getGoalLoopService().generateLoopProgramConnect(
      toBinary(GenerateLoopProgramRequestSchema, request),
    ),
  );
}

export async function requestLoopAIRepair(
  input: LoopAIRepairRequest,
): Promise<Uint8Array> {
  await initWasmCore();
  assertModelResourceId(input.modelResourceId);
  if (!input.diagnosticCode || !input.nodeId || !input.fieldPath) {
    throw new Error("invalid repair target");
  }
  const request = create(RepairLoopProgramRequestSchema, {
    orgSlug: input.orgSlug,
    source: input.source,
    modelResourceId: BigInt(input.modelResourceId),
    locale: input.locale,
    revision: BigInt(input.revision),
    diagnosticCode: input.diagnosticCode,
    nodeId: input.nodeId,
    fieldPath: input.fieldPath,
    prompt: input.prompt,
  });
  return new Uint8Array(
    await getGoalLoopService().repairLoopProgramConnect(
      toBinary(RepairLoopProgramRequestSchema, request),
    ),
  );
}

export function decodeLoopAIDraft(response: Uint8Array) {
  const proposal = fromBinary(CompileLoopProgramResponseSchema, response);
  assertValidProposal(proposal);
  return proposal;
}

export function decodeLoopAIRepair(
  response: Uint8Array,
  expected: LoopAIRepairExpectation,
) {
  const repaired = fromBinary(RepairLoopProgramResponseSchema, response);
  if (!repaired.proposal || !repaired.patch) {
    throw new Error("AI returned an incomplete Loop repair");
  }
  assertValidProposal(repaired.proposal);
  if (repaired.proposal.revision !== BigInt(expected.revision)) {
    throw new Error("AI returned a stale Loop repair");
  }
  if (
    repaired.patch.nodeId !== expected.nodeId ||
    repaired.patch.fieldPath !== expected.fieldPath
  ) {
    throw new Error("AI returned a repair for a different target");
  }
  if (repaired.patch.oldValue === repaired.patch.newValue) {
    throw new Error("AI returned an unchanged Loop repair");
  }
  return {
    proposal: repaired.proposal,
    proposalBytes: toBinary(
      CompileLoopProgramResponseSchema,
      repaired.proposal,
    ),
    patch: repaired.patch,
  };
}

export async function applyLoopAIDraft(
  response: Uint8Array,
): Promise<{ applied: boolean; snapshot: LoopWorkbenchSnapshot }> {
  await initWasmCore();
  const applied = getLoopBuilderState().apply_ai_draft_response(response);
  return { applied, snapshot: await readLoopSnapshot() };
}

function assertValidProposal(
  proposal: ReturnType<typeof decodeLoopAIDraft>,
): void {
  if (!proposal.canonicalSource || !proposal.program || proposal.diagnostics.length > 0) {
    throw new Error("AI returned an invalid Loop proposal");
  }
}
