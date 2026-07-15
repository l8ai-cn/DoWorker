import { IssueSeverity } from "@proto/orchestration_resource/v1/orchestration_resource_pb";
import { goalLoopHasIntegerErrors } from "./goal-loop-integer-draft";
import type { ResourceDraftState } from "./resource-draft-reducer";

export function resourceDraftCanSubmit(state: ResourceDraftState): boolean {
  return state.draft.kind !== "GoalLoop" ||
    !goalLoopHasIntegerErrors(state.draft);
}

export function resourceDraftCanApply(state: ResourceDraftState): boolean {
  if (!resourceDraftCanSubmit(state)) return false;
  if (state.source.dirty || state.source.error || state.plan.status !== "ready") {
    return false;
  }
  if (state.plan.version !== state.version || !state.plan.response.plan) {
    return false;
  }
  const expiresAt = Date.parse(state.plan.response.plan.expiresAt);
  if (!Number.isFinite(expiresAt) || expiresAt <= Date.now()) return false;
  return !state.plan.response.issues.some(
    (issue) => issue.severity === IssueSeverity.BLOCKING,
  );
}
