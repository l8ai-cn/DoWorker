import { describe, expect, it, vi, beforeEach } from "vitest";
import { registerServiceProvider, markServiceReady } from "@agentsmesh/service-runtime";

// The adapter delegates all connection management to getRelayManager(). These
// tests pin the adapter's remaining responsibilities: endpoint delegation, the
// per-pod listener fan-out, the legacy "none" status baseline, and the sync
// status cache that isConnected()/getStatus() read. We drive getRelayManager()
// through the real service registry rather than vi.mock.
const mgr = {
  subscribe: vi.fn().mockResolvedValue(undefined),
  unsubscribe: vi.fn().mockResolvedValue(undefined),
  send: vi.fn().mockResolvedValue(undefined),
  send_resize: vi.fn().mockResolvedValue(undefined),
  force_resize: vi.fn().mockResolvedValue(undefined),
  send_acp_command: vi.fn().mockResolvedValue(undefined),
  disconnect: vi.fn().mockResolvedValue(undefined),
  disconnect_all: vi.fn().mockResolvedValue(undefined),
  get_status: vi.fn().mockResolvedValue("disconnected"),
  is_runner_disconnected: vi.fn().mockResolvedValue(false),
  get_pod_size: vi.fn().mockResolvedValue(null),
  on_status_change: vi.fn().mockResolvedValue(undefined),
  on_acp_message: vi.fn().mockResolvedValue(undefined),
  on_pod_disconnected: vi.fn(),
};

vi.mock("@/lib/api/facade/podConnect", () => ({
  getPodConnection: vi.fn().mockResolvedValue({
    relay_url: "wss://relay.example.com",
    token: "test-token",
    pod_key: "pod-1",
  }),
}));

type StatusRaw = { status: string; runnerDisconnected: boolean };

async function freshPool() {
  delete (globalThis as Record<string, unknown>).__relayPool;
  vi.resetModules();
  return (await import("@/stores/relayConnection")).relayPool;
}

function lastStatusCb(): (raw: StatusRaw) => void {
  return mgr.on_status_change.mock.calls.at(-1)![1] as (raw: StatusRaw) => void;
}
function lastAcpCb(): (mt: number, pl: unknown) => void {
  return mgr.on_acp_message.mock.calls.at(-1)![1] as (mt: number, pl: unknown) => void;
}

