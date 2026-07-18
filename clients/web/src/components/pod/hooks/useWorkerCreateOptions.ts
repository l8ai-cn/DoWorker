import { useEffect, useState } from "react";
import type {
  WorkerCreateOptions,
  WorkerCreateOptionsFilter,
} from "@/lib/api/facade/podConnect";
import { listWorkerCreateOptions } from "@/lib/api/facade/podConnect";
import { safeServiceErrorMessage } from "@/lib/errors/safeServiceErrorMessage";
import type { AsyncState } from "./workerCreateDraft";

interface WorkerCreateOptionsSelection {
  workerTypeSlug: string;
  computeTargetId: number;
  deploymentMode: string;
}

interface LoadedWorkerCreateOptions {
  requestKey: string;
  state: AsyncState<WorkerCreateOptions>;
}

export function useWorkerCreateOptions(
  enabled: boolean,
  orgSlug: string,
  selection: WorkerCreateOptionsSelection,
): AsyncState<WorkerCreateOptions> {
  const { workerTypeSlug, computeTargetId, deploymentMode } = selection;
  const requestKey = enabled
    ? `${orgSlug}:${workerTypeSlug}:${computeTargetId}:${deploymentMode}`
    : "disabled";
  const [loaded, setLoaded] = useState<LoadedWorkerCreateOptions>({
    requestKey: "",
    state: { status: "idle" },
  });

  useEffect(() => {
    if (!enabled) return;
    let cancelled = false;
    const currentSelection = {
      workerTypeSlug,
      computeTargetId,
      deploymentMode,
    };
    void loadWorkerCreateOptions(orgSlug, currentSelection)
      .then((data) => {
        if (!cancelled) {
          setLoaded({
            requestKey,
            state: { status: "ready", data },
          });
        }
      })
      .catch((error: unknown) => {
        if (!cancelled) {
          setLoaded({
            requestKey,
            state: {
              status: "error",
              error: safeServiceErrorMessage(
                error,
                "Failed to load Worker options",
              ),
            },
          });
        }
      });
    return () => {
      cancelled = true;
    };
  }, [
    computeTargetId,
    deploymentMode,
    enabled,
    orgSlug,
    requestKey,
    workerTypeSlug,
  ]);

  if (!enabled) return { status: "idle" };
  return loaded.requestKey === requestKey
    ? loaded.state
    : { status: "loading" };
}

export async function loadWorkerCreateOptions(
  orgSlug: string,
  selection: WorkerCreateOptionsSelection,
): Promise<WorkerCreateOptions> {
  const base = await listWorkerCreateOptions(orgSlug);
  if (!hasFilter(selection)) return base;
  const filtered = await listWorkerCreateOptions(
    orgSlug,
    optionFilter(selection),
  );
  return mergeOptions(base, filtered);
}

function hasFilter(selection: WorkerCreateOptionsSelection): boolean {
  return Boolean(
    selection.workerTypeSlug ||
      selection.computeTargetId > 0 ||
      selection.deploymentMode,
  );
}

function optionFilter(
  selection: WorkerCreateOptionsSelection,
): WorkerCreateOptionsFilter {
  return {
    worker_type_slug: selection.workerTypeSlug || undefined,
    compute_target_id:
      selection.computeTargetId > 0 ? selection.computeTargetId : undefined,
    deployment_mode: selection.deploymentMode || undefined,
  };
}

function mergeOptions(
  base: WorkerCreateOptions,
  filtered: WorkerCreateOptions,
): WorkerCreateOptions {
  return {
    revision: filtered.revision,
    worker_types: base.worker_types,
    runtime_images: filtered.runtime_images,
    compute_targets: base.compute_targets,
    deployment_modes: filtered.deployment_modes,
    resource_profiles: base.resource_profiles,
  };
}
