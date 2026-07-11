import { describe, it, expect, beforeEach, vi } from "vitest";
import { act } from "@testing-library/react";
import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import {
  CreatePodRequestSchema,
  CreatePodResponseSchema,
  ListPodsResponseSchema,
  PodSchema,
} from "@proto/pod/v1/pod_pb";
import { usePodStore } from "../pod";
import { getPodService } from "@/lib/wasm-core";
import { useAuthStore } from "@/stores/auth";
import {
  mockPod,
  mockPod2,
  resetPodStore,
  seedPods,
  readPods,
  readCurrentPod,
  lastReplaceCachedPods,
  lastInsertCreatedPod,
  podStateMock,
} from "./pod-test-utils";

interface MockService {
  list_pods_connect: ReturnType<typeof vi.fn>;
  get_pod_connect: ReturnType<typeof vi.fn>;
  create_pod_connect: ReturnType<typeof vi.fn>;
}

function svc(): MockService {
  return getPodService() as unknown as MockService;
}

function mockListPodsConnect(pods: unknown[], total = pods.length) {
  const items = pods.map((p) =>
    create(PodSchema, {
      id: BigInt((p as { id: number }).id),
      podKey: (p as { pod_key: string }).pod_key,
      status: (p as { status: string }).status,
      agentStatus: (p as { agent_status?: string }).agent_status ?? "",
      createdAt: (p as { created_at?: string }).created_at ?? "",
    }),
  );
  const resp = create(ListPodsResponseSchema, {
    items,
    total: BigInt(total),
    limit: 0,
    offset: 0,
  });
  vi.mocked(svc().list_pods_connect).mockResolvedValue(toBinary(ListPodsResponseSchema, resp));
}

function mockGetPodConnect(pod: unknown) {
  const protoPod = create(PodSchema, {
    id: BigInt((pod as { id: number }).id),
    podKey: (pod as { pod_key: string }).pod_key,
    status: (pod as { status: string }).status,
    agentStatus: (pod as { agent_status?: string }).agent_status ?? "",
    createdAt: (pod as { created_at?: string }).created_at ?? "",
  });
  vi.mocked(svc().get_pod_connect).mockResolvedValue(toBinary(PodSchema, protoPod));
}

function mockWakePodConnect(pod: unknown) {
  const protoPod = create(PodSchema, {
    id: BigInt((pod as { id: number }).id),
    podKey: (pod as { pod_key: string }).pod_key,
    status: (pod as { status: string }).status,
    agentStatus: (pod as { agent_status?: string }).agent_status ?? "",
    createdAt: (pod as { created_at?: string }).created_at ?? "",
  });
  const response = create(CreatePodResponseSchema, { pod: protoPod });
  vi.mocked(svc().create_pod_connect).mockResolvedValue(
    toBinary(CreatePodResponseSchema, response),
  );
}

