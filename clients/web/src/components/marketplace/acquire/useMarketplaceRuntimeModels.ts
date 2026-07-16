"use client";

import { useEffect, useState } from "react";

import {
  listMarketplaceModelResources,
  type MarketplaceModelResource,
} from "@/lib/marketplace-model-resources";
import {
  listMarketplaceToolModelResources,
  type MarketplaceToolModelGroup,
} from "@/lib/marketplace-tool-model-resources";

interface ModelLoadResult {
  requestKey: string;
  resources: MarketplaceModelResource[];
  toolGroups: MarketplaceToolModelGroup[];
  error: boolean;
}

export function useMarketplaceRuntimeModels(
  organizationSlug: string | undefined,
  agentSlug: string | undefined,
) {
  const contextKey = `${organizationSlug ?? ""}\u0000${agentSlug ?? ""}`;
  const [selection, setSelection] = useState({ contextKey: "", id: "" });
  const [toolSelection, setToolSelection] = useState<{
    contextKey: string;
    ids: Record<string, string>;
  }>({ contextKey: "", ids: {} });
  const [loadResult, setLoadResult] = useState<ModelLoadResult>({
    requestKey: "",
    resources: [],
    toolGroups: [],
    error: false,
  });
  const [reloadKey, setReloadKey] = useState(0);
  const requestKey = `${contextKey}\u0000${reloadKey}`;
  const canLoad = Boolean(organizationSlug && agentSlug);
  const currentResult = loadResult.requestKey === requestKey;
  const incompatibleListing = Boolean(organizationSlug && agentSlug === "");

  useEffect(() => {
    let cancelled = false;
    if (!organizationSlug || !agentSlug) return;
    Promise.all([
      listMarketplaceModelResources(organizationSlug, agentSlug),
      listMarketplaceToolModelResources(organizationSlug, agentSlug),
    ])
      .then(([resources, toolGroups]) => {
        if (!cancelled) {
          setLoadResult({ requestKey, resources, toolGroups, error: false });
        }
      })
      .catch(() => {
        if (!cancelled) {
          setLoadResult({
            requestKey,
            resources: [],
            toolGroups: [],
            error: true,
          });
        }
      });
    return () => {
      cancelled = true;
    };
  }, [agentSlug, organizationSlug, requestKey]);

  const toolIDs =
    toolSelection.contextKey === contextKey ? toolSelection.ids : {};
  const toolGroups = currentResult ? loadResult.toolGroups : [];
  return {
    modelResourceID: selection.contextKey === contextKey ? selection.id : "",
    setModelResourceID: (id: string) => setSelection({ contextKey, id }),
    modelResources: currentResult ? loadResult.resources : [],
    toolModelGroups: toolGroups,
    toolModelResourceIDs: toolIDs,
    setToolModelResourceID: (role: string, id: string) =>
      setToolSelection((current) => ({
        contextKey,
        ids:
          current.contextKey === contextKey
            ? { ...current.ids, [role]: id }
            : { [role]: id },
      })),
    toolSelectionComplete: toolGroups.every(
      (group) => Boolean(toolIDs[group.role]),
    ),
    missingCompatibleResource:
      currentResult &&
      (loadResult.resources.length === 0 ||
        toolGroups.some((group) => group.resources.length === 0)),
    loadingModels: canLoad && !currentResult,
    modelError: currentResult && loadResult.error,
    incompatibleListing,
    reloadModels: () => setReloadKey((value) => value + 1),
  };
}
