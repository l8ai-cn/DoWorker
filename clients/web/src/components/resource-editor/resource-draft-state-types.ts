import type {
  PlanResourceResponse,
  ValidateResourceResponse,
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
    | {
      status: "expired";
      response: PlanResourceResponse;
      version: number;
    }
    | { status: "ready"; response: PlanResourceResponse; version: number };
  apply:
    | Idle
    | { status: "loading"; planId: string; version: number }
    | Failed
    | { status: "ready"; result: ResourceApplyResult };
}

export type ResourceDraftAction =
  | { type: "set_mode"; mode: ResourceDraftState["mode"] }
  | { type: "replace_draft"; draft: ResourceDraft }
  | { type: "open_yaml"; text: string; version: number }
  | { type: "source_changed"; text: string }
  | { type: "source_invalid"; error: string; version: number }
  | { type: "source_parsed"; draft: ResourceDraft; version: number }
  | {
    type: "yaml_form_loaded";
    draft: ResourceDraft;
    requestId: string;
    version: number;
    response: ValidateResourceResponse;
  }
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
    draft: ResourceDraft;
    sourceText: string;
  }
  | { type: "plan_failed"; requestId: string; error: string }
  | { type: "plan_expired" }
  | { type: "apply_loading"; planId: string; version: number }
  | {
    type: "apply_succeeded";
    planId: string;
    version: number;
    result: ResourceApplyResult;
  }
  | { type: "apply_failed"; planId: string; version: number; error: string };

export type ResourceRequestState =
  | Idle
  | Loading
  | Failed
  | { status: "expired" }
  | { status: "ready" };
