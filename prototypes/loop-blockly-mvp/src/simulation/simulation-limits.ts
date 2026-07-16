import type { GoalLoopProgram } from "../domain/loop-types";

export type SimulationStopReason =
  | "max_iterations"
  | "token_budget"
  | "timeout"
  | "no_progress"
  | "same_error"
  | "scenario_exhausted";

export function estimatedTokens(program: GoalLoopProgram): number {
  const text = [
    program.objective,
    ...program.acceptance_criteria,
    program.verification.command,
  ].join("\n");
  return Math.max(500, text.length * 4);
}

export function failureStopReason(
  program: GoalLoopProgram,
  iteration: number,
  noProgressCount: number,
  sameErrorCount: number,
): SimulationStopReason | undefined {
  if (noProgressCount >= program.limits.no_progress_limit) {
    return "no_progress";
  }
  if (sameErrorCount >= program.limits.same_error_limit) {
    return "same_error";
  }
  if (iteration >= program.limits.max_iterations) return "max_iterations";
  return undefined;
}

export function timeoutReached(
  program: GoalLoopProgram,
  startedAt: number,
): boolean {
  return Date.now() - startedAt >= program.limits.timeout_minutes * 60_000;
}
