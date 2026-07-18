import type { Dispatch } from "react";
import {
  IssueSeverity,
  SourceFormat,
} from "@proto/orchestration_resource/v1/orchestration_resource_types_pb";
import { validateResource } from "@/lib/api/facade/orchestrationResource";
import { safeServiceErrorMessage } from "@/lib/errors/safeServiceErrorMessage";
import { resourceDraftCanSubmitDraft } from "./resource-draft-selectors";
import {
  assertResourceDraftIdentity,
  type ResourceIdentity,
} from "./resource-draft-identity";
import type { ResourceEditorKind } from "./resource-editor-types";
import type { ResourceDraftAction } from "./resource-draft-reducer";

export async function switchYamlToForm(
  orgSlug: string,
  source: string,
  expectedKind: ResourceEditorKind,
  version: number,
  dispatch: Dispatch<ResourceDraftAction>,
  isCurrent: () => boolean,
  lockedIdentity?: ResourceIdentity,
): Promise<boolean> {
  const codec = await import("./resource-yaml-codec");
  if (!isCurrent()) return false;
  try {
    const draft = codec.parseResourceYaml(source, expectedKind);
    assertResourceDraftIdentity(draft, lockedIdentity);
    if (!resourceDraftCanSubmitDraft(draft)) {
      throw new Error("GoalLoop YAML contains invalid integer fields.");
    }
  } catch (error) {
    if (isCurrent()) {
      dispatch({
        type: "source_invalid",
        error: safeResourceError(error, "YAML validation failed."),
        version,
      });
    }
    return false;
  }

  if (!isCurrent()) return false;
  const requestId = crypto.randomUUID();
  dispatch({ type: "validation_loading", requestId, version });
  try {
    const response = await validateResource(orgSlug, {
      format: SourceFormat.YAML,
      content: source,
    });
    if (!isCurrent()) return false;
    if (response.issues.some(
      (issue) => issue.severity === IssueSeverity.BLOCKING,
    )) {
      dispatch({
        type: "validation_succeeded",
        requestId,
        version,
        response,
      });
      dispatch({
        type: "source_invalid",
        error: "Resource document failed validation.",
        version,
      });
      return false;
    }
    const draft = codec.parseCanonicalResourceJson(
      response.canonicalJson,
      expectedKind,
    );
    assertResourceDraftIdentity(draft, lockedIdentity);
    if (!isCurrent()) return false;
    dispatch({
      type: "yaml_form_loaded",
      requestId,
      version,
      response,
      draft,
    });
    return true;
  } catch (error) {
    if (isCurrent()) {
      dispatch({
        type: "validation_failed",
        requestId,
        error: safeResourceError(error, "Resource validation failed."),
      });
    }
    return false;
  }
}

export async function switchYamlToPlan(
  source: string,
  expectedKind: ResourceEditorKind,
  version: number,
  dispatch: Dispatch<ResourceDraftAction>,
  isCurrent: () => boolean,
  lockedIdentity?: ResourceIdentity,
): Promise<boolean> {
  const codec = await import("./resource-yaml-codec");
  if (!isCurrent()) return false;
  try {
    const draft = codec.parseResourceYaml(source, expectedKind);
    assertResourceDraftIdentity(draft, lockedIdentity);
    if (!resourceDraftCanSubmitDraft(draft)) {
      throw new Error("GoalLoop YAML contains invalid integer fields.");
    }
    if (!isCurrent()) return false;
    dispatch({ type: "source_parsed", draft, version });
    dispatch({ type: "set_mode", mode: "plan" });
    return true;
  } catch (error) {
    if (isCurrent()) {
      dispatch({
        type: "source_invalid",
        error: safeResourceError(error, "YAML validation failed."),
        version,
      });
    }
    return false;
  }
}

export function safeResourceError(
  error: unknown,
  fallback: string,
): string {
  return safeServiceErrorMessage(error, fallback);
}
