import { relayPool } from "@/stores/relayConnection";

export function doagentRpc(
  podKey: string,
  method: string,
  params: Record<string, unknown> = {},
): boolean {
  if (!relayPool.isConnected(podKey)) return false;
  relayPool.sendAcpCommand(podKey, {
    type: "control_request",
    subtype: "doagent.rpc",
    payload: { method, params },
  });
  return true;
}

export function doagentControl(
  podKey: string,
  subtype: string,
  payload: Record<string, unknown> = {},
): boolean {
  if (!relayPool.isConnected(podKey)) return false;
  relayPool.sendAcpCommand(podKey, { type: "control_request", subtype, payload });
  return true;
}
