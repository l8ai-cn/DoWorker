import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { useWorkerControlLease } from "../useWorkerControlLease";

const mocks = vi.hoisted(() => ({
  acquireControl: vi.fn(),
  renewControl: vi.fn(),
  releaseControl: vi.fn(),
  relay: {
    status: "connected",
    runnerDisconnected: false,
    controlLease: { status: "observer" },
  },
}));

vi.mock("@/stores/relayConnection", () => ({
  relayPool: {
    acquireControl: mocks.acquireControl,
    renewControl: mocks.renewControl,
    releaseControl: mocks.releaseControl,
  },
}));

vi.mock("@/hooks/useTerminalStatus", () => ({
  useTerminalStatus: () => mocks.relay,
}));

describe("useWorkerControlLease", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.useRealTimers();
    mocks.relay = {
      status: "connected",
      runnerDisconnected: false,
      controlLease: { status: "observer" },
    };
    mocks.acquireControl.mockResolvedValue(undefined);
    mocks.renewControl.mockResolvedValue(undefined);
    mocks.releaseControl.mockResolvedValue(undefined);
  });

  it("only acquires control after an explicit user action", async () => {
    const { result } = renderHook(() =>
      useWorkerControlLease("pod-1", "mobile"),
    );

    expect(mocks.acquireControl).not.toHaveBeenCalled();
    await act(() => result.current.acquire());

    expect(mocks.acquireControl).toHaveBeenCalledWith("pod-1", "mobile");
  });

  it("renews a granted lease before it expires", async () => {
    vi.useFakeTimers();
    mocks.relay = {
      status: "connected",
      runnerDisconnected: false,
      controlLease: {
        status: "granted",
        leaseId: "lease-1",
        expiresAt: Date.now() + 30_000,
      },
    };

    renderHook(() => useWorkerControlLease("pod-1", "mobile"));
    await vi.advanceTimersByTimeAsync(20_000);

    expect(mocks.renewControl).toHaveBeenCalledWith("pod-1", "lease-1");
  });

  it("releases control once when the page is hidden and unmounted", () => {
    mocks.relay = {
      status: "connected",
      runnerDisconnected: false,
      controlLease: {
        status: "granted",
        leaseId: "lease-1",
        expiresAt: Date.now() + 30_000,
      },
    };
    const { unmount } = renderHook(() =>
      useWorkerControlLease("pod-1", "mobile"),
    );

    Object.defineProperty(document, "visibilityState", {
      configurable: true,
      value: "hidden",
    });
    document.dispatchEvent(new Event("visibilitychange"));
    unmount();

    expect(mocks.releaseControl).toHaveBeenCalledTimes(1);
    expect(mocks.releaseControl).toHaveBeenCalledWith("pod-1", "lease-1");
  });
});
