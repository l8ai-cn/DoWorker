import type { ResourceDraftState } from "./resource-draft-reducer";

export function isCurrentYaml(
  state: ResourceDraftState,
  version: number,
  source: string,
): boolean {
  return state.mode === "yaml" &&
    state.version === version &&
    state.source.text === source;
}

export function hasCurrentPlan(
  state: ResourceDraftState,
  planId: string,
  version: number,
): boolean {
  return state.version === version &&
    state.plan.status === "ready" &&
    state.plan.response.plan?.planId === planId;
}
