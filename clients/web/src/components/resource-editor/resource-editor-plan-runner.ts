import type { Dispatch } from "react";
import { IssueSeverity } from "@proto/orchestration_resource/v1/orchestration_resource_pb";
import {
  planResource,
} from "@/lib/api/facade/orchestrationResource";
import {
  assertResourceDraftIdentity,
  type ResourceIdentity,
} from "./resource-draft-identity";
import {
  resourceDraftCanSubmit,
  resourceDraftCanSubmitDraft,
} from "./resource-draft-selectors";
import { safeResourceError } from "./resource-editor-source-transition";
import type {
  ResourceDraftAction,
  ResourceDraftState,
} from "./resource-draft-state-types";
import type { ResourceEditorKind } from "./resource-editor-types";
import type { PreparedResourceDocument } from "./resource-editor-document";

interface RunResourcePlanOptions {
  orgSlug: string;
  kind: ResourceEditorKind;
  state: ResourceDraftState;
  lockedIdentity?: ResourceIdentity;
  dispatch: Dispatch<ResourceDraftAction>;
  prepareDocument: () => Promise<PreparedResourceDocument>;
}

export async function runResourcePlan({
  orgSlug,
  kind,
  state,
  lockedIdentity,
  dispatch,
  prepareDocument,
}: RunResourcePlanOptions) {
  if (!resourceDraftCanSubmit(state)) return null;
  const requestId = crypto.randomUUID();
  const version = state.version;
  dispatch({ type: "plan_loading", requestId, version });
  try {
    const prepared = await prepareDocument();
    if (prepared.draft.kind === "WorkerTemplate") {
      const { assertWorkerTemplatePlanReady } = await import(
        "./worker-template-plan-readiness"
      );
      await assertWorkerTemplatePlanReady(orgSlug, prepared.draft);
    }
    const response = await planResource(orgSlug, prepared.document);
    const codec = await import("./resource-yaml-codec");
    const blocking = response.issues.some(
      (issue) => issue.severity === IssueSeverity.BLOCKING,
    );
    if (!response.plan && !blocking) {
      throw new Error("Resource planning response did not include a plan.");
    }
    if (!response.plan) {
      dispatch({
        type: "plan_succeeded",
        requestId,
        version,
        response,
        draft: prepared.draft,
        sourceText: codec.stringifyResourceYaml(prepared.draft),
      });
      return response;
    }
    if (response.canonicalJson.length === 0) {
      throw new Error(
        "Resource planning response did not include a canonical document.",
      );
    }
    const draft = codec.parseCanonicalResourceJson(response.canonicalJson, kind);
    assertResourceDraftIdentity(draft, lockedIdentity);
    if (!resourceDraftCanSubmitDraft(draft)) {
      throw new Error("Canonical GoalLoop contains invalid integer fields.");
    }
    dispatch({
      type: "plan_succeeded",
      requestId,
      version,
      response,
      draft,
      sourceText: codec.stringifyResourceYaml(draft),
    });
    return response;
  } catch (error) {
    dispatch({
      type: "plan_failed",
      requestId,
      error: safeResourceError(error, "Resource planning failed."),
    });
    return null;
  }
}
