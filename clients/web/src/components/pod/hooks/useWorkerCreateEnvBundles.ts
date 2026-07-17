import { useEffect, useState } from "react";
import { listEnvBundles } from "@/lib/api/facade/envBundleConnect";
import type { EnvBundleSummary } from "@/lib/api";
import type { AsyncState } from "./workerCreateDraft";

interface WorkerEnvBundles {
  runtime: AsyncState<EnvBundleSummary[]>;
  credential: AsyncState<EnvBundleSummary[]>;
  config: AsyncState<EnvBundleSummary[]>;
}

interface LoadedWorkerEnvBundles extends WorkerEnvBundles {
  workerTypeSlug: string;
}

export function useWorkerCreateEnvBundles(
  workerTypeSlug: string,
): WorkerEnvBundles {
  const [loaded, setLoaded] = useState<LoadedWorkerEnvBundles>({
    workerTypeSlug: "",
    runtime: { status: "idle" },
    credential: { status: "idle" },
    config: { status: "idle" },
  });

  useEffect(() => {
    if (!workerTypeSlug) return;

    let cancelled = false;
    void Promise.all([
      loadBundles("runtime", workerTypeSlug),
      loadBundles("credential", workerTypeSlug),
      loadBundles("config", workerTypeSlug),
    ]).then(([runtime, credential, config]) => {
      if (!cancelled) setLoaded({ workerTypeSlug, runtime, credential, config });
    });
    return () => {
      cancelled = true;
    };
  }, [workerTypeSlug]);

  if (!workerTypeSlug) {
    return {
      runtime: { status: "ready", data: [] },
      credential: { status: "ready", data: [] },
      config: { status: "ready", data: [] },
    };
  }
  if (loaded.workerTypeSlug !== workerTypeSlug) {
    return {
      runtime: { status: "loading" },
      credential: { status: "loading" },
      config: { status: "loading" },
    };
  }
  return loaded;
}

async function loadBundles(
  kind: string,
  workerTypeSlug: string,
): Promise<AsyncState<EnvBundleSummary[]>> {
  try {
    const response = await listEnvBundles({
      kind,
      agentSlug: workerTypeSlug,
    });
    return {
      status: "ready",
      data: response.items.map((bundle) => ({
        id: Number(bundle.id),
        name: bundle.name,
        agent_slug: bundle.agentSlug ?? workerTypeSlug,
        kind: bundle.kind,
        kind_primary: bundle.kindPrimary,
        configured_fields:
          bundle.configuredFields.length > 0
            ? bundle.configuredFields
            : undefined,
      })),
    };
  } catch (error) {
    return {
      status: "error",
      error:
        error instanceof Error
          ? error.message
          : `Failed to load ${kind} environment bundles`,
    };
  }
}
