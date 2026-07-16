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
const listeners = vi.hoisted(() => ({
  acp: undefined as ((messageType: number, payload: unknown) => void) | undefined,
  status: undefined as ((payload: unknown) => void) | undefined,
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
    listeners.acp = undefined;
    listeners.status = undefined;
    relay.set_acp_listener.mockImplementation(async (_podKey, _listenerId, listener) => {
      listeners.acp = listener;
    });
    relay.set_status_listener.mockImplementation(async (_podKey, _listenerId, listener) => {
      listeners.status = listener;
    });
    relay.subscribe.mockResolvedValue(undefined);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it("uses the Pod as the listener key and removes both listeners on unmount", async () => {
    const { unmount } = renderHook(() => useMobileAcpRelay("pod-1"));

    await waitFor(() => expect(relay.subscribe).toHaveBeenCalledOnce());
    const statusListenerId = relay.set_status_listener.mock.calls[0][1];
    const acpListenerId = relay.set_acp_listener.mock.calls[0][1];
    expect(statusListenerId).toBe("mobile-acp-pod-1");
    expect(statusListenerId).toBe(acpListenerId);

    await act(async () => {
      unmount();
    });

    expect(relay.remove_status_listener).toHaveBeenCalledWith("pod-1", statusListenerId);
    expect(relay.remove_acp_listener).toHaveBeenCalledWith("pod-1", acpListenerId);
  });

  it("waits for Runner prompt acceptance before resolving the send", async () => {
    const { result } = renderHook(() => useMobileAcpRelay("pod-1"));

    await waitFor(() => expect(relay.subscribe).toHaveBeenCalledOnce());
    await act(async () => {
      listeners.status?.({
        status: "connected",
        controlLeaseExpiresAt: Date.now() + 30_000,
        controlLeaseId: "lease-1",
        controlLeaseStatus: "granted",
      });
    });
    const sent = result.current.sendPrompt("hello");
    await waitFor(() => expect(relay.send_acp_command).toHaveBeenCalledOnce());
    const command = JSON.parse(relay.send_acp_command.mock.calls[0][1]);

    await act(async () => {
      listeners.acp?.(0x0b, {
        requestId: command.requestId,
        role: "user",
        text: "hello",
        type: "contentChunk",
      });
    });

    await expect(sent).resolves.toBeUndefined();
  });
});
