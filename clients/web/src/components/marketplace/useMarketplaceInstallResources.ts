"use client";

import { useEffect, useState } from "react";

import { discoverFirstOrgSlug } from "@/lib/light-auth";
import { updateLightSessionOrgSlug } from "@/lib/light-session";
import {
  listMarketplaceModelResources,
  type MarketplaceModelResource,
} from "@/lib/marketplace-model-resources";
import {
  listMarketplaceToolModelResources,
  type MarketplaceToolModelGroup,
} from "@/lib/marketplace-tool-model-resources";

interface InstallResourceState {
  requestKey: string;
  orgSlug: string | null;
  models: MarketplaceModelResource[];
  tools: MarketplaceToolModelGroup[];
  error: boolean;
}

export function useMarketplaceInstallResources(
  orgSlug: string | null,
  agentSlug: string,
) {
  const [reloadKey, setReloadKey] = useState(0);
  const requestKey = `${orgSlug ?? ""}\u0000${agentSlug}\u0000${reloadKey}`;
  const [state, setState] = useState<InstallResourceState>({
    requestKey: "",
    orgSlug,
    models: [],
    tools: [],
    error: false,
  });
  const [modelSelection, setModelSelection] = useState({
    requestKey: "",
    id: "",
  });
  const [toolSelection, setToolSelection] = useState<{
    requestKey: string;
    ids: Record<string, string>;
  }>({ requestKey: "", ids: {} });

  useEffect(() => {
    let cancelled = false;
    void (async () => {
      let resolvedOrgSlug = orgSlug;
      try {
        if (!resolvedOrgSlug) {
          const discovery = await discoverFirstOrgSlug();
          if (cancelled) return;
          if (discovery.status === "unavailable") {
            throw new Error("organization discovery unavailable");
          }
          resolvedOrgSlug =
            discovery.status === "found" ? discovery.slug : null;
        }
        if (!resolvedOrgSlug) {
          if (!cancelled) {
            setState({
              requestKey,
              orgSlug: null,
              models: [],
              tools: [],
              error: false,
            });
          }
          return;
        }
        if (cancelled) return;
        updateLightSessionOrgSlug(resolvedOrgSlug);
        const [models, tools] = await Promise.all([
          listMarketplaceModelResources(resolvedOrgSlug, agentSlug),
          listMarketplaceToolModelResources(resolvedOrgSlug, agentSlug),
        ]);
        if (!cancelled) {
          setState({
            requestKey,
            orgSlug: resolvedOrgSlug,
            models,
            tools,
            error: false,
          });
        }
      } catch {
        if (!cancelled) {
          setState({
            requestKey,
            orgSlug: resolvedOrgSlug,
            models: [],
            tools: [],
            error: true,
          });
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [agentSlug, orgSlug, requestKey]);

  const current = state.requestKey === requestKey;
  const modelID =
    modelSelection.requestKey === requestKey ? modelSelection.id : "";
  const toolIDs =
    toolSelection.requestKey === requestKey ? toolSelection.ids : {};
  const tools = current ? state.tools : [];
  const selectionComplete =
    Boolean(modelID) && tools.every((group) => Boolean(toolIDs[group.role]));
  return {
    orgSlug: current ? state.orgSlug : orgSlug,
    models: current ? state.models : [],
    tools,
    loading: !current,
    error: current && state.error,
    modelID,
    setModelID: (id: string) => setModelSelection({ requestKey, id }),
    toolIDs,
    setToolID: (role: string, id: string) =>
      setToolSelection((selection) => ({
        requestKey,
        ids:
          selection.requestKey === requestKey
            ? { ...selection.ids, [role]: id }
            : { [role]: id },
      })),
    selectionComplete,
    missingCompatibleResource:
      current &&
      (state.models.length === 0 ||
        tools.some((group) => group.resources.length === 0)),
    reload: () => setReloadKey((value) => value + 1),
  };
}
