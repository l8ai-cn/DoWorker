import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useMobileAcpRelay } from "./use-mobile-acp-relay";

const relay = vi.hoisted(() => ({
  disconnect: vi.fn().mockResolvedValue(undefined),
  get_status: vi.fn().mockResolvedValue("disconnected"),
  release_control: vi.fn().mockResolvedValue(undefined),
  remove_acp_listener: vi.fn(),
  remove_status_listener: vi.fn(),
  send_acp_command: vi.fn().mockResolvedValue(undefined),
  set_acp_listener: vi.fn(),
  set_status_listener: vi.fn(),
  subscribe: vi.fn().mockResolvedValue(undefined),
}));

const acp = vi.hoisted(() => ({
  add_content_chunk: vi.fn(),
  add_log: vi.fn(),
  add_permission_request: vi.fn(),
  clear_session: vi.fn(),
  get_session_json: vi.fn(() => JSON.stringify({})),
  mark_last_message_complete: vi.fn(),
  update_session_state: vi.fn(),
}));

vi.mock("@/lib/mobile-relay-manager", () => ({
  getMobileRelayManager: vi.fn().mockResolvedValue(relay),
}));
vi.mock("@/lib/mobile-pod-api", () => ({
  getMobilePodConnection: vi.fn().mockResolvedValue({
    podKey: "pod-1",
    relayUrl: "wss://relay.example",
    token: "relay-token",
  }),
}));
vi.mock("@/lib/mobile-relay-reset", () => ({
  resetMobileRelayConnection: vi.fn().mockResolvedValue(undefined),
}));
vi.mock("@/lib/mobile-wasm", () => ({
  getMobileAcpManager: vi.fn().mockResolvedValue(acp),
}));

describe("useMobileAcpRelay", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    relay.set_acp_listener.mockResolvedValue(undefined);
    relay.set_status_listener.mockResolvedValue(undefined);
    relay.subscribe.mockResolvedValue(undefined);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it("registers one named listener pair and removes both on unmount", async () => {
    const { unmount } = renderHook(() => useMobileAcpRelay("pod-1"));

    await waitFor(() => expect(relay.subscribe).toHaveBeenCalledOnce());
    const statusListenerId = relay.set_status_listener.mock.calls[0][1];
    const acpListenerId = relay.set_acp_listener.mock.calls[0][1];
    expect(statusListenerId).toBe(acpListenerId);

    await act(async () => {
      unmount();
    });

    expect(relay.remove_status_listener).toHaveBeenCalledWith("pod-1", statusListenerId);
    expect(relay.remove_acp_listener).toHaveBeenCalledWith("pod-1", acpListenerId);
  });
});
