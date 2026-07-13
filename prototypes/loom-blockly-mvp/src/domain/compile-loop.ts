import type {
  CompileDiagnostic,
  CompileLoopResult,
  GoalLoopProgram,
  LoopDraft,
  LoopLimits,
} from "./loop-types";

function diagnostic(
  code: string,
  message: string,
  blockId?: string,
): CompileDiagnostic {
  return { code, message, ...(blockId ? { blockId } : {}) };
}

function validPositiveInteger(value: number | undefined): boolean {
  return value !== undefined && Number.isInteger(value) && value > 0;
}

function validLimits(limits: LoopLimits): boolean {
  return (
    validPositiveInteger(limits.maxIterations) &&
    limits.maxIterations <= 100 &&
    (limits.tokenBudget === undefined ||
      validPositiveInteger(limits.tokenBudget)) &&
    validPositiveInteger(limits.timeoutMinutes) &&
    validPositiveInteger(limits.noProgressLimit) &&
    validPositiveInteger(limits.sameErrorLimit)
  );
}

function executionBlockIds(draft: LoopDraft): string[] {
  return [
    draft.rootBlockId,
    draft.worker?.blockId,
    ...draft.instructions.map(({ blockId }) => blockId),
    ...draft.acceptanceCriteria.map(({ blockId }) => blockId),
    draft.verification?.blockId,
    draft.limits?.blockId,
    draft.escalationPolicy?.blockId,
  ].filter((blockId): blockId is string => blockId !== undefined);
}

export function compileLoop(draft: LoopDraft): CompileLoopResult {
  const diagnostics: CompileDiagnostic[] = [];

  if (!draft.worker) {
    diagnostics.push(diagnostic("missing-worker", "A worker is required."));
  }
  if (draft.instructions.length === 0) {
    diagnostics.push(
      diagnostic("missing-instructions", "At least one task is required."),
    );
  }
  if (draft.acceptanceCriteria.length === 0) {
    diagnostics.push(
      diagnostic(
        "missing-acceptance-criteria",
        "At least one acceptance criterion is required.",
      ),
    );
  }
  if (!draft.verification) {
    diagnostics.push(
      diagnostic("missing-verification", "A verification command is required."),
    );
  }
  if (!draft.limits) {
    diagnostics.push(diagnostic("missing-limits", "Loop limits are required."));
  } else if (!validLimits(draft.limits.value)) {
    diagnostics.push(
      diagnostic(
        "invalid-limits",
        "Limits must be positive integers and max iterations cannot exceed 100.",
        draft.limits.blockId,
      ),
    );
  }
  if (!draft.escalationPolicy) {
    diagnostics.push(
      diagnostic("missing-escalation-policy", "An escalation policy is required."),
    );
  }

  for (const blockId of draft.looseBlockIds) {
    diagnostics.push(
      diagnostic("loose-block", "Every block must connect to the root.", blockId),
    );
  }
  for (const unknown of draft.unknownBlockTypes) {
    diagnostics.push(
      diagnostic(
        "unknown-block-type",
        `Unknown block type: ${unknown.type}.`,
        unknown.blockId,
      ),
    );
  }

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
    name: draft.name,
    worker: {
      snapshot_id: worker.value.snapshotId,
      label: worker.value.label,
    },
    objective: draft.instructions.map(({ value }) => value).join("\n"),
    acceptance_criteria: draft.acceptanceCriteria.map(({ value }) => value),
    verification: { kind: "command", command: verification.value },
    limits: {
      max_iterations: limits.value.maxIterations,
      ...(limits.value.tokenBudget === undefined
        ? {}
        : { token_budget: limits.value.tokenBudget }),
      timeout_minutes: limits.value.timeoutMinutes,
      no_progress_limit: limits.value.noProgressLimit,
      same_error_limit: limits.value.sameErrorLimit,
    },
    escalation_policy: escalationPolicy.value,
  };

  return { diagnostics, program, executionBlockIds: blockIds };
}
