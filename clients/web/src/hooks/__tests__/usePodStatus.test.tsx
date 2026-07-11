import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const { fetchPod } = vi.hoisted(() => ({ fetchPod: vi.fn() }));
let storedPod: { status: string; error_message?: string } | undefined;

vi.mock("@/stores/pod", () => ({
  usePod: vi.fn(() => storedPod),
  usePodStore: (selector: (state: { fetchPod: typeof fetchPod }) => unknown) =>
    selector({ fetchPod }),
}));

import { usePodStatus } from "../usePodStatus";
import { ApiError } from "@/lib/api/api-types";

describe("usePodStatus", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    fetchPod.mockReset();
    storedPod = undefined;
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("retries a transient pod lookup so a worker can become connectable after auth is ready", async () => {
    fetchPod
      .mockRejectedValueOnce(new Error("session is not ready"))
      .mockImplementationOnce(async () => {
        storedPod = { status: "running" };
      });

    const { result, rerender } = renderHook(() => usePodStatus("pod-123"));

    await act(async () => {});
    expect(fetchPod).toHaveBeenCalledTimes(1);
    expect(result.current.podError).toBeNull();

    await act(async () => {
      await vi.advanceTimersByTimeAsync(1000);
    });
    rerender();

    expect(fetchPod).toHaveBeenCalledTimes(2);
    expect(result.current.isPodReady).toBe(true);
  });

  it("restarts the lookup budget when the pane switches to another worker", async () => {
    fetchPod.mockRejectedValue(new Error("network unavailable"));
    const { rerender } = renderHook(
      ({ podKey }) => usePodStatus(podKey),
      { initialProps: { podKey: "pod-one" } },
    );

    await act(async () => {});
    await act(async () => {
      await vi.advanceTimersByTimeAsync(1000);
    });
    await act(async () => {
      await vi.advanceTimersByTimeAsync(1000);
    });
    expect(fetchPod).toHaveBeenCalledTimes(3);

    rerender({ podKey: "pod-two" });
    await act(async () => {});

    expect(fetchPod).toHaveBeenLastCalledWith("pod-two");
  });

  it("does not retry a worker that the server says does not exist", async () => {
    fetchPod.mockRejectedValueOnce(new ApiError(404, "Not Found"));
    const { result } = renderHook(() => usePodStatus("missing-pod"));

    await act(async () => {});
    expect(result.current.podError).toBe("Pod not found");

    await act(async () => {
      await vi.advanceTimersByTimeAsync(1000);
    });
    expect(fetchPod).toHaveBeenCalledTimes(1);
  });

  it("does not schedule a retry after the pane unmounts during a pending lookup", async () => {
    let rejectLookup: (reason?: unknown) => void = () => {};
    fetchPod.mockImplementationOnce(
      () =>
        new Promise<void>((_, reject) => {
          rejectLookup = reject;
        }),
    );
    const { unmount } = renderHook(() => usePodStatus("pod-123"));

    await act(async () => {});
    unmount();
    await act(async () => {
      rejectLookup(new Error("network unavailable"));
    });
    await act(async () => {
      await vi.advanceTimersByTimeAsync(1000);
    });

    expect(fetchPod).toHaveBeenCalledTimes(1);
  });
});
