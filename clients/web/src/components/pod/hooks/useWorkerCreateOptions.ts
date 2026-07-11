import { useEffect, useState } from "react";
import { podApi } from "@/lib/api";
import type {
  WorkerCreateOptions,
  WorkerCreateOptionsFilter,
} from "@/lib/api/facade/podConnect";
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
  selection: WorkerCreateOptionsSelection,
): AsyncState<WorkerCreateOptions> {
  const { workerTypeSlug, computeTargetId, deploymentMode } = selection;
  const requestKey = enabled
    ? `${workerTypeSlug}:${computeTargetId}:${deploymentMode}`
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
    void loadOptions(currentSelection)
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
            state: { status: "error", error: errorMessage(error) },
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
    requestKey,
    workerTypeSlug,
  ]);

  if (!enabled) return { status: "idle" };
  return loaded.requestKey === requestKey
    ? loaded.state
    : { status: "loading" };
}

async function loadOptions(
  selection: WorkerCreateOptionsSelection,
): Promise<WorkerCreateOptions> {
  const base = await podApi.listWorkerCreateOptions();
  if (!hasFilter(selection)) return base;
  const filtered = await podApi.listWorkerCreateOptions(optionFilter(selection));
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

function errorMessage(error: unknown): string {
  return error instanceof Error
    ? error.message
    : "Failed to load Worker options";
}
