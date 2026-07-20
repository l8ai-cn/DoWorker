import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ListPodsResponseSchema, PodSchema } from "@proto/pod/v1/pod_pb";
import { ReplaceCachedPodsRequestSchema } from "@proto/pod_state/v1/pod_state_pb";

const mocks = vi.hoisted(() => ({
  queryPods: new Map<string, Uint8Array>(),
  listPodsRaw: vi.fn(),
  setState: vi.fn(),
}));

vi.mock("@/lib/wasm-core", () => ({
  getPodState: () => ({
    apply_fetched_pod_query: (key: string, bytes: Uint8Array) => {
      const response = fromBinary(ListPodsResponseSchema, bytes);
      mocks.queryPods.set(
        key,
        toBinary(
          ReplaceCachedPodsRequestSchema,
          create(ReplaceCachedPodsRequestSchema, { pods: response.items }),
        ),
      );
    },
    apply_fetched_pods: () => undefined,
    query_pods_bytes: (key: string) =>
      mocks.queryPods.get(key)
      ?? toBinary(ReplaceCachedPodsRequestSchema, create(ReplaceCachedPodsRequestSchema)),
  }),
  initWasmCore: vi.fn(),
}));

vi.mock("@/lib/api/facade/podConnect", () => ({
  listPodsRaw: mocks.listPodsRaw,
}));

vi.mock("@/stores/pod", () => ({
  usePodStore: Object.assign(
    (selector: (state: { _tick: number }) => unknown) => selector({ _tick: 0 }),
    { setState: mocks.setState },
  ),
}));

import {
  fetchMobileWorkerPods,
  MOBILE_WORKER_STATUSES,
  readMobileWorkerPods,
} from "./mobileWorkerPodQuery";
import { getPodState } from "@/lib/wasm-core";

function response(
  items: Array<{ podKey: string; status: string }>,
  total: number,
  offset: number,
): Uint8Array {
  return toBinary(ListPodsResponseSchema, create(ListPodsResponseSchema, {
    items: items.map((pod) => create(PodSchema, pod)),
    total: BigInt(total),
    limit: 20,
    offset,
  }));
}

describe("mobileWorkerPodQuery", () => {
  beforeEach(() => {
    mocks.queryPods.clear();
    mocks.listPodsRaw.mockReset();
    mocks.setState.mockReset();
  });

  it("keeps a second-page completed Worker after the sidebar replaces its baseline", async () => {
    const firstPage = Array.from({ length: 20 }, (_, index) => ({
      podKey: `active-${index}`,
      status: "running",
    }));
    mocks.listPodsRaw
      .mockResolvedValueOnce(response(firstPage, 21, 0))
      .mockResolvedValueOnce(response([{ podKey: "pattern-r3", status: "completed" }], 21, 20));

    await fetchMobileWorkerPods("dev-org");

    const sidebarResponse = response([{ podKey: "sidebar-active", status: "running" }], 1, 0);
    getPodState().apply_fetched_pods(sidebarResponse);

    expect(mocks.listPodsRaw).toHaveBeenNthCalledWith(1, "dev-org", {
      status: MOBILE_WORKER_STATUSES, limit: 20, offset: 0,
    });
    expect(mocks.listPodsRaw).toHaveBeenNthCalledWith(2, "dev-org", {
      status: MOBILE_WORKER_STATUSES, limit: 20, offset: 20,
    });
    expect(readMobileWorkerPods("dev-org")).toEqual(expect.arrayContaining([
      expect.objectContaining({ pod_key: "pattern-r3", status: "completed" }),
    ]));
  });

  it("fails closed when a later page changes the total", async () => {
    mocks.listPodsRaw
      .mockResolvedValueOnce(response(Array.from({ length: 20 }, (_, index) => ({
        podKey: `active-${index}`, status: "running",
      })), 21, 0))
      .mockResolvedValueOnce(response([{ podKey: "pattern-r3", status: "completed" }], 22, 20));

    await expect(fetchMobileWorkerPods("dev-org-drift")).rejects.toThrow(
      "Worker list changed while loading; retry",
    );
    expect(readMobileWorkerPods("dev-org-drift")).toEqual([]);
  });

  it("fails closed when a later page is empty before completion", async () => {
    mocks.listPodsRaw
      .mockResolvedValueOnce(response(Array.from({ length: 20 }, (_, index) => ({
        podKey: `active-${index}`, status: "running",
      })), 21, 0))
      .mockResolvedValueOnce(response([], 21, 20));

    await expect(fetchMobileWorkerPods("dev-org-empty")).rejects.toThrow(
      "Worker list returned an empty page before completion",
    );
    expect(readMobileWorkerPods("dev-org-empty")).toEqual([]);
  });
});
