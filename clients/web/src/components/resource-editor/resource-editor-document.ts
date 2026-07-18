import type { Dispatch, MutableRefObject } from "react";
import { SourceFormat } from "@proto/orchestration_resource/v1/orchestration_resource_types_pb";
import type { ResourceDocument } from "@/lib/api/facade/orchestrationResource";
import {
  resourceDraftCanSubmitDraft,
  type ResourceDraftAction,
  type ResourceDraftState,
} from "./resource-draft-reducer";
import {
  assertResourceDraftIdentity,
  type ResourceIdentity,
} from "./resource-draft-identity";
import { isCurrentYaml } from "./resource-editor-request-guards";
import { safeResourceError } from "./resource-editor-source-transition";
import type { ResourceEditorKind } from "./resource-editor-types";

interface PrepareResourceDocumentOptions {
  state: ResourceDraftState;
  stateRef: MutableRefObject<ResourceDraftState>;
  kind: ResourceEditorKind;
  lockedIdentity?: ResourceIdentity;
  dispatch: Dispatch<ResourceDraftAction>;
}

export interface PreparedResourceDocument {
  document: ResourceDocument;
  draft: ResourceDraftState["draft"];
}

export async function prepareResourceDocument({
  state,
  stateRef,
  kind,
  lockedIdentity,
  dispatch,
}: PrepareResourceDocumentOptions): Promise<PreparedResourceDocument> {
  if (state.source.dirty && state.mode !== "yaml") {
    throw new Error("YAML changes must be parsed before submission.");
  }
  if (state.mode !== "yaml") {
    assertResourceDraftIdentity(state.draft, lockedIdentity);
    return {
      draft: state.draft,
      document: {
        format: SourceFormat.JSON,
        content: JSON.stringify(state.draft),
      },
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
    assertResourceDraftIdentity(draft, lockedIdentity);
    if (!resourceDraftCanSubmitDraft(draft)) {
      throw new Error("GoalLoop YAML contains invalid integer fields.");
    }
    dispatch({ type: "source_parsed", draft, version });
    return {
      draft,
      document: { format: SourceFormat.YAML, content: source },
    };
  } catch (error) {
    if (isCurrentYaml(stateRef.current, version, source)) {
      dispatch({
        type: "source_invalid",
        error: safeResourceError(error, "YAML validation failed."),
        version,
      });
    }
    throw error;
  }
}
