import type { Dispatch } from "react";
import {
  IssueSeverity,
  SourceFormat,
} from "@proto/orchestration_resource/v1/orchestration_resource_pb";
import { validateResource } from "@/lib/api/facade/orchestrationResource";
import type { ResourceEditorKind } from "./resource-editor-types";
import type { ResourceDraftAction } from "./resource-draft-reducer";

export async function switchYamlToForm(
  orgSlug: string,
  source: string,
  expectedKind: ResourceEditorKind,
  version: number,
  dispatch: Dispatch<ResourceDraftAction>,
): Promise<boolean> {
  const codec = await import("./resource-yaml-codec");
  try {
    codec.parseResourceYaml(source, expectedKind);
  } catch (error) {
    dispatch({
      type: "source_invalid",
      error: safeResourceError(error, "YAML validation failed."),
    });
    return false;
  }

  const requestId = crypto.randomUUID();
  dispatch({ type: "validation_loading", requestId, version });
  try {
    const response = await validateResource(orgSlug, {
      format: SourceFormat.YAML,
      content: source,
    });
    dispatch({
      type: "validation_succeeded",
      requestId,
      version,
      response,
    });
    if (response.issues.some(
      (issue) => issue.severity === IssueSeverity.BLOCKING,
    )) {
      dispatch({
        type: "source_invalid",
        error: "Resource document failed validation.",
      });
      return false;
    }
    dispatch({
      type: "source_parsed",
      draft: codec.parseCanonicalResourceJson(
        response.canonicalJson,
        expectedKind,
      ),
    });
    dispatch({ type: "set_mode", mode: "form" });
    return true;
  } catch (error) {
    dispatch({
      type: "validation_failed",
      requestId,
      error: safeResourceError(error, "Resource validation failed."),
    });
    return false;
  }
}

export function safeResourceError(
  error: unknown,
  fallback: string,
): string {
  return error instanceof Error && error.message ? error.message : fallback;
}
