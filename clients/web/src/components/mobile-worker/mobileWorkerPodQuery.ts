"use client";

import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import {
  ListPodsResponseSchema,
  type Pod as ProtoPod,
} from "@proto/pod/v1/pod_pb";
import { ReplaceCachedPodsRequestSchema } from "@proto/pod_state/v1/pod_state_pb";
import { listPodsRaw } from "@/lib/api/facade/podConnect";
import { fromProtoPod } from "@/lib/api/podProtoMap";
import { getPodState, initWasmCore } from "@/lib/wasm-core";
import { usePodStore, type Pod } from "@/stores/pod";

export const MOBILE_WORKER_STATUSES =
  "running,initializing,paused,disconnected,orphaned,completed";

const MOBILE_WORKER_PAGE_SIZE = 20;
const requestSequences = new Map<string, number>();

function queryKey(orgSlug: string): string {
  return `mobile-worker-list:${orgSlug}`;
}

function bumpPodSelectors() {
  usePodStore.setState((state) => ({ _tick: state._tick + 1 }));
}

function totalFor(value: bigint): number {
  const total = Number(value);
  if (!Number.isSafeInteger(total) || total < 0) {
    throw new Error("Worker list returned an invalid total");
  }
  return total;
}

async function listAllMobileWorkerPods(
  orgSlug: string,
  sequence: number,
  key: string,
): Promise<Uint8Array | null> {
  const pods: ProtoPod[] = [];
  const podKeys = new Set<string>();
  let expectedTotal: number | undefined;
  let offset = 0;

  for (;;) {
    if (requestSequences.get(key) !== sequence) return null;

    const responseBytes = await listPodsRaw(orgSlug, {
      status: MOBILE_WORKER_STATUSES,
      limit: MOBILE_WORKER_PAGE_SIZE,
      offset,
    });
    const response = fromBinary(ListPodsResponseSchema, responseBytes);
    const total = totalFor(response.total);

    if (expectedTotal === undefined) {
      expectedTotal = total;
    } else if (total !== expectedTotal) {
      throw new Error("Worker list changed while loading; retry");
    }
    if (response.offset !== offset) {
      throw new Error("Worker list returned an unexpected page");
    }
    if (total === 0) {
      if (response.items.length !== 0) {
        throw new Error("Worker list returned items beyond its total");
      }
      return toBinary(ListPodsResponseSchema, create(ListPodsResponseSchema));
    }
    if (response.items.length === 0) {
      throw new Error("Worker list returned an empty page before completion");
    }

    for (const pod of response.items) {
      if (!pod.podKey || podKeys.has(pod.podKey)) {
        throw new Error("Worker list returned an invalid page");
      }
      podKeys.add(pod.podKey);
      pods.push(pod);
    }
    offset += response.items.length;
    if (offset > total) {
      throw new Error("Worker list returned items beyond its total");
    }
    if (offset === total) {
      return toBinary(ListPodsResponseSchema, create(ListPodsResponseSchema, {
        items: pods,
        total: BigInt(total),
        limit: MOBILE_WORKER_PAGE_SIZE,
        offset: 0,
      }));
    }
  }
}

export function readMobileWorkerPods(orgSlug: string): Pod[] {
  const bytes = getPodState().query_pods_bytes(queryKey(orgSlug));
  return fromBinary(ReplaceCachedPodsRequestSchema, bytes).pods.map(fromProtoPod) as Pod[];
}

export function useMobileWorkerPods(orgSlug: string): Pod[] {
  usePodStore((state) => state._tick);
  return readMobileWorkerPods(orgSlug);
}

export async function fetchMobileWorkerPods(orgSlug: string): Promise<void> {
  const key = queryKey(orgSlug);
  const sequence = (requestSequences.get(key) ?? 0) + 1;
  requestSequences.set(key, sequence);

  await initWasmCore();
  const response = await listAllMobileWorkerPods(orgSlug, sequence, key);
  if (response === null || requestSequences.get(key) !== sequence) return;

  getPodState().apply_fetched_pod_query(key, response);
  bumpPodSelectors();
}
