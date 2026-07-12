import { apiFetch } from "./api-fetch";

export interface SessionRelayConnection {
  relayUrl: string;
  token: string;
  podKey: string;
}

export async function getSessionRelayConnection(
  sessionId: string,
): Promise<SessionRelayConnection> {
  const res = await apiFetch(`/v1/sessions/${encodeURIComponent(sessionId)}/relay-connection`);
  if (!res.ok) {
    throw new Error((await res.text()) || `Relay connection failed (${res.status})`);
  }
  const body = (await res.json()) as Partial<{
    relay_url: string;
    token: string;
    pod_key: string;
  }>;
  if (!body.relay_url || !body.token || !body.pod_key) {
    throw new Error("Invalid relay connection");
  }
  return { relayUrl: body.relay_url, token: body.token, podKey: body.pod_key };
}
