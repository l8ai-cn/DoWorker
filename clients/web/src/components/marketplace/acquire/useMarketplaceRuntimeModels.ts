"use client";

import { useEffect, useState } from "react";

import {
  listMarketplaceModelResources,
  type MarketplaceModelResource,
} from "@/lib/marketplace-model-resources";

interface ModelLoadResult {
  requestKey: string;
  resources: MarketplaceModelResource[];
  error: boolean;
}

export function useMarketplaceRuntimeModels(
  organizationSlug: string | undefined,
  agentSlug: string | undefined,
) {
  const contextKey = `${organizationSlug ?? ""}\u0000${agentSlug ?? ""}`;
  const [selection, setSelection] = useState({ contextKey: "", id: "" });
  const [loadResult, setLoadResult] = useState<ModelLoadResult>({
    requestKey: "",
    resources: [],
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
    listMarketplaceModelResources(organizationSlug, agentSlug)
      .then((items) => {
        if (!cancelled) {
          setLoadResult({ requestKey, resources: items, error: false });
        }
      })
      .catch(() => {
        if (!cancelled) {
          setLoadResult({ requestKey, resources: [], error: true });
        }
      });
    return () => {
      cancelled = true;
    };
  }, [agentSlug, organizationSlug, requestKey]);

  return {
    modelResourceID: selection.contextKey === contextKey ? selection.id : "",
    setModelResourceID: (id: string) => setSelection({ contextKey, id }),
    modelResources: currentResult ? loadResult.resources : [],
    loadingModels: canLoad && !currentResult,
    modelError: currentResult && loadResult.error,
    incompatibleListing,
    reloadModels: () => setReloadKey((value) => value + 1),
  };
}
