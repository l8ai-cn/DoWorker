import type { TerminalControlLease } from "@do-worker/agent-ui";

export type ControlAction = "acquire" | "renew" | "release";

export interface PendingControl {
  action: ControlAction;
  leaseId?: string;
  reject(error: Error): void;
  resolve(lease?: TerminalControlLease): void;
  timer: number;
}

export interface EmbeddedTerminalConnection {
  lease: TerminalControlLease | null;
  pending: PendingControl | null;
  ready: Promise<void>;
  rejectReady(error: Error): void;
  resolveReady(): void;
  socket: WebSocket | null;
}

export function clearPendingControl(
  connection: EmbeddedTerminalConnection,
): PendingControl | null {
  const pending = connection.pending;
  if (!pending) return null;
  window.clearTimeout(pending.timer);
  connection.pending = null;
  return pending;
}
