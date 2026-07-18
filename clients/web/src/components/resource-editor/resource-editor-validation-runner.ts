import type { Dispatch } from "react";
import {
  validateResource,
} from "@/lib/api/facade/orchestrationResource";
import { resourceDraftCanSubmit } from "./resource-draft-selectors";
import { safeResourceError } from "./resource-editor-source-transition";
import type {
  ResourceDraftAction,
  ResourceDraftState,
} from "./resource-draft-state-types";
import type { PreparedResourceDocument } from "./resource-editor-document";

interface RunResourceValidationOptions {
  orgSlug: string;
  state: ResourceDraftState;
  dispatch: Dispatch<ResourceDraftAction>;
  prepareDocument: () => Promise<PreparedResourceDocument>;
}

export async function runResourceValidation({
  orgSlug,
  state,
  dispatch,
  prepareDocument,
}: RunResourceValidationOptions) {
  if (!resourceDraftCanSubmit(state)) return null;
  const requestId = crypto.randomUUID();
  const version = state.version;
  dispatch({ type: "validation_loading", requestId, version });
  try {
    const prepared = await prepareDocument();
    const response = await validateResource(orgSlug, prepared.document);
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
}
