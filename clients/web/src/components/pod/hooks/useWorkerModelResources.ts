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
  type WorkerModelResourceRequirement,
} from "../CreatePodForm/workerModelResources";

export function requiresModelResource(agentSlug: string | null | undefined): boolean {
  return agentRequiresModelResource(agentSlug ?? null);
}

export function useWorkerModelResources(
  agentSlug: string | null | undefined,
  initialModelResourceId: number | null = null,
  includeToolModels = false,
  requirement?: WorkerModelResourceRequirement,
) {
  const [resources, setResources] = useState<EffectiveResource[]>([]);
  const [toolResources, setToolResources] = useState<EffectiveResource[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loadedAgentSlug, setLoadedAgentSlug] = useState("");
  const [selectedModelResourceId, setSelectedModelResourceId] = useState<number | null>(null);
  const requestAgentSlug = agentSlug ?? "";
  const usesDefinitionRequirement = requirement !== undefined;
  const modelRequired = requirement?.required ?? requiresModelResource(agentSlug);
  const protocolAdapterKey = requirement?.protocolAdapters.join(",") ?? "";

  useEffect(() => {
    setSelectedModelResourceId(initialModelResourceId);
    if (!modelRequired) {
      setResources([]);
      setToolResources([]);
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
            ? listOrganizationEffectiveResources(
              orgSlug,
              includeToolModels ? ["chat", "video"] : ["chat"],
            )
            : listPersonalEffectiveResources(
              includeToolModels ? ["chat", "video"] : ["chat"],
            ),
        ]);
        if (cancelled) return;
        const deduped = dedupeResources(effective);
        const definitionRequirement = usesDefinitionRequirement
          ? {
            required: modelRequired,
            protocolAdapters: protocolAdapterKey
              ? protocolAdapterKey.split(",")
              : [],
          }
          : undefined;
        setResources(
          compatibleModelResources(
            agentSlug ?? null,
            deduped,
            catalog,
            definitionRequirement,
          ),
        );
        setToolResources(includeToolModels ? deduped : []);
      } catch (err) {
        if (cancelled) return;
        setResources([]);
        setToolResources([]);
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
  }, [
    agentSlug,
    includeToolModels,
    initialModelResourceId,
    modelRequired,
    protocolAdapterKey,
    requestAgentSlug,
    usesDefinitionRequirement,
  ]);

  const current = loadedAgentSlug === requestAgentSlug;
  const visibleResources = current ? resources : [];
  const selectedModelResource = visibleResources.find(
    (item) => item.resource?.id === selectedModelResourceId,
  );

  return {
    modelResources: visibleResources,
    toolModelResources: current ? toolResources : [],
    loadingModelResources:
      modelRequired && (!current || loading),
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
