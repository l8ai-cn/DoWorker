"use client";

import { useCallback, useEffect, useReducer } from "react";
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
  resourceDraftReducer,
} from "./resource-draft-reducer";
import type {
  ResourceDraft,
  ResourceEditorKind,
} from "./resource-editor-types";
import {
  safeResourceError,
  switchYamlToForm,
} from "./resource-editor-source-transition";
import { createResourceDraft } from "./resource-draft-factory";

export function useResourceEditorController(
  orgSlug: string,
  kind: ResourceEditorKind,
) {
  const [state, dispatch] = useReducer(
    resourceDraftReducer,
    orgSlug,
    (namespace) => createResourceDraftState(createResourceDraft(kind, namespace)),
  );

  useEffect(() => {
    if (state.plan.status !== "ready" || !state.plan.response.plan) return;
    const expiresAt = Date.parse(state.plan.response.plan.expiresAt);
    const delay = Number.isFinite(expiresAt)
      ? Math.max(0, expiresAt - Date.now())
      : 0;
    const timer = window.setTimeout(() => {
      dispatch({ type: "plan_expired" });
    }, Math.min(delay, 2_147_483_647));
    return () => window.clearTimeout(timer);
  }, [state.plan]);

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
    const codec = await import("./resource-yaml-codec");
    try {
      const draft = codec.parseResourceYaml(state.source.text, kind);
      dispatch({ type: "source_parsed", draft });
      return { format: SourceFormat.YAML, content: state.source.text };
    } catch (error) {
      const message = safeResourceError(error, "YAML validation failed.");
      dispatch({ type: "source_invalid", error: message });
      throw error;
    }
  }, [kind, state.draft, state.mode, state.source.text]);

  const runValidation = useCallback(async () => {
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
  }, [orgSlug, prepareDocument, state.version]);

  const runPlan = useCallback(async () => {
    const requestId = crypto.randomUUID();
    const version = state.version;
    dispatch({ type: "plan_loading", requestId, version });
    try {
      const document = await prepareDocument();
      const response = await planResource(orgSlug, document);
      dispatch({ type: "plan_succeeded", requestId, version, response });
      dispatch({ type: "set_mode", mode: "plan" });
      return response;
    } catch (error) {
      dispatch({
        type: "plan_failed",
        requestId,
        error: safeResourceError(error, "Resource planning failed."),
      });
      return null;
    }
  }, [orgSlug, prepareDocument, state.version]);

  const setMode = useCallback(async (
    mode: "form" | "yaml" | "plan",
  ): Promise<boolean> => {
    if (mode === state.mode) return true;
    if (mode === "yaml") {
      const codec = await import("./resource-yaml-codec");
      dispatch({
        type: "source_synced",
        text: codec.stringifyResourceYaml(state.draft),
      });
      dispatch({ type: "set_mode", mode });
      return true;
    }
    if (mode === "form" && state.mode === "yaml") {
      return switchYamlToForm(
        orgSlug,
        state.source.text,
        kind,
        state.version,
        dispatch,
      );
    }
    dispatch({ type: "set_mode", mode });
    return true;
  }, [kind, orgSlug, state.draft, state.mode, state.source.text, state.version]);

  const apply = useCallback(async () => {
    if (!resourceDraftCanApply(state) || state.plan.status !== "ready") return null;
    const planId = state.plan.response.plan?.planId;
    if (!planId) return null;
    dispatch({ type: "apply_loading", planId });
    try {
      const result = await applyResourcePlan(orgSlug, state.draft.kind, planId);
      dispatch({ type: "apply_succeeded", result });
      return result;
    } catch (error) {
      dispatch({
        type: "apply_failed",
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
    canApply: resourceDraftCanApply(state),
  };
}
