import { renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

const mocks = vi.hoisted(() => ({
  dispatchAcp: vi.fn(),
  dispatchDoAgent: vi.fn(),
  dispatchLoopal: vi.fn(() => false),
  onAcpMessage: vi.fn(),
  subscribe: vi.fn(),
  unsubscribe: vi.fn(),
}));

vi.mock("@/stores/relayConnection", () => ({
  relayPool: {
    onAcpMessage: mocks.onAcpMessage,
    subscribe: mocks.subscribe,
    unsubscribe: mocks.unsubscribe,
  },
}));

vi.mock("@/stores/acpEventDispatcher", () => ({
  dispatchAcpRelayEvent: mocks.dispatchAcp,
}));

vi.mock("@/stores/doagentDispatcher", () => ({
  dispatchDoAgentRelayEvent: mocks.dispatchDoAgent,
}));

vi.mock("@/stores/loopalDispatcher", () => ({
  dispatchLoopalRelayEvent: mocks.dispatchLoopal,
}));

import { useAcpRelay } from "./useAcpRelay";

describe("useAcpRelay", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.onAcpMessage.mockReturnValue(vi.fn());
  });

  it("registers the ACP listener before subscribe can synchronously deliver its snapshot", async () => {
    mocks.subscribe.mockImplementation(async () => {
      const listener = mocks.onAcpMessage.mock.calls[0]?.[1];
      listener(13, { sessionId: "session-1", configuration: {} });
    });

    renderHook(() => useAcpRelay("pod-1", "pane-1", true));

    await waitFor(() => {
      expect(mocks.dispatchAcp).toHaveBeenCalledWith(
        "pod-1",
        13,
        { sessionId: "session-1", configuration: {} },
      );
    });
    expect(mocks.onAcpMessage).toHaveBeenCalledBefore(mocks.subscribe);
  });
});
