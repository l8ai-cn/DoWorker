export type RelayConnectionState = "connecting" | "connected" | "reconnecting" | "failed";

export function relayConnectionState(status: unknown): RelayConnectionState {
  if (status === "connected") return "connected";
  if (status === "connecting") return "connecting";
  if (status === "error") return "failed";
  return "reconnecting";
}
