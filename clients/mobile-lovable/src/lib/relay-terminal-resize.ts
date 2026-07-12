import type { RelayLease } from "@/hooks/use-mobile-control-lease";

export interface RelayTerminalResizer {
  send_resize(podKey: string, columns: number, rows: number): Promise<void>;
}

export function sendGrantedTerminalResize(
  relay: RelayTerminalResizer | null,
  podKey: string | null,
  lease: RelayLease,
  columns: number,
  rows: number,
) {
  if (!relay || !podKey || lease.status !== "granted") return false;
  void relay.send_resize(podKey, columns, rows);
  return true;
}
