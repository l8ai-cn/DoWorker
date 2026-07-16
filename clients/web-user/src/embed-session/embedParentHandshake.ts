export const EMBED_OPEN_MESSAGE = "agentsmesh.embed.open";
export const EMBED_READY_MESSAGE = "agentsmesh.embed.ready";

export function isEmbedReadyMessage(data: unknown): boolean {
  if (typeof data !== "object" || data === null) {
    return false;
  }
  const payload = data as { type?: unknown; version?: unknown };
  return payload.type === EMBED_READY_MESSAGE && payload.version === 1;
}

export function readAllowedEmbedOpenProof(
  event: Pick<MessageEvent, "origin" | "source" | "data">,
  parentWindow: Window,
  allowedOrigins: readonly string[],
): string | null {
  if (event.source !== parentWindow || !allowedOrigins.includes(event.origin)) {
    return null;
  }
  if (typeof event.data !== "object" || event.data === null) {
    return null;
  }
  const payload = event.data as {
    type?: unknown;
    version?: unknown;
    redemptionProof?: unknown;
  };
  if (
    payload.type !== EMBED_OPEN_MESSAGE ||
    payload.version !== 1 ||
    typeof payload.redemptionProof !== "string" ||
    payload.redemptionProof === ""
  ) {
    return null;
  }
  return payload.redemptionProof;
}