describe("Pod Store — basic reads", () => {
  beforeEach(resetPodStore);

  describe("initial state", () => {
    it("should have default values", () => {
      const state = usePodStore.getState();
      expect(readPods()).toEqual([]);
      expect(readCurrentPod()).toBeNull();
      expect(state.loading).toBe(false);
      expect(state.error).toBeNull();
    });
  });

  describe("fetchPods", () => {
    it("should fetch pods successfully", async () => {
      mockListPodsConnect([mockPod, mockPod2], 2);

      await act(async () => { await usePodStore.getState().fetchPods(); });

      // Assert against the bridge call — proto-bytes mutators are opaque
      // no-ops in the mock, so we decode the request body the store sent.
      const sent = lastReplaceCachedPods();
      expect(sent).toHaveLength(2);
      expect(sent[0].pod_key).toBe("pod-abc-123");
      expect(usePodStore.getState().loading).toBe(false);
      expect(usePodStore.getState().error).toBeNull();
    });

    it("should call Connect with status filter", async () => {
      mockListPodsConnect([]);

      await act(async () => {
        await usePodStore.getState().fetchPods({ status: "running" });
      });

      expect(svc().list_pods_connect).toHaveBeenCalled();
    });

    it("should handle empty response", async () => {
      mockListPodsConnect([]);
      await act(async () => { await usePodStore.getState().fetchPods(); });
      expect(lastReplaceCachedPods()).toEqual([]);
    });

    it("should handle fetch error", async () => {
      vi.mocked(svc().list_pods_connect).mockRejectedValue(new Error("Network error"));

      await act(async () => { await usePodStore.getState().fetchPods(); });

      const state = usePodStore.getState();
      expect(state.error).toBe("Network error");
      expect(state.loading).toBe(false);
    });

    it("should use default error message when no message provided", async () => {
      vi.mocked(svc().list_pods_connect).mockRejectedValue({});

      await act(async () => { await usePodStore.getState().fetchPods(); });
      expect(usePodStore.getState().error).toBe("Failed to fetch pods");
    });
  });

  describe("fetchPod", () => {
    it("should fetch single pod successfully", async () => {
      mockGetPodConnect(mockPod);

      await act(async () => { await usePodStore.getState().fetchPod("pod-abc-123"); });

      // fetchPod routes to `insert_created_pod` — current pod is unchanged.
      expect(readCurrentPod()).toBeNull();
      expect(usePodStore.getState().loading).toBe(false);
    });

    it("should call insert_created_pod with the fetched pod payload", async () => {
      mockGetPodConnect(mockPod);
      expect(podStateMock().insert_created_pod).not.toHaveBeenCalled();

      await act(async () => { await usePodStore.getState().fetchPod("pod-abc-123"); });

      expect(podStateMock().insert_created_pod).toHaveBeenCalledTimes(1);
      const sent = lastInsertCreatedPod();
      expect(sent?.pod_key).toBe(mockPod.pod_key);
    });

    it("should re-route through insert_created_pod when pod already exists", async () => {
      // Production behaviour: insert_created_pod is upsert-by-pod_key on
      // the wasm side; the store doesn't branch on cache contents. Assert
      // the same bridge call carries the updated payload.
      const updatedPod = { ...mockPod, status: "terminated" as const };
      seedPods(mockPod, mockPod2);
      mockGetPodConnect(updatedPod);

      await act(async () => { await usePodStore.getState().fetchPod("pod-abc-123"); });

      const sent = lastInsertCreatedPod();
      expect(sent?.pod_key).toBe("pod-abc-123");
      expect(sent?.status).toBe("terminated");
    });

    it("should handle fetch error", async () => {
      vi.mocked(svc().get_pod_connect).mockRejectedValue({ message: "Pod not found" });

      await act(async () => {
        await usePodStore.getState().fetchPod("non-existent").catch(() => {});
      });

      const state = usePodStore.getState();
      // fetchPod doesn't write `error` on failure — it just rethrows.
      expect(state.error).toBeNull();
      expect(state.loading).toBe(false);
    });
  });

  describe("wakePod", () => {
    it("creates a resumed pod from the stopped pod and caches the new pod", async () => {
      await useAuthStore.getState().setCurrentOrg({ id: 1, name: "Dev", slug: "dev-org" });
      const resumedPod = {
        ...mockPod,
        id: 3,
        pod_key: "pod-resumed-789",
        status: "initializing" as const,
      };
      mockWakePodConnect(resumedPod);

      let result: typeof resumedPod | undefined;
      await act(async () => {
        result = await usePodStore.getState().wakePod("pod-abc-123");
      });

      const request = fromBinary(
        CreatePodRequestSchema,
        svc().create_pod_connect.mock.calls[0][0],
      );
      expect(request).toMatchObject({
        orgSlug: "dev-org",
        agentSlug: "",
        sourcePodKey: "pod-abc-123",
        resumeAgentSession: true,
      });
      expect(result?.pod_key).toBe("pod-resumed-789");
      expect(lastInsertCreatedPod()?.pod_key).toBe("pod-resumed-789");
    });
  });

  describe("setCurrentPod", () => {
    // Production `setCurrentPod` is a stub that just bumps `_tick`
    // (production has no caller). Tests only assert the bump path.
    it("should bump _tick when called", () => {
      const before = usePodStore.getState()._tick;
      act(() => { usePodStore.getState().setCurrentPod(mockPod); });
      expect(usePodStore.getState()._tick).toBe(before + 1);
    });

    it("should accept null without throwing", () => {
      act(() => { usePodStore.getState().setCurrentPod(null); });
      // No assertion on cache — setCurrentPod is a no-op stub.
    });
  });

  describe("clearError", () => {
    it("should clear error", () => {
      usePodStore.setState({ error: "Some error" });
      act(() => { usePodStore.getState().clearError(); });
      expect(usePodStore.getState().error).toBeNull();
    });
  });
});
