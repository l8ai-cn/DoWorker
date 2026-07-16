import { act, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { RelayTerminalPanel } from "./relay-terminal-panel";
import { getMobilePodConnection } from "@/lib/mobile-pod-api";

const relayState = vi.hoisted(() => ({
  statusListener: undefined as ((value: unknown) => void) | undefined,
  relay: {
    acquire_control: vi.fn().mockResolvedValue(undefined),
    disconnect: vi.fn().mockResolvedValue(undefined),
    get_status: vi.fn().mockResolvedValue("disconnected"),
    remove_status_listener: vi.fn(),
    release_control: vi.fn().mockResolvedValue(undefined),
    renew_control: vi.fn().mockResolvedValue(undefined),
    send: vi.fn().mockResolvedValue(undefined),
    send_resize: vi.fn().mockResolvedValue(undefined),
    set_status_listener: vi.fn(),
    subscribe: vi.fn().mockResolvedValue(undefined),
  },
}));

vi.mock("@/lib/mobile-relay-manager", () => ({
  getMobileRelayManager: vi.fn().mockResolvedValue(relayState.relay),
}));

vi.mock("@/lib/session-relay-api", () => ({
  getSessionRelayConnection: vi.fn().mockResolvedValue({
    podKey: "pod-1",
    relayUrl: "wss://relay.example",
    token: "relay-token",
  }),
}));

vi.mock("@/lib/mobile-pod-api", () => ({
  getMobilePodConnection: vi.fn(),
}));

vi.mock("@xterm/addon-fit", () => ({
  FitAddon: class {
    fit() {}
  },
}));

vi.mock("@xterm/xterm", () => ({
  Terminal: class {
    cols = 80;
    rows = 24;
    options = { disableStdin: true };

    dispose() {}
    focus() {}
    loadAddon() {}
    onData() {
      return { dispose() {} };
    }
    open() {}
    write() {}
  },
}));

class ResizeObserverMock {
  disconnect() {}
  observe() {}
}

describe("RelayTerminalPanel", () => {
  beforeEach(() => {
    relayState.statusListener = undefined;
    vi.clearAllMocks();
    relayState.relay.set_status_listener.mockImplementation(
      async (_podKey: string, _listenerId: string, listener: (value: unknown) => void) => {
        relayState.statusListener = listener;
      },
    );
    vi.stubGlobal("ResizeObserver", ResizeObserverMock);
    vi.stubGlobal("requestAnimationFrame", (callback: FrameRequestCallback) => {
      callback(0);
      return 1;
    });
    vi.stubGlobal("cancelAnimationFrame", vi.fn());
    vi.mocked(getMobilePodConnection).mockResolvedValue({
      podKey: "pod-1",
      relayUrl: "wss://relay.example",
      token: "pod-relay-token",
    });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("keeps terminal controls locked until Relay reports data-ready", async () => {
    render(<RelayTerminalPanel sessionId="session-1" />);

    await waitFor(() => expect(relayState.relay.subscribe).toHaveBeenCalledOnce());
    expect(relayState.relay.set_status_listener).toHaveBeenCalledWith(
      "pod-1",
      "mobile-terminal-session-session-1",
      expect.any(Function),
    );
    expect(screen.getByText("正在连接 Worker…")).not.toBeNull();
    expect((screen.getByRole("button", { name: "接管输入" }) as HTMLButtonElement).disabled).toBe(
      true,
    );

    await act(async () => {
      relayState.statusListener?.({ status: "connected", controlLeaseStatus: "observer" });
    });

    await waitFor(() => {
      expect(screen.queryByText("正在连接 Worker…")).toBeNull();
      expect((screen.getByRole("button", { name: "接管输入" }) as HTMLButtonElement).disabled).toBe(
        false,
      );
    });
  });

  it("uses the Pod connection endpoint for a Worker deep link", async () => {
    render(<RelayTerminalPanel podKey="pod-1" />);

    await waitFor(() => expect(getMobilePodConnection).toHaveBeenCalledWith("pod-1"));
    expect(relayState.relay.subscribe).toHaveBeenCalledWith(
      "pod-1",
      expect.any(String),
      "wss://relay.example",
      "pod-relay-token",
      expect.any(Function),
    );
    expect(relayState.relay.set_status_listener).toHaveBeenCalledWith(
      "pod-1",
      "mobile-terminal-pod-pod-1",
      expect.any(Function),
    );
  });
});
