import { IssueSeverity } from "@proto/orchestration_resource/v1/orchestration_resource_pb";
import type { ResourceDraftState } from "./resource-draft-reducer";

export function resourceDraftCanApply(state: ResourceDraftState): boolean {
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
