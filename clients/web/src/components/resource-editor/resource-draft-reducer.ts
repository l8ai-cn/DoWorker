import type { ResourceDraft } from "./resource-editor-types";
import type {
  ResourceDraftAction,
  ResourceDraftState,
  ResourceRequestState,
} from "./resource-draft-state-types";

export type {
  ResourceDraftAction,
  ResourceDraftState,
} from "./resource-draft-state-types";

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
    case "open_yaml":
      if (action.version !== state.version) return state;
      return {
        ...state,
        mode: "yaml",
        source: { text: action.text, dirty: false, error: null },
      };
    case "source_changed":
      return changed(state, state.draft, {
        text: action.text,
        dirty: true,
        error: null,
      });
    case "source_invalid":
      if (action.version !== state.version) return state;
      return {
        ...state,
        source: { ...state.source, dirty: true, error: action.error },
      };
    case "source_parsed":
      if (action.version !== state.version) return state;
      return {
        ...state,
        draft: action.draft,
        source: { ...state.source, dirty: false, error: null },
      };
    case "yaml_form_loaded":
      if (!matches(
        state.validation,
        action.requestId,
        action.version,
        state.version,
      )) return state;
      return {
        ...state,
        mode: "form",
        draft: action.draft,
        source: { ...state.source, dirty: false, error: null },
        validation: {
          status: "ready",
          response: action.response,
          version: action.version,
        },
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
        apply: { status: "idle" },
        validation: { status: "idle" },
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
        mode: "plan",
        draft: action.draft,
        source: {
          text: action.sourceText,
          dirty: false,
          error: null,
        },
        plan: { status: "ready", response: action.response, version: action.version },
      };
    case "plan_failed":
      if (!matchesRequest(state.plan, action.requestId)) return state;
      return { ...state, plan: { status: "error", error: action.error } };
    case "plan_expired":
      if (state.apply.status === "loading" || state.plan.status !== "ready") {
        return state;
      }
      return {
        ...state,
        plan: { status: "expired", response: state.plan.response,
          version: state.plan.version },
      };
    case "apply_loading":
      if (action.version !== state.version) return state;
      return {
        ...state,
        apply: {
          status: "loading",
          planId: action.planId,
          version: action.version,
        },
      };
    case "apply_succeeded":
      if (!matchesApply(state, action.planId, action.version)) return state;
      return { ...state, apply: { status: "ready", result: action.result } };
    case "apply_failed":
      if (!matchesApply(state, action.planId, action.version)) return state;
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
  state: ResourceRequestState,
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
  state: ResourceRequestState,
  requestId: string,
): boolean {
  return state.status === "loading" && state.requestId === requestId;
}

function matchesApply(
  state: ResourceDraftState,
  planId: string,
  version: number,
): boolean {
  return state.apply.status === "loading" &&
    state.apply.planId === planId &&
    state.apply.version === version &&
    state.version === version;
}

export { resourceDraftCanApply, resourceDraftCanSubmit,
  resourceDraftCanSubmitDraft } from "./resource-draft-selectors";
