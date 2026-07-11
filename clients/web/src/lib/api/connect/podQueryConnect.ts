import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import {
  GetPodRequestSchema,
  ListPodsRequestSchema,
  ListPodsResponseSchema,
  PodSchema,
} from "@proto/pod/v1/pod_pb";

import type { PodData } from "@/lib/api/facade/pod";
import { podToCache } from "@/lib/api/projections";
import { getPodService } from "@/lib/wasm-core";

export interface ListPodsOptions {
  status?: string;
  created_by_id?: number;
  runner_id?: number;
  limit?: number;
  offset?: number;
}

export async function listPods(
  orgSlug: string,
  options: ListPodsOptions = {},
): Promise<{ items: PodData[]; total: number; limit: number; offset: number }> {
  const request = create(ListPodsRequestSchema, {
    orgSlug,
    status: options.status,
    createdById:
      options.created_by_id === undefined ? undefined : BigInt(options.created_by_id),
    runnerId: options.runner_id === undefined ? undefined : BigInt(options.runner_id),
    limit: options.limit,
    offset: options.offset,
  });
  const responseBytes = await getPodService().list_pods_connect(
    toBinary(ListPodsRequestSchema, request),
  );
  const response = fromBinary(ListPodsResponseSchema, new Uint8Array(responseBytes));
  return {
    items: response.items.map(podToCache),
    total: Number(response.total),
    limit: response.limit,
    offset: response.offset,
  };
}

export async function listPodsRaw(
  orgSlug: string,
  options: ListPodsOptions = {},
): Promise<Uint8Array> {
  const request = create(ListPodsRequestSchema, {
    orgSlug,
    status: options.status,
    createdById:
      options.created_by_id === undefined ? undefined : BigInt(options.created_by_id),
    runnerId: options.runner_id === undefined ? undefined : BigInt(options.runner_id),
    limit: options.limit,
    offset: options.offset,
  });
  return new Uint8Array(
    await getPodService().list_pods_connect(toBinary(ListPodsRequestSchema, request)),
  );
}

export async function getPod(orgSlug: string, podKey: string): Promise<PodData> {
  const request = create(GetPodRequestSchema, { orgSlug, podKey });
  const responseBytes = await getPodService().get_pod_connect(
    toBinary(GetPodRequestSchema, request),
  );
  return podToCache(fromBinary(PodSchema, new Uint8Array(responseBytes)));
}
