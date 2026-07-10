"use client";

import { useEffect, useState } from "react";
import { listEnvBundles, type EnvBundle } from "@/lib/api/facade/envBundleConnect";
import type { EnvBundleSummary } from "@/lib/viewModels/envBundleSummary";

export function useLoopEnvBundles(args: {
  open: boolean;
  agentSlug: string | null;
}): {
  envBundles: EnvBundleSummary[];
  loadingBundles: boolean;
} {
  const { open, agentSlug } = args;
  const [envBundles, setEnvBundles] = useState<EnvBundleSummary[]>([]);
  const [loadingBundles, setLoadingBundles] = useState(false);

  useEffect(() => {
    if (!open || !agentSlug) {
      setEnvBundles([]);
      return;
    }
    let cancelled = false;
    const load = async () => {
      setLoadingBundles(true);
      try {
        const runtimeRes = await listEnvBundles({ kind: "runtime", agentSlug }).catch(() => ({ items: [] }));
        if (cancelled) return;
        const mapBundle = (b: EnvBundle): EnvBundleSummary => ({
          id: Number(b.id),
          name: b.name,
          agent_slug: b.agentSlug ?? agentSlug,
          kind: b.kind,
          kind_primary: b.kindPrimary,
          configured_fields:
            b.configuredFields.length > 0 ? b.configuredFields : undefined,
        });
        const runtimeBundles: EnvBundleSummary[] = (runtimeRes.items ?? []).map(mapBundle);
        setEnvBundles(runtimeBundles);
      } catch {
        if (!cancelled) setEnvBundles([]);
      } finally {
        if (!cancelled) setLoadingBundles(false);
      }
    };
    load();
    return () => {
      cancelled = true;
    };
  }, [open, agentSlug]);

  return { envBundles, loadingBundles };
}
