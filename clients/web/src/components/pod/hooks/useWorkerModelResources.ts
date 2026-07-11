import { useEffect, useState } from "react";
import {
  getCatalog,
  listOrganizationEffectiveResources,
  listPersonalEffectiveResources,
} from "@/lib/api/facade/aiResourceConnect";
import type { EffectiveResource } from "@/lib/api/facade/aiResource";
import { readCurrentOrg } from "@/stores/auth";
import {
  agentRequiresModelResource,
  compatibleModelResources,
} from "../CreatePodForm/workerModelResources";

export function requiresModelResource(agentSlug: string | null | undefined): boolean {
  return agentRequiresModelResource(agentSlug ?? null);
}

export function useWorkerModelResources(
  agentSlug: string | null | undefined,
  initialModelResourceId: number | null = null,
) {
  const [resources, setResources] = useState<EffectiveResource[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loadedAgentSlug, setLoadedAgentSlug] = useState("");
  const [selectedModelResourceId, setSelectedModelResourceId] = useState<number | null>(null);
  const requestAgentSlug = agentSlug ?? "";

  useEffect(() => {
    setSelectedModelResourceId(initialModelResourceId);
    if (!requiresModelResource(agentSlug)) {
      setResources([]);
      setError(null);
      setLoading(false);
      setLoadedAgentSlug(requestAgentSlug);
      return;
    }

    let cancelled = false;
    const load = async () => {
      setLoading(true);
      setError(null);
      try {
        const orgSlug = readCurrentOrg()?.slug ?? "";
        const [catalog, effective] = await Promise.all([
          getCatalog(),
          orgSlug
            ? listOrganizationEffectiveResources(orgSlug, ["chat"])
            : listPersonalEffectiveResources(["chat"]),
        ]);
        if (cancelled) return;
        setResources(compatibleModelResources(agentSlug ?? null, dedupeResources(effective), catalog));
      } catch (err) {
        if (cancelled) return;
        setResources([]);
        setError(err instanceof Error ? err.message : "Failed to load model resources");
      } finally {
        if (!cancelled) {
          setLoading(false);
          setLoadedAgentSlug(requestAgentSlug);
        }
      }
    };

    void load();
    return () => {
      cancelled = true;
    };
  }, [agentSlug, initialModelResourceId, requestAgentSlug]);

  const current = loadedAgentSlug === requestAgentSlug;
  const visibleResources = current ? resources : [];
  const selectedModelResource = visibleResources.find(
    (item) => item.resource?.id === selectedModelResourceId,
  );

  return {
    modelResources: visibleResources,
    loadingModelResources:
      requiresModelResource(agentSlug) && (!current || loading),
    modelResourceError: current ? error : null,
    selectedModelResource,
    selectedModelResourceId,
    setSelectedModelResourceId,
  };
}

function dedupeResources(items: EffectiveResource[]): EffectiveResource[] {
  const seen = new Set<number>();
  const out: EffectiveResource[] = [];
  for (const item of items) {
    const id = item.resource?.id;
    if (!id || seen.has(id)) continue;
    seen.add(id);
    out.push(item);
  }
  return out;
}