describe("relayConnection adapter", () => {
  let pool: Awaited<ReturnType<typeof freshPool>>;

  beforeEach(async () => {
    vi.clearAllMocks();
    registerServiceProvider({ relayManager: mgr as never });
    markServiceReady();
    pool = await freshPool();
  });

  describe("subscribe", () => {
    it("selects the endpoint then delegates to the manager and returns a handle", async () => {
      const onMessage = vi.fn();
      const handle = await pool.subscribe("pod-1", "sub-1", onMessage);

      expect(mgr.subscribe).toHaveBeenCalledWith(
        "pod-1", "sub-1", "wss://relay.example.com", "test-token", onMessage,
      );
      expect(handle).toHaveProperty("send");
      expect(handle).toHaveProperty("unsubscribe");
    });

    it("registers exactly one upstream status listener per pod", async () => {
      await pool.subscribe("pod-1", "sub-1", vi.fn());
      pool.onStatusChange("pod-1", vi.fn());
      await pool.subscribe("pod-1", "sub-2", vi.fn());

      expect(mgr.on_status_change).toHaveBeenCalledTimes(1);
    });

    it("handle.send / handle.unsubscribe delegate to the manager", async () => {
      const handle = await pool.subscribe("pod-1", "sub-1", vi.fn());
      handle.send("x");
      handle.unsubscribe();
      expect(mgr.send).toHaveBeenCalledWith("pod-1", "x");
      expect(mgr.unsubscribe).toHaveBeenCalledWith("pod-1", "sub-1");
    });
  });

  describe("input / resize delegation (debounce + dedup live in the pool now)", () => {
    it("send / sendResize / forceResize delegate; non-positive sizes are dropped", () => {
      pool.send("pod-1", "data");
      pool.sendResize("pod-1", 80, 24);
      pool.forceResize("pod-1", 100, 40);
      pool.sendResize("pod-1", 0, 24);
      pool.forceResize("pod-1", 80, 0);

      expect(mgr.send).toHaveBeenCalledWith("pod-1", "data");
      expect(mgr.send_resize).toHaveBeenCalledExactlyOnceWith("pod-1", 80, 24);
      expect(mgr.force_resize).toHaveBeenCalledExactlyOnceWith("pod-1", 100, 40);
    });

    it("sendAcpCommand JSON-encodes the command for the string-typed manager", () => {
      pool.sendAcpCommand("pod-1", { type: "prompt", prompt: "hi" });
      expect(mgr.send_acp_command).toHaveBeenCalledWith(
        "pod-1", JSON.stringify({ type: "prompt", prompt: "hi" }),
      );
    });
  });

  describe("status fan-out + 'none' baseline", () => {
    it("emits 'none' immediately for an unknown pod and maps pre-connect 'disconnected' to 'none'", () => {
      const listener = vi.fn();
      pool.onStatusChange("pod-1", listener);
      expect(listener).toHaveBeenCalledWith({ status: "none", runnerDisconnected: false });

      lastStatusCb()({ status: "disconnected", runnerDisconnected: false });
      expect(listener).toHaveBeenLastCalledWith({ status: "none", runnerDisconnected: false });
    });

    it("passes through real statuses once subscribed and updates isConnected/getStatus", async () => {
      const listener = vi.fn();
      pool.onStatusChange("pod-1", listener);
      await pool.subscribe("pod-1", "sub-1", vi.fn());

      lastStatusCb()({ status: "connected", runnerDisconnected: false });
      expect(listener).toHaveBeenLastCalledWith({ status: "connected", runnerDisconnected: false });
      expect(pool.isConnected("pod-1")).toBe(true);
      expect(pool.getStatus("pod-1")).toBe("connected");

      lastStatusCb()({ status: "disconnected", runnerDisconnected: true });
      expect(pool.isConnected("pod-1")).toBe(false);
      expect(pool.isRunnerDisconnected("pod-1")).toBe(true);
    });

    it("stops notifying a removed listener", () => {
      const listener = vi.fn();
      const off = pool.onStatusChange("pod-1", listener);
      listener.mockClear();
      off();
      lastStatusCb()({ status: "connected", runnerDisconnected: false });
      expect(listener).not.toHaveBeenCalled();
    });
  });

  describe("acp fan-out", () => {
    it("routes manager ACP messages to registered listeners until removed", () => {
      const listener = vi.fn();
      const off = pool.onAcpMessage("pod-1", listener);
      lastAcpCb()(0x0b, { type: "contentChunk" });
      expect(listener).toHaveBeenCalledWith(0x0b, { type: "contentChunk" });

      off();
      listener.mockClear();
      lastAcpCb()(0x0b, { type: "more" });
      expect(listener).not.toHaveBeenCalled();
    });
  });

  describe("defaults for unknown pods", () => {
    it("getStatus/isConnected/isRunnerDisconnected return safe defaults", () => {
      expect(pool.getStatus("unknown")).toBe("none");
      expect(pool.isConnected("unknown")).toBe(false);
      expect(pool.isRunnerDisconnected("unknown")).toBe(false);
    });

    it("disconnect / disconnectAll delegate to the manager", async () => {
      await pool.subscribe("pod-1", "sub-1", vi.fn());
      pool.disconnect("pod-1");
      pool.disconnectAll();
      expect(mgr.disconnect).toHaveBeenCalledWith("pod-1");
      expect(mgr.disconnect_all).toHaveBeenCalled();
    });
  });
});
