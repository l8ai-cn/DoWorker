"use client";

import { useCallback, useRef } from "react";
import { applyResourcePlan } from "./apply-resource-plan";
import { resourceDraftCanApply, resourceDraftCanSubmit } from "./resource-draft-reducer";
import type { ResourceDraft, ResourceEditorKind } from "./resource-editor-types";
import type { ResourceIdentity } from "./resource-draft-identity";
import { prepareResourceDocument } from "./resource-editor-document";
import {
  safeResourceError,
  switchYamlToForm,
  switchYamlToPlan,
} from "./resource-editor-source-transition";
import {
  hasCurrentApply,
  isCurrentYaml,
} from "./resource-editor-request-guards";
import { useResourcePlanExpiry } from "./use-resource-plan-expiry";
import { useResourceEditorSession } from "./ResourceEditorSessionProvider";
import type { ResourceDraftState } from "./resource-draft-state-types";
import { createResourceEditorInitialState } from "./resource-editor-initial-state";
import { resourceEditorInstanceIdentity } from "./resource-editor-instance-identity";
import { runResourcePlan } from "./resource-editor-plan-runner";
import { runResourceValidation } from "./resource-editor-validation-runner";
import { useLocalResourceDraftState } from "./use-local-resource-draft-state";
export function useResourceEditorController(
  orgSlug: string,
  kind: ResourceEditorKind,
  initialDraft?: ResourceDraft,
  lockedIdentity?: ResourceIdentity,
  sessionKey?: string,
) {
  const identity = resourceEditorInstanceIdentity(
    orgSlug,
    kind,
    sessionKey,
    lockedIdentity,
  );
  const initialStateRef = useRef<{
    identity: string;
    state: ResourceDraftState;
  } | null>(null);
  if (initialStateRef.current?.identity !== identity) {
    initialStateRef.current = {
      identity,
      state: createResourceEditorInitialState(
        orgSlug,
        kind,
        initialDraft,
        lockedIdentity,
      ),
    };
  }
  const initialState = initialStateRef.current.state;
  const [localState, localDispatch] = useLocalResourceDraftState(
    identity,
    initialState,
  );
  const session = useResourceEditorSession(
    sessionKey ? `${identity}:${sessionKey}` : undefined,
    initialState,
  );
  const state = session?.state ?? localState;
  const dispatch = session?.dispatch ?? localDispatch;
  const stateRef = useRef(state);
  stateRef.current = state;
  useResourcePlanExpiry(state.plan, state.apply.status === "loading", dispatch);
  const replaceDraft = useCallback((draft: ResourceDraft) => {
    dispatch({ type: "replace_draft", draft });
  }, [dispatch]);
  const setSource = useCallback((text: string) => {
    dispatch({ type: "source_changed", text });
  }, [dispatch]);
  const prepareDocument = useCallback(
    () => prepareResourceDocument({
      state,
      stateRef,
      kind,
      lockedIdentity,
      dispatch,
    }),
    [dispatch, kind, lockedIdentity, state],
  );
  const runValidation = useCallback(
    () => runResourceValidation({
      orgSlug,
      state,
      dispatch,
      prepareDocument,
    }),
    [dispatch, orgSlug, prepareDocument, state],
  );
  const runPlan = useCallback(
    () => runResourcePlan({
      orgSlug,
      kind,
      state,
      lockedIdentity,
      dispatch,
      prepareDocument,
    }),
    [dispatch, kind, lockedIdentity, orgSlug, prepareDocument, state],
  );
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
    if (mode === "plan" && state.mode === "yaml") {
      return switchYamlToPlan(
        state.source.text,
        kind,
        state.version,
        dispatch,
        () => isCurrentYaml(
          stateRef.current,
          state.version,
          state.source.text,
        ),
        lockedIdentity,
      );
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
        lockedIdentity,
      );
    }
    dispatch({ type: "set_mode", mode });
    return true;
  }, [
    dispatch,
    kind,
    lockedIdentity,
    orgSlug,
    state.draft,
    state.mode,
    state.source.text,
    state.version,
  ]);
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
  }, [dispatch, orgSlug, state]);
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
