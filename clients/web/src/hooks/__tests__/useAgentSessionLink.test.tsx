import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const { fetchSessionByPodKey } = vi.hoisted(() => ({
  fetchSessionByPodKey: vi.fn(),
}));

vi.mock("@/lib/api/sessionImportApi", () => ({
  fetchSessionByPodKey,
}));

import { useAgentSessionLink } from "../useAgentSessionLink";

describe("useAgentSessionLink", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    fetchSessionByPodKey.mockReset();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("uses the platform conversation associated with the Worker", async () => {
    fetchSessionByPodKey.mockResolvedValue({
      id: "conv_platform",
      title: "Worker",
    });

    const { result } = renderHook(() =>
      useAgentSessionLink("worker-pod", true),
    );
    await act(async () => {});

    expect(result.current).toEqual({
      error: null,
      loading: false,
      sessionId: "conv_platform",
    });
  });

  it("retries while the ACP event path is creating the association", async () => {
    fetchSessionByPodKey
      .mockResolvedValueOnce(null)
      .mockResolvedValueOnce({ id: "conv_ready", title: null });

    const { result } = renderHook(() =>
      useAgentSessionLink("worker-pod", true),
    );
    await act(async () => {});

    expect(result.current.loading).toBe(true);
    await act(async () => {
      await vi.advanceTimersByTimeAsync(500);
    });

    expect(fetchSessionByPodKey).toHaveBeenCalledTimes(2);
    expect(result.current.sessionId).toBe("conv_ready");
  });

  it("surfaces lookup failures instead of treating them as no session", async () => {
    fetchSessionByPodKey.mockRejectedValue(new Error("database unavailable"));

    const { result } = renderHook(() =>
      useAgentSessionLink("worker-pod", true),
    );
    await act(async () => {});

    expect(result.current).toEqual({
      error: "database unavailable",
      loading: false,
      sessionId: null,
    });
  });

  it("does not resolve a session while the Worker is not connectable", async () => {
    const { result } = renderHook(() =>
      useAgentSessionLink("worker-pod", false),
    );
    await act(async () => {});

    expect(fetchSessionByPodKey).not.toHaveBeenCalled();
    expect(result.current).toEqual({
      error: null,
      loading: false,
      sessionId: null,
    });
  });
});
