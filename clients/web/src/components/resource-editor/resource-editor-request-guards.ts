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

export function hasCurrentApply(
  state: ResourceDraftState,
  planId: string,
  version: number,
): boolean {
  return state.version === version &&
    state.apply.status === "loading" &&
    state.apply.planId === planId &&
    state.apply.version === version;
}
