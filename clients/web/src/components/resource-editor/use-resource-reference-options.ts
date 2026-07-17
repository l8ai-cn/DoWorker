"use client";

import { useEffect, useState } from "react";
import { listResources } from "@/lib/api/facade/orchestrationResource";
import { getErrorMessage } from "@/lib/utils";
import type {
  ResourceReferenceCatalog,
  ResourceReferenceOption,
} from "./resource-reference-options";

const REFERENCE_KINDS = [
  "WorkerTemplate",
  "Prompt",
  "ModelBinding",
  "ToolBinding",
  "Repository",
  "Skill",
  "KnowledgeBase",
  "EnvironmentBundle",
  "ComputeTarget",
  "ResourceProfile",
] as const;

export function useResourceReferenceOptions(
  orgSlug: string,
): ResourceReferenceCatalog {
  const [state, setState] = useState<ResourceReferenceCatalog & {
    orgSlug: string;
  }>({
    orgSlug,
    loading: true,
    error: null,
    byKind: {},
  });

  useEffect(() => {
    let active = true;
    Promise.all(REFERENCE_KINDS.map(async (kind) => {
      const response = await listResources(orgSlug, {
        kind,
        limit: 100,
        offset: 0,
      });
      return [kind, response.items.map(toOption)] as const;
    })).then((entries) => {
      if (!active) return;
      setState({
        orgSlug,
        loading: false,
        error: null,
        byKind: Object.fromEntries(entries),
      });
    }).catch((error) => {
      if (!active) return;
      setState({
        orgSlug,
        loading: false,
        error: getErrorMessage(error, "Failed to load resource references."),
        byKind: {},
      });
    });
    return () => {
      active = false;
    };
  }, [orgSlug]);

  if (state.orgSlug !== orgSlug) {
    return { loading: true, error: null, byKind: {} };
  }
  return state;
}

function toOption(resource: {
  identity?: { target?: { name?: string } };
  displayName: string;
  revision: bigint;
}): ResourceReferenceOption {
  return {
    name: resource.identity?.target?.name ?? "",
    displayName: resource.displayName,
    revision: Number(resource.revision),
  };
}
