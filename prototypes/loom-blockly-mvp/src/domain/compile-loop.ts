import type { CompileResult, GoalLoopProgram, LoopDraft } from "./loop-types";
import { validateLoopDraft } from "./validate-loop-draft";

function blockId(value: unknown): string | undefined {
  if (typeof value !== "object" || value === null) return undefined;
  const id = (value as { blockId?: unknown }).blockId;
  return typeof id === "string" && id.trim() !== "" ? id : undefined;
}

function blockIds(value: unknown): string[] {
  if (!Array.isArray(value)) return [];
  return value.map(blockId).filter((id): id is string => id !== undefined);
}

function executionBlockIds(draft: unknown): string[] {
  if (typeof draft !== "object" || draft === null) return [];
  const value = draft as Record<string, unknown>;
  return [
    typeof value.rootBlockId === "string" && value.rootBlockId.trim() !== ""
      ? value.rootBlockId
      : undefined,
    blockId(value.worker),
    blockId(value.limits),
    blockId(value.escalationPolicy),
    ...blockIds(value.instructions),
    ...blockIds(value.acceptanceCriteria),
    blockId(value.verification),
  ].filter((id): id is string => id !== undefined);
}

export function compileLoop(draft: LoopDraft): CompileResult {
  const diagnostics = validateLoopDraft(draft);
  const blockIds = executionBlockIds(draft);
  if (diagnostics.length > 0) {
    return { diagnostics, executionBlockIds: blockIds };
  }

  const worker = draft.worker!;
  const verification = draft.verification!;
  const limits = draft.limits!;
  const escalationPolicy = draft.escalationPolicy!;
  const program: GoalLoopProgram = {
    kind: "goal-loop-program",
    schema_version: 1,
    name: draft.name.trim(),
    worker: {
      snapshot_id: worker.value.snapshotId,
      label: worker.value.label,
    },
    objective: draft.instructions.map(({ value }) => value.trim()).join("\n"),
    acceptance_criteria: draft.acceptanceCriteria.map(({ value }) => value.trim()),
    verification: { kind: "command", command: verification.value.trim() },
    limits: {
      max_iterations: limits.value.maxIterations,
      token_budget: limits.value.tokenBudget,
      timeout_minutes: limits.value.timeoutMinutes,
      no_progress_limit: limits.value.noProgressLimit,
      same_error_limit: limits.value.sameErrorLimit,
    },
    escalation_policy: escalationPolicy.value,
  };

  return { diagnostics, program, executionBlockIds: blockIds };
}
