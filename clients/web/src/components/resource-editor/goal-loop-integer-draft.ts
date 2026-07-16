import type {
  GoalLoopDraft,
  GoalLoopIntegerDraft,
} from "./resource-editor-types";

export type GoalLoopIntegerField =
  | "maxIterations"
  | "tokenBudget"
  | "timeoutMinutes"
  | "noProgressLimit"
  | "sameErrorLimit";

export type GoalLoopIntegerError =
  | { code: "required" }
  | { code: "integer" }
  | { code: "safeInteger" }
  | { code: "range"; min: number; max?: number };

const CONSTRAINTS: Record<
  GoalLoopIntegerField,
  { min: number; max?: number; optional?: boolean }
> = {
  maxIterations: { min: 1, max: 100 },
  tokenBudget: { min: 1, optional: true },
  timeoutMinutes: { min: 1, max: 1440 },
  noProgressLimit: { min: 1, max: 20 },
  sameErrorLimit: { min: 1, max: 20 },
};

export function parseGoalLoopIntegerDraft(
  value: string,
): GoalLoopIntegerDraft {
  if (!/^-?\d+$/.test(value)) return value;
  const parsed = Number(value);
  return Number.isSafeInteger(parsed) ? parsed : value;
}

export function goalLoopIntegerError(
  field: GoalLoopIntegerField,
  value: GoalLoopIntegerDraft | undefined,
): GoalLoopIntegerError | null {
  const constraint = CONSTRAINTS[field];
  if (value === undefined && constraint.optional) return null;
  if (value === undefined) return { code: "required" };
  if (value === "") return { code: "required" };
  if (typeof value === "string") {
    if (!/^-?\d+$/.test(value)) return { code: "integer" };
    return { code: "safeInteger" };
  }
  if (!Number.isSafeInteger(value)) return { code: "safeInteger" };
  if (
    value < constraint.min ||
    (constraint.max !== undefined && value > constraint.max)
  ) {
    return {
      code: "range",
      min: constraint.min,
      max: constraint.max,
    };
  }
  return null;
}

export function goalLoopHasIntegerErrors(draft: GoalLoopDraft): boolean {
  return (Object.keys(CONSTRAINTS) as GoalLoopIntegerField[]).some(
    (field) => goalLoopIntegerError(field, draft.spec[field]) !== null,
  );
}
