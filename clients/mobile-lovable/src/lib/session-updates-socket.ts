import { resolveWebSocketUrl } from "./ws-url";

export type SessionUpdatesFrame =
  | { type: "snapshot"; items: Array<Record<string, unknown> & { id: string }> }
  | { type: "changed"; items: Array<Record<string, unknown> & { id: string }> }
  | { type: "removed"; ids: string[] }
  | { type: "heartbeat" };

type FrameListener = (frame: SessionUpdatesFrame) => void;

const RECONNECT_BASE_MS = 250;
const RECONNECT_MAX_MS = 5_000;

function nextDelay(attempts: number): number {
  const base = Math.min(RECONNECT_BASE_MS * 2 ** (attempts - 1), RECONNECT_MAX_MS);
  return base / 2 + Math.random() * (base / 2);
}

class SessionUpdatesSocket {
  private ws: WebSocket | null = null;
  private watched: string[] = [];
  private watchedKey = "";
  private readonly listeners = new Set<FrameListener>();
  private started = false;
  private failedAttempts = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private generation = 0;

  start(): void {
    if (this.started) return;
    this.started = true;
    this.connect();
  }

  stop(): void {
    this.started = false;
    this.generation += 1;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    const ws = this.ws;
    this.ws = null;
    if (ws) {
      ws.onopen = ws.onmessage = ws.onclose = null;
      if (ws.readyState === WebSocket.OPEN) ws.close();
    }
  }

  setWatched(ids: string[]): void {
    const key = [...ids].sort().join(",");
    if (key === this.watchedKey) return;
    this.watchedKey = key;
    this.watched = ids;
    this.sendWatch();
  }

  subscribe(listener: FrameListener): () => void {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  private connect(): void {
    const gen = this.generation;
    let ws: WebSocket;
    try {
      ws = new WebSocket(resolveWebSocketUrl("/v1/sessions/updates"));
    } catch {
      this.scheduleReconnect();
      return;
    }
    this.ws = ws;
    ws.onopen = () => {
      if (!this.started || gen !== this.generation) {
        ws.close();
        return;
      }
      this.failedAttempts = 0;
      this.sendWatch();
    };
    ws.onmessage = (event) => {
      if (gen !== this.generation || typeof event.data !== "string") return;
      try {
        const frame = JSON.parse(event.data) as SessionUpdatesFrame;
        for (const listener of this.listeners) listener(frame);
      } catch {
        /* skip */
      }
    };
    ws.onclose = () => {
      if (gen !== this.generation) return;
      this.ws = null;
      if (this.started) this.scheduleReconnect();
    };
  }

  private scheduleReconnect(): void {
    if (!this.started || this.reconnectTimer) return;
    this.failedAttempts += 1;
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      if (this.started) this.connect();
    }, nextDelay(this.failedAttempts));
  }

  private sendWatch(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: "watch", session_ids: this.watched }));
    }
  }
}

export const sessionUpdatesSocket = new SessionUpdatesSocket();
