import { getPodConnection } from "@/lib/api/facade/podConnect";
import { readCurrentOrg } from "@/stores/auth";
import { getRelayManager } from "@do-worker/service-runtime";
import type {
  ConnectionStatus, ConnectionHandle, RelayStatusInfo, StatusListener,
} from "./relayConnectionTypes";

export type { ConnectionStatus, ConnectionHandle, RelayStatusInfo } from "./relayConnectionTypes";

type OnMessage = (data: Uint8Array | string) => void;
type AcpListener = (msgType: number, payload: unknown) => void;
type StatusInfoRaw = { status: string; runnerDisconnected: boolean };

const NONE: RelayStatusInfo = { status: "none", runnerDisconnected: false };

// Thin adapter over the Rust relay pool (via getRelayManager(): WasmRelayManager
// on web, WasmRelayManager). All connection
// management — reconnect/backoff, input dedup, resize debounce, snapshot replay,
// codec — lives in the Rust pool now. This layer keeps the legacy "none" status
// baseline + per-pod listener fan-out — the managers expose no off_* method, so
// we register one upstream listener per pod and fan out here, mirroring the
// previous pool's public contract.
class RelayConnectionPool {
  private statusListeners = new Map<string, Set<StatusListener>>();
  private acpListeners = new Map<string, Set<AcpListener>>();
  private statusCache = new Map<string, RelayStatusInfo>();
  private connectedPods = new Set<string>();
  private statusUpstream = new Set<string>();
  private acpUpstream = new Set<string>();
  private disconnectHookWired = false;

  constructor() {
    if (typeof window !== "undefined") {
      window.addEventListener("beforeunload", () => this.disconnectAll());
    }
  }

  private mgr() {
    return getRelayManager();
  }

  async subscribe(podKey: string, subscriptionId: string, onMessage: OnMessage): Promise<ConnectionHandle> {
    this.ensureStatusUpstream(podKey);
    const info = await getPodConnection(readCurrentOrg()?.slug ?? "", podKey);
    this.connectedPods.add(podKey);
    await this.mgr().subscribe(podKey, subscriptionId, info.relay_url, info.token, onMessage);
    return {
      send: (data) => this.send(podKey, data),
      unsubscribe: () => this.unsubscribe(podKey, subscriptionId),
    };
  }

  send(podKey: string, data: string): void {
    void this.mgr().send(podKey, data);
  }

  sendResize(podKey: string, cols: number, rows: number): void {
    if (cols > 0 && rows > 0) void this.mgr().send_resize(podKey, cols, rows);
  }

  forceResize(podKey: string, cols: number, rows: number): void {
    if (cols > 0 && rows > 0) void this.mgr().force_resize(podKey, cols, rows);
  }

  unsubscribe(podKey: string, subscriptionId: string): void {
    void this.mgr().unsubscribe(podKey, subscriptionId);
  }

  disconnect(podKey: string): void {
    this.connectedPods.delete(podKey);
    void this.mgr().disconnect(podKey);
  }

  disconnectAll(): void {
    this.connectedPods.clear();
    void this.mgr().disconnect_all();
  }

  onStatusChange(podKey: string, listener: StatusListener): () => void {
    let set = this.statusListeners.get(podKey);
    if (!set) { set = new Set(); this.statusListeners.set(podKey, set); }
    set.add(listener);
    this.ensureStatusUpstream(podKey);
    listener(this.statusCache.get(podKey) ?? NONE);
    return () => {
      set!.delete(listener);
      if (set!.size === 0) this.statusListeners.delete(podKey);
    };
  }

  onAcpMessage(podKey: string, listener: AcpListener): () => void {
    let set = this.acpListeners.get(podKey);
    if (!set) { set = new Set(); this.acpListeners.set(podKey, set); }
    set.add(listener);
    this.ensureAcpUpstream(podKey);
    return () => {
      set!.delete(listener);
      if (set!.size === 0) this.acpListeners.delete(podKey);
    };
  }

  sendAcpCommand(podKey: string, command: Record<string, unknown>): void {
    void this.mgr().send_acp_command(podKey, JSON.stringify(command));
  }

  getStatus(podKey: string): ConnectionStatus | "none" {
    return this.statusCache.get(podKey)?.status ?? "none";
  }

  isConnected(podKey: string): boolean {
    return this.statusCache.get(podKey)?.status === "connected";
  }

  isRunnerDisconnected(podKey: string): boolean {
    return this.statusCache.get(podKey)?.runnerDisconnected ?? false;
  }

  getPodSize(): { rows: number; cols: number } | undefined {
    // podSize lives in the Rust pool; no synchronous consumer needs it.
    return undefined;
  }

  // The Rust pool clears a pod's status/ACP listeners on full teardown
  // (disconnect_inner) and fires this once. Drop our register-once guard so the
  // next subscribe re-registers the upstream listeners; local listeners persist
  // (the terminal component stays mounted across a transient drop).
  private ensureDisconnectHook(): void {
    if (this.disconnectHookWired) return;
    this.disconnectHookWired = true;
    void this.mgr().on_pod_disconnected((podKey: string) => {
      this.statusUpstream.delete(podKey);
      this.acpUpstream.delete(podKey);
      this.statusCache.delete(podKey);
    });
  }

  private ensureStatusUpstream(podKey: string): void {
    this.ensureDisconnectHook();
    if (this.statusUpstream.has(podKey)) return;
    this.statusUpstream.add(podKey);
    void this.mgr().on_status_change(podKey, (raw: StatusInfoRaw) => {
      // The pool reports "disconnected" for a pod that has never connected;
      // preserve the legacy "none" baseline so the terminal stays "connecting"
      // (yellow) instead of flashing "disconnected" (red) before first connect.
      const info: RelayStatusInfo =
        raw.status === "disconnected" && !this.connectedPods.has(podKey)
          ? { status: "none", runnerDisconnected: raw.runnerDisconnected }
          : { status: raw.status as ConnectionStatus, runnerDisconnected: raw.runnerDisconnected };
      this.statusCache.set(podKey, info);
      const set = this.statusListeners.get(podKey);
      if (set) for (const l of set) l(info);
    });
  }

  private ensureAcpUpstream(podKey: string): void {
    this.ensureDisconnectHook();
    if (this.acpUpstream.has(podKey)) return;
    this.acpUpstream.add(podKey);
    void this.mgr().on_acp_message(podKey, (msgType: number, payload: unknown) => {
      const set = this.acpListeners.get(podKey);
      if (set) for (const l of set) l(msgType, payload);
    });
  }
}

function getOrCreatePool(): RelayConnectionPool {
  const key = "__relayPool" as keyof typeof globalThis;
  const existing = globalThis[key] as RelayConnectionPool | undefined;
  if (existing) {
    if (process.env.NODE_ENV === "development") existing.disconnectAll();
    else return existing;
  }
  const pool = new RelayConnectionPool();
  (globalThis as Record<string, unknown>)[key] = pool;
  return pool;
}

export const relayPool = getOrCreatePool();
