import { useEffect, useState } from "react";
import {
  getCatalog,
  listOrganizationEffectiveResources,
  listPersonalEffectiveResources,
} from "@/lib/api/facade/aiResourceConnect";
import type {
  EffectiveResource,
  ProviderDefinition,
} from "@/lib/api/facade/aiResource";
import { readCurrentOrg } from "@/stores/auth";
import type { AsyncState } from "./workerCreateDraft";

export interface WorkerCreateModelResources {
  modelResources: AsyncState<EffectiveResource[]>;
  modelProviders: AsyncState<ProviderDefinition[]>;
}

export function useWorkerCreateModelResources(): WorkerCreateModelResources {
  const [resources, setResources] = useState<AsyncState<EffectiveResource[]>>({
    status: "loading",
  });
  const [providers, setProviders] = useState<AsyncState<ProviderDefinition[]>>({
    status: "loading",
  });

  useEffect(() => {
    let cancelled = false;

    const load = async () => {
      try {
        const orgSlug = readCurrentOrg()?.slug ?? "";
        const [catalog, effective] = await Promise.all([
          getCatalog(),
          orgSlug
            ? listOrganizationEffectiveResources(orgSlug)
            : listPersonalEffectiveResources(),
        ]);
        if (cancelled) return;
        setProviders({ status: "ready", data: catalog });
        setResources({ status: "ready", data: dedupeResources(effective) });
      } catch (error) {
        if (cancelled) return;
        const message = error instanceof Error
          ? error.message
          : "Failed to load model resources";
        setProviders({ status: "error", error: message });
        setResources({ status: "error", error: message });
      }
    };

    void load();
    return () => {
      cancelled = true;
    };
  }, []);

  return { modelResources: resources, modelProviders: providers };
}

function dedupeResources(items: EffectiveResource[]): EffectiveResource[] {
  const seen = new Set<number>();
  return items.filter((item) => {
    const id = item.resource?.id;
    if (!id || seen.has(id)) return false;
    seen.add(id);
    return true;
  });
}
