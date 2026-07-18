import type { PodData } from "@/lib/api";
import type {
  WorkerDraftFillResult,
  WorkerPreflightResult,
  WorkerSpecDraft,
} from "@/lib/api/facade/podConnect";
import { createInitialWorkerDraftState } from "./workerCreateInitialState";
export { createInitialWorkerDraftState } from "./workerCreateInitialState";
export type WorkerCreateStepId = 1 | 2 | 3 | 4;
export type AsyncState<T> =
  | { status: "idle" }
  | { status: "loading" }
  | { status: "ready"; data: T }
  | { status: "error"; error: string };
export interface WorkerCreateDraftState {
  instanceId: string;
  step: WorkerCreateStepId;
  fillPrompt: string;
  generationModelResourceId: number;
  draft: WorkerSpecDraft;
  fill: AsyncState<WorkerDraftFillResult>;
  fillRequestId: string | null;
  preflight: AsyncState<WorkerPreflightResult>;
  preflightRequestId: string | null;
  create: AsyncState<PodData>;
}

export type WorkerCreateDraftAction =
  | { type: "reset"; draft?: Partial<WorkerSpecDraft> }
  | { type: "set_fill_prompt"; prompt: string }
  | { type: "set_generation_model"; resourceId: number }
  | { type: "patch_draft"; patch: Partial<WorkerSpecDraft> }
  | {
      type: "change_worker_type";
      workerTypeSlug: string;
      schemaVersion: number;
    }
  | {
      type: "set_lifecycle";
      terminationPolicy: string;
      idleTimeoutMinutes: number;
    }
  | { type: "set_step"; step: WorkerCreateStepId }
  | { type: "fill_loading"; requestId: string }
  | {
      type: "fill_succeeded";
      requestId: string;
      result: WorkerDraftFillResult;
    }
  | { type: "fill_failed"; requestId: string; error: string }
  | { type: "preflight_loading"; requestId: string }
  | {
      type: "preflight_succeeded";
      requestId: string;
      result: WorkerPreflightResult;
    }
  | { type: "preflight_failed"; requestId: string; error: string }
  | { type: "create_loading" }
  | { type: "create_succeeded"; pod: PodData }
  | { type: "create_failed"; error: string };

export function createInitialWorkerDraftState(
  initial?: Partial<WorkerSpecDraft>,
): WorkerCreateDraftState {
  return {
    instanceId: crypto.randomUUID(),
    step: 1,
    fillPrompt: "",
    draft: {
      model_resource_id: 0,
      tool_model_resource_ids: {},
      worker_type_slug: "",
      runtime_image_id: 0,
      placement_policy: "automatic",
      compute_target_id: 0,
      deployment_mode: "",
      resource_profile_id: 0,
      type_schema_version: 0,
      type_config_values: {},
      secret_refs: [],
      interaction_mode: "acp",
      automation_level: "autonomous",
      branch: "",
      skill_ids: [],
      knowledge_mounts: [],
      env_bundle_ids: [],
      config_document_bindings: [],
      instructions: "",
      initial_task: "",
      termination_policy: "manual",
      idle_timeout_minutes: 0,
      alias: "",
      options_revision: "",
      ...initial,
    },
    fill: { status: "idle" },
    fillRequestId: null,
    preflight: { status: "idle" },
    preflightRequestId: null,
    create: { status: "idle" },
  };
}
export function workerCreateDraftReducer(
  state: WorkerCreateDraftState,
  action: WorkerCreateDraftAction,
): WorkerCreateDraftState {
  switch (action.type) {
    case "reset":
      return createInitialWorkerDraftState(action.draft);
    case "set_fill_prompt":
      return { ...state, fillPrompt: action.prompt };
    case "set_generation_model":
      return { ...state, generationModelResourceId: action.resourceId };
    case "patch_draft":
      return invalidatePreflight(state, { ...state.draft, ...action.patch });
    case "change_worker_type":
      return invalidatePreflight(state, {
        ...state.draft,
        model_resource_id: 0,
        tool_model_resource_ids: {},
        worker_type_slug: action.workerTypeSlug,
        runtime_image_id: 0,
        type_schema_version: action.schemaVersion,
        type_config_values: {},
        secret_refs: [],
        skill_ids: [],
        env_bundle_ids: [],
        config_document_bindings: [],
      });
    case "set_lifecycle":
      return invalidatePreflight(state, {
        ...state.draft,
        termination_policy: action.terminationPolicy,
        idle_timeout_minutes:
          action.terminationPolicy === "idle" ? action.idleTimeoutMinutes : 0,
      });
    case "set_step":
      return { ...state, step: action.step };
    case "fill_loading":
      return {
        ...state,
        fill: { status: "loading" },
        fillRequestId: action.requestId,
        preflight: { status: "idle" },
        preflightRequestId: null,
        create: { status: "idle" },
      };
    case "fill_succeeded":
      if (state.fillRequestId !== action.requestId) return state;
      return {
        ...state,
        draft: action.result.draft,
        fill: { status: "ready", data: action.result },
        fillRequestId: null,
        preflight: { status: "idle" },
        preflightRequestId: null,
        create: { status: "idle" },
      };
    case "fill_failed":
      if (state.fillRequestId !== action.requestId) return state;
      return {
        ...state,
        fill: { status: "error", error: action.error },
        fillRequestId: null,
      };
    case "preflight_loading":
      return {
        ...state,
        preflight: { status: "loading" },
        preflightRequestId: action.requestId,
      };
    case "preflight_succeeded":
      if (state.preflightRequestId !== action.requestId) return state;
      return {
        ...state,
        preflight: { status: "ready", data: action.result },
        preflightRequestId: null,
      };
    case "preflight_failed":
      if (state.preflightRequestId !== action.requestId) return state;
      return {
        ...state,
        preflight: { status: "error", error: action.error },
        preflightRequestId: null,
      };
    case "create_loading":
      return { ...state, create: { status: "loading" } };
    case "create_succeeded":
      return { ...state, create: { status: "ready", data: action.pod } };
    case "create_failed":
      return { ...state, create: { status: "error", error: action.error } };
  }
}

function invalidatePreflight(
  state: WorkerCreateDraftState,
  draft: WorkerSpecDraft,
): WorkerCreateDraftState {
  return {
    ...state,
    draft,
    fill: { status: "idle" },
    fillRequestId: null,
    preflight: { status: "idle" },
    preflightRequestId: null,
    create: { status: "idle" },
  };
}
