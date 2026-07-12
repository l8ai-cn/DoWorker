import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { useMobileControlLease, type RelayLease } from "./use-mobile-control-lease";

function manager() {
  return {
    acquire_control: vi.fn().mockResolvedValue(undefined),
    renew_control: vi.fn().mockResolvedValue(undefined),
    release_control: vi.fn().mockResolvedValue(undefined),
  };
}

describe("useMobileControlLease", () => {
  it("requires an explicit acquire and releases a granted lease on unmount", async () => {
    const relay = manager();
    const { result, rerender, unmount } = renderHook(
      ({ lease }) => useMobileControlLease(relay, "pod-1", lease),
      { initialProps: { lease: { status: "observer" } as RelayLease } },
    );

    await act(async () => {
      await result.current.acquire();
    });
    expect(relay.acquire_control).toHaveBeenCalledWith("pod-1", "mobile");

    rerender({
      lease: { status: "granted" as const, leaseId: "lease-1", expiresAt: Date.now() + 60_000 },
    });
    unmount();

    expect(relay.release_control).toHaveBeenCalledWith("pod-1", "lease-1");
  });
});
