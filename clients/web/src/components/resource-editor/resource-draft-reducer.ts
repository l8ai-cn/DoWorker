import {
  type PlanResourceResponse,
  type ValidateResourceResponse,
} from "@proto/orchestration_resource/v1/orchestration_resource_pb";
import type { ResourceApplyResult } from "./resource-apply-result";
import type { ResourceDraft } from "./resource-editor-types";

type Idle = { status: "idle" };
type Loading = { status: "loading"; requestId: string; version: number };
type Failed = { status: "error"; error: string };

export interface ResourceDraftState {
  draft: ResourceDraft;
  mode: "form" | "yaml" | "plan";
  version: number;
  source: {
    text: string;
    dirty: boolean;
    error: string | null;
  };
  validation:
    | Idle
    | Loading
    | Failed
    | { status: "ready"; response: ValidateResourceResponse; version: number };
  plan:
    | Idle
    | Loading
    | Failed
    | { status: "expired" }
    | { status: "ready"; response: PlanResourceResponse; version: number };
  apply:
    | Idle
    | { status: "loading"; planId: string }
    | Failed
    | {
      status: "ready";
      result: ResourceApplyResult;
    };
}

export type ResourceDraftAction =
  | { type: "set_mode"; mode: ResourceDraftState["mode"] }
  | { type: "replace_draft"; draft: ResourceDraft }
  | { type: "source_synced"; text: string }
  | { type: "source_changed"; text: string }
  | { type: "source_invalid"; error: string }
  | { type: "source_parsed"; draft: ResourceDraft }
  | { type: "validation_loading"; requestId: string; version: number }
  | {
    type: "validation_succeeded";
    requestId: string;
    version: number;
    response: ValidateResourceResponse;
  }
  | { type: "validation_failed"; requestId: string; error: string }
  | { type: "plan_loading"; requestId: string; version: number }
  | {
    type: "plan_succeeded";
    requestId: string;
    version: number;
    response: PlanResourceResponse;
  }
  | { type: "plan_failed"; requestId: string; error: string }
  | { type: "plan_expired" }
  | { type: "apply_loading"; planId: string }
  | {
    type: "apply_succeeded";
    result: ResourceApplyResult;
  }
  | { type: "apply_failed"; error: string };

export function createResourceDraftState(
  draft: ResourceDraft,
): ResourceDraftState {
  return {
    draft,
    mode: "form",
    version: 0,
    source: { text: "", dirty: false, error: null },
    validation: { status: "idle" },
    plan: { status: "idle" },
    apply: { status: "idle" },
  };
}

export function resourceDraftReducer(
  state: ResourceDraftState,
  action: ResourceDraftAction,
): ResourceDraftState {
  switch (action.type) {
    case "set_mode":
      return { ...state, mode: action.mode };
    case "replace_draft":
      return changed(state, action.draft, state.source);
    case "source_synced":
      return {
        ...state,
        source: { text: action.text, dirty: false, error: null },
      };
    case "source_changed":
      return changed(state, state.draft, {
        text: action.text,
        dirty: true,
        error: null,
      });
    case "source_invalid":
      return {
        ...state,
        source: { ...state.source, dirty: true, error: action.error },
      };
    case "source_parsed":
      return {
        ...state,
        draft: action.draft,
        source: { ...state.source, dirty: false, error: null },
      };
    case "validation_loading":
      return {
        ...state,
        validation: {
          status: "loading",
          requestId: action.requestId,
          version: action.version,
        },
      };
    case "validation_succeeded":
      if (!matches(state.validation, action.requestId, action.version, state.version)) return state;
      return {
        ...state,
        validation: { status: "ready", response: action.response, version: action.version },
      };
    case "validation_failed":
      if (!matchesRequest(state.validation, action.requestId)) return state;
      return { ...state, validation: { status: "error", error: action.error } };
    case "plan_loading":
      return {
        ...state,
        plan: {
          status: "loading",
          requestId: action.requestId,
          version: action.version,
        },
      };
    case "plan_succeeded":
      if (!matches(state.plan, action.requestId, action.version, state.version)) return state;
      return {
        ...state,
        plan: { status: "ready", response: action.response, version: action.version },
      };
    case "plan_failed":
      if (!matchesRequest(state.plan, action.requestId)) return state;
      return { ...state, plan: { status: "error", error: action.error } };
    case "plan_expired":
      if (state.plan.status !== "ready") return state;
      return { ...state, plan: { status: "expired" } };
    case "apply_loading":
      return { ...state, apply: { status: "loading", planId: action.planId } };
    case "apply_succeeded":
      return { ...state, apply: { status: "ready", result: action.result } };
    case "apply_failed":
      return { ...state, apply: { status: "error", error: action.error } };
  }
}

function changed(
  state: ResourceDraftState,
  draft: ResourceDraft,
  source: ResourceDraftState["source"],
): ResourceDraftState {
  return {
    ...state,
    draft,
    source,
    version: state.version + 1,
    validation: { status: "idle" },
    plan: { status: "idle" },
    apply: { status: "idle" },
  };
}

function matches(
  state: Idle | Loading | Failed | { status: "expired" } | { status: "ready" },
  requestId: string,
  version: number,
  currentVersion: number,
): boolean {
  return state.status === "loading" &&
    state.requestId === requestId &&
    state.version === version &&
    version === currentVersion;
}
function matchesRequest(
  state: Idle | Loading | Failed | { status: "expired" } | { status: "ready" },
  requestId: string,
): boolean {
  return state.status === "loading" && state.requestId === requestId;
}
export { resourceDraftCanApply } from "./resource-draft-selectors";
