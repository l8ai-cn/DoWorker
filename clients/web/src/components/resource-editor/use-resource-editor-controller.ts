"use client";

import { useCallback, useReducer, useRef } from "react";
import { SourceFormat } from "@proto/orchestration_resource/v1/orchestration_resource_pb";
import {
  planResource,
  validateResource,
  type ResourceDocument,
} from "@/lib/api/facade/orchestrationResource";
import { applyResourcePlan } from "./apply-resource-plan";
import {
  createResourceDraftState,
  resourceDraftCanApply,
  resourceDraftCanSubmit,
  resourceDraftCanSubmitDraft,
  resourceDraftReducer,
} from "./resource-draft-reducer";
import type { ResourceDraft, ResourceEditorKind } from "./resource-editor-types";
import {
  safeResourceError,
  switchYamlToForm,
} from "./resource-editor-source-transition";
import { createResourceDraft } from "./resource-draft-factory";
import {
  hasCurrentApply,
  isCurrentYaml,
} from "./resource-editor-request-guards";
import { useResourcePlanExpiry } from "./use-resource-plan-expiry";

export function useResourceEditorController(
  orgSlug: string,
  kind: ResourceEditorKind,
) {
  const [state, dispatch] = useReducer(
    resourceDraftReducer,
    orgSlug,
    (namespace) => createResourceDraftState(createResourceDraft(kind, namespace)),
  );
  const stateRef = useRef(state);
  stateRef.current = state;

  useResourcePlanExpiry(state.plan, state.apply.status === "loading", dispatch);

  const replaceDraft = useCallback((draft: ResourceDraft) => {
    dispatch({ type: "replace_draft", draft });
  }, []);

  const setSource = useCallback((text: string) => {
    dispatch({ type: "source_changed", text });
  }, []);

  const prepareDocument = useCallback(async (): Promise<ResourceDocument> => {
    if (state.mode !== "yaml") {
      return {
        format: SourceFormat.JSON,
        content: JSON.stringify(state.draft),
      };
    }
    const version = state.version;
    const source = state.source.text;
    const codec = await import("./resource-yaml-codec");
    if (!isCurrentYaml(stateRef.current, version, source)) {
      throw new Error("Resource draft changed while preparing YAML.");
    }
    try {
      const draft = codec.parseResourceYaml(source, kind);
      if (!resourceDraftCanSubmitDraft(draft)) {
        throw new Error("GoalLoop YAML contains invalid integer fields.");
      }
      dispatch({ type: "source_parsed", draft, version });
      return { format: SourceFormat.YAML, content: source };
    } catch (error) {
      if (isCurrentYaml(stateRef.current, version, source)) {
        const message = safeResourceError(error, "YAML validation failed.");
        dispatch({ type: "source_invalid", error: message, version });
      }
      throw error;
    }
  }, [kind, state.draft, state.mode, state.source.text, state.version]);

  const runValidation = useCallback(async () => {
    if (!resourceDraftCanSubmit(state)) return null;
    const requestId = crypto.randomUUID();
    const version = state.version;
    dispatch({ type: "validation_loading", requestId, version });
    try {
      const document = await prepareDocument();
      const response = await validateResource(orgSlug, document);
      dispatch({
        type: "validation_succeeded",
        requestId,
        version,
        response,
      });
      return response;
    } catch (error) {
      dispatch({
        type: "validation_failed",
        requestId,
        error: safeResourceError(error, "Resource validation failed."),
      });
      return null;
    }
  }, [orgSlug, prepareDocument, state]);

  const runPlan = useCallback(async () => {
    if (!resourceDraftCanSubmit(state)) return null;
    const requestId = crypto.randomUUID();
    const version = state.version;
    dispatch({ type: "plan_loading", requestId, version });
    try {
      const document = await prepareDocument();
      const response = await planResource(orgSlug, document);
      dispatch({ type: "plan_succeeded", requestId, version, response });
      return response;
    } catch (error) {
      dispatch({
        type: "plan_failed",
        requestId,
        error: safeResourceError(error, "Resource planning failed."),
      });
      return null;
    }
  }, [orgSlug, prepareDocument, state]);

  const setMode = useCallback(async (
    mode: "form" | "yaml" | "plan",
  ): Promise<boolean> => {
    if (mode === state.mode) return true;
    if (mode === "yaml") {
      const version = state.version;
      const currentMode = state.mode;
      const codec = await import("./resource-yaml-codec");
      if (
        stateRef.current.version !== version ||
        stateRef.current.mode !== currentMode
      ) return false;
      dispatch({
        type: "open_yaml",
        text: codec.stringifyResourceYaml(state.draft),
        version,
      });
      return true;
    }
    if (mode === "form" && state.mode === "yaml") {
      return switchYamlToForm(
        orgSlug,
        state.source.text,
        kind,
        state.version,
        dispatch,
        () => isCurrentYaml(
          stateRef.current,
          state.version,
          state.source.text,
        ),
      );
    }
    dispatch({ type: "set_mode", mode });
    return true;
  }, [kind, orgSlug, state.draft, state.mode, state.source.text, state.version]);

  const apply = useCallback(async () => {
    if (!resourceDraftCanApply(state) || state.plan.status !== "ready") return null;
    const planId = state.plan.response.plan?.planId;
    if (!planId) return null;
    const version = state.version;
    dispatch({ type: "apply_loading", planId, version });
    try {
      const result = await applyResourcePlan(orgSlug, state.draft.kind, planId);
      if (!hasCurrentApply(stateRef.current, planId, version)) return null;
      dispatch({ type: "apply_succeeded", planId, version, result });
      return result;
    } catch (error) {
      if (!hasCurrentApply(stateRef.current, planId, version)) return null;
      dispatch({
        type: "apply_failed",
        planId,
        version,
        error: safeResourceError(error, "Resource apply failed."),
      });
      return null;
    }
  }, [orgSlug, state]);

  return {
    state,
    replaceDraft,
    setSource,
    setMode,
    runValidation,
    runPlan,
    apply,
    canSubmit: resourceDraftCanSubmit(state),
    canApply: resourceDraftCanApply(state),
  };
}
