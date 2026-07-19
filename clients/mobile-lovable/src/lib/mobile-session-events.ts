import { apiFetch } from "./api-fetch";

export type MessageContentBlock =
  | { type: "input_text"; text: string }
  | { type: "input_image"; file_id: string; filename?: string }
  | { type: "input_file"; file_id: string; filename: string };

export async function postMessageContent(
  sessionId: string,
  content: MessageContentBlock[],
): Promise<void> {
  const response = await apiFetch(`/v1/sessions/${encodeURIComponent(sessionId)}/events`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ type: "message", data: { role: "user", content } }),
  });
  if (!response.ok) throw new Error(await response.text());
}

export async function postMessage(sessionId: string, text: string): Promise<void> {
  await postMessageContent(sessionId, [{ type: "input_text", text }]);
}

export async function stopSession(sessionId: string): Promise<void> {
  const response = await apiFetch(`/v1/sessions/${encodeURIComponent(sessionId)}/events`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ type: "stop_session", data: {} }),
  });
  if (!response.ok) throw new Error(await response.text());
}

export async function resolveElicitation(
  sessionId: string,
  elicitationId: string,
  accept: boolean,
): Promise<void> {
  const response = await apiFetch(
    `/v1/sessions/${encodeURIComponent(sessionId)}/elicitations/${encodeURIComponent(elicitationId)}/resolve`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ action: accept ? "accept" : "decline" }),
    },
  );
  if (!response.ok) throw new Error(await response.text());
}

export function openSessionStream(sessionId: string, signal: AbortSignal): Promise<Response> {
  return apiFetch(`/v1/sessions/${encodeURIComponent(sessionId)}/stream`, {
    headers: { Accept: "text/event-stream" },
    signal,
  });
}

export async function fetchSessionItems(
  sessionId: string,
  limit = 50,
): Promise<Array<Record<string, unknown>>> {
  const response = await apiFetch(
    `/v1/sessions/${encodeURIComponent(sessionId)}/items?limit=${limit}&order=desc`,
  );
  if (!response.ok) throw new Error((await response.text()) || `HTTP ${response.status}`);
  const page = (await response.json()) as { data?: Array<Record<string, unknown>> };
  return [...(page.data ?? [])].reverse();
}
